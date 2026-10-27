<?php
// Admin-only. POST + CSRF. (It used to be an unauthenticated GET that anyone could call to lock/unlock any account.)
require __DIR__ . '/lib/bootstrap.php';
require_post();
$admin = require_role('admin');
csrf_verify();

$email = strtolower(clean($_POST['email'] ?? '', 254));
$action = $_POST['action'] ?? '';
if (!valid_email($email)) {
    fail(400, 'Invalid email format.');
}
$status = ['deactivate' => 'inactive', 'activate' => 'active'][$action] ?? null;
if ($status === null) {
    fail(400, 'Invalid action.');
}
if ($email === strtolower($admin['email'])) {
    fail(400, 'You cannot change your own status.');
}
$stmt = db()->prepare('UPDATE users SET status = ? WHERE email = ?');
$stmt->bind_param('ss', $status, $email);
$stmt->execute();
if ($stmt->affected_rows < 1) {
    fail(404, 'No such user, or status unchanged.');
}
error_log('[bogoboom] admin ' . $admin['id'] . ' set status=' . $status . ' for a user');
header('Location: /backend/profiles/login.php');
exit;
