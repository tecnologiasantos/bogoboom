<?php

return [
    'callback' => 'http://15.229.154.73/backend/profiles/social_callback_newcomer.php',
    'providers' => [
        'Google' => [
            'enabled' => true,
            'keys' => [
                'id' => 'GOOGLE_CLIENT_ID',
                'secret' => 'GOOGLE_CLIENT_SECRET'
            ],
            'scope' => 'email profile'
        ],
        'Facebook' => [
            'enabled' => true,
            'keys' => [
                'id' => 'FACEBOOK_APP_ID',
                'secret' => 'FACEBOOK_APP_SECRET'
            ],
            'scope' => 'email, public_profile'
        ],
        'LinkedIn' => [
            'enabled' => true,
            'keys' => [
                'id' => 'LINKEDIN_CLIENT_ID',
                'secret' => 'LINKEDIN_CLIENT_SECRET'
            ],
            'scope' => 'r_liteprofile r_emailaddress'
        ]
    ]
];
