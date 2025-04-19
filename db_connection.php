<?php
$servername = "localhost";
$username = "bogoboom_user";
$password = "Gottmitmir2025!";
$database = "bogoboom";

// Create connection
$conn = new mysqli($servername, $username, $password, $database);

// Check connection
if ($conn->connect_error) {
    die("Connection failed: " . $conn->connect_error);
}
?>
