<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

session_start();

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
    $country = mysqli_real_escape_string($conn, trim($_POST['country']));
    $state = mysqli_real_escape_string($conn, trim($_POST['state']));
    $city = mysqli_real_escape_string($conn, trim($_POST['city']));
    $company_password = $_POST['company_password'];
    $confirm_password = $_POST['confirm_password'];

    // Validate passwords
    if ($company_password !== $confirm_password) {
        echo "❌ Error: Passwords do not match.";
        exit;
    }

    $hashed_password = password_hash($company_password, PASSWORD_BCRYPT);
    $dataConsent = isset($_POST['dataConsent']) ? 1 : 0;

    // Generate unique ID
    $company_id = uniqid('comp_', true);

    // 1️⃣ Insert into users table
    $query_user = "INSERT INTO users (
            id, email, company_email, password_hash, role,
            company_name, company_phone, company_address, company_website
        ) VALUES (
            '$company_id', '$company_email', '$company_email', '$hashed_password', 'companies',
            '$company_name', '$company_phone', '$company_address', '$company_website'
        )";

    if (mysqli_query($conn, $query_user)) {
        // 2️⃣ Insert into companies table (with location)
        $query = "INSERT INTO companies (
            id, user_id, company_name, company_email, company_phone, responsible_name,
            company_address, company_website, company_industry, country, state, city,
            company_password_hash, data_consent, allow_agency_contact, created_at
        ) VALUES (
            '$company_id', '$company_id', '$company_name', '$company_email', '$company_phone', '$responsible_name',
            '$company_address', '$company_website', '$company_industry', '$country', '$state', '$city',
            '$hashed_password', '$dataConsent', 1, NOW()
        )";

        if (mysqli_query($conn, $query)) {
            echo "✅ Company registered successfully.";
            echo '<meta http-equiv="refresh" content="5;url=https://bogoboom.com">';
        } else {
            echo "❌ Error inserting into companies table: " . mysqli_error($conn);
        }
    } else {
        echo "❌ Error inserting into users table: " . mysqli_error($conn);
    }
}

$conn->close();
?>
