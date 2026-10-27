<?php
require __DIR__ . '/lib/bootstrap.php';
require __DIR__ . '/lib/layout.php';
$u = require_login();

page_start('Default Profile | Bogoboom', 'Welcome to Your Profile');
echo '<section><h3>' . h($u['email']) . '</h3>'
    . '<p>Your account does not have a specific role profile yet. Please complete your registration.</p></section>';
page_end();
