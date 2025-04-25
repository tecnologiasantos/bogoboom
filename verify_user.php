<?php
require_once('../backend/config/db_connection.php'); // Update this path to your actual DB connection script

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $email = $_POST['email'];
    $password = $_POST['password'];

    // Prepare and execute the query
    $stmt = $conn->prepare("SELECT id, password_hash, role, is_verified FROM users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $result = $stmt->get_result();

    // Check if user exists
    if ($result->num_rows === 1) {
        $user = $result->fetch_assoc();

        if (password_verify($password, $user['password_hash'])) {
            echo "<h2>✅ Login Successful</h2>";
            echo "<p>User Role: <strong>" . htmlspecialchars($user['role']) . "</strong></p>";
            echo "<p>Email Verified: <strong>" . ($user['is_verified'] ? "Yes" : "No") . "</strong></p>";
        } else {
            echo "<h2>❌ Invalid password.</h2>";
        }
    } else {
        echo "<h2>❌ No user found with that email.</h2>";
    }

    $stmt->close();
    $conn->close();
}
?>

<!DOCTYPE html>
<html>
<head>
    <title>User Verification</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Verify User Login</h1>
    <form method="POST" action="verify_user.php">
        <label for="email">Email:</label><br>
        <input type="email" name="email" required><br><br>

        <label for="password">Password:</label><br>
        <input type="password" name="password" required><br><br>

        <input type="submit" value="Verify">
    </form>
</body>
</html>
