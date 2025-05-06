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
    $name = mysqli_real_escape_string($conn, trim($_POST['name']));
    $email = mysqli_real_escape_string($conn, trim($_POST['email']));
    $phone = mysqli_real_escape_string($conn, trim($_POST['phone']));
    $responsible_name = mysqli_real_escape_string($conn, trim($_POST['responsible_name']));
    $address = mysqli_real_escape_string($conn, trim($_POST['address']));
    $website = mysqli_real_escape_string($conn, trim($_POST['website']));
    $industry = mysqli_real_escape_string($conn, trim($_POST['industry']));
    $password = $_POST['password'];
    $confirm_password = $_POST['confirm_password'];

    // Password check
    if ($password !== $confirm_password) {
        echo "❌ Error: Passwords do not match.";
        exit;
    }

    // Hash the password before saving it to the database
    $hashed_password = password_hash($password, PASSWORD_BCRYPT);

    $dataConsent = isset($_POST['dataConsent']) ? 1 : 0;

    // Generate UUID for both user and company (same id)
    $company_id = uniqid('comp_', true); // Use UUID or a better unique id generator if needed

    // 1️⃣ First Insert into users table
    $query_user = "INSERT INTO users (id, email, password_hash, role) 
                   VALUES ('$company_id', '$email', '$hashed_password', 'companies')";

    if (mysqli_query($conn, $query_user)) {
        // 2️⃣ Then Insert into companies table
        $query = "INSERT INTO companies 
            (id, user_id, company_name, company_email, company_phone, responsible_name, company_address, company_website, company_industry, company_password_hash, data_consent, allow_agency_contact, created_at) 
            VALUES ('$company_id', '$company_id', '$name', '$email', '$phone', '$responsible_name', '$address', '$website', '$industry', '$hashed_password', '$dataConsent', 1, NOW())";

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
