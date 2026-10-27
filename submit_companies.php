<?php
// Rewritten with prepared statements (it used string-built SQL with mysqli_real_escape_string).
require __DIR__ . '/lib/bootstrap.php';
require_post();
csrf_verify();
rate_limit_or_fail('register_company', 10, 3600);

$name = clean($_POST['company_name'] ?? '', 160);
$email = strtolower(clean($_POST['company_email'] ?? '', 254));
$phone = clean($_POST['company_phone'] ?? '', 40);
$responsible = clean($_POST['responsible_name'] ?? '', 160);
$address = clean($_POST['company_address'] ?? '', 255);
$website = clean($_POST['company_website'] ?? '', 255);
$industry = clean($_POST['company_industry'] ?? '', 120);
$country = clean($_POST['country'] ?? '', 80);
$state = clean($_POST['state'] ?? '', 80);
$city = clean($_POST['city'] ?? '', 80);
$password = (string) ($_POST['company_password'] ?? '');
$confirm = (string) ($_POST['confirm_password'] ?? '');

foreach ([$name, $phone, $responsible, $address, $country, $city] as $required) {
    if ($required === '') {
        fail(400, 'Please fill in all required fields.');
    }
}
if (!valid_email($email)) {
    fail(400, 'Invalid email format.');
}
if ($website !== '' && !preg_match('#^https?://#i', $website)) {
    fail(400, 'Website must start with http:// or https://');
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

$conn = db();
$id = uuid_v4();                         // same id for user and company (as before), now unpredictable
$hash = password_hash($password, PASSWORD_BCRYPT);
$consent = 1;
$allowAgencyContact = 1;                 // unchanged product behaviour; see docs/SECURITY-DEPLOY.md (privacy note)

$conn->begin_transaction();
try {
    $u = $conn->prepare("INSERT INTO users (id, email, company_email, password_hash, role, company_name, company_phone, company_address, company_website) VALUES (?, ?, ?, ?, 'companies', ?, ?, ?, ?)");
    $u->bind_param('ssssssss', $id, $email, $email, $hash, $name, $phone, $address, $website);
    $u->execute();
    $c = $conn->prepare('INSERT INTO companies (id, user_id, company_name, company_email, company_phone, responsible_name, company_address, company_website, company_industry, country, state_or_department, city, company_password_hash, data_consent, allow_agency_contact, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())');
    $c->bind_param('sssssssssssssii', $id, $id, $name, $email, $phone, $responsible, $address, $website, $industry, $country, $state, $city, $hash, $consent, $allowAgencyContact);
    $c->execute();
    $conn->commit();
} catch (mysqli_sql_exception $e) {
    $conn->rollback();
    if ($e->getCode() === 1062) {
        fail(409, 'This e-mail is already registered.');
    }
    throw $e;
}
echo '✅ Company registered successfully.';
echo '<meta http-equiv="refresh" content="5;url=/">';
