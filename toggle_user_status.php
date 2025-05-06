<?php
require '/var/www/html/bogoboom_project/safe/db_connection.php';

if (isset($_GET['email']) && isset($_GET['action'])) {
    $user_email = trim($_GET['email']);
    $action = $_GET['action'];

    // Validate email format
    if (!filter_var($user_email, FILTER_VALIDATE_EMAIL)) {
        die('Invalid email format.');
    }

    if ($action === 'deactivate') {
        $new_status = 'inactive';
    } elseif ($action === 'activate') {
        $new_status = 'active';
    } else {
        die('Invalid action.');
    }

    // Prepare the query
    $stmt = $conn->prepare("UPDATE users SET status = ? WHERE email = ?");
    $stmt->bind_param('ss', $new_status, $user_email);
    $stmt->execute();

    if ($stmt->affected_rows > 0) {
        header('Location: list_database.php');
        exit();
    } else {
        echo "Failed to update status.";
    }
} else {
    echo "Invalid request.";
}
?>

