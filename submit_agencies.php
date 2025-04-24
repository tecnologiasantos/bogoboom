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

    // Check if passwords match
    if ($company_password !== $confirm_password) {
        echo "Error: Passwords do not match.";
        exit;
    }

    // Generate a unique ID for the agency registration
    $agency_id = uniqid('', true);

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

    // SQL query to insert data into the agencies table (updated to match the form)
    $query = "INSERT INTO agencies (id, name, email, phone, location, responsible_name, password, document, created_at) 
              VALUES ('$agency_id', '$agency_name', '$company_email', '$company_phone', '$company_address', '$responsible_name', '" . password_hash($company_password, PASSWORD_DEFAULT) . "', '$document', NOW())";

    // Execute the query using the database connection
    if (mysqli_query($conn, $query)) {
        echo "✅ Agency registered successfully.";
    } else {
        echo "Error: " . mysqli_error($conn);
    }

    // Close the database connection
    mysqli_close($conn);
} else {
    echo "Invalid request method.";
}
?>
