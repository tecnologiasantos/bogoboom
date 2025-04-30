<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

session_start();

// Function to sanitize user inputs
function cleanInput($data) {
    return htmlspecialchars(stripslashes(trim($data)));
}

// Check if the user is already logged in
if (isset($_SESSION['user_id'])) {
    // Redirect logged-in users to their profile page based on their role
    if ($_SESSION['user_role'] == 'newcomer') {
        header("Location: /backend/profiles/newcomer_profile.php");
    } elseif ($_SESSION['user_role'] == 'company') {
        header("Location: /backend/profiles/company_profile.php");
    } else {
        header("Location: /backend/profiles/default_profile.php");
    }
    exit();
}

$error_message = ''; // Variable to store error messages

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    // Clean input data
    $email = cleanInput($_POST["email"]);
    $password = cleanInput($_POST["password"]);

    // Prepare the SQL query to check if the user exists
    $stmt = $conn->prepare("SELECT id, email, password_hash, role FROM users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $result = $stmt->get_result();

    // Check if the user exists and validate the password
    if ($result->num_rows > 0) {
        $user = $result->fetch_assoc();

        // Verify the password
        if (password_verify($password, $user['password_hash'])) {
            // Start the session and store user information
            $_SESSION['user_id'] = $user['id'];
            $_SESSION['user_email'] = $user['email'];
            $_SESSION['user_role'] = $user['role'];

            // Redirect based on user role
            if ($user['role'] == 'newcomer') {
                header("Location: /backend/profiles/newcomer_profile.php");
            } elseif ($user['role'] == 'company') {
                header("Location: /backend/profiles/company_profile.php");
            } else {
                header("Location: /backend/profiles/default_profile.php");
            }
            exit();
        } else {
            $error_message = "Incorrect password. Please try again.";
        }
    } else {
        $error_message = "No user found with that email address.";
    }

    $stmt->close();
}
?>

<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>Bogoboom | Login</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      margin: 0;
      padding: 0;
      background: #f5f5f5;
    }
    /* Logo styles */
    .image-container,
    .image-container-right {
      position: fixed;
      top: 50%;
      transform: translateY(-50%);
      z-index: 10;
    }
    .image-container {
      left: 10px;
    }
    .image-container-right {
      right: 10px;
    }
    .image-container img,
    .image-container-right img {
      width: 360px;
      height: auto;
    }

    @media (max-width: 1024px) {
      .image-container,
      .image-container-right {
        display: none;
      }
    }

    header {
      background-color: #004a8f;
      color: white;
      padding: 1rem 2rem;
      text-align: center;
    }
    main {
      padding: 2rem;
      text-align: center;
    }
    .intro {
      max-width: 600px;
      margin: 0 auto 2rem;
      font-size: 1.2rem;
    }
    section {
      max-width: 700px;
      margin: 2rem auto;
      text-align: left;
      background: #fff;
      padding: 2rem;
      border-radius: 8px;
      box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    section h3 {
      color: #004a8f;
      margin-bottom: 1rem;
    }
    section form {
      display: grid;
      gap: 1rem;
      max-width: 600px;
      margin: 0 auto;
    }
    section form div {
      margin-bottom: 1.5rem;
    }
    label {
      font-weight: bold;
      color: #004a8f;
    }
    input, button {
      width: 100%;
      padding: 0.8rem;
      border-radius: 6px;
      font-size: 1rem;
      border: 1px solid #ddd;
      box-sizing: border-box;
    }
    input[type="text"], input[type="email"], input[type="password"] {
      background-color: #f9f9f9;
    }
    button {
      background-color: #004a8f;
      color: white;
      font-size: 1.1rem;
      border: none;
      cursor: pointer;
      transition: background-color 0.3s ease;
    }
    button:hover {
      background-color: #003c6e;
    }
    footer {
      margin-top: 2rem;
      text-align: center;
      font-size: 0.9rem;
      color: #666;
    }
    .error-message {
      color: red;
      font-weight: bold;
      margin-bottom: 1rem;
    }
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
    <h1>Welcome to Bogoboom</h1>
    <p>Connecting newcomers in Bogotá with jobs and homes</p>
  </header>

  <main>
    <div class="intro">
      <p>Bogoboom is your bridge to a smooth landing in Bogotá. Whether you're a newcomer, an HR department, or a real estate agency, our platform helps you connect and thrive.</p>
    </div>

    <!-- Login Section -->
    <section>
      <h3>Login</h3>
      <?php if (!empty($error_message)): ?>
        <div class="error-message"><?php echo $error_message; ?></div>
      <?php endif; ?>
      <form action="/backend/profiles/login.php" method="POST">
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

  </main>

  <footer>
    &copy; 2025 Bogoboom. All rights reserved.
  </footer>

</body>
</html>
