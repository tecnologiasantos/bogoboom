<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Update the path to db_connection.php based on your structure
require '../../safe/db_connection.php'; // Make sure the path is correct relative to the current file

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
    // Sanitize and validate input
    $email = filter_var($_POST['email'], FILTER_SANITIZE_EMAIL);
    $password = $_POST['password'];
    $role = $_POST['role'];

    // Validate email format
    if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
        echo "Formato de email inválido.";
        exit;
    }

    // Ensure password is not empty
    if (empty($password)) {
        echo "La contraseña no puede estar vacía.";
        exit;
    }

    // Hash the password
    $password_hash = password_hash($password, PASSWORD_DEFAULT);
    
    // Generate UUID for the new user
    $uuid = generate_uuid();

    // Prepare the SQL statement to check if the email already exists
    $stmt = $conn->prepare("SELECT COUNT(*) FROM Users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $stmt->bind_result($email_count);
    $stmt->fetch();
    $stmt->close();

    // If email already exists, show an error message
    if ($email_count > 0) {
        echo "Este correo ya está registrado. Por favor intenta con otro.";
        exit;
    }

    // Prepare the insert query to create the new user
    $stmt = $conn->prepare("INSERT INTO Users (id, email, password_hash, role) VALUES (?, ?, ?, ?)");
    
    // Check if the query preparation is successful
    if (!$stmt) {
        die("Error al preparar la consulta: " . $conn->error);
    }

    // Bind parameters and execute the insert query
    $stmt->bind_param("ssss", $uuid, $email, $password_hash, $role);

    if ($stmt->execute()) {
        // After successful registration, redirect to the login page
        header('Location: /public/login.html');
        exit;
    } else {
        // If an error occurs during the insert, show the error
        echo "Error al insertar en la base de datos: " . $stmt->error;
    }

    // Close the prepared statement
    $stmt->close();
}
?>

