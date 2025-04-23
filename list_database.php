<?php
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// DB connection
require '/var/www/html/bogoboom_project/safe/db_connection.php';

// List of tables to fetch data from
$tables = [
    "Newcomers", "Agencies", "Companies"
];

echo "<!DOCTYPE html><html lang='en'><head>";
echo "<meta charset='UTF-8' />";
echo "<meta name='viewport' content='width=device-width, initial-scale=1.0'/>";
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
      overflow-x: auto; /* Allows horizontal scrolling if content overflows */
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
      table-layout: fixed; /* Prevents wide columns */
      overflow-wrap: break-word; /* Ensures long words break into the next line */
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
</header>";

echo "<main>
    <div class='intro'>
        <p>Welcome to the database viewer for Bogoboom. Below are the records from different tables in our system.</p>
    </div>";

foreach ($tables as $table) {
    echo "<section>
            <h3>Records from $table</h3>";

    // Check if the table exists before querying
    $check_table_query = "SHOW TABLES LIKE '$table'";
    $check_result = $conn->query($check_table_query);

    if ($check_result && $check_result->num_rows > 0) {
        // Table exists, proceed with the query
        $query = "SELECT * FROM $table";
        $result = $conn->query($query);

        if ($result && $result->num_rows > 0) {
            echo "<table><thead><tr>";
            while ($field = $result->fetch_field()) {
                echo "<th>" . htmlspecialchars($field->name) . "</th>";
            }
            echo "</tr></thead><tbody>";

            while ($row = $result->fetch_assoc()) {
                echo "<tr>";
                foreach ($row as $col) {
                    $display = is_null($col) ? "NULL" : htmlspecialchars($col);
                    echo "<td>$display</td>";
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
        <a href='https://tecnologiasantos.com' target='_blank'>Visit MIRA IT job portal - a Tecnologia Santos IT project</a>
    </div>
</footer>";

$conn->close();

echo "</body></html>";
?>
