<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Use absolute path to the db_connection.php
require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

// Assuming session is used for user authentication and user_id is stored in the session
session_start();  // Start session to retrieve user ID

// Check if the user is logged in, otherwise handle error
if (!isset($_SESSION['user_id'])) {
    die("Error: User is not logged in.");
}

// Retrieve the logged-in user's ID
$user_id = $_SESSION['user_id'];
$created_by = $user_id;  // Assuming the creator of the profile is the logged-in user

// Set default values for optional fields if not provided
$fullname = isset($_POST['fullname']) ? mysqli_real_escape_string($conn, trim($_POST['fullname'])) : '';
$email = isset($_POST['email']) ? mysqli_real_escape_string($conn, trim($_POST['email'])) : '';
$phone = isset($_POST['phone']) ? mysqli_real_escape_string($conn, trim($_POST['phone'])) : '';
$country = isset($_POST['country']) ? mysqli_real_escape_string($conn, trim($_POST['country'])) : '';
$languages = isset($_POST['languages']) ? mysqli_real_escape_string($conn, trim($_POST['languages'])) : '';
$budget = isset($_POST['budget']) ? (float)$_POST['budget'] : 0.00;
$allow_contact = isset($_POST['allow_contact']) ? 1 : 0;  // Assuming checkbox for allow_contact
$city_of_contract = isset($_POST['city_of_contract']) ? mysqli_real_escape_string($conn, trim($_POST['city_of_contract'])) : ''; 
$arrival_date = isset($_POST['arrival_date']) && $_POST['arrival_date'] != '' ? mysqli_real_escape_string($conn, $_POST['arrival_date']) : null; 
$preferences = isset($_POST['preferences']) ? json_encode($_POST['preferences']) : ''; // Encode preferences as JSON

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
