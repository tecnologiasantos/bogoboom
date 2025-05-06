<?php
session_start();
require_once '../config/db_connection.php'; // adjust the path if needed

$title = "Bogoboom | User Panel";
$userRole = "";
$users = [];

if ($_SERVER["REQUEST_METHOD"] === "POST") {
    $email = $_POST['email'] ?? '';
    $password = $_POST['password'] ?? '';

    $stmt = $conn->prepare("SELECT id, role, password FROM users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $stmt->store_result();

    if ($stmt->num_rows === 1) {
        $stmt->bind_result($userId, $role, $hashedPassword);
        $stmt->fetch();

        if (password_verify($password, $hashedPassword)) {
            $userRole = ucfirst($role);
            $title = "Bogoboom | Logged in as $userRole";

            // Fetch all users
            $result = $conn->query("SELECT id, name, email, role FROM users ORDER BY role, name");
            if ($result && $result->num_rows > 0) {
                while ($row = $result->fetch_assoc()) {
                    $users[] = $row;
                }
            }
        } else {
            $error = "Invalid password.";
        }
    } else {
        $error = "No user found with that email.";
    }
    $stmt->close();
}
?>
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title><?= htmlspecialchars($title) ?></title>
  <style>
    /* (Same CSS as before — unchanged for brevity) */
    body { font-family: Arial, sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
    .image-container, .image-container-right {
      position: fixed; top: 50%; transform: translateY(-50%); z-index: 10;
    }
    .image-container { left: 10px; }
    .image-container-right { right: 10px; }
    .image-container img, .image-container-right img {
      width: 360px; height: auto;
    }
    @media (max-width: 1024px) {
      .image-container, .image-container-right { display: none; }
    }
    header {
      background-color: #004a8f; color: white; padding: 1rem 2rem; text-align: center;
    }
    main { padding: 2rem; text-align: center; }
    .intro { max-width: 600px; margin: 0 auto 2rem; font-size: 1.2rem; }
    section {
      max-width: 700px; margin: 2rem auto; text-align: left;
      background: #fff; padding: 2rem; border-radius: 8px;
      box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    section h3 { color: #004a8f; margin-bottom: 1rem; }
    label { font-weight: bold; color: #004a8f; }
    input, button {
      width: 100%; padding: 0.8rem; border-radius: 6px; font-size: 1rem;
      border: 1px solid #ddd; box-sizing: border-box;
    }
    input[type="text"], input[type="email"], input[type="password"] {
      background-color: #f9f9f9;
    }
    button {
      background-color: #004a8f; color: white; font-size: 1.1rem;
      border: none; cursor: pointer; transition: background-color 0.3s ease;
    }
    button:hover { background-color: #003c6e; }
    footer {
      margin-top: 2rem; text-align: center; font-size: 0.9rem; color: #666;
    }
    table {
      width: 100%; border-collapse: collapse; margin-top: 1rem;
    }
    th, td {
      padding: 0.8rem; text-align: left; border-bottom: 1px solid #ddd;
    }
    th { background-color: #004a8f; color: white; }
  </style>
</head>
<body>

  <!-- Logo containers -->
  <div class="image-container">
    <img src="https://tecnologiasantos-mira.s3.sa-east-1.amazonaws.com/mira/MIRA-powered.png" alt="MIRA Project Logo">
  </div>

  <div class="image-container-right">
    <img src="https://tecnologiasantos-mira.s3.sa-east-1.amazonaws.com/mira/logo-print-hd.jpg" alt="MIRA Project Logo Right">
  </div>

  <header>
    <h1><?= htmlspecialchars($title) ?></h1>
    <p>Connecting newcomers in Bogotá with jobs and homes</p>
  </header>

  <main>
    <div class="intro">
      <p>
        <?php if (!empty($userRole)): ?>
          Logged in as <strong><?= htmlspecialchars($userRole) ?></strong>.
        <?php else: ?>
          Please log in to view users.
        <?php endif; ?>
      </p>
    </div>

    <!-- Login Section -->
    <section>
      <h3>Login</h3>
      <?php if (!empty($error)): ?>
        <p style="color: red;"><?= htmlspecialchars($error) ?></p>
      <?php endif; ?>
      <form method="POST">
        <div>
          <label for="login-email">📧 Email:</label>
          <input type="email" id="login-email" name="email" required>
        </div>
        <div>
          <label for="login-password">🔒 Password:</label>
          <input type="password" id="login-password" name="password" required>
        </div>
        <div style="text-align: center;">
          <button type="submit">🔑 Login</button>
        </div>
      </form>
    </section>

    <!-- Display all users if logged in -->
    <?php if (!empty($users)): ?>
    <section>
      <h3>All Users</h3>
      <table>
        <thead>
          <tr>
            <th>ID</th><th>Name</th><th>Email</th><th>Role</th>
          </tr>
        </thead>
        <tbody>
          <?php foreach ($users as $u): ?>
            <tr>
              <td><?= $u['id'] ?></td>
              <td><?= htmlspecialchars($u['name']) ?></td>
              <td><?= htmlspecialchars($u['email']) ?></td>
              <td><?= ucfirst($u['role']) ?></td>
            </tr>
          <?php endforeach; ?>
        </tbody>
      </table>
    </section>
    <?php endif; ?>

  </main>

  <footer>
    &copy; 2025 Bogoboom. All rights reserved.
  </footer>

</body>
</html>

