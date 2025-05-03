<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// Start the session for logout functionality
session_start();

// Prompt for login
$authenticated = false;
$current_user_email = "";

// Check if the email and password are passed via a POST request
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['action']) && $_POST['action'] == 'login') {
    $email = $_POST['email'];
    $password = $_POST['password'];

    // Fetch user from DB
    $stmt = $conn->prepare("SELECT * FROM users WHERE email = ? LIMIT 1");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $result = $stmt->get_result();

    if ($result && $result->num_rows === 1) {
        $user = $result->fetch_assoc();
        // Check if the user is active and password is correct
        if ($user['status'] === 'active' && password_verify($password, $user['password_hash'])) {
            $authenticated = true;
            $_SESSION['user_email'] = $user['email']; // Store user email in session
            $current_user_email = $user['email'];
        }
    }

    if (!$authenticated) {
        echo "Access denied. Incorrect credentials or inactive user.";
        exit;
    }
} elseif (isset($_POST['action']) && $_POST['action'] == 'logout') {
    // Logout the user
    session_destroy();
    header("Location: ".$_SERVER['PHP_SELF']);
    exit;
} elseif (isset($_SESSION['user_email'])) {
    // User is logged in, set the current user email
    $current_user_email = $_SESSION['user_email'];
} else {
    // If the form is not submitted, ask for login details
    echo "<!DOCTYPE html>
<html lang='en'>
<head>
  <meta charset='UTF-8' />
  <meta name='viewport' content='width=device-width, initial-scale=1.0'/>
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
    input[type='text'], input[type='email'], input[type='password'] {
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
  </style>
</head>
<body>
  

  <header>
    <h1>Welcome to Bogoboom</h1>
    <p>Connecting newcomers in Bogotá with jobs and homes</p>
  </header>

  <main>
    <div class='intro'>
      <p>Bogoboom is your bridge to a smooth landing in Bogotá. Whether you're a newcomer, an HR department, or a real estate agency, our platform helps you connect and thrive.</p>
    </div>

    <!-- Login Section -->
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
        <div style='text-align: center;'>
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

// Friendly field labels
$field_labels = [
    "id" => "ID",
    "email" => "Email Address",
    "role" => "User Role",
    "status" => "Account Status",
    "created_at" => "Creation Date",
    "name" => "Name",
    "company_email" => "Company Email",
    "company_phone" => "Company Phone Number",
    "phone" => "Phone Number",
    "country" => "Country"
];

// Tables and fields to display
$tables = [
    "users" => ["id", "email", "role", "status", "created_at"],
    "companies" => ["id", "company_name AS name", "company_email AS email", "company_phone", "created_at"],
    "agencies" => ["id", "company_name AS name", "company_email AS email", "company_phone", "created_at"],
    "newcomers" => ["id", "full_name AS name", "email", "phone", "country", "created_at"]
];

// Start HTML
echo "<!DOCTYPE html><html lang='en'><head>";
echo "<meta charset='UTF-8' />";
echo "<meta name='viewport' content='width=device-width, initial-scale=1.0' />";
echo "<title>Bogoboom | Database Viewer</title>";
echo "<style>
    body {
      font-family: Arial, sans-serif;
      margin: 0;
      padding: 0;
      background: #f5f5f5;
    }
    header {
      background-color: #004a8f;
      color: white;
      padding: 1rem 2rem;
      text-align: center;
      position: relative;
    }
    .user-info {
      position: absolute;
      right: 2rem;
      top: 1rem;
      font-size: 0.9rem;
      font-style: italic;
    }
    .logoff-button {
      position: absolute;
      left: 2rem;
      top: 1rem;
      background-color: #ff4d4d;
      color: white;
      padding: 0.5rem 1rem;
      border: none;
      border-radius: 5px;
      cursor: pointer;
    }
    main {
      padding: 2rem;
      text-align: center;
      max-width: 1200px;
      margin: 0 auto;
    }
    .intro {
      max-width: 600px;
      margin: 0 auto 2rem;
      font-size: 1.2rem;
    }
    section {
      width: 100%;
      margin: 2rem auto;
      text-align: left;
      background: #fff;
      padding: 1.5rem;
      border-radius: 8px;
      box-shadow: 0 4px 8px rgba(0,0,0,0.1);
      overflow-x: auto;
    }
    section h3 {
      color: #004a8f;
      margin-bottom: 1rem;
    }
    .no-records {
      font-style: italic;
      color: #888;
      margin-bottom: 20px;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      margin-top: 1.5rem;
      background-color: #fff;
      table-layout: fixed;
      overflow-wrap: break-word;
    }
    table th, table td {
      padding: 1rem;
      border: 1px solid #ddd;
      text-align: left;
      word-wrap: break-word;
    }
    table th {
      background-color: #004a8f;
      color: white;
    }
    footer {
      margin-top: 2rem;
      text-align: center;
      font-size: 0.9rem;
      color: #666;
    }
    footer .links {
      margin-top: 0.5rem;
    }
    footer .links a {
      margin: 0 1rem;
      text-decoration: none;
      color: #004a8f;
    }
</style></head><body>";

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
        <p>Welcome to the database viewer for Bogoboom. Below are the records from selected tables.</p>
    </div>";

// Loop through tables
foreach ($tables as $table => $fields) {
    $fieldList = implode(", ", $fields);

    echo "<section>
            <h3>Records from $table</h3>";

    $check_table_query = "SHOW TABLES LIKE '$table'";
    $check_result = $conn->query($check_table_query);

    if ($check_result && $check_result->num_rows > 0) {
        $query = "SELECT $fieldList FROM $table LIMIT 100";
        $result = $conn->query($query);

        if ($result && $result->num_rows > 0) {
            echo "<table><thead><tr>";

            // Get column names and apply friendly labels
            $fieldsInfo = $result->fetch_fields();
            foreach ($fieldsInfo as $field) {
                $field_name = $field->name;
                echo "<th>" . htmlspecialchars($field_labels[$field_name] ?? $field_name) . "</th>";
            }

            // Add an extra column for "Action" buttons only for users table
            if ($table === 'users') {
                echo "<th>Action</th>";
            }
            echo "</tr></thead><tbody>";

            while ($row = $result->fetch_assoc()) {
                echo "<tr>";
                foreach ($row as $col) {
                    $display = is_null($col) ? "NULL" : htmlspecialchars($col);
                    echo "<td>$display</td>";
                }

                // Add toggle action
                if ($table === 'users') {
                    $user_email = $row['email'];
                    $current_status = strtolower($row['status']);

                    if ($current_status == 'active') {
                        echo "<td><a href='toggle_user_status.php?email=$user_email&action=deactivate' style='color: red;'>Deactivate</a></td>";
                    } else {
                        echo "<td><a href='toggle_user_status.php?email=$user_email&action=activate' style='color: green;'>Activate</a></td>";
                    }
                }

                echo "</tr>";
            }

            echo "</tbody></table>";
        } else {
            echo "<p class='no-records'>No records found in <strong>$table</strong>.</p>";
        }
    } else {
        echo "<p class='no-records'>The table <strong>$table</strong> does not exist.</p>";
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
</footer>";

echo "</body></html>";
?>
