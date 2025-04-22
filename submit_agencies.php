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
$agent_name = isset($_POST['agent_name']) ? mysqli_real_escape_string($conn, trim($_POST['agent_name'])) : '';
$email = isset($_POST['email']) ? mysqli_real_escape_string($conn, trim($_POST['email'])) : '';
$phone = isset($_POST['phone']) ? mysqli_real_escape_string($conn, trim($_POST['phone'])) : '';
$region = isset($_POST['region']) ? mysqli_real_escape_string($conn, trim($_POST['region'])) : '';

// Generate a unique ID for the agency registration (you can adjust this depending on the use case)
$registration_id = uniqid('', true);

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

// SQL query to insert data into the real_estate table (updated to match the form)
$query = "INSERT INTO real_estate (registration_id, agency_name, agent_name, email, phone, region, document_path, registration_date) 
          VALUES ('$registration_id', '$agency_name', '$agent_name', '$email', '$phone', '$region', '$document', NOW())";

// Execute the query using the database connection
if (mysqli_query($conn, $query)) {
    echo "✅ Company registered successfully.";
} else {
    echo "Error: " . mysqli_error($conn);
}

// Close the database connection
mysqli_close($conn);
?>
