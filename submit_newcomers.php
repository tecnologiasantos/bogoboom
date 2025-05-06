<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session
session_start();

// Include the DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';



$user_id = $_SESSION['user_id'];

// Collect form data with checks
$name = isset($_POST['fullname']) ? trim($_POST['fullname']) : '';
$email = isset($_POST['email']) ? trim($_POST['email']) : '';
$phone = isset($_POST['phone']) ? trim($_POST['phone']) : '';
$country = isset($_POST['country']) ? trim($_POST['country']) : '';
$city = isset($_POST['city']) ? trim($_POST['city']) : ''; // Added check for city
$languages = isset($_POST['languages']) ? trim($_POST['languages']) : '';
$password = isset($_POST['password']) ? trim($_POST['password']) : '';
$confirm_password = isset($_POST['confirm_password']) ? trim($_POST['confirm_password']) : '';

// Validate if passwords match
if ($password !== $confirm_password) {
    echo "Passwords do not match!";
    exit;
}

// Handle file upload (if any)
$document = NULL;  // Default to NULL if no file is uploaded
if (isset($_FILES['document']) && $_FILES['document']['error'] == UPLOAD_ERR_OK) {
    $document_name = $_FILES['document']['name'];
    $document_tmp_name = $_FILES['document']['tmp_name'];
    $document_type = $_FILES['document']['type'];

    // Validate document type (PDF, DOC, DOCX)
    $allowed_types = ['application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
    if (!in_array($document_type, $allowed_types)) {
        echo "Invalid document type. Only PDF, DOC, and DOCX are allowed.";
        exit;
    }

    // Set the upload directory (make sure the directory exists and is writable)
    $upload_dir = '/var/www/html/bogoboom_project/uploads/documents/';
    $document_path = $upload_dir . basename($document_name);

    // Move the uploaded file to the specified directory
    if (move_uploaded_file($document_tmp_name, $document_path)) {
        $document = $document_path; // Store the file path in the database
    } else {
        echo "Error uploading document.";
        exit;
    }
}

// Prepare SQL query to insert newcomer data into the database
$sql = "INSERT INTO newcomers (id, user_id, name, email, phone, country, city, languages, document, contact, created_by, files)
        VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?, ?, TRUE, ?, ?)";

// Prepare the statement
if ($stmt = $conn->prepare($sql)) {
    // Prepare files data (you can store more than just document paths here if necessary)
    $files = json_encode(['document' => $document]);

    $stmt->bind_param("ssssssssss", $user_id, $name, $email, $phone, $country, $city, $languages, $document, $user_id, $files);

    // Execute the query
    if ($stmt->execute()) {
        echo "Registration successful! Welcome to Bogoboom.";
    } else {
        // Output the error message if the query fails
        echo "Error executing query: " . $stmt->error;
    }

    // Close the statement
    $stmt->close();
} else {
    // Output the error message if preparing the statement fails
    echo "Error preparing the statement: " . $conn->error;
}

// Close the database connection
$conn->close();
?>
