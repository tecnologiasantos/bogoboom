<?php



// Load config
require 'config.php';

header('Content-Type: application/json');

// Get the user's message
$data = json_decode(file_get_contents('php://input'), true);
$userMessage = $data['message'] ?? '';

if (!$userMessage) {
    echo json_encode(['reply' => 'Please enter a message.']);
    exit;
}

// Prepare Gemini API request
$apiKey = GEMINI_API_KEY;
$endpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key={$apiKey}";

$payload = json_encode([
    'contents' => [
        ['parts' => [['text' => $userMessage]]]
    ]
]);

// Send the request to Gemini
$ch = curl_init($endpoint);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, $payload);

$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

if (curl_errno($ch)) {
    echo json_encode(['reply' => 'Error talking to Gemini API.']);
    exit;
}
curl_close($ch);

// Decode response
$result = json_decode($response, true);

// Handle errors
if ($httpCode !== 200 || !isset($result['candidates'][0]['content']['parts'][0]['text'])) {
    echo json_encode([
        'reply' => 'Gemini API error.',
        'http_code' => $httpCode,
        'raw_response' => $response
    ]);
    exit;
}

// Extract reply
$reply = $result['candidates'][0]['content']['parts'][0]['text'];
echo json_encode(['reply' => $reply]);
?>

