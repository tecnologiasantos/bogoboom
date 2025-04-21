<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start session if needed
session_start();

// Securely include DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Sanitize and collect inputs
    $company_name = trim($_POST['company_name']);
    $company_email = trim($_POST['company_email']);
    $company_phone = trim($_POST['company_phone']);
    $responsible_name = trim($_POST['responsible_name']);
    $company_address = trim($_POST['company_address']);
    $company_website = trim($_POST['company_website']);
    $company_industry = trim($_POST['company_industry']);
    $dataConsent = isset($_POST['dataConsent']) ? 1 : 0;

    // Prepare and execute insert
    $stmt = $conn->prepare("INSERT INTO companies 
        (name, email, phone, contact_person, address, website, industry, allow_agency_contact, created_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())");

    if (!$stmt) {
        die("Prepare failed: " . $conn->error);
    }

    $stmt->bind_param("sssssssi", 
        $company_name, 
        $company_email, 
        $company_phone, 
        $responsible_name, 
        $company_address, 
        $company_website, 
        $company_industry, 
        $dataConsent
    );

    if ($stmt->execute()) {
        echo "✅ Company registered successfully.";
    } else {
        echo "❌ Database error: " . $stmt->error;
    }

    $stmt->close();
}

$conn->close();
?>
