<?php
require __DIR__ . '/lib/bootstrap.php';
require __DIR__ . '/lib/layout.php';

if (empty($_SESSION['user_id'])) {
    header('Location: /backend/profiles/login.php');
    exit;
}
$userId = $_SESSION['user_id'];
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    csrf_verify();
    rate_limit_or_fail('chgpw:' . $userId, 5, 900);
    $current = (string) ($_POST['current_password'] ?? '');
    $new = (string) ($_POST['new_password'] ?? '');
    $confirm = (string) ($_POST['confirm_password'] ?? '');

    if ($new !== $confirm) {
        $message = 'New passwords do not match.';
    } elseif (($problem = password_problem($new)) !== null) {
        $message = $problem;
    } elseif (hash_equals($current, $new)) {
        $message = 'Choose a password different from the current one.';
    } else {
        $stmt = db()->prepare('SELECT password_hash FROM users WHERE id = ? LIMIT 1');
        $stmt->bind_param('s', $userId);
        $stmt->execute();
        $row = $stmt->get_result()->fetch_assoc();
        if (!$row || !password_verify($current, $row['password_hash'])) {
            $message = 'Current password is incorrect.';
        } else {
            $hash = password_hash($new, PASSWORD_DEFAULT);
            $upd = db()->prepare('UPDATE users SET password_hash = ? WHERE id = ?');
            $upd->bind_param('ss', $hash, $userId);
            $upd->execute();
            logout_user();
            page_start('Password changed');
            echo '<section><p class="ok">✅ Password changed successfully. Please log in again.</p><p><a href="/login.html">Go to login</a></p></section>';
            page_end();
            exit;
        }
    }
}

page_start('Change Password', 'Change Password');
echo '<section><form class="grid" method="POST" action="">' . csrf_field()
    . '<div><label for="cp">Current Password</label><input type="password" id="cp" name="current_password" required autocomplete="current-password" maxlength="128"></div>'
    . '<div><label for="np">New Password (min. 10 characters)</label><input type="password" id="np" name="new_password" required minlength="10" maxlength="128" autocomplete="new-password"></div>'
    . '<div><label for="cf">Confirm New Password</label><input type="password" id="cf" name="confirm_password" required minlength="10" maxlength="128" autocomplete="new-password"></div>'
    . '<button type="submit">🔒 Update Password</button></form>';
if ($message !== '') {
    echo '<p class="err" role="alert">' . h($message) . '</p>';
}
echo '</section>';
page_end();
