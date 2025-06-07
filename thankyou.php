<?php
// On backend/partner_inquiry.php
if ($_SERVER["REQUEST_METHOD"] === "POST") {
    // Process input and save
    // ...

    header("Location: /thank-you.html");
    exit;
}
?>
