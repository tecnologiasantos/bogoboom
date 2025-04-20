<?php
require '../secret/db_connection.php'; // Adjust the path if necessary

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
    // Validate user inputs (you can expand these validations)
    $email = filter_var($_POST['email'], FILTER_SANITIZE_EMAIL);
    $password = $_POST['password'];
    $role = $_POST['role'];

    // Basic validation (you can expand this)
    if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
        echo "Invalid email format.";
        exit;
    }

    if (empty($password)) {
        echo "Password cannot be empty.";
        exit;
    }

    // Hash the password
    $password_hash = password_hash($password, PASSWORD_DEFAULT);

    // Generate a UUID for the new user
    $uuid = generate_uuid();

    // Prepare and execute the insert query
    $stmt = $conn->prepare("INSERT INTO Users (id, email, password_hash, role) VALUES (?, ?, ?, ?)");
    $stmt->bind_param("ssss", $uuid, $email, $password_hash, $role);

    if ($stmt->execute()) {
        // Optionally, redirect after successful registration
        header('Location: login.html'); // Redirect to login page or another appropriate page
        exit;
    } else {
        echo "Error: " . $stmt->error;
    }

    // Close the statement
    $stmt->close();
}
?>
