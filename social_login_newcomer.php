<?php
require_once '../../vendor/autoload.php';

$provider = $_GET['provider'];
$role = 'newcomer'; // Role for newcomer

$config = include 'hybridauth_newcomer.php'; // Hybridauth config for newcomer

try {
    $hybridauth = new Hybridauth\Hybridauth($config);
    $adapter = $hybridauth->authenticate($provider);
    $userProfile = $adapter->getUserProfile();
    
    // Handle user profile (save or login the user)
    header('Location: /backend/profiles/social_callback_newcomer.php?provider=' . $provider);
    exit;
} catch (\Exception $e) {
    echo 'Error: ' . $e->getMessage();
}
