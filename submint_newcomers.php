<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session
session_start();

if (!isset($_SESSION['user_id'])) {
    die("Error: User is not logged in.");
}

// Include DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

$user_id = $_SESSION['user_id'];
$created_by = $user_id;

// Sanitize form inputs
$fullname = mysqli_real_escape_string($conn, trim($_POST['fullname'] ?? ''));
$email = mysqli_real_escape_string($conn, trim($_POST['email'] ?? ''));
$phone = mysqli_real_escape_string($conn, trim($_POST['phone'] ?? ''));
$country = mysqli_real_escape_string($conn, trim($_POST['country'] ?? ''));
$languages = mysqli_real_escape_string($conn, trim($_POST['languages'] ?? ''));
$budget = isset($_POST['budget']) ? (float)$_POST['budget'] : 0.00;

// Generate unique ID
$id = uniqid('', true);

// Handle optional document upload
$document = null;
if (isset($_FILES['document']) && $_FILES['document']['error'] === 0) {
    $uploadDir = '/var/www/html/bogoboom_project/uploads/temp_files/';
    if (!file_exists($uploadDir)) {
        mkdir($uploadDir, 0777, true);
    }

    $uploadFile = $uploadDir . basename($_FILES['document']['name']);
    $allowedTypes = [
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ];
    $fileType = mime_content_type($_FILES['document']['tmp_name']);

    if (in_array($fileType, $allowedTypes)) {
        if (move_uploaded_file($_FILES['document']['tmp_name'], $uploadFile)) {
            $document = $uploadFile;
        } else {
            die("Error: Could not upload the document.");
        }
    } else {
        die("Error: Invalid file type. Only PDF and DOC/DOCX allowed.");
    }
}

// Final SQL INSERT (matches fields you actually use)
$query = "INSERT INTO newcomers (
    id, user_id, full_name, email, phone, country, languages, document, created_by, housing_budget
) VALUES (
    '$id', '$user_id', '$fullname', '$email', '$phone', '$country', '$languages', '$document', '$created_by', '$budget'
)";

if (mysqli_query($conn, $query)) {
    echo "✅ Newcomer registered successfully!";
} else {
    echo "❌ Error: " . mysqli_error($conn);
}
?>

