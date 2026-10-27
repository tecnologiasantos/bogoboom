<?php
require __DIR__ . '/lib/bootstrap.php';
use PHPMailer\PHPMailer\Exception as MailException;
use PHPMailer\PHPMailer\PHPMailer;

require_post();
csrf_verify();
if (!empty($_POST['website'] ?? '')) {   // honeypot: bots fill hidden fields
    exit;
}
rate_limit_or_fail('partner_inquiry', 5, 3600);

$fullName = clean($_POST['fullName'] ?? '', 160);
$email = strtolower(clean($_POST['emailAddress'] ?? '', 254));
$phone = clean($_POST['phoneNumber'] ?? '', 40);
$country = clean($_POST['country'] ?? '', 80);
$state = clean($_POST['stateOrDepartment'] ?? '', 80);
$city = clean($_POST['city'] ?? '', 80);
$collab = clean($_POST['collaborationType'] ?? '', 120);

if ($fullName === '' || $country === '' || $city === '' || $collab === '' || !valid_email($email)) {
    fail(400, 'Please fill in all required fields with valid values.');
}

$conn = db();
$honeypot = '';
$stmt = $conn->prepare('INSERT INTO partner_inquiries (full_name, email_address, phone_number, country, state_or_department, city, collaboration_type, honeypot_website_field) VALUES (?, ?, ?, ?, ?, ?, ?, ?)');
$stmt->bind_param('ssssssss', $fullName, $email, $phone, $country, $state, $city, $collab, $honeypot);
$stmt->execute();

// Notification e-mail. Credentials come from the environment (SMTP_USER / SMTP_PASS); if absent we only store the inquiry.
$smtpUser = env('SMTP_USER');
$smtpPass = env('SMTP_PASS');
$autoload = rtrim(env('APP_ROOT', '/var/www/html/bogoboom_project'), '/') . '/vendor/autoload.php';
if ($smtpUser && $smtpPass && is_readable($autoload)) {
    require_once $autoload;
    try {
        $mail = new PHPMailer(true);
        $mail->isSMTP();
        $mail->Host = env('SMTP_HOST', 'smtp.gmail.com');
        $mail->SMTPAuth = true;
        $mail->Username = $smtpUser;
        $mail->Password = $smtpPass;
        $mail->SMTPSecure = PHPMailer::ENCRYPTION_STARTTLS;
        $mail->Port = (int) env('SMTP_PORT', '587');
        $mail->setFrom(env('SMTP_FROM', $smtpUser), 'Bogoboom Support');
        $mail->addAddress(env('NOTIFY_TO', $smtpUser), 'Bogoboom Team');
        $mail->addReplyTo($email, $fullName);
        $mail->isHTML(true);
        $mail->Subject = 'New Partner Inquiry from ' . preg_replace('/[\r\n]+/', ' ', $fullName);
        $mail->Body = '<h2>New Partner Inquiry Details</h2>'
            . '<p><strong>Full Name:</strong> ' . h($fullName) . '</p><p><strong>Email:</strong> ' . h($email) . '</p>'
            . '<p><strong>Phone:</strong> ' . h($phone) . '</p><p><strong>Country:</strong> ' . h($country) . '</p>'
            . '<p><strong>State/Department:</strong> ' . h($state) . '</p><p><strong>City:</strong> ' . h($city) . '</p>'
            . '<p><strong>Collaboration Type:</strong> ' . h($collab) . '</p>';
        $mail->send();
    } catch (MailException $e) {
        error_log('[bogoboom] partner mail failed: ' . $e->getMessage());   // inquiry is already saved
    }
} else {
    error_log('[bogoboom] partner inquiry stored; SMTP not configured, no notification sent');
}

echo '✅ Your inquiry was successfully registered. Our team will reach out to you in a few days.';
echo '<meta http-equiv="refresh" content="10;url=/">';
