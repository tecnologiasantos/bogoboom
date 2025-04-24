<?php
session_start();
require_once '../../vendor/autoload.php';
require '../safe/db_connection.php';

$provider = $_GET['provider'];
$role = 'company'; // Role is predefined for this callback

$config = include 'hybridauth_company.php'; // Hybridauth config for company

try {
    $hybridauth = new Hybridauth\Hybridauth($config);
    $adapter = $hybridauth->authenticate($provider);
    $userProfile = $adapter->getUserProfile();
    
    // Check if company exists in the DB
    $email = $userProfile->email;
    $query = "SELECT * FROM companies WHERE email = '$email'";
    $result = mysqli_query($conn, $query);

    if (mysqli_num_rows($result) > 0) {
        // User exists, log them in (session-based login example)
        $_SESSION['user'] = mysqli_fetch_assoc($result);
        header("Location: /backend/profiles/company_profile.php");
    } else {
        // New user, create account (for example)
        $name = $userProfile->displayName;
        $insertQuery = "INSERT INTO companies (name, email) VALUES ('$name', '$email')";
        mysqli_query($conn, $insertQuery);

        // Log them in after registration
        $_SESSION['user'] = ['name' => $name, 'email' => $email];
        header("Location: /backend/profiles/company_profile.php");
    }
    exit;
} catch (\Exception $e) {
    echo 'Error: ' . $e->getMessage();
}
