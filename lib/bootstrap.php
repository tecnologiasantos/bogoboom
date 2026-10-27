<?php
declare(strict_types=1);
/**
 * Single entry point for every public script:  require __DIR__ . '/lib/bootstrap.php';
 * - no error output to visitors (logged instead)  - security headers  - hardened session  - helpers
 * Define BB_NO_SESSION before requiring for pure JSON endpoints that need no session.
 */
require_once __DIR__ . '/env.php';

error_reporting(E_ALL);
ini_set('display_errors', '0');
ini_set('display_startup_errors', '0');
ini_set('log_errors', '1');
mysqli_report(MYSQLI_REPORT_ERROR | MYSQLI_REPORT_STRICT);

set_exception_handler(static function (Throwable $e): void {
    error_log('[bogoboom] ' . get_class($e) . ': ' . $e->getMessage() . ' @ ' . basename($e->getFile()) . ':' . $e->getLine());
    if (!headers_sent()) {
        http_response_code(500);
        header('Content-Type: text/plain; charset=UTF-8');
    }
    echo 'Something went wrong. Please try again later.';
});

if (!headers_sent()) {
    header('X-Content-Type-Options: nosniff');
    header('X-Frame-Options: SAMEORIGIN');
    header('Referrer-Policy: strict-origin-when-cross-origin');
    header('Permissions-Policy: camera=(), microphone=(), geolocation=()');
    header('Cache-Control: no-store');
    header_remove('X-Powered-By');
}

