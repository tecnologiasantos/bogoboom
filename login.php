<?php
session_start();
require_once '../db_connection.php';

// Function to sanitize user inputs
function cleanInput($data) {
    return htmlspecialchars(stripslashes(trim($data)));
}

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    // Clean input data
    $email = cleanInput($_POST["email"]);
    $password = cleanInput($_POST["password"]);
    
    // Prepare the SQL query to check if the user exists
    $stmt = $conn->prepare("SELECT id, email, password, role FROM users WHERE email = ?");
    $stmt->bind_param("s", $email);
    $stmt->execute();
    $result = $stmt->get_result();

    // Check if the user exists and validate the password
    if ($result->num_rows > 0) {
        $user = $result->fetch_assoc();
        
        // Verify the password
        if (password_verify($password, $user['password'])) {
            // Start the session
            $_SESSION['user_id'] = $user['id'];
            $_SESSION['user_email'] = $user['email'];
            $_SESSION['user_role'] = $user['role']; // Store the user role (newcomer or company)
            
            // Redirect based on user role
            if ($user['role'] == 'newcomer') {
                header("Location: /backend/profiles/newcomer_profile.php"); // For newcomers
            } elseif ($user['role'] == 'company') {
                header("Location: /backend/profiles/company_profile.php"); // For companies
            } else {
                header("Location: /backend/profiles/default_profile.php"); // Default profile page
            }
            exit();
        } else {
            echo "Incorrect password.";
        }
    } else {
        echo "No user found with that email address.";
    }
    
    $stmt->close();
}
?>

