<?php
require '../safe/db_connection.php'; // Adjust the path if needed

if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $fullname = $_POST['fullname'];
    $email = $_POST['email'];
    $phone = $_POST['phone'];
    $country = $_POST['country'];
    $languages = $_POST['languages'];
    $budget = isset($_POST['budget']) ? number_format((float)$_POST['budget'], 2, '.', '') : 0.00;

    // Handle file upload (optional)
    $documentPath = null;
    if (isset($_FILES['document']) && $_FILES['document']['error'] == 0) {
        $tmp = $_FILES['document']['tmp_name'];
        $name = basename($_FILES['document']['name']);
        $destination = '../../uploads/temp_files/' . $name;
        move_uploaded_file($tmp, $destination);
        $documentPath = $destination;
    }

    $sql = "INSERT INTO Newcomers (full_name, email, phone, country, languages, housing_budget)
            VALUES (?, ?, ?, ?, ?, ?)";

    $stmt = $conn->prepare($sql);
    $stmt->bind_param("sssssd", $fullname, $email, $phone, $country, $languages, $budget);

    if ($stmt->execute()) {
        echo "✅ Newcomer registered successfully!";
    } else {
        echo "❌ Error: " . $stmt->error;
    }
}
?>
