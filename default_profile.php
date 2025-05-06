<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

session_start();

if (!isset($_SESSION['user_id'])) {
    // Redirect to login page if not logged in
    header("Location: /login.php");
    exit();
}

$user_id = $_SESSION['user_id'];
$user_email = $_SESSION['user_email'];
$user_role = $_SESSION['user_role'];

?>

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Default Profile | Bogoboom</title>
    <style>
        /* Add your styles here */
    </style>
</head>
<body>

<header>
    <h1>Welcome to Your Profile, <?php echo $user_email; ?>!</h1>
</header>

<main>
    <h2>Default Profile</h2>
    <p>It seems like your account doesn't have a specific role assigned yet. Please check your account settings to complete your profile.</p>

    <!-- Provide options to complete their profile -->
    <a href="/profile_settings.php">Update Profile</a>
</main>

<footer>
    &copy; 2025 Bogoboom. All rights reserved.
</footer>

</body>
</html>

