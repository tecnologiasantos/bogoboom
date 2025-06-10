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
$user_id = $_SESSION['user_id'] ?? null; // Assuming user_id stored in session after login
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
        $uploadFileDir = '/var/www/html/bogoboom_project/uploads/agency_docs/'; // Ensure this dir exists and is writable
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

// Generate UUID for the agency
$agency_id = bin2hex(random_bytes(16));

// Prepare SQL Insert
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

// Bind parameters
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

// Execute
if ($stmt->execute()) {
    echo "✅ Agency registered successfully.";
    // Redirect after 5 seconds
    echo '<meta http-equiv="refresh" content="5;url=https://bogoboom.com">';
} else {
    echo "Error: " . $stmt->error;
}

$stmt->close();
$conn->close();
?>
