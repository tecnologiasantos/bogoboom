<?php
require '../config.php';

header('Content-Type: application/json');

$data = json_decode(file_get_contents('php://input'), true);
$userMessage = $data['message'] ?? '';

if (!$userMessage) {
    echo json_encode(['reply' => 'Please enter a message.']);
    exit;
}

// Make API request to OpenAI
$ch = curl_init('https://api.openai.com/v1/chat/completions');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
    'Authorization: Bearer ' . OPENAI_API_KEY
]);
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
    "model" => "gpt-3.5-turbo",
    "messages" => [
        ["role" => "system", "content" => "You are the Bogoboom assistant. Help newcomers, HR, and agencies with their questions about relocating to Bogotá."],
        ["role" => "user", "content" => $userMessage]
    ]
]));

$response = curl_exec($ch);
if (curl_errno($ch)) {
    echo json_encode(['reply' => 'Error talking to AI.']);
    exit;
}
curl_close($ch);

$result = json_decode($response, true);
$reply = $result['choices'][0]['message']['content'] ?? 'Sorry, I could not generate a reply.';

echo json_encode(['reply' => $reply]);
?>
