<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session (if you're using sessions, you can keep this)
session_start();

// Include the DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

// Retrieve and sanitize form data
$agency_name = isset($_POST['agency_name']) ? mysqli_real_escape_string($conn, trim($_POST['agency_name'])) : '';
$location = isset($_POST['location']) ? mysqli_real_escape_string($conn, trim($_POST['location'])) : '';
$description = isset($_POST['description']) ? mysqli_real_escape_string($conn, trim($_POST['description'])) : '';
$subscription_status = isset($_POST['subscription_status']) ? mysqli_real_escape_string($conn, trim($_POST['subscription_status'])) : 'inactive'; // Default value

// Generate a unique ID for the agency registration
$agency_id = uniqid('', true);

// Handle file upload if there is a document uploaded
$document = null;
if (isset($_FILES['document']) && $_FILES['document']['error'] == 0) {
    // Define the absolute file upload directory
    $uploadDir = '/var/www/html/bogoboom_project/uploads/temp_files/';

    // Ensure the directory exists
    if (!file_exists($uploadDir)) {
        mkdir($uploadDir, 0777, true);  // Create the directory if it doesn't exist
    }

    // Set the path to where the uploaded file will be moved
    $uploadFile = $uploadDir . basename($_FILES['document']['name']);

    // Validate file type (optional)
    $allowedTypes = ['application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
    $fileType = mime_content_type($_FILES['document']['tmp_name']);

    if (in_array($fileType, $allowedTypes)) {
        // Move the uploaded file
        if (move_uploaded_file($_FILES['document']['tmp_name'], $uploadFile)) {
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
$query = "INSERT INTO agencies (id, name, location, description, subscription_status, created_at) 
          VALUES ('$agency_id', '$agency_name', '$location', '$description', '$subscription_status', NOW())";

// Execute the query using the database connection
if (mysqli_query($conn, $query)) {
    echo "✅ Agency registered successfully.";
} else {
    echo "Error: " . mysqli_error($conn);
}

// Close the database connection
mysqli_close($conn);
?>
