<?php
require __DIR__ . '/lib/bootstrap.php';
require_post();
csrf_verify();
rate_limit_or_fail('register', 10, 3600);

$email = strtolower(clean($_POST['email'] ?? '', 254));
$password = (string) ($_POST['password'] ?? '');
$role = (string) ($_POST['role'] ?? '');

// The role was taken straight from the form: anyone could register as any role. Whitelist only.
if (!in_array($role, ['newcomers', 'companies', 'agencies'], true)) {
    fail(400, 'Invalid account type.');
}
if (!valid_email($email)) {
    fail(400, 'Invalid email format.');
}
if (($problem = password_problem($password)) !== null) {
    fail(400, $problem);
}

$conn = db();
$check = $conn->prepare('SELECT 1 FROM users WHERE email = ? LIMIT 1');
$check->bind_param('s', $email);
$check->execute();
if ($check->get_result()->num_rows > 0) {
    fail(409, 'This e-mail is already registered.');
}

$id = uuid_v4();
$hash = password_hash($password, PASSWORD_DEFAULT);
$ins = $conn->prepare('INSERT INTO users (id, email, password_hash, role) VALUES (?, ?, ?, ?)');
$ins->bind_param('ssss', $id, $email, $hash, $role);
try {
    $ins->execute();
} catch (mysqli_sql_exception $e) {
    if ($e->getCode() === 1062) {
        fail(409, 'This e-mail is already registered.');
    }
    throw $e;
}
header('Location: /login.html');
exit;
