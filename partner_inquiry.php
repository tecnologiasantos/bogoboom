<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Include your DB config file
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// Load PHPMailer classes (make sure you installed PHPMailer via Composer)
use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

require '/var/www/html/bogoboom_project/vendor/autoload.php'; // Adjust path if needed

// Check if the form was submitted via POST
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    
    // Honeypot anti-bot field
    if (!empty($_POST['website'])) {
        // Bot detected — silently ignore
        exit;
    }

    // Sanitize and assign form data
    $fullName = trim($_POST['fullName']);
    $email = filter_var(trim($_POST['emailAddress']), FILTER_SANITIZE_EMAIL);
    $phone = trim($_POST['phoneNumber']);
    $country = trim($_POST['country']);
    $state = trim($_POST['stateOrDepartment'] ?? '');
    $city = trim($_POST['city']);
    $collabType = trim($_POST['collaborationType']);

    // Validate required fields
    if (empty($fullName) || empty($email) || empty($country) || empty($city) || empty($collabType)) {
        http_response_code(400);
        echo "Please fill in all required fields.";
        exit;
    }

    // Validate email format
    if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
        http_response_code(400);
        echo "Invalid email address.";
        exit;
    }

    // Prepare and execute SQL
    $stmt = $conn->prepare("
        INSERT INTO partner_inquiries 
        (full_name, email_address, phone_number, country, state_or_department, city, collaboration_type, honeypot_website_field) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ");

    $honeypot = ''; // Always empty because real users won't fill it
    $stmt->bind_param(
        "ssssssss",
        $fullName,
        $email,
        $phone,
        $country,
        $state,
        $city,
        $collabType,
        $honeypot
    );

    if ($stmt->execute()) {
        // Send notification email via PHPMailer
        $mail = new PHPMailer(true);
        try {
            // SMTP settings for Gmail with app password
            $mail->isSMTP();
            $mail->Host = 'smtp.gmail.com';
            $mail->SMTPAuth = true;
            $mail->Username = 'support@tecnologiasantos.com';
            $mail->Password = 'jzxu cxip rwuv ichj'; // Your app password
            $mail->SMTPSecure = PHPMailer::ENCRYPTION_STARTTLS;
            $mail->Port = 587;

            $mail->setFrom('support@tecnologiasantos.com', 'Bogoboom Support');
            $mail->addAddress('support@tecnologiasantos.com', 'Bogoboom Team');

            $mail->isHTML(true);
            $mail->Subject = 'New Partner Inquiry from ' . htmlspecialchars($fullName);
            $mail->Body = "
                <h2>New Partner Inquiry Details</h2>
                <p><strong>Full Name:</strong> " . htmlspecialchars($fullName) . "</p>
                <p><strong>Email:</strong> " . htmlspecialchars($email) . "</p>
                <p><strong>Phone:</strong> " . htmlspecialchars($phone) . "</p>
                <p><strong>Country:</strong> " . htmlspecialchars($country) . "</p>
                <p><strong>State/Department:</strong> " . htmlspecialchars($state) . "</p>
                <p><strong>City:</strong> " . htmlspecialchars($city) . "</p>
                <p><strong>Collaboration Type:</strong> " . htmlspecialchars($collabType) . "</p>
            ";

            $mail->send();
        } catch (Exception $e) {
            // You can log this error or handle it silently
            error_log("Mail error: " . $mail->ErrorInfo);
        }

        echo "✅ Your inquiry was successfully registered. Have in mind that our stuff will reach out to you in some days.";
        echo '<meta http-equiv="refresh" content="10;url=https://bogoboom.com">';
    } else {
        http_response_code(500);
        echo "Error saving your information. Please try again later.";
    }

    $stmt->close();
    $conn->close();

} else {
    http_response_code(405);
    echo "Method Not Allowed";
}
?>

