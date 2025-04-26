# **📦 Bogoboom Project Structure & Setup**

Welcome to the **Bogoboom** project! Below is the organized folder structure for our project, broken down into clear sections with a focus on scalability and ease of development. 🌱

---

## **🗂️ Project Folder Structure**

```plaintext
/var/www/html/bogoboom_project
├── backend
│   ├── auth
│   │   ├── login.php/     (empty directory)
│   │   ├── logout.php/    (empty directory)
│   │   └── register.php
│   ├── config
│   │   └── constants.php/ (empty directory)
│   ├── middleware
│   │   └── validate_input.php/ (empty directory)
│   ├── notifications
│   │   └── send_notification.php/ (empty directory)
│   ├── profiles
│   │   ├── composer.json
│   │   ├── composer.lock
│   │   ├── list_database.php
│   │   ├── login.php
│   │   ├── register.php
│   │   ├── submint_newcomers.php
│   │   ├── submit_agencies.php
│   │   ├── submit_companies.php
│   │   ├── submit_newcomers.php
│   │   ├── testconnection.php
│   │   ├── verify_user.php
│   │   └── vendor
│   │       ├── autoload.php
│   │       ├── composer
│   │       │   ├── (files like ClassLoader.php, InstalledVersions.php, etc.)
│   │       └── hybridauth
│   │           ├── hybridauth
│   │           │   ├── CHANGELOG.md
│   │           │   ├── CODE_OF_CONDUCT.md
│   │           │   ├── CONTRIBUTING.md
│   │           │   ├── COPYING.md
│   │           │   ├── ISSUE_TEMPLATE.md
│   │           │   ├── PULL_REQUEST_TEMPLATE.md
│   │           │   ├── README.md
│   │           │   ├── composer.json
│   │           │   ├── phpunit.xml
│   │           │   ├── examples
│   │           │   │   ├── example_01.php
│   │           │   │   ├── example_02.php
│   │           │   │   ├── example_03.php
│   │           │   │   ├── example_04.php
│   │           │   │   ├── example_05.php
│   │           │   │   ├── example_06/
│   │           │   │   │   ├── callback.php
│   │           │   │   │   ├── config.php
│   │           │   │   │   └── index.php
│   │           │   │   ├── example_07/
│   │           │   │   │   ├── callback.php
│   │           │   │   │   ├── config.php
│   │           │   │   │   └── index.php
│   │           │   │   └── index.html
│   │           │   ├── src
│   │           │   │   ├── Adapter/
│   │           │   │   │   ├── AbstractAdapter.php, OAuth1.php, OAuth2.php, etc.
│   │           │   │   ├── Data/
│   │           │   │   ├── Exception/
│   │           │   │   ├── HttpClient/
│   │           │   │   ├── Logger/
│   │           │   │   ├── Provider/
│   │           │   │   ├── Storage/
│   │           │   │   ├── Thirdparty/
│   │           │   │   └── User/
│   │           │   └── tests/
│   │           │       └── (test files)
│   │           └── index.html
│   └── utils
│       └── validate_input.php/ (empty directory)
├── public
│   └── (public-facing files, no detailed listing)
├── safe
│   └── (no detailed listing)
└── uploads
    └── (no detailed listing)
