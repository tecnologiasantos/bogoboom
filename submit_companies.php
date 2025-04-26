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
    $company_name = mysqli_real_escape_string($conn, trim($_POST['company_name']));
    $company_email = mysqli_real_escape_string($conn, trim($_POST['company_email']));
    $company_phone = mysqli_real_escape_string($conn, trim($_POST['company_phone']));
    $responsible_name = mysqli_real_escape_string($conn, trim($_POST['responsible_name']));
    $company_address = mysqli_real_escape_string($conn, trim($_POST['company_address']));
    $company_website = mysqli_real_escape_string($conn, trim($_POST['company_website']));
    $company_industry = mysqli_real_escape_string($conn, trim($_POST['company_industry']));
    $company_password = $_POST['company_password'];
    $confirm_password = $_POST['confirm_password'];

    // Password check
    if ($company_password !== $confirm_password) {
        echo "❌ Error: Passwords do not match.";
        exit;
    }

    // Hash the password before saving it to the database
    $hashed_password = password_hash($company_password, PASSWORD_BCRYPT);

    $dataConsent = isset($_POST['dataConsent']) ? 1 : 0;

    // Generate UUID for the company id
    $company_id = uniqid('comp_', true); // Unique company ID

    // Prepare the insert query for the companies table
    $query = "INSERT INTO companies 
        (id, company_name, company_email, company_phone, responsible_name, company_address, company_website, company_industry, company_password_hash, data_consent, allow_agency_contact, created_at) 
        VALUES ('$company_id', '$company_name', '$company_email', '$company_phone', '$responsible_name', '$company_address', '$company_website', '$company_industry', '$hashed_password', '$dataConsent', 1, NOW())";

    // Prepare the insert query for the users table
    $query_user = "INSERT INTO users (id, email, password_hash, role) 
                   VALUES ('$company_id', '$company_email', '$hashed_password', 'companies')"; NOW())";

    // Execute the company insert query
    if (mysqli_query($conn, $query)) {
        // Execute the user insert query if company insertion was successful
        if (mysqli_query($conn, $query_user)) {
            echo "✅ Company registered successfully.";
        } else {
            echo "❌ Error inserting into users table: " . mysqli_error($conn);
        }
    } else {
        echo "❌ Error: " . mysqli_error($conn);
    }
}

$conn->close();
?>
