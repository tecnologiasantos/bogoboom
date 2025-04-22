<?php
// Step 1: Create a connection to the database
$conn = new mysqli('localhost', 'bogoboom_user', 'Gottmitmir2025!', 'bogoboom');

// Step 2: Check if the connection is successful
if ($conn->connect_error) {
    die("Connection failed: " . $conn->connect_error);  // If there’s an error, stop the script and show the error
}

// Step 3: If the connection is successful, print a success message
echo "Connected successfully";
?>
