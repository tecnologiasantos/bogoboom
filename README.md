# **📦 Bogoboom Project Structure & Setup**

Welcome to the **Bogoboom** project! Below is the organized folder structure for our project, broken down into clear sections with a focus on scalability and ease of development. 🌱

---

## **🗂️ Project Folder Structure**

```plaintext
/bogoboom_project
├── /public/                         # Publicly accessible HTML files
│   ├── index.html                   # Home page for the platform
│   ├── register.html                # Newcomer registration form
│   ├── real_estate_company.html     # Real estate company details page
│   ├── register_newcomers.html      # Newcomer registration form
│   ├── company_hiring.html          # HR team hiring page
│   ├── privacy_policy.html          # Privacy policy page
│   └── /api/                         # API endpoints for form submissions
│       ├── submit_companies.php      # Handle company submissions
│       ├── submit_real_estate.php    # Handle real estate agency submissions
│       └── submit_newcomers.php      # Handle newcomer submissions
│   └── whoami.html                  # User info page
│
├── /backend/                         # Backend logic and API
│   ├── /auth/
│   │   ├── register.php              # Handles user account registration
│   │   ├── login.php                 # Login logic
│   │   └── logout.php                # Logout logic
│   │
│   ├── /profiles/
│   │   ├── testconnection.php       # Test DB connection
│   │   ├── submit_companies.php      # Final submission of HR/company
│   │   ├── submit_real_estate.php    # Final submission of real estate agency
│   │   └── submit_newcomers.php      # Final submission of newcomer profile
│   │
│   ├── /utils/
│   │   
│   │   
│   │
│   ├── /middleware/
│   │   └── validate_input.php        # Optional form validation logic
│   │
│   ├── /notifications/
│   │   └── send_notification.php     # Optional: alerts/messages/emails
│   │
│   └── /config/
│       └── constants.php             # Constants (e.g., S3 bucket, DB credentials)
│
├── /uploads/
│   └── /temp_files/                  # Used only if not fully relying on S3
│       └──                           # Example uploaded file
│
├── /safe/                            # Safe directory (for sensitive files)
│        ├── db_connection.php         # ✅ USE THIS: Central DB connection
├── /logs/
│   └── error.log                     # Logs for debugging and audits
│
├── /assets/                          # Static assets
│
├── .env                              # Environment variables for configuration
├── .gitignore                       # Ignore sensitive/system files
└── README.md                        # Project documentation (optional)
