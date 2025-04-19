<?php
// Assuming you're already handling the form submission with POST method
if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $fullname = $_POST['fullname'];
    $email = $_POST['email'];
    $phone = $_POST['phone'];
    $country = $_POST['country'];
    $languages = $_POST['languages'];
    $budget = $_POST['budget']; // The budget field from the form
    $document = $_FILES['document']; // If you handle file upload

    // Validate and sanitize the budget input (optional but recommended)
    if ($budget) {
        $budget = number_format((float)$budget, 2, '.', ''); // Ensure two decimal places
    } else {
        $budget = 0.00; // Default if not provided
    }

    // Example SQL query to insert into the database (adjust based on your table structure)
    $query = "INSERT INTO Newcomers (full_name, email, phone, country, languages, housing_budget) 
              VALUES ('$fullname', '$email', '$phone', '$country', '$languages', '$budget')";
    
    // Execute the query using your database connection
    if (mysqli_query($conn, $query)) {
        echo "Newcomer registered successfully!";
    } else {
        echo "Error: " . mysqli_error($conn);
    }
}
?>
