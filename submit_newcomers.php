<?php
// Enable error reporting for debugging
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Debug: Check if the file exists and provide error message if not
$connection_file = '../../safe/db_connection.php';  // Adjusted path based on your folder structure

if (!file_exists($connection_file)) {
    die("Error: db_connection.php file not found.");
}

require $connection_file; // Include database connection file

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    // Collect and sanitize form inputs
    $fullname = mysqli_real_escape_string($conn, trim($_POST['fullname']));
    $email = mysqli_real_escape_string($conn, trim($_POST['email']));
    $phone = mysqli_real_escape_string($conn, trim($_POST['phone']));
    $country = mysqli_real_escape_string($conn, trim($_POST['country']));
    $languages = mysqli_real_escape_string($conn, trim($_POST['languages']));
    $budget = isset($_POST['budget']) ? (float)$_POST['budget'] : 0.00; // Ensure it's a float and default to 0.00 if not set

    // Sanitize and validate budget to ensure it's a number
    $budget = number_format($budget, 2, '.', '');

    // Handle file upload if there is a document uploaded
    $document = null;
    if (isset($_FILES['document']) && $_FILES['document']['error'] == 0) {
        // Define file upload destination
        $uploadDir = '../../uploads/temp_files/'; // Adjusted path for file upload location
        $uploadFile = $uploadDir . basename($_FILES['document']['name']);
        
        // Validate file type (optional, adjust as needed)
        $allowedTypes = ['application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
        $fileType = mime_content_type($_FILES['document']['tmp_name']);

        if (in_array($fileType, $allowedTypes)) {
            // Move uploaded file to destination
            if (move_uploaded_file($_FILES['document']['tmp_name'], $uploadFile)) {
                $document = $uploadFile; // Store the file path in the database (optional)
            } else {
                echo "Error: Could not upload the document.";
                exit;
            }
        } else {
            echo "Error: Invalid file type. Only PDF and DOC files are allowed.";
            exit;
        }
    }

    // SQL query to insert data into the database
    $query = "INSERT INTO Newcomers (full_name, email, phone, country, languages, housing_budget, document) 
              VALUES ('$fullname', '$email', '$phone', '$country', '$languages', '$budget', '$document')";

    // Execute the query using your database connection
    if (mysqli_query($conn, $query)) {
        echo "Newcomer registered successfully!";
    } else {
        echo "Error: " . mysqli_error($conn);
    }
}
?>
