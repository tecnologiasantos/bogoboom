<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

session_start();

require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

$full_name = trim($_POST['full_name'] ?? '');
$email = trim($_POST['email'] ?? '');
$phone = trim($_POST['phone'] ?? '');
$country = trim($_POST['country'] ?? '');
$languages = trim($_POST['languages'] ?? '');
$password = $_POST['password'] ?? '';
$confirm_password = $_POST['confirm_password'] ?? '';

if ($password !== $confirm_password) {
    echo "❌ Passwords do not match.";
    exit;
}

$password_hash = password_hash($password, PASSWORD_DEFAULT);
$id = bin2hex(random_bytes(16));

$document = null;
if (!empty($_FILES['document']) && $_FILES['document']['error'] === UPLOAD_ERR_OK) {
    $uploadDir = '/var/www/html/bogoboom_project/uploads/temp_files/';
    
    if (!file_exists($uploadDir)) {
        mkdir($uploadDir, 0755, true);
    }

    $allowedMimeTypes = [
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ];
    
    $fileType = mime_content_type($_FILES['document']['tmp_name']);
    $extension = pathinfo($_FILES['document']['name'], PATHINFO_EXTENSION);
    
    if (in_array($fileType, $allowedMimeTypes)) {
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

try {
    $stmtUser = $conn->prepare("INSERT INTO users (id, email, full_name, phone, password_hash, role) VALUES (?, ?, ?, ?, ?, 'newcomers')");
    $stmtUser->bind_param("sssss", $id, $email, $full_name, $phone, $password_hash);
    $stmtUser->execute();

    $stmtNewcomer = $conn->prepare("INSERT INTO newcomers (id, full_name, email, phone, country, languages, document) VALUES (?, ?, ?, ?, ?, ?, ?)");
    
    // Avoid passing a function directly (like move_uploaded_file() return value) to bind_param
    $doc_path = $document;
    $stmtNewcomer->bind_param("sssssss", $id, $full_name, $email, $phone, $country, $languages, $doc_path);
    $stmtNewcomer->execute();

    echo "✅ Newcomer registered successfully.";
    echo '<meta http-equiv="refresh" content="5;url=https://bogoboom.com">';

    $stmtUser->close();
    $stmtNewcomer->close();
} catch (Exception $e) {
    // Development-time debugging: Show exact error
    echo "❌ Exception: " . $e->getMessage();
}

$conn->close();
?>
