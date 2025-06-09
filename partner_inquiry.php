<?php
// Include your DB config file
require '../safe/db_connection.php'; // Adjust the path if needed

// Check if the form was submitted via POST
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    
    // Honeypot anti-bot field
    if (!empty($_POST['website'])) {
        // Bot detected — silently ignore
        exit;
    }

    // Sanitize and assign form data
    $fullName = trim($_POST['fullName']);
    $email = filter_var(trim($_POST['emailAddress']), FILTER_SANITIZE_EMAIL);
    $phone = trim($_POST['phoneNumber']);
    $country = trim($_POST['country']);
    $state = trim($_POST['stateOrDepartment'] ?? '');
    $city = trim($_POST['city']);
    $collabType = trim($_POST['collaborationType']);

    // Validate required fields
    if (empty($fullName) || empty($email) || empty($country) || empty($city) || empty($collabType)) {
        http_response_code(400);
        echo "Please fill in all required fields.";
        exit;
    }

    // Validate email format
    if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
        http_response_code(400);
        echo "Invalid email address.";
        exit;
    }

    // Prepare and execute SQL
    $stmt = $conn->prepare("
        INSERT INTO partner_inquiries 
        (full_name, email_address, phone_number, country, state_or_department, city, collaboration_type, honeypot_website_field) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ");

    $honeypot = ''; // Always empty because real users won't fill it
    $stmt->bind_param(
        "ssssssss",
        $fullName,
        $email,
        $phone,
        $country,
        $state,
        $city,
        $collabType,
        $honeypot
    );

    if ($stmt->execute()) {
        echo "Thank you! Your inquiry has been submitted.";
    } else {
        http_response_code(500);
        echo "Error saving your information. Please try again later.";
    }

    $stmt->close();
    $conn->close();

} else {
    http_response_code(405);
    echo "Method Not Allowed";
}
?>
