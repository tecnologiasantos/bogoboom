<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session
session_start();

// Check if the user is logged in
if (!isset($_SESSION['user_id'])) {
    die("Error: User is not logged in.");
}

// Include the DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

// Retrieve the logged-in user's ID
$user_id = $_SESSION['user_id'];
$created_by = $user_id;  // Assuming the creator of the profile is the logged-in user

// Retrieve and sanitize form data
$fullname = isset($_POST['fullname']) ? mysqli_real_escape_string($conn, trim($_POST['fullname'])) : '';
$email = isset($_POST['email']) ? mysqli_real_escape_string($conn, trim($_POST['email'])) : '';
$phone = isset($_POST['phone']) ? mysqli_real_escape_string($conn, trim($_POST['phone'])) : '';
$country = isset($_POST['country']) ? mysqli_real_escape_string($conn, trim($_POST['country'])) : '';
$languages = isset($_POST['languages']) ? mysqli_real_escape_string($conn, trim($_POST['languages'])) : '';
// Generate a unique ID for the newcomer
$id = uniqid('', true);

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

// SQL query to insert data into the Newcomers table
$query = "INSERT INTO Newcomers (id, user_id, full_name, email, phone, country, city_of_contract, arrival_date, preferences, languages, document, allow_contact, created_by, uploaded_documents, housing_budget) 
          VALUES ('$id', '$user_id', '$fullname', '$email', '$phone', '$country', '$city_of_contract', '$arrival_date', '$preferences', '$languages', '$document', '$allow_contact', '$created_by', '$uploaded_documents', '$housing_budget')";

// Execute the query using the database connection
if (mysqli_query($conn, $query)) {
    echo "Newcomer registered successfully!";
} else {
    echo "Error: " . mysqli_error($conn);
}
?>
