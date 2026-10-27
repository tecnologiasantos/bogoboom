<?php
// Kept for backward compatibility with scripts that `require` this file.
// Credentials now come from the environment (DB_HOST, DB_USER, DB_PASS, DB_NAME) — never from source code.
require_once __DIR__ . '/lib/bootstrap.php';
$conn = db();
