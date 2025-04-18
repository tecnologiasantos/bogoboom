<?php
require 'vendor/autoload.php';
use Aws\S3\S3Client;
use Aws\Exception\AwsException;

$bucket = 'tecnologiasantos-mira';
$region = 'us-east-1';
$folder = 'companies/documents/';

$host = 'localhost';
$db = 'bogoboom_db';
$user = 'bogoboom_user';
$pass = 'Gottmitmir2025!';

$conn = new mysqli($host, $user, $pass, $db);
if ($conn->connect_error) {
    die("DB connection failed: " . $conn->connect_error);
}

if ($_SERVER['REQUEST_METHOD'] == 'POST') {
    if (isset($_FILES['document']) && $_FILES['document']['error'] == 0) {
        $file = $_FILES['document'];
        $filename = basename($file['name']);
        $extension = strtolower(pathinfo($filename, PATHINFO_EXTENSION));
        $allowed = ['pdf', 'docx'];

        if (!in_array($extension, $allowed)) {
            die("Only PDF and DOCX files are allowed.");
        }

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

        $stmt = $conn->prepare("INSERT INTO companies (company_name, contact_person, email, phone, website, document_path, registration_date) VALUES (?, ?, ?, ?, ?, ?, NOW())");
        $stmt->bind_param("ssssss", 
            $_POST['company_name'], $_POST['contact_person'], $_POST['email'], 
            $_POST['phone'], $_POST['website'], $key
        );

        if ($stmt->execute()) {
            echo "Company registered successfully.";
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
