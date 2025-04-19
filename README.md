```markdown
# 🌍 Bogoboom

**Connecting newcomers in Colombia with homes and jobs.**  
Bogoboom is a digital platform that bridges newcomers, HR departments, and real estate agencies across Colombia, streamlining housing and employment integration.

---

## 📁 Project Structure

```
/bogoboom_project
├── /public/                        # Publicly accessible HTML files
│   ├── index.html
│   ├── register.html
│   ├── login.html
│   ├── newcomer_profile.html
│   ├── hr_profile.html
│   └── agency_profile.html
│
├── /backend/                      # All backend logic (PHP)
│   ├── /auth/
│   │   ├── register.php
│   │   ├── login.php
│   │   └── logout.php
│   │
│   ├── /profiles/
│   │   ├── save_newcomer.php
│   │   ├── save_hr.php
│   │   └── save_agency.php
│   │
│   ├── /utils/
│   │   ├── db_connection.php
│   │   ├── s3_upload.php
│   │   └── helpers.php             # (Optional: reusable functions)
│   │
│   ├── /middleware/               # (Optional: form validation, auth checks)
│   │   └── validate_input.php
│   │
│   ├── /notifications/            # (Optional: alert logic)
│   │   └── send_notification.php
│   │
│   └── /config/
│       └── constants.php
│
├── /uploads/                      # (Optional: temp uploads if not using S3)
│   └── temp_files/
│
├── /logs/                         # (Optional: for audits, debugging, etc.)
│   └── error.log
│
├── /assets/                       # Images, JS, CSS
│   ├── /css/
│   │   └── styles.css
│   ├── /js/
│   │   └── main.js
│   └── /images/
│       └── logo.png
│
├── .env                           # (Optional: environment variables)
├── composer.json                  # AWS SDK dependencies, etc.
├── .gitignore                     # To ignore sensitive files and folders
└── README.md                      # Project setup instructions
```

---

## 🚀 Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/bogoboom.git
cd bogoboom_project
```

### 2. Backend Requirements

- PHP >= 7.4
- Composer
- MariaDB or MySQL
- AWS credentials (IAM Role or Access Key)

### 3. Install Dependencies

```bash
composer install
```

This installs the **AWS SDK for PHP** and other PHP packages.

### 4. Configure Environment

Create a `.env` file in the root directory:

```env
DB_HOST=localhost
DB_NAME=bogoboom_db
DB_USER=bogoboom_user
DB_PASS=your_password

AWS_BUCKET=tecnologiasantos-mira
AWS_REGION=us-east-1
```

Alternatively, you can define these in `/backend/config/constants.php`.

### 5. Set Up the Database

Import your schema:

```bash
mysql -u bogoboom_user -p bogoboom_db < schema.sql
```

### 6. Configure Your Web Server

- Set `/public/` as the web root.
- Ensure `uploads/` is writable (if used).
- Enable PHP, mod_rewrite, and file permissions as needed.

---

## 🌐 Main Pages

| File                            | Description                    |
|----------------------------------|--------------------------------|
| `/public/index.html`            | Landing page                   |
| `/public/register.html`         | User registration              |
| `/public/login.html`            | Login form                     |
| `/public/newcomer_profile.html` | Newcomer dashboard             |
| `/public/hr_profile.html`       | HR dashboard                   |
| `/public/agency_profile.html`   | Real estate agency dashboard   |

---

## ✅ Features

- Secure newcomer onboarding
- Role-specific profile creation
- AWS S3 document uploads
- Clean folder structure for scalability
- Optional modules for notifications and validation
- Modern PHP best practices

---

## 🔐 Security Recommendations

- Never commit your `.env` file, `uploads/`, or log files.
- Sanitize and validate all user input.
- Protect endpoints using session or token-based auth.

---

## 🛠️ Development Workflow

1. Create a new branch: `git checkout -b feature/your-feature-name`
2. Make changes and commit: `git commit -m "Add your message"`
3. Push to GitHub: `git push origin feature/your-feature-name`
4. Open a Pull Request

---



MIRA job portal and Bogoboom are assets from Tecnologia Santos company.
