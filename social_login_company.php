<?php
require_once '../../vendor/autoload.php';

$provider = $_GET['provider'];
$role = 'company'; // Role for company

$config = include 'hybridauth_company.php'; // Hybridauth config for company

try {
    $hybridauth = new Hybridauth\Hybridauth($config);
    $adapter = $hybridauth->authenticate($provider);
    $userProfile = $adapter->getUserProfile();
    
    // Handle user profile (save or login the user)
    header('Location: /backend/profiles/social_callback_company.php?provider=' . $provider);
    exit;
} catch (\Exception $e) {
    echo 'Error: ' . $e->getMessage();
}
