<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session (optional unless you plan to use it)
session_start();

// Include the DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

// Retrieve and sanitize form data
$name = trim($_POST['name'] ?? '');
$email = trim($_POST['email'] ?? '');
$phone = trim($_POST['phone'] ?? '');
$country = trim($_POST['country'] ?? '');
$languages = trim($_POST['languages'] ?? '');
$password = $_POST['password'] ?? '';
$confirm_password = $_POST['confirm_password'] ?? '';

// Check if passwords match
if ($password !== $confirm_password) {
    echo "❌ Passwords do not match.";
    exit;
}

// Hash the password securely
$password_hash = password_hash($password, PASSWORD_DEFAULT);

// Generate a strong unique ID
$id = bin2hex(random_bytes(16)); // 32-character unique ID

// Handle file upload securely
$document = null;
if (!empty($_FILES['document']) && $_FILES['document']['error'] === UPLOAD_ERR_OK) {
    $uploadDir = '/var/www/html/bogoboom_project/uploads/temp_files/';
    
    if (!file_exists($uploadDir)) {
        mkdir($uploadDir, 0755, true); // More secure permissions
    }

    $allowedMimeTypes = [
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ];
    
    $fileType = mime_content_type($_FILES['document']['tmp_name']);
    $extension = pathinfo($_FILES['document']['name'], PATHINFO_EXTENSION);
    
    if (in_array($fileType, $allowedMimeTypes)) {
        // Rename file to avoid overwriting and security issues
        $newFileName = $id . '.' . $extension;
        $uploadFile = $uploadDir . $newFileName;

        if (move_uploaded_file($_FILES['document']['tmp_name'], $uploadFile)) {
            $document = $uploadFile;
        } else {
            echo "❌ Error uploading document.";
            exit;
        }
    } else {
        echo "❌ Invalid document type.";
        exit;
    }
}

// Use prepared statements to avoid SQL Injection
try {
    // Insert into users table
    $stmtUser = $conn->prepare("INSERT INTO users (id, email, password, role) VALUES (?, ?, ?, 'newcomers')");
    $stmtUser->bind_param("sss", $id, $email, $password_hash);

    // Insert into newcomers table
    $stmtNewcomer = $conn->prepare("INSERT INTO newcomers (id, name, email, phone, country, languages, document) VALUES (?, ?, ?, ?, ?, ?, ?)");
    $stmtNewcomer->bind_param("sssssss", $id, $name, $email, $phone, $country, $languages, $document);

    if ($stmtUser->execute()) {
        if ($stmtNewcomer->execute()) {
            echo "✅ Newcomer registered successfully.";
            // Redirect after 5 seconds
            echo '<meta http-equiv="refresh" content="5;url=https://bogoboom.com">';
        } else {
            echo "❌ Error saving newcomer: " . $stmtNewcomer->error;
        }
    } else {
        echo "❌ Error creating user: " . $stmtUser->error;
    }

    $stmtUser->close();
    $stmtNewcomer->close();
} catch (Exception $e) {
    error_log("Registration Error: " . $e->getMessage());
    echo "❌ An unexpected error occurred. Please try again later.";
}

$conn->close();
?>
