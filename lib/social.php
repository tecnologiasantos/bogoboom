<?php
declare(strict_types=1);
require_once __DIR__ . '/bootstrap.php';

/** DB role => URL slug. Roles come from the endpoint, never from user input. */
const SOCIAL_SLUGS = ['newcomers' => 'newcomer', 'companies' => 'company', 'agencies' => 'agency'];
const SOCIAL_PROVIDERS = ['Google', 'Facebook', 'LinkedIn'];

function hybridauth_config(string $slug): array
{
    $base = rtrim(env_required('APP_BASE_URL'), '/');       // e.g. https://bogoboom.com
    if (!str_starts_with($base, 'https://') && env('APP_ENV') !== 'dev') {
        throw new RuntimeException('APP_BASE_URL must be https');
    }
    $providers = [];
    foreach ([['Google', 'GOOGLE', 'email profile'], ['Facebook', 'FACEBOOK', 'email, public_profile'], ['LinkedIn', 'LINKEDIN', 'r_liteprofile r_emailaddress']] as [$name, $prefix, $scope]) {
        $id = env($prefix . '_CLIENT_ID');
        $secret = env($prefix . '_CLIENT_SECRET');
        if ($id && $secret) {
            $providers[$name] = ['enabled' => true, 'keys' => ['id' => $id, 'secret' => $secret], 'scope' => $scope];
        }
    }
    return ['callback' => $base . '/backend/profiles/social_callback_' . $slug . '.php', 'providers' => $providers];
}

function social_provider(): string
{
    $p = $_GET['provider'] ?? '';
    if (!is_string($p) || !in_array($p, SOCIAL_PROVIDERS, true)) {
        fail(400, 'Unknown login provider.');
    }
    return $p;
}

function social_hybridauth(string $slug): \Hybridauth\Hybridauth
{
    require_once rtrim(env('APP_ROOT', '/var/www/html/bogoboom_project'), '/') . '/vendor/autoload.php';
    return new \Hybridauth\Hybridauth(hybridauth_config($slug));
}

function social_start(string $slug): void
{
    rate_limit_or_fail('social_start', 30, 600);
    $provider = social_provider();
    social_hybridauth($slug)->authenticate($provider);       // redirects to the provider
    header('Location: /backend/profiles/social_callback_' . $slug . '.php?provider=' . rawurlencode($provider));
    exit;
}

/** Logs in an EXISTING active account of the right role. Never creates accounts or trusts unverified e-mails. */
function social_finish(string $dbRole, string $slug): void
{
    rate_limit_or_fail('social_cb', 30, 600);
    $provider = social_provider();
    try {
        $adapter = social_hybridauth($slug)->authenticate($provider);
        $profile = $adapter->getUserProfile();
        $adapter->disconnect();
    } catch (\Throwable $e) {
        error_log('[bogoboom] social login failed: ' . get_class($e));
        fail(400, 'Social login failed. Please try again or use your e-mail and password.');
    }
    $email = strtolower((string) ($profile->emailVerified ?? ''));   // only a provider-VERIFIED e-mail counts
    if (!valid_email($email)) {
        fail(400, 'Your provider did not share a verified e-mail address.');
    }
    $stmt = db()->prepare('SELECT id, email, role, status FROM users WHERE email = ? AND role = ? LIMIT 1');
    $stmt->bind_param('ss', $email, $dbRole);
    $stmt->execute();
    $user = $stmt->get_result()->fetch_assoc();
    if ($user && $user['status'] === 'active') {
        login_user($user);
        header('Location: /backend/profiles/login.php');
    } else {
        header('Location: /register_' . ($slug === 'company' ? 'companies' : $slug . 's') . '.html?social=notfound');
    }
    exit;
}
