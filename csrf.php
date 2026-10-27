<?php
// Hands the CSRF token to static HTML pages (see csrf.js). Same-origin only.
require __DIR__ . '/lib/bootstrap.php';
if (!origin_ok()) {
    fail(403, 'Forbidden.');
}
json_out(['token' => csrf_token()]);
