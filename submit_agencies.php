<?php
// Enable full error reporting
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Start the session
session_start();

// Include the DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// Initialize variables safely
$company_name = $_POST['company_name'] ?? '';
$company_email = $_POST['company_email'] ?? '';
$company_phone = $_POST['company_phone'] ?? '';
$responsible_name = $_POST['responsible_name'] ?? '';
$company_address = $_POST['company_address'] ?? '';
$country = $_POST['country'] ?? '';
$state_or_department = $_POST['state'] ?? ''; // From the HTML form field named 'state'
$city = $_POST['city'] ?? '';
$company_website = $_POST['company_website'] ?? '';
$company_password = $_POST['company_password'] ?? '';
$confirm_password = $_POST['confirm_password'] ?? '';
$data_consent = isset($_POST['dataConsent']) ? 1 : 0;

// Validate required fields
if (
    empty($company_name) || empty($company_email) || empty($company_phone) ||
    empty($responsible_name) || empty($company_address) || empty($country) ||
    empty($state_or_department) || empty($city) || empty($company_password) || empty($confirm_password)
) {
    die('Please fill in all required fields.');
}

// Validate password match
if ($company_password !== $confirm_password) {
    die('Passwords do not match.');
}

// Password hashing
$company_password_hash = password_hash($company_password, PASSWORD_DEFAULT);

// Handle file upload
$company_document = null;
if (isset($_FILES['company_document']) && $_FILES['company_document']['error'] === UPLOAD_ERR_OK) {
    $fileTmpPath = $_FILES['company_document']['tmp_name'];
    $fileName = $_FILES['company_document']['name'];
    $fileSize = $_FILES['company_document']['size'];
    $fileType = $_FILES['company_document']['type'];
    $fileNameCmps = explode(".", $fileName);
    $fileExtension = strtolower(end($fileNameCmps));

    // Allowed file extensions
    $allowedExtensions = ['pdf', 'doc', 'docx'];
    if (in_array($fileExtension, $allowedExtensions)) {
        $newFileName = uniqid('doc_', true) . '.' . $fileExtension;
        $uploadFileDir = '/var/www/html/bogoboom_project/uploads/agency_docs/';
        $dest_path = $uploadFileDir . $newFileName;

        if (move_uploaded_file($fileTmpPath, $dest_path)) {
            $company_document = $newFileName;
        } else {
            die('Error moving the uploaded file.');
        }
    } else {
        die('Invalid file type. Only PDF, DOC, DOCX allowed.');
    }
}

// Generate UUID for the user (users table)
$user_id = bin2hex(random_bytes(16));

// Insert into 'users' table
$stmtUser = $conn->prepare("INSERT INTO users (id, email, full_name, phone, password_hash, role) VALUES (?, ?, ?, ?, ?, 'agencies')");
if (!$stmtUser) {
    die('Prepare failed for users table: ' . $conn->error);
}

$stmtUser->bind_param(
    "sssss",
    $user_id,
    $company_email,
    $company_name,
    $company_phone,
    $company_password_hash
);

if (!$stmtUser->execute()) {
    die('Error inserting into users table: ' . $stmtUser->error);
}

$stmtUser->close();

// Generate UUID for the agency (agencies table)
$agency_id = bin2hex(random_bytes(16));

// Prepare SQL Insert for 'agencies' table
$sql = "INSERT INTO agencies (
            id, user_id, company_name, company_email, company_phone,
            responsible_name, company_address, country, state_or_department,
            city, company_website, company_document, company_password_hash,
            data_consent
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)";

$stmt = $conn->prepare($sql);
if (!$stmt) {
    die('Prepare failed: ' . $conn->error);
}

// Bind parameters for 'agencies' table
$stmt->bind_param(
    "sssssssssssssi",
    $agency_id,
    $user_id,
    $company_name,
    $company_email,
    $company_phone,
    $responsible_name,
    $company_address,
    $country,
    $state_or_department,
    $city,
    $company_website,
    $company_document,
    $company_password_hash,
    $data_consent
);

// Execute insert for 'agencies' table
if ($stmt->execute()) {
    echo "✅ Agency registered successfully.";
    echo '<meta http-equiv="refresh" content="5;url=https://bogoboom.com">';
} else {
    echo "Error: " . $stmt->error;
}

$stmt->close();
$conn->close();
?>

