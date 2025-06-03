<?php
// Include database connection
require_once '../backend/config/db_connection.php'; // Adjust path as needed

if ($_SERVER["REQUEST_METHOD"] === "POST") {
    // Sanitize and validate input
    $name     = trim($_POST["name"] ?? '');
    $email    = trim($_POST["email"] ?? '');
    $phone    = trim($_POST["phone"] ?? '');
    $country  = trim($_POST["country"] ?? '');
    $state    = trim($_POST["state"] ?? '');
    $city     = trim($_POST["city"] ?? '');
    $interest = trim($_POST["interest"] ?? '');

    // Optional: Combine state and city if you want to store more specific location
    $full_city = $state ? "$state - $city" : $city;

    // Validate required fields
    if (empty($name) || empty($email) || empty($country) || empty($city) || empty($interest)) {
        http_response_code(400);
        echo "Please fill in all required fields.";
        exit;
    }

    // Prepare and execute insert statement
    try {
        $stmt = $pdo->prepare("INSERT INTO leads (name, email, phone, city, country, interest)
                               VALUES (:name, :email, :phone, :city, :country, :interest)");
        $stmt->execute([
            ':name'     => $name,
            ':email'    => $email,
            ':phone'    => $phone,
            ':city'     => $full_city,
            ':country'  => $country,
            ':interest' => $interest
        ]);

        echo "Thank you! Your application has been received.";
    } catch (PDOException $e) {
        http_response_code(500);
        echo "Error saving lead: " . $e->getMessage();
    }
} else {
    http_response_code(405);
    echo "Method not allowed.";
}
?>
