<?php
session_start();
require '/var/www/html/bogoboom_project/safe/db_connection.php';

if (!isset($_SESSION['user_email'])) {
    echo "Access denied.";
    exit;
}

$message = "";

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $email = $_SESSION['user_email'];
    $current_pass = $_POST['current_password'];
    $new_pass = $_POST['new_password'];

    $stmt = $conn->prepare("SELECT password_hash FROM users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $stmt->bind_result($password_hash);
    $stmt->fetch();
    $stmt->close();

    if (password_verify($current_pass, $password_hash)) {
        $new_hash = password_hash($new_pass, PASSWORD_DEFAULT);
        $stmt = $conn->prepare("UPDATE users SET password_hash = ? WHERE email = ?");
        $stmt->bind_param("ss", $new_hash, $email);
        if ($stmt->execute()) {
            $message = "✅ Password changed successfully.";
        } else {
            $message = "❌ Failed to update password.";
        }
        $stmt->close();
    } else {
        $message = "❌ Current password incorrect.";
    }
}
?>

<!DOCTYPE html>
<html>
<head>
  <title>Change Password</title>
  <style>
    body { font-family: Arial; padding: 2rem; background: #f4f4f4; }
    form { background: white; padding: 2rem; border-radius: 8px; max-width: 400px; margin: auto; }
    input { display: block; width: 100%; padding: 0.8rem; margin-bottom: 1rem; }
    button { background-color: #004a8f; color: white; padding: 0.8rem; width: 100%; border: none; border-radius: 5px; }
    .message { text-align: center; margin-top: 1rem; color: #d00; }
  </style>
</head>
<body>
  <h2 style='text-align:center;'>Change Password</h2>
  <form method="POST">
    <label for="current_password">Current Password:</label>
    <input type="password" name="current_password" required>
    <label for="new_password">New Password:</label>
    <input type="password" name="new_password" required>
    <button type="submit">Update Password</button>
  </form>
  <div class="message"><?php echo htmlspecialchars($message); ?></div>
</body>
</html>
