<?php
require 'vendor/autoload.php';
use Aws\S3\S3Client;
use Aws\Exception\AwsException;

// AWS and MariaDB configuration
$bucket = 'tecnologiasantos-mira';
$region = 'us-east-1';
$folder = 'newcomers/documents/';

// MariaDB credentials
$host = 'localhost';
$db = 'bogoboom';
$user = 'mira_admin';
$pass = 'Gottmitmir2025!';

// Connect to MariaDB
$conn = new mysqli($host, $user, $pass, $db);
if ($conn->connect_error) {
    die("DB connection failed: " . $conn->connect_error);
}

// Validate and upload file
if ($_SERVER['REQUEST_METHOD'] == 'POST') {
    if (isset($_FILES['document']) && $_FILES['document']['error'] == 0) {
        $file = $_FILES['document'];
        $filename = basename($file['name']);
        $extension = strtolower(pathinfo($filename, PATHINFO_EXTENSION));
        $allowed = ['pdf', 'docx'];

        if (!in_array($extension, $allowed)) {
            die("Only PDF and DOCX files are allowed.");
        }

        // S3 client with IAM role
        $s3 = new S3Client([
            'version' => 'latest',
            'region'  => $region,
        ]);

        $key = $folder . time() . '-' . $filename;

        try {
            $s3->putObject([
                'Bucket' => $bucket,
                'Key'    => $key,
                'SourceFile' => $file['tmp_name'],
                'ACL'    => 'private',
            ]);
        } catch (AwsException $e) {
            die('Upload error: ' . $e->getMessage());
        }

        // Insert data
        $stmt = $conn->prepare("INSERT INTO newcomers (full_name, email, phone, country, languages, document_path, housing_budget, data_consent, registration_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())");
        $stmt->bind_param("ssssssds", 
            $_POST['full_name'], $_POST['email'], $_POST['phone'], $_POST['country'],
            $_POST['languages'], $key, $_POST['housing_budget'], $_POST['data_consent']
        );

        if ($stmt->execute()) {
            echo "Newcomer registered successfully.";
        } else {
            echo "DB error: " . $stmt->error;
        }

        $stmt->close();
    } else {
        echo "Please upload a valid document.";
    }
}

$conn->close();
?>
