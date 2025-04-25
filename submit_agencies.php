<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session
session_start();

// Include the DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

// Process form data only if it's a POST request
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Sanitize and collect inputs
    $agency_name = trim($_POST['company_name']);
    $company_email = trim($_POST['company_email']);
    $company_phone = trim($_POST['company_phone']);
    $responsible_name = trim($_POST['responsible_name']);
    $company_address = trim($_POST['company_address']);
    $company_password = $_POST['company_password'];
    $confirm_password = $_POST['confirm_password'];
    $dataConsent = isset($_POST['dataConsent']) ? 1 : 0;  // Handling the consent checkbox

    // Check if passwords match
    if ($company_password !== $confirm_password) {
        echo "Error: Passwords do not match.";
        exit;
    }

    // Generate a unique ID for the agency registration
    $agency_id = uniqid('', true);
    $user_id = uniqid('', true); // Unique ID for the user as well

    // Handle file upload if there is a document uploaded
    $document = null;
    if (isset($_FILES['company_document']) && $_FILES['company_document']['error'] == 0) {
        // Define the absolute file upload directory
        $uploadDir = '/var/www/html/bogoboom_project/uploads/temp_files/';

        // Ensure the directory exists
        if (!file_exists($uploadDir)) {
            mkdir($uploadDir, 0777, true);  // Create the directory if it doesn't exist
        }

        // Set the path to where the uploaded file will be moved
        $uploadFile = $uploadDir . basename($_FILES['company_document']['name']);

        // Validate file type (optional)
        $allowedTypes = ['application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
        $fileType = mime_content_type($_FILES['company_document']['tmp_name']);

        if (in_array($fileType, $allowedTypes)) {
            // Move the uploaded file
            if (move_uploaded_file($_FILES['company_document']['tmp_name'], $uploadFile)) {
                $document = $uploadFile; // Store file path in the database
            } else {
                echo "Error: Could not upload the document.";
                exit;
            }
        } else {
            echo "Error: Invalid file type. Only PDF and DOC files are allowed.";
            exit;
        }
    }

    // Hash the password before storing
    $hashed_password = password_hash($company_password, PASSWORD_DEFAULT);

    // SQL query to insert data into the agencies table (updated to match the form and table schema)
    $query_agency = "INSERT INTO agencies (id, company_name, company_email, company_phone, responsible_name, company_address, company_password_hash, company_document, data_consent, created_at) 
                     VALUES ('$agency_id', '$agency_name', '$company_email', '$company_phone', '$responsible_name', '$company_address', '$hashed_password', '$document', '$dataConsent', NOW())";

    // SQL query to insert data into the users table (adjusted to match your schema)
    $query_user = "INSERT INTO users (id, email, password_hash, created_at) 
                   VALUES ('$user_id', '$company_email', '$hashed_password', NOW())";

    // Execute both queries using the database connection
    if (mysqli_query($conn, $query_agency) && mysqli_query($conn, $query_user)) {
        echo "✅ Agency and User registered successfully.";
    } else {
        echo "Error: " . mysqli_error($conn);
    }

    // Close the database connection
    mysqli_close($conn);
} else {
    echo "Invalid request method.";
}
?>
