<?php
require '/var/www/html/bogoboom_project/safe/db_connection.php';

if (isset($_GET['id']) && isset($_GET['action'])) {
    $user_id = intval($_GET['id']);
    $action = $_GET['action'];

    if ($action === 'deactivate') {
        $new_status = 'inactive';
    } elseif ($action === 'activate') {
        $new_status = 'active';
    } else {
        die('Invalid action');
    }

    $stmt = $conn->prepare("UPDATE users SET status = ? WHERE id = ?");
    $stmt->bind_param('si', $new_status, $user_id);
    $stmt->execute();

    if ($stmt->affected_rows > 0) {
        header('Location: list_database.php'); // redirect back after update
        exit();
    } else {
        echo "Failed to update status.";
    }
} else {
    echo "Invalid request.";
}
?>
