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
            $_SESSION['user_role'] = $user['role'];
            $current_user_email = $user['email'];
            $current_user_role = $user['role'];
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
    $current_user_role = $_SESSION['user_role'];
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

  <script>
    window.onbeforeunload = function() {
      // Trigger logoff when the user closes the page
      var form = document.createElement('form');
      form.method = 'POST';
      form.action = '';
      var actionInput = document.createElement('input');
      actionInput.type = 'hidden';
      actionInput.name = 'action';
      actionInput.value = 'logout';
      form.appendChild(actionInput);
      document.body.appendChild(form);
      form.submit();
    };
  </script>
</body>
</html>";
    exit;
}

// Field labels (removed id, status, created_at)
$field_labels = [
    "email" => "Email Address",
    "role" => "User Role",
    "full_name" => "Client's name",
    "company_name" => "Company name",
    "phone" => "Phone number",
    "company_website" => "Web site"  
];

// Define the fields based on user role
if ($current_user_role === 'newcomers') {
    $tables = [
        "users" => ["company_name", "email", "phone", "company_website"]
    ];
    $stmt = $conn->prepare("SELECT " . implode(", ", $tables['users']) . " FROM users WHERE role = 'agencies' AND status = 'active' LIMIT 100");
} elseif ($current_user_role === 'agencies') {
    $tables = [
        "users" => ["full_name", "email", "phone"]
    ];

    if (isset($_POST['generate_report']) && isset($_POST['start_date']) && isset($_POST['end_date'])) {
        $start_date = $_POST['start_date'] . ' 00:00:00';
        $end_date = $_POST['end_date'] . ' 23:59:59';

        $stmt = $conn->prepare(
            "SELECT " . implode(", ", $tables['users']) . "
             FROM users
             WHERE role = 'newcomers' AND status = 'active' AND created_at BETWEEN ? AND ?
             LIMIT 100"
        );
        $stmt->bind_param("ss", $start_date, $end_date);
    } else {
        $stmt = $conn->prepare("SELECT " . implode(", ", $tables['users']) . " FROM users WHERE role = 'newcomers' AND status = 'active' LIMIT 100");
    }
} elseif ($current_user_role === 'companies') {
    $tables = [
        "users" => ["company_name", "email", "phone", "company_website"]
    ];
    $stmt = $conn->prepare("SELECT " . implode(", ", $tables['users']) . " FROM users WHERE role = 'agencies' AND status = 'active' LIMIT 100");
} else {
    $tables = [
        "users" => ["full_name", "email", "phone"]
    ];
    $stmt = $conn->prepare("SELECT " . implode(", ", $tables['users']) . " FROM users LIMIT 100");
}

// Execute the query
$stmt->execute();
$result = $stmt->get_result();

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

<div class='header-actions' style='display: flex; justify-content: space-between; align-items: center; margin-top: 1rem;'>
  <div class='user-info' style='font-size: 0.9rem; font-style: italic;'>Logged in as: " . htmlspecialchars($current_user_email) . "</div>
  <div style='display: flex; gap: 1rem;'>
    <form method='POST' action=''>
      <input type='hidden' name='action' value='logout'>
      <button type='submit' style='background-color: #ff4d4d; color: white; padding: 0.5rem 1rem; border: none; border-radius: 5px; cursor: pointer;'>Log off</button>
    </form>
    <form method='GET' action='change_password.php'>
      <button type='submit' style='background-color: #ffa500; color: white; padding: 0.5rem 1rem; border: none; border-radius: 5px; cursor: pointer;'>Change Password</button>
    </form>
  </div>
</div>

  
</header>";

echo "<main>
  <div class='intro'>
    <p>Welcome to the database viewer. You can see user records below.</p>
  </div>";

if ($current_user_role === 'agencies') {
    echo "<section>
        <h3>Generate Report for Newcomers</h3>
        <form method='POST' action=''>
            <label for='start_date'>Start Date:</label>
            <input type='date' id='start_date' name='start_date' required>
            <label for='end_date'>End Date:</label>
            <input type='date' id='end_date' name='end_date' required>
            <input type='hidden' name='generate_report' value='1'>
            <button type='submit'>📄 Generate Report</button>
        </form>
    </section>";
}

if (isset($_POST['generate_report'])) {
    $startFormatted = date("d/m/Y", strtotime($_POST['start_date']));
    $endFormatted = date("d/m/Y", strtotime($_POST['end_date']));
    echo "<section><strong>Showing registration date from " . $startFormatted . " to " . $endFormatted . "</strong></section>";
}

if ($result && $result->num_rows > 0) {
    while ($row = $result->fetch_assoc()) {
        echo "<section><table><thead><tr>";
        foreach ($row as $key => $value) {
            $label = isset($field_labels[$key]) ? $field_labels[$key] : $key;
            echo "<th>" . htmlspecialchars($label) . "</th>";
        }
        echo "</tr></thead><tbody><tr>";
        foreach ($row as $value) {
            echo "<td>" . htmlspecialchars($value ?? '') . "</td>";
        }
        echo "</tr></tbody></table></section>";
    }
} else {
    echo "<p class='no-records'>No records found.</p>";
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
