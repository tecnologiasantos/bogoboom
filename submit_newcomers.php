<?php
require __DIR__ . '/lib/bootstrap.php';
require_post();
csrf_verify();
rate_limit_or_fail('register_newcomer', 10, 3600);

// the HTML form posts "fullname"; older code expected "full_name" — accept both
$fullName = clean($_POST['fullname'] ?? $_POST['full_name'] ?? '', 160);
$email = strtolower(clean($_POST['email'] ?? '', 254));
$phone = clean($_POST['phone'] ?? '', 40);
$country = clean($_POST['country'] ?? '', 80);
$languages = clean($_POST['languages'] ?? '', 255);
$password = (string) ($_POST['password'] ?? '');
$confirm = (string) ($_POST['confirm_password'] ?? '');

if ($fullName === '' || $phone === '' || $country === '' || !valid_email($email)) {
    fail(400, 'Please fill in all required fields with valid values.');
}
if (empty($_POST['dataConsent'])) {
    fail(400, 'You must accept the privacy policy to register.');
}
if ($password !== $confirm) {
    fail(400, 'Passwords do not match.');
}
if (($problem = password_problem($password)) !== null) {
    fail(400, $problem);
}

$document = save_upload($_FILES['document'] ?? ['error' => UPLOAD_ERR_NO_FILE], uploads_dir('temp_files'));

$conn = db();
$userId = uuid_v4();
$newcomerId = uuid_v4();
$hash = password_hash($password, PASSWORD_DEFAULT);

$conn->begin_transaction();
try {
    $u = $conn->prepare("INSERT INTO users (id, email, full_name, phone, password_hash, role) VALUES (?, ?, ?, ?, ?, 'newcomers')");
    $u->bind_param('sssss', $userId, $email, $fullName, $phone, $hash);
    $u->execute();
    $n = $conn->prepare('INSERT INTO newcomers (id, user_id, full_name, email, phone, country, languages, document) VALUES (?, ?, ?, ?, ?, ?, ?, ?)');
    $n->bind_param('ssssssss', $newcomerId, $userId, $fullName, $email, $phone, $country, $languages, $document);
    $n->execute();
    $conn->commit();
} catch (mysqli_sql_exception $e) {
    $conn->rollback();
    if ($document) {
        @unlink(uploads_dir('temp_files') . '/' . $document);
    }
    if ($e->getCode() === 1062) {
        fail(409, 'This e-mail is already registered.');
    }
    throw $e;
}
echo '✅ Newcomer registered successfully.';
echo '<meta http-equiv="refresh" content="5;url=/">';
