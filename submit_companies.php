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

    // Generate UUID for the company id (you can use uniqid() or a library like RAMSEY\UUID)
    $company_id = uniqid('comp_', true); // You can use a better UUID library if needed

    // Prepare the insert query
    $query = "INSERT INTO companies 
        (id, name, email, phone, contact_person, address, website, industry, allow_agency_contact, created_at) 
        VALUES ('$company_id', '$company_name', '$company_email', '$company_phone', '$responsible_name', '$company_address', '$company_website', '$company_industry', '$dataConsent', NOW())";

    // Execute the query using the database connection
    if (mysqli_query($conn, $query)) {
        echo "✅ Company registered successfully.";
    } else {
        echo "❌ Error: " . mysqli_error($conn);
    }
}

$conn->close();
?>
