<?php
require __DIR__ . '/lib/bootstrap.php';
require_post();
if (!origin_ok()) {
    fail(403, 'Forbidden.');
}
if (!empty($_POST['website'] ?? '')) {   // honeypot
    exit;
}
rate_limit_or_fail('lead', 10, 3600);

$name = clean($_POST['name'] ?? '', 160);
$email = strtolower(clean($_POST['email'] ?? '', 254));
$phone = clean($_POST['phone'] ?? '', 40);
$country = clean($_POST['country'] ?? '', 80);
$state = clean($_POST['state'] ?? '', 80);
$city = clean($_POST['city'] ?? '', 80);
$interest = clean($_POST['interest'] ?? '', 255);
$fullCity = $state !== '' ? "$state - $city" : $city;

if ($name === '' || $country === '' || $city === '' || $interest === '' || !valid_email($email)) {
    fail(400, 'Please fill in all required fields with valid values.');
}

$stmt = pdo()->prepare('INSERT INTO leads (name, email, phone, city, country, interest) VALUES (:name, :email, :phone, :city, :country, :interest)');
$stmt->execute([':name' => $name, ':email' => $email, ':phone' => $phone, ':city' => $fullCity, ':country' => $country, ':interest' => $interest]);
echo 'Thank you! Your application has been received.';
