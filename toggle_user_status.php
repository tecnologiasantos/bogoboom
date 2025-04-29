<?php
require '/var/www/html/bogoboom_project/safe/db_connection.php';

if (isset($_GET['id']) && isset($_GET['action'])) {
    $user_id = trim($_GET['id']);
    $action = $_GET['action'];

    // Validate 32-character hex UUID (no dashes)
    if (!preg_match('/^[0-9a-fA-F]{32}$/', $user_id)) {
        die('Invalid user ID format.');
    }

    if ($action === 'deactivate') {
        $new_status = 'inactive';
    } elseif ($action === 'activate') {
        $new_status = 'active';
    } else {
        die('Invalid action.');
    }

    // Prepare the query
    $stmt = $conn->prepare("UPDATE users SET status = ? WHERE id = ?");
    $stmt->bind_param('ss', $new_status, $user_id);
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
