<?php
require 'db_connection.php'; // file with mysqli connection code

function generate_uuid() {
    return sprintf('%04x%04x-%04x-%04x-%04x-%04x%04x%04x',
        mt_rand(0, 0xffff), mt_rand(0, 0xffff),
        mt_rand(0, 0xffff),
        mt_rand(0, 0x0fff) | 0x4000,
        mt_rand(0, 0x3fff) | 0x8000,
        mt_rand(0, 0xffff), mt_rand(0, 0xffff), mt_rand(0, 0xffff)
    );
}

if ($_SERVER["REQUEST_METHOD"] === "POST") {
    $email = $_POST['email'];
    $password = password_hash($_POST['password'], PASSWORD_DEFAULT);
    $role = $_POST['role'];
    $uuid = generate_uuid();

    $stmt = $conn->prepare("INSERT INTO Users (id, email, password_hash, role) VALUES (?, ?, ?, ?)");
    $stmt->bind_param("ssss", $uuid, $email, $password, $role);

    if ($stmt->execute()) {
        echo "User registered!";
        // Optionally redirect to role-specific profile form
    } else {
        echo "Error: " . $stmt->error;
    }
    $stmt->close();
}
?>
