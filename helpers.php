<?php
// Kept for backward compatibility; the implementation lives in lib/bootstrap.php (CSPRNG based).
require_once __DIR__ . '/lib/bootstrap.php';

function generateUUID(): string
{
    return uuid_v4();
}
