<?php
require __DIR__ . '/lib/bootstrap.php';
require_post();
csrf_verify();
rate_limit_or_fail('register_agency', 10, 3600);

$name = clean($_POST['company_name'] ?? '', 160);
$email = strtolower(clean($_POST['company_email'] ?? '', 254));
$phone = clean($_POST['company_phone'] ?? '', 40);
$responsible = clean($_POST['responsible_name'] ?? '', 160);
$address = clean($_POST['company_address'] ?? '', 255);
$country = clean($_POST['country'] ?? '', 80);
$state = clean($_POST['state'] ?? '', 80);
$city = clean($_POST['city'] ?? '', 80);
$website = clean($_POST['company_website'] ?? '', 255);
$password = (string) ($_POST['company_password'] ?? '');
$confirm = (string) ($_POST['confirm_password'] ?? '');

foreach ([$name, $phone, $responsible, $address, $country, $state, $city] as $required) {
    if ($required === '') {
        fail(400, 'Please fill in all required fields.');
    }
}
if (!valid_email($email)) {
    fail(400, 'Invalid email format.');
}
if ($website !== '' && !preg_match('#^https?://#i', $website)) {
    fail(400, 'Website must start with http:// or https://');   // blocks javascript: URLs later rendered as links
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

$document = save_upload($_FILES['company_document'] ?? ['error' => UPLOAD_ERR_NO_FILE], uploads_dir('agency_docs'));

$conn = db();
$userId = uuid_v4();
$agencyId = uuid_v4();
$hash = password_hash($password, PASSWORD_DEFAULT);
$consent = 1;

$conn->begin_transaction();
try {
    $u = $conn->prepare("INSERT INTO users (id, email, full_name, phone, password_hash, role, company_name, company_website, created_at) VALUES (?, ?, ?, ?, ?, 'agencies', ?, ?, NOW())");
    $u->bind_param('sssssss', $userId, $email, $name, $phone, $hash, $name, $website);
    $u->execute();
    $a = $conn->prepare('INSERT INTO agencies (id, user_id, company_name, company_email, company_phone, responsible_name, company_address, country, state_or_department, city, company_website, company_document, company_password_hash, data_consent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)');
    $a->bind_param('sssssssssssssi', $agencyId, $userId, $name, $email, $phone, $responsible, $address, $country, $state, $city, $website, $document, $hash, $consent);
    $a->execute();
    $conn->commit();
} catch (mysqli_sql_exception $e) {
    $conn->rollback();
    if ($document) {
        @unlink(uploads_dir('agency_docs') . '/' . $document);
    }
    if ($e->getCode() === 1062) {
        fail(409, 'This e-mail is already registered.');
    }
    throw $e;
}
echo '✅ Agency registered successfully.';
echo '<meta http-equiv="refresh" content="5;url=/">';