// ---------- request helpers ----------
function h(?string $s): string
{
    return htmlspecialchars((string) $s, ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8');
}

function is_https(): bool
{
    if (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') {
        return true;
    }
    // only trust the proxy header when explicitly enabled (Cloudflare / ALB in front)
    return env('TRUST_PROXY') === '1' && ($_SERVER['HTTP_X_FORWARDED_PROTO'] ?? '') === 'https';
}

function client_ip(): string
{
    $ip = $_SERVER['REMOTE_ADDR'] ?? '0.0.0.0';
    if (env('TRUST_PROXY') === '1') {
        $cand = $_SERVER['HTTP_CF_CONNECTING_IP'] ?? trim(explode(',', $_SERVER['HTTP_X_FORWARDED_FOR'] ?? '')[0]);
        if ($cand !== '' && filter_var($cand, FILTER_VALIDATE_IP)) {
            $ip = $cand;
        }
    }
    return $ip;
}

function fail(int $code, string $message): never
{
    http_response_code($code);
    header('Content-Type: text/plain; charset=UTF-8');
    echo $message;
    exit;
}

function json_out(array $data, int $code = 200): never
{
    http_response_code($code);
    header('Content-Type: application/json; charset=UTF-8');
    echo json_encode($data, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
    exit;
}

function uuid_v4(): string
{
    $b = random_bytes(16);               // CSPRNG (the old mt_rand UUIDs were predictable)
    $b[6] = chr((ord($b[6]) & 0x0f) | 0x40);
    $b[8] = chr((ord($b[8]) & 0x3f) | 0x80);
    return vsprintf('%s%s-%s-%s-%s-%s%s%s', str_split(bin2hex($b), 4));
}

// ---------- database ----------
function db(): mysqli
{
    static $conn = null;
    if ($conn === null) {
        $conn = new mysqli(env('DB_HOST', 'localhost'), env_required('DB_USER'), env_required('DB_PASS'), env_required('DB_NAME'), (int) env('DB_PORT', '3306'));
        $conn->set_charset('utf8mb4');
    }
    return $conn;
}

function pdo(): PDO
{
    static $pdo = null;
    if ($pdo === null) {
        $dsn = sprintf('mysql:host=%s;port=%d;dbname=%s;charset=utf8mb4', env('DB_HOST', 'localhost'), (int) env('DB_PORT', '3306'), env_required('DB_NAME'));
        $pdo = new PDO($dsn, env_required('DB_USER'), env_required('DB_PASS'), [
            PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
            PDO::ATTR_EMULATE_PREPARES => false,
        ]);
    }
    return $pdo;
}

// ---------- session ----------
function start_secure_session(): void
{
    if (session_status() === PHP_SESSION_ACTIVE) {
        return;
    }
    ini_set('session.use_strict_mode', '1');
    ini_set('session.use_only_cookies', '1');
    session_name('bb_sess');
    session_set_cookie_params(['lifetime' => 0, 'path' => '/', 'secure' => is_https(), 'httponly' => true, 'samesite' => 'Lax']);
    session_start();
    $now = time();
    if (isset($_SESSION['_created']) && ($now - $_SESSION['_last'] > 7200 || $now - $_SESSION['_created'] > 43200)) {
        $_SESSION = [];                  // idle > 2h or older than 12h: start over
    }
    $_SESSION['_created'] = $_SESSION['_created'] ?? $now;
    $_SESSION['_last'] = $now;
}

function login_user(array $u): void
{
    session_regenerate_id(true);         // prevents session fixation
    $_SESSION = ['_created' => time(), '_last' => time(), 'user_id' => $u['id'], 'user_email' => $u['email'], 'user_role' => $u['role']];
    unset($_SESSION['csrf']);
}

function logout_user(): void
{
    $_SESSION = [];
    if (ini_get('session.use_cookies')) {
        $p = session_get_cookie_params();
        setcookie(session_name(), '', ['expires' => time() - 3600, 'path' => $p['path'], 'secure' => $p['secure'], 'httponly' => true, 'samesite' => 'Lax']);
    }
    session_destroy();
}

function require_login(): array
{
    if (empty($_SESSION['user_id'])) {
        fail(401, 'Please log in.');
    }
    return ['id' => $_SESSION['user_id'], 'email' => $_SESSION['user_email'], 'role' => $_SESSION['user_role']];
}

function require_role(string ...$roles): array
{
    $u = require_login();
    if (!in_array($u['role'], $roles, true)) {
        fail(403, 'Forbidden.');
    }
    return $u;
}

// ---------- CSRF ----------
function csrf_token(): string
{
    start_secure_session();
    return $_SESSION['csrf'] ??= bin2hex(random_bytes(32));
}

function csrf_field(): string
{
    return '<input type="hidden" name="csrf_token" value="' . h(csrf_token()) . '">';
}

function origin_ok(): bool
{
    $host = strtolower(preg_replace('/:\d+$/', '', $_SERVER['HTTP_HOST'] ?? ''));
    $allowed = array_filter(array_map('trim', explode(',', strtolower(env('ALLOWED_HOSTS', '')))));
    $allowed[] = $host;
    foreach (['HTTP_ORIGIN', 'HTTP_REFERER'] as $hdr) {
        if (!empty($_SERVER[$hdr])) {
            $h = strtolower((string) parse_url($_SERVER[$hdr], PHP_URL_HOST));
            return $h !== '' && in_array($h, $allowed, true);
        }
    }
    return true;                         // non-browser client: the CSRF token below is still required
}

function csrf_verify(): void
{
    start_secure_session();
    $given = $_POST['csrf_token'] ?? ($_SERVER['HTTP_X_CSRF_TOKEN'] ?? '');
    $ok = isset($_SESSION['csrf']) && is_string($given) && $given !== '' && hash_equals($_SESSION['csrf'], $given);
    if (!$ok || !origin_ok()) {
        fail(403, 'Your session expired or the request was not allowed. Reload the page and try again.');
    }
}

function require_post(): void
{
    if (($_SERVER['REQUEST_METHOD'] ?? '') !== 'POST') {
        header('Allow: POST');
        fail(405, 'Method not allowed.');
    }
}

// ---------- abuse control ----------
/** Sliding-window limiter stored in files (works without Redis/APCu). Returns false when over the limit. */
function rate_limit(string $bucket, int $max, int $windowSec): bool
{
    $dir = env('RATE_LIMIT_DIR', sys_get_temp_dir() . '/bogoboom_rl');
    if (!is_dir($dir)) {
        @mkdir($dir, 0700, true);
    }
    $f = fopen($dir . '/' . hash('sha256', $bucket . '|' . client_ip()), 'c+');
    if (!$f) {
        return true;                     // fail open rather than lock everyone out
    }
    flock($f, LOCK_EX);
    $now = time();
    $hits = array_values(array_filter(json_decode(stream_get_contents($f) ?: '[]', true) ?: [], static fn ($t) => $t > $now - $windowSec));
    $allowed = count($hits) < $max;
    if ($allowed) {
        $hits[] = $now;
    }
    ftruncate($f, 0);
    rewind($f);
    fwrite($f, json_encode($hits));
    flock($f, LOCK_UN);
    fclose($f);
    return $allowed;
}

function rate_limit_or_fail(string $bucket, int $max, int $windowSec): void
{
    if (!rate_limit($bucket, $max, $windowSec)) {
        header('Retry-After: ' . $windowSec);
        fail(429, 'Too many attempts. Please wait a few minutes and try again.');
    }
}

// ---------- validation ----------
function password_problem(string $p): ?string
{
    if (strlen($p) < 10) {
        return 'Password must be at least 10 characters.';
    }
    if (strlen($p) > 128) {
        return 'Password is too long (max 128 characters).';
    }
    return null;
}

function clean(?string $s, int $max = 255): string
{
    $s = trim((string) $s);
    $s = preg_replace('/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/u', '', $s) ?? '';
    return mb_substr($s, 0, $max);
}

function valid_email(string $e): bool
{
    return strlen($e) <= 254 && filter_var($e, FILTER_VALIDATE_EMAIL) !== false;
}

/**
 * Safe upload: size cap, real MIME sniffing, extension derived from the MIME (never from the client name),
 * random file name, stored outside the web root with 0640 permissions. Returns the stored file name.
 */
function save_upload(array $file, string $dir, int $maxBytes = 5_242_880): ?string
{
    if (($file['error'] ?? UPLOAD_ERR_NO_FILE) === UPLOAD_ERR_NO_FILE) {
        return null;
    }
    if ($file['error'] !== UPLOAD_ERR_OK || !is_uploaded_file($file['tmp_name']) || $file['size'] > $maxBytes) {
        fail(400, 'Upload failed or file too large (max 5 MB).');
    }
    $map = ['application/pdf' => 'pdf', 'application/msword' => 'doc', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' => 'docx'];
    $mime = (new finfo(FILEINFO_MIME_TYPE))->file($file['tmp_name']);
    if (!isset($map[$mime])) {
        fail(400, 'Invalid document type. Only PDF, DOC and DOCX are allowed.');
    }
    if (!is_dir($dir) && !mkdir($dir, 0750, true) && !is_dir($dir)) {
        throw new RuntimeException('Cannot create upload directory');
    }
    $name = bin2hex(random_bytes(16)) . '.' . $map[$mime];
    if (!move_uploaded_file($file['tmp_name'], rtrim($dir, '/') . '/' . $name)) {
        throw new RuntimeException('Cannot store upload');
    }
    chmod(rtrim($dir, '/') . '/' . $name, 0640);
    return $name;
}

function uploads_dir(string $sub): string
{
    return rtrim(env('UPLOADS_DIR', '/var/www/html/bogoboom_project/uploads'), '/') . '/' . $sub;
}

if (!defined('BB_NO_SESSION')) {
    start_secure_session();
}
