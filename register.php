<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

require '../../safe/db_connection.php';

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
    $email = filter_var($_POST['email'], FILTER_SANITIZE_EMAIL);
    $password = $_POST['password'];
    $role = $_POST['role'];

    if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
        echo "Formato de email inválido.";
        exit;
    }

    if (empty($password)) {
        echo "La contraseña no puede estar vacía.";
        exit;
    }

    $password_hash = password_hash($password, PASSWORD_DEFAULT);
    $uuid = generate_uuid();

    $stmt = $conn->prepare("INSERT INTO Users (id, email, password_hash, role) VALUES (?, ?, ?, ?)");

    if (!$stmt) {
        die("Error al preparar la consulta: " . $conn->error);
    }

    $stmt->bind_param("ssss", $uuid, $email, $password_hash, $role);

    if ($stmt->execute()) {
        header('Location: /public/login.html');
        exit;
    } else {
        if ($stmt->errno === 1062) {
            echo "Este correo ya está registrado. Por favor intenta con otro.";
        } else {
            echo "Error al insertar en la base de datos: " . $stmt->error;
        }
    }

    $stmt->close();
}
?>
