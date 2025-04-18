<?php
// Database config
$host = "localhost";
$dbname = "bogoboom_db";
$username = "bogoboom_user";  // Updated username
$password = "Gottmitmir2025!";

// Load AWS SDK for PHP
require 'vendor/autoload.php'; // Composer autoload

use Aws\S3\S3Client;
use Aws\Exception\AwsException;

// Create S3 client using EC2 IAM role (no credentials required)
$s3 = new S3Client([
    'region'  => 'us-east-1',
    'version' => 'latest',
]);

$bucket = 'tecnologiasantos-mira';
$folder = 'newcomers/documents/';

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    // Sanitize form data
    $fullname = $_POST['fullname'] ?? '';
    $email = $_POST['email'] ?? '';
    $phone = $_POST['phone'] ?? '';
    $country = $_POST['country'] ?? '';
    $languages = $_POST['languages'] ?? '';
    $budget = $_POST['budget'] ?? '';
    $data_consent = isset($_POST['data_consent']) ? 1 : 0;
    $registration_date = date('Y-m-d H:i:s');

    // Handle file upload
    $file = $_FILES['document'];
    $file_name = basename($file['name']);
    $file_tmp = $file['tmp_name'];
    $file_ext = strtolower(pathinfo($file_name, PATHINFO_EXTENSION));

    // Allow only PDF and DOCX
    if (!in_array($file_ext, ['pdf', 'docx'])) {
        die("❌ Invalid file type. Only PDF and DOCX files are allowed.");
    }

    $new_file_name = uniqid("doc_") . "." . $file_ext;
    $s3_key = $folder . $new_file_name;

    try {
        // Upload to S3
        $result = $s3->putObject([
            'Bucket' => $bucket,
            'Key'    => $s3_key,
            'SourceFile' => $file_tmp,
            'ACL'    => 'private'
        ]);

        $document_path = $result['ObjectURL'];
    } catch (AwsException $e) {
        die("❌ File upload failed: " . $e->getMessage());
    }

    // Save to MariaDB
    try {
        $pdo = new PDO("mysql:host=$host;dbname=$dbname", $username, $password);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

        $stmt = $pdo->prepare("INSERT INTO newcomers 
            (fullname, email, phone, country, languages, document_path, housing_budget, data_consent, registration_date) 
            VALUES 
            (:fullname, :email, :phone, :country, :languages, :document_path, :budget, :data_consent, :registration_date)");

        $stmt->execute([
            ':fullname' => $fullname,
            ':email' => $email,
            ':phone' => $phone,
            ':country' => $country,
            ':languages' => $languages,
            ':document_path' => $document_path,
            ':budget' => $budget,
            ':data_consent' => $data_consent,
            ':registration_date' => $registration_date
        ]);

        echo "✅ Registration successful!";
    } catch (PDOException $e) {
        die("❌ Database error: " . $e->getMessage());
    }
} else {
    // HTML form with correct field names and file input
    echo <<<HTML
    <form action="submit_newcomers.php" method="POST" enctype="multipart/form-data">
        <input type="text" name="fullname" placeholder="Full Name" required><br>
        <input type="email" name="email" placeholder="Email" required><br>
        <input type="tel" name="phone" placeholder="Phone Number" required><br>
        <input type="text" name="country" placeholder="Country" required><br>
        <input type="text" name="languages" placeholder="Languages Spoken"><br>
        <input type="number" step="0.01" name="budget" placeholder="Housing Budget"><br>

        <label for="document">Upload CV or ID (PDF or DOCX):</label><br>
        <input type="file" name="document" accept=".pdf,.docx" required><br><br>

        <label>
            <input type="checkbox" name="data_consent" value="1"> I consent to data collection
        </label><br><br>

        <button type="submit">Register</button>
    </form>
    HTML;
}
?>
