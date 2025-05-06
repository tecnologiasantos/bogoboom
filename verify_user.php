<?php
require '/var/www/html/bogoboom_project/safe/db_connection.php';

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $email = $_POST['email'];
    $password = $_POST['password'];

    // Prepare and execute the query to fetch user details
    $stmt = $conn->prepare("SELECT id, email, password_hash, role, status FROM users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $result = $stmt->get_result();

    // Check if user exists
    if ($result->num_rows === 1) {
        $user = $result->fetch_assoc();

        if (password_verify($password, $user['password_hash'])) {
            echo "<h2>✅ Login Successful</h2>";
            echo "<p>Email: <strong>" . htmlspecialchars($user['email']) . "</strong></p>";

            // Role-based Name fetch
            $nameValue = '';
            if ($user['role'] === 'newcomers') {
                $q = $conn->prepare("SELECT full_name AS name FROM newcomers WHERE user_id = ?");
            } elseif ($user['role'] === 'companies') {
                $q = $conn->prepare("SELECT company_name AS name FROM companies WHERE user_id = ?");
            } else { // 'agencies'
                $q = $conn->prepare("SELECT company_name AS name FROM agencies WHERE user_id = ?");
            }
            $q->bind_param("s", $user['id']);
            $q->execute();
            $r = $q->get_result();
            if ($r->num_rows === 1) {
                $row = $r->fetch_assoc();
                $nameValue = htmlspecialchars($row['name']);
            }
            $q->close();

            echo "<p>Name: <strong>{$nameValue}</strong></p>";
            echo "<p>User Role: <strong>" . htmlspecialchars($user['role']) . "</strong></p>";
            echo "<p>Account Status: <strong>" . htmlspecialchars($user['status']) . "</strong></p>";
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
    <title>User Status Checker</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Check User Login</h1>
    <form method="POST" action="verify_user.php">
        <label for="email">Email:</label><br>
        <input type="email" name="email" required><br><br>

        <label for="password">Password:</label><br>
        <input type="password" name="password" required><br><br>

        <input type="submit" value="Check User">
    </form>

    <form method="GET" action="verify_user.php" style="margin-top: 20px;">
        <button type="submit">🔄 Refresh / Check Another User</button>
    </form>
</body>
</html>

