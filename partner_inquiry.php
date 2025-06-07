<?php
// partner_inquiry.php

// CONFIGURATION
$adminEmail = "contact@bogoboom.com"; // Change to your preferred email
$sendEmail = true;                    // Set to false if not using email for now

// UTILITIES
function sanitize($data) {
    return htmlspecialchars(trim($data), ENT_QUOTES, 'UTF-8');
}

// HONEYPOT CHECK
if (!empty($_POST['website'])) {
    http_response_code(403);
    exit('Spam detected.');
}

// VALIDATE FIELDS
$requiredFields = ['fullName', 'emailAddress', 'country', 'city', 'collaborationType'];
$errors = [];

foreach ($requiredFields as $field) {
    if (empty($_POST[$field])) {
        $errors[] = $field . ' is required.';
    }
}

if (!empty($errors)) {
    http_response_code(400);
    echo json_encode(['status' => 'error', 'message' => 'Validation error', 'errors' => $errors]);
    exit;
}

// SANITIZE INPUT
$fullName = sanitize($_POST['fullName']);
$emailAddress = filter_var($_POST['emailAddress'], FILTER_SANITIZE_EMAIL);
$phoneNumber = sanitize($_POST['phoneNumber'] ?? '');
$country = sanitize($_POST['country']);
$city = sanitize($_POST['city']);
$collaborationType = sanitize($_POST['collaborationType']);

// EMAIL TEMPLATE
$emailSubject = "New Partner Inquiry from $fullName";
$emailBody = "
New partner inquiry submitted on Bogoboom:

Full Name: $fullName
Email: $emailAddress
Phone: $phoneNumber
Country: $country
City: $city
Collaboration Type: $collaborationType

Date: " . date("Y-m-d H:i:s") . "
";

// (OPTIONAL) SEND EMAIL
if ($sendEmail) {
    $headers = "From: noreply@bogoboom.com\r\n";
    $headers .= "Reply-To: $emailAddress\r\n";
    $headers .= "Content-Type: text/plain; charset=UTF-8\r\n";

    mail($adminEmail, $emailSubject, $emailBody, $headers);
}

// LOG OR STORE (You can extend this to save to MariaDB)
file_put_contents(__DIR__ . "/partner_inquiries.log", $emailBody . "\n-----------------\n", FILE_APPEND);

// SUCCESS RESPONSE
http_response_code(200);
echo json_encode(['status' => 'success', 'message' => 'Thank you for your inquiry. We will get in touch soon.']);
exit;
