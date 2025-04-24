<?php
require_once '../../vendor/autoload.php';

$provider = $_GET['provider'];
$role = 'agency'; // Role for agency

$config = include 'hybridauth_agency.php'; // Hybridauth config for agency

try {
    $hybridauth = new Hybridauth\Hybridauth($config);
    $adapter = $hybridauth->authenticate($provider);
    $userProfile = $adapter->getUserProfile();
    
    // Handle user profile (save or login the user)
    header('Location: /backend/profiles/social_callback_agency.php?provider=' . $provider);
    exit;
} catch (\Exception $e) {
    echo 'Error: ' . $e->getMessage();
}
