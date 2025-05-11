<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

session_start();

require '/var/www/html/bogoboom_project/safe/db_connection.php'; 

// UUIDv4 generator
function generate_uuid_v4() {
    return sprintf(
        '%04x%04x-%04x-%04x-%04x-%04x%04x%04x',
        mt_rand(0, 0xffff), mt_rand(0, 0xffff),
        mt_rand(0, 0xffff),
        mt_rand(0, 0x0fff) | 0x4000,
        mt_rand(0, 0x3fff) | 0x8000,
        mt_rand(0, 0xffff), mt_rand(0, 0xffff), mt_rand(0, 0xffff)
    );
}

// Collect and sanitize input
$full_name = trim($_POST['full_name'] ?? '');
$email = trim($_POST['email'] ?? '');
$phone = trim($_POST['phone'] ?? '');
$country = trim($_POST['country'] ?? '');
$languages = trim($_POST['languages'] ?? '');
$password = $_POST['password'] ?? '';
$confirm_password = $_POST['confirm_password'] ?? '';

// Password match check
if ($password !== $confirm_password) {
    echo "❌ Passwords do not match.";
    exit;
}

// Secure password hash
$password_hash = password_hash($password, PASSWORD_BCRYPT);

// Generate UUIDs
$user_id = generate_uuid_v4();
$newcomer_id = generate_uuid_v4();

// Document upload handling
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
        $newFileName = $newcomer_id . '.' . $extension;
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
    // Insert into users table
    $stmtUser = $conn->prepare("INSERT INTO users (id, email, full_name, phone, password_hash, role) VALUES (?, ?, ?, ?, ?, 'newcomers')");
    $stmtUser->bind_param("sssss", $user_id, $email, $full_name, $phone, $password_hash);
    $stmtUser->execute();

    // Insert into newcomers table
    $stmtNewcomer = $conn->prepare("INSERT INTO newcomers (id, user_id, full_name, email, phone, country, languages, document) VALUES (?, ?, ?, ?, ?, ?, ?, ?)");
    $stmtNewcomer->bind_param("ssssssss", $newcomer_id, $user_id, $full_name, $email, $phone, $country, $languages, $document);
    $stmtNewcomer->execute();

    echo "✅ Newcomer registered successfully.";
    echo '<meta http-equiv="refresh" content="5;url=https://bogoboom.com">';

    $stmtUser->close();
    $stmtNewcomer->close();
} catch (Exception $e) {
    echo "❌ Exception: " . $e->getMessage();
}

$conn->close();
?>
