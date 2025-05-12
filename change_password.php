<?php
session_start();
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// Redirect if not logged in
if (!isset($_SESSION['user_email'])) {
    header("Location: index.html");
    exit;
}

$email = $_SESSION['user_email'];
$message = "";

// Handle form submission
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $current_password = $_POST['current_password'];
    $new_password = $_POST['new_password'];
    $confirm_password = $_POST['confirm_password'];

    // Check if new passwords match
    if ($new_password !== $confirm_password) {
        $message = "❌ New passwords do not match.";
    } else {
        $stmt = $conn->prepare("SELECT password_hash FROM users WHERE email = ?");
        $stmt->bind_param("s", $email);
        $stmt->execute();
        $stmt->bind_result($stored_hash);
        $stmt->fetch();
        $stmt->close();

        // Check current password
        if (!password_verify($current_password, $stored_hash)) {
            $message = "❌ Current password is incorrect.";
        } else {
            // Hash and update the new password
            $new_hash = password_hash($new_password, PASSWORD_DEFAULT);
            $update_stmt = $conn->prepare("UPDATE users SET password_hash = ? WHERE email = ?");
            $update_stmt->bind_param("ss", $new_hash, $email);
            $update_stmt->execute();
            $update_stmt->close();

            echo "<!DOCTYPE html>
            <html><head><meta charset='UTF-8'>
            <meta http-equiv='refresh' content='5;url=index.html'>
            <title>Password Changed</title>
            <style>
              body { font-family: Arial; background: #f9f9f9; text-align: center; padding: 50px; }
              .message { background: #d4edda; color: #155724; padding: 20px; border: 1px solid #c3e6cb; border-radius: 8px; display: inline-block; }
            </style>
            </head><body>
            <div class='message'>
              ✅ Password changed successfully. You will be redirected in 5 seconds...
            </div>
            </body></html>";
            exit;
        }
    }
}
?>

<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Change Password</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f2f2f2; padding: 2rem; }
        .container { max-width: 400px; margin: auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        h2 { text-align: center; color: #004a8f; }
        label { font-weight: bold; display: block; margin-top: 1rem; }
        input { width: 100%; padding: 0.8rem; margin-top: 0.5rem; border-radius: 6px; border: 1px solid #ccc; }
        button { margin-top: 1.5rem; background: #004a8f; color: white; padding: 0.8rem; border: none; border-radius: 6px; width: 100%; cursor: pointer; }
        button:hover { background: #003366; }
        .error { color: red; margin-top: 1rem; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Change Password</h2>
        <form method="POST">
            <label for="current_password">Current Password</label>
            <input type="password" id="current_password" name="current_password" required>

            <label for="new_password">New Password</label>
            <input type="password" id="new_password" name="new_password" required>

            <label for="confirm_password">Confirm New Password</label>
            <input type="password" id="confirm_password" name="confirm_password" required>

            <button type="submit">🔒 Update Password</button>
        </form>
        <?php if ($message): ?>
            <div class="error"><?= htmlspecialchars($message) ?></div>
        <?php endif; ?>
    </div>
</body>
</html>
