<?php
// Hybridauth settings come from the environment (see lib/social.php) — no secrets in source.
require_once __DIR__ . '/lib/social.php';
return hybridauth_config('company');
