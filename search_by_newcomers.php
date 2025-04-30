<?php
// Start the session to retrieve the user role
session_start();

// Include the database connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// Check if the user is logged in
if (!isset($_SESSION['user_email'])) {
    // Redirect to the login page if not logged in
    header('Location: https://bogoboom.com/login.html');
    exit();
}

// Get the logged-in user information from session
$user_email = $_SESSION['user_email'];

// Query the database to get the user role
$query = "SELECT role FROM users WHERE email = ?";
$stmt = $conn->prepare($query);
$stmt->bind_param('s', $user_email);
$stmt->execute();
$stmt->bind_result($user_role);
$stmt->fetch();
$stmt->close();

// Start the HTML
echo "<!DOCTYPE html>
<html lang='en'>
<head>
  <meta charset='UTF-8' />
  <meta name='viewport' content='width=device-width, initial-scale=1.0'/>
  <title>Bogoboom | Search Newcomers</title>
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

  <!-- Logo containers -->
  <div class='image-container'>
    <img src='https://tecnologiasantos-mira.s3.sa-east-1.amazonaws.com/mira/MIRA-powered.png' alt='MIRA Project Logo'>
  </div>

  <div class='image-container-right'>
    <img src='https://tecnologiasantos-mira.s3.sa-east-1.amazonaws.com/mira/logo-print-hd.jpg' alt='MIRA Project Logo Right'>
  </div>

  <header>
    <h1>Welcome to Bogoboom</h1>
    <p>Connecting newcomers in Bogotá with jobs and homes</p>
  </header>

  <main>
    <div class='intro'>
      <p>Bogoboom is your bridge to a smooth landing in Bogotá. Whether you're a newcomer, an HR department, or a real estate agency, our platform helps you connect and thrive.</p>
    </div>";

    // Check if the logged-in user is a newcomer
    if ($user_role === 'newcomer') {
        // Show agencies for newcomers
        echo "<section>
                <h3>Available Agencies</h3>
                <p>Here are the agencies that are available for newcomers:</p>";

        // Query the database to get all agencies
        $agency_query = "SELECT company_name, company_email, company_phone FROM agencies";
        $agency_result = $conn->query($agency_query);

        if ($agency_result->num_rows > 0) {
            echo "<table style='width:100%; border-collapse: collapse;'>
                    <thead>
                        <tr>
                            <th>Agency Name</th>
                            <th>Email</th>
                            <th>Phone</th>
                        </tr>
                    </thead>
                    <tbody>";

            while ($row = $agency_result->fetch_assoc()) {
                echo "<tr>
                        <td>" . htmlspecialchars($row['company_name']) . "</td>
                        <td>" . htmlspecialchars($row['company_email']) . "</td>
                        <td>" . htmlspecialchars($row['company_phone']) . "</td>
                      </tr>";
            }

            echo "</tbody></table>";
        } else {
            echo "<p>No agencies available at the moment.</p>";
        }

        echo "</section>";
    } else {
        // Show message for non-newcomer roles
        echo "<section>
                <h3>Access Denied</h3>
                <p>This page is only accessible to newcomers. Please contact support if you need further assistance.</p>
              </section>";
    }

  echo "</main>

  <footer>
    &copy; 2025 Bogoboom. All rights reserved.
  </footer>

</body>
</html>";
?>
