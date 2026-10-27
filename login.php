<?php
require __DIR__ . '/lib/bootstrap.php';
require __DIR__ . '/lib/layout.php';

const DUMMY_HASH = '$2y$10$/tflA4snGTGItFcN37wfUONALBkgIdkarCvU.NugjLCXm5az2EOay'; // bcrypt of a random string: equalises timing when the account does not exist

function render_login(string $error = '', int $code = 200): never
{
    http_response_code($code);
    page_start('Bogoboom | Login', 'Welcome to Bogoboom', 'Connecting newcomers with jobs and homes');
    echo '<section><h3>Login</h3>';
    if ($error !== '') {
        echo '<p class="err" role="alert">' . h($error) . '</p>';
    }
    echo '<form class="grid" method="POST" action="/backend/profiles/login.php">' . csrf_field()
        . '<input type="hidden" name="action" value="login">'
        . '<div><label for="email">📧 Email:</label><input type="email" id="email" name="email" required autocomplete="username" maxlength="254"></div>'
        . '<div><label for="password">🔒 Password:</label><input type="password" id="password" name="password" required autocomplete="current-password" maxlength="128"></div>'
        . '<button type="submit">🔑 Login</button></form></section>';
    page_end();
    exit;
}

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    csrf_verify();
    $action = $_POST['action'] ?? '';

    if ($action === 'logout') {
        logout_user();
        header('Location: /');
        exit;
    }

    if ($action === 'login') {
        $email = strtolower(clean($_POST['email'] ?? '', 254));
        $password = (string) ($_POST['password'] ?? '');
        rate_limit_or_fail('login_ip', 20, 900);
        rate_limit_or_fail('login_acct:' . $email, 5, 900);

        $user = null;
        if (valid_email($email) && strlen($password) <= 128) {
            $stmt = db()->prepare('SELECT id, email, role, status, password_hash FROM users WHERE email = ? LIMIT 1');
            $stmt->bind_param('s', $email);
            $stmt->execute();
            $user = $stmt->get_result()->fetch_assoc();
        }
        $hash = $user['password_hash'] ?? DUMMY_HASH;           // always verify: no timing/enumeration signal
        $ok = password_verify($password, $hash) && $user && $user['status'] === 'active';
        if (!$ok) {
            render_login('Incorrect credentials or inactive account.', 401);
        }
        login_user($user);
        header('Location: /backend/profiles/login.php');         // post/redirect/get
        exit;
    }
    fail(400, 'Bad request.');
}

if (empty($_SESSION['user_id'])) {
    render_login();
}

// ---------- logged-in dashboard ----------
$u = require_login();
$labels = ['company_name' => 'Company', 'email' => 'Email', 'phone' => 'Phone', 'company_website' => 'Website', 'country' => 'Country', 'state_or_department' => 'State/Department', 'city' => 'City', 'full_name' => 'Name'];

if (in_array($u['role'], ['newcomers', 'companies'], true)) {
    $title = 'Agencies';
    $sql = "SELECT u.company_name, u.email, u.phone, u.company_website, a.country, a.state_or_department, a.city
              FROM users u JOIN agencies a ON u.id = a.user_id WHERE u.role = 'agencies' AND u.status = 'active' LIMIT 100";
} elseif (in_array($u['role'], ['agencies', 'admin'], true)) {
    $title = 'Newcomers';
    $sql = "SELECT full_name, email, phone FROM users WHERE role = 'newcomers' AND status = 'active' LIMIT 100";
} else {
    fail(403, 'Forbidden.');
}
$result = db()->query($sql);

page_start('Bogoboom | Directory', 'Bogoboom ' . $title, 'Logged in as ' . $u['email']);
echo '<div class="actions"><form method="POST" action="/backend/profiles/login.php">' . csrf_field()
    . '<input type="hidden" name="action" value="logout"><button class="danger" type="submit">Log off</button></form>'
    . '<form method="GET" action="/backend/profiles/change_password.php"><button class="warn" type="submit">Change Password</button></form></div>';
echo '<section><h3>' . h($title) . '</h3>';
if ($result && $result->num_rows > 0) {
    $rows = $result->fetch_all(MYSQLI_ASSOC);
    echo '<table><thead><tr>';
    foreach (array_keys($rows[0]) as $k) {
        echo '<th>' . h($labels[$k] ?? $k) . '</th>';
    }
    echo '</tr></thead><tbody>';
    foreach ($rows as $row) {
        echo '<tr>';
        foreach ($row as $v) {
            echo '<td>' . h((string) $v) . '</td>';
        }
        echo '</tr>';
    }
    echo '</tbody></table>';
} else {
    echo '<p>No records found.</p>';
}
echo '</section>';
page_end();
