# **📦 Bogoboom Project Structure & Setup**

Welcome to the **Bogoboom** project! Below is the organized folder structure for our project, broken down into clear sections with a focus on scalability and ease of development. 🌱

---

## **🗂️ Project Folder Structure**

```plaintext
/bogoboom_project
├── /public/                         # Publicly accessible HTML files
│   ├── index.html                   # Home page for the platform
│   ├── register.html                # Newcomer registration form
│   ├── login.html                   # Login page for users
│   ├── newcomer_profile.html        # Profile page for newcomers
│   ├── hr_profile.html              # Profile page for HR teams
│   └── agency_profile.html          # Profile page for real estate agencies
│
├── /backend/                         # Backend logic and API
│   ├── /auth/
│   │   ├── register.php              # Handles user account registration
│   │   ├── login.php                 # Login logic
│   │   └── logout.php                # Logout logic
│   │
│   ├── /profiles/
│   │   ├── save_newcomer.php         # Save draft or update newcomer profile
│   │   ├── save_hr.php               # Save HR profile
│   │   ├── save_agency.php           # Save agency profile
│   │   ├── submit_newcomers.php      # Final submission of newcomer profile
│   │   ├── submit_companies.php      # Final submission of HR/company
│   │   └── submit_real_estate.php    # Final submission of agency
│   │
│   ├── /utils/
│   │   ├── db_connection.php         # ✅ USE THIS: Central DB connection
│   │   ├── connection.php            # 🔁 DEPRECATED: Likely safe to remove if unused
│   │   ├── s3_upload.php             # Upload files to AWS S3
│   │   └── helpers.php               # Reusable helper functions
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
│       └── leidy.docx                # Example uploaded file
│
├── /logs/
│   └── error.log                     # Logs for debugging and audits
│
├── /assets/
│   ├── /css/
│   │   └── styles.css                # Project styling
│   ├── /js/
│   │   └── main.js                   # Client-side logic
│   └── /images/
│       └── logo.png                  # Site branding
│
├── /database/
│   └── bogoboom_mvp_schema.sql      # SQL schema of your DB
│
├── composer.json                    # AWS SDK and other PHP packages
├── .gitignore                       # Ignore sensitive/system files
└── README.md                        # Project documentation (optional)
```


