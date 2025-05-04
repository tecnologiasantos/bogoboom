<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// Start the session
session_start();

$authenticated = false;
$current_user_email = "";
$current_user_role = "";

// Login logic
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['action']) && $_POST['action'] == 'login') {
    $email = $_POST['email'];
    $password = $_POST['password'];

    $stmt = $conn->prepare("SELECT * FROM users WHERE email = ? LIMIT 1");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $result = $stmt->get_result();

    if ($result && $result->num_rows === 1) {
        $user = $result->fetch_assoc();
        if ($user['status'] === 'active' && password_verify($password, $user['password_hash'])) {
            $authenticated = true;
            $_SESSION['user_email'] = $user['email'];
            $_SESSION['user_role'] = $user['role'];  // Store user role
            $current_user_email = $user['email'];
            $current_user_role = $user['role'];  // Set current user role
        }
    }

    if (!$authenticated) {
        echo "Access denied. Incorrect credentials or inactive user.";
        exit;
    }
} elseif (isset($_POST['action']) && $_POST['action'] == 'logout') {
    session_destroy();
    header("Location: " . $_SERVER['PHP_SELF']);
    exit;
} elseif (isset($_SESSION['user_email'])) {
    $current_user_email = $_SESSION['user_email'];
    $current_user_role = $_SESSION['user_role'];  // Get the current user role from session
} else {
    echo "<!DOCTYPE html>
<html lang='en'>
<head>
  <meta charset='UTF-8' />
  <meta name='viewport' content='width=device-width, initial-scale=1.0'/>
  <title>Bogoboom | Login</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
    header { background-color: #004a8f; color: white; padding: 1rem 2rem; text-align: center; }
    main { padding: 2rem; text-align: center; }
    .intro { max-width: 600px; margin: 0 auto 2rem; font-size: 1.2rem; }
    section { max-width: 700px; margin: 2rem auto; background: #fff; padding: 2rem; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1); }
    h3 { color: #004a8f; }
    form { display: grid; gap: 1rem; }
    label { font-weight: bold; color: #004a8f; }
    input, button { width: 100%; padding: 0.8rem; border-radius: 6px; font-size: 1rem; border: 1px solid #ddd; }
    button { background-color: #004a8f; color: white; cursor: pointer; }
    button:hover { background-color: #003c6e; }
    footer { text-align: center; font-size: 0.9rem; color: #666; margin-top: 2rem; }
  </style>
</head>
<body>
  <header>
    <h1>Welcome to Bogoboom</h1>
    <p>Connecting newcomers in Bogotá with jobs and homes</p>
  </header>
  <main>
    <div class='intro'>
      <p>Bogoboom is your bridge to a smooth landing in Bogotá.</p>
    </div>
    <section>
      <h3>Login</h3>
      <form method='POST' action=''>
        <div>
          <label for='email'>📧 Email:</label>
          <input type='email' id='email' name='email' required>
        </div>
        <div>
          <label for='password'>🔒 Password:</label>
          <input type='password' id='password' name='password' required>
        </div>
        <div>
          <button type='submit'>🔑 Login</button>
        </div>
        <input type='hidden' name='action' value='login'>
      </form>
    </section>
  </main>
  <footer>
    &copy; 2025 Bogoboom. All rights reserved.
  </footer>
</body>
</html>";
    exit;
}

// Field labels
$field_labels = [
    "id" => "ID",
    "email" => "Email Address",
    "role" => "User Role",
    "status" => "Account Status",
    "created_at" => "Creation Date"
];

// Only display users table
$tables = [
    "users" => ["id", "email", "role", "status", "created_at"]
];

// Start HTML output
echo "<!DOCTYPE html><html lang='en'><head>
<meta charset='UTF-8' />
<meta name='viewport' content='width=device-width, initial-scale=1.0' />
<title>Bogoboom | Database Viewer</title>
<style>
  body { font-family: Arial, sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
  header { background-color: #004a8f; color: white; padding: 1rem 2rem; text-align: center; position: relative; }
  .user-info { position: absolute; right: 2rem; top: 1rem; font-size: 0.9rem; font-style: italic; }
  .logoff-button { position: absolute; left: 2rem; top: 1rem; background-color: #ff4d4d; color: white; padding: 0.5rem 1rem; border: none; border-radius: 5px; cursor: pointer; }
  main { padding: 2rem; max-width: 1200px; margin: 0 auto; }
  .intro { max-width: 600px; margin: 0 auto 2rem; font-size: 1.2rem; }
  section { margin: 2rem auto; background: #fff; padding: 1.5rem; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1); overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; margin-top: 1.5rem; background-color: #fff; }
  th, td { padding: 1rem; border: 1px solid #ddd; text-align: left; }
  th { background-color: #004a8f; color: white; }
  .no-records { font-style: italic; color: #888; }
  footer { text-align: center; font-size: 0.9rem; color: #666; margin-top: 2rem; }
  .links a { margin: 0 1rem; text-decoration: none; color: #004a8f; }
</style>
</head><body>";

echo "<header>
  <h1>Bogoboom Database Viewer</h1>
  <p>Check the records of our platform tables</p>
  <div class='user-info'>Logged in as: " . htmlspecialchars($current_user_email) . "</div>
  <form method='POST' action='' style='display: inline;'>
    <input type='hidden' name='action' value='logout'>
    <button type='submit' class='logoff-button'>Log off</button>
  </form>
</header>";

echo "<main>
  <div class='intro'>
    <p>Welcome to the database viewer. You can see user records below.</p>
  </div>";

foreach ($tables as $table => $fields) {
    $fieldList = implode(", ", $fields);
    echo "<section><h3>Records from $table</h3>";

    // If role is newcomers, show only active agencies
    if ($current_user_role === "newcomers") {
        $stmt = $conn->prepare("SELECT $fieldList FROM users WHERE role = 'agencies' AND status = 'active' LIMIT 100");
    } else {
        $stmt = $conn->prepare("SELECT $fieldList FROM users LIMIT 100");
    }

    $stmt->execute();
    $result = $stmt->get_result();

    if ($result && $result->num_rows > 0) {
        echo "<table><thead><tr>";
        $fieldsInfo = $result->fetch_fields();
        foreach ($fieldsInfo as $field) {
            $field_name = $field->name;
            // Skip Account Status, Creation Date, and Action columns for newcomers
            if ($current_user_role === "newcomers" && in_array($field_name, ["status", "created_at", "action"])) {
                continue;
            }
            echo "<th>" . htmlspecialchars($field_labels[$field_name] ?? $field_name) . "</th>";
        }
        echo "</tr></thead><tbody>";

        while ($row = $result->fetch_assoc()) {
            echo "<tr>";
            foreach ($row as $field_name => $col) {
                // Skip Account Status, Creation Date, and Action columns for newcomers
                if ($current_user_role === "newcomers" && in_array($field_name, ["status", "created_at", "action"])) {
                    continue;
                }
                echo "<td>" . (is_null($col) ? "NULL" : htmlspecialchars($col)) . "</td>";
            }
            echo "</tr>";
        }

        echo "</tbody></table>";
    } else {
        echo "<p class='no-records'>No records found in <strong>$table</strong>.</p>";
    }

    echo "</section>";
}

echo "</main>";

echo "<footer>
  &copy; 2025 Bogoboom. All rights reserved.
  <div class='links'>
    <a href='https://tecnologiasantos.com'>About Us</a>
    <a href='/terms'>Terms & Conditions</a>
    <a href='/privacy'>Privacy Policy</a>
  </div>
</footer></body></html>";
?>
