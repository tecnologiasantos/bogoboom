-- Schema INFERRED from the PHP code (the original dump was never committed).
-- Compare with your production database before using; add only what is missing.
-- MySQL 8 / MariaDB 10.6+, utf8mb4.
CREATE TABLE IF NOT EXISTS users (
  id CHAR(36) PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  full_name VARCHAR(255) NULL,
  phone VARCHAR(50) NULL,
  password_hash VARCHAR(255) NOT NULL,           -- bcrypt/argon2 via password_hash()
  role ENUM('newcomers','companies','agencies','admin') NOT NULL,
  status ENUM('active','inactive','blocked') NOT NULL DEFAULT 'active',
  company_name VARCHAR(255) NULL,
  company_email VARCHAR(255) NULL,
  company_phone VARCHAR(50) NULL,
  company_address VARCHAR(500) NULL,
  company_website VARCHAR(500) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) CHARACTER SET utf8mb4;

CREATE TABLE IF NOT EXISTS newcomers (
  id CHAR(36) PRIMARY KEY, user_id CHAR(36) NOT NULL,
  full_name VARCHAR(255), email VARCHAR(255), phone VARCHAR(50),
  country VARCHAR(100), languages VARCHAR(500), document VARCHAR(255) NULL,  -- stored file name only
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) CHARACTER SET utf8mb4;

CREATE TABLE IF NOT EXISTS agencies (
  id CHAR(36) PRIMARY KEY, user_id CHAR(36) NOT NULL,
  company_name VARCHAR(255), company_email VARCHAR(255), company_phone VARCHAR(50),
  responsible_name VARCHAR(255), company_address VARCHAR(500),
  country VARCHAR(100), state_or_department VARCHAR(100), city VARCHAR(100),
  company_website VARCHAR(500), company_document VARCHAR(255) NULL,
  company_password_hash VARCHAR(255), data_consent TINYINT(1) NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) CHARACTER SET utf8mb4;

CREATE TABLE IF NOT EXISTS companies (
  id CHAR(36) PRIMARY KEY, user_id CHAR(36) NOT NULL,
  company_name VARCHAR(255), company_email VARCHAR(255), company_phone VARCHAR(50),
  responsible_name VARCHAR(255), company_address VARCHAR(500), company_website VARCHAR(500),
  company_industry VARCHAR(255), country VARCHAR(100), state_or_department VARCHAR(100), city VARCHAR(100),
  company_password_hash VARCHAR(255), data_consent TINYINT(1) NOT NULL DEFAULT 0,
  allow_agency_contact TINYINT(1) NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) CHARACTER SET utf8mb4;

CREATE TABLE IF NOT EXISTS leads (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255), email VARCHAR(255), phone VARCHAR(50),
  city VARCHAR(255), country VARCHAR(100), interest VARCHAR(255),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) CHARACTER SET utf8mb4;

CREATE TABLE IF NOT EXISTS partner_inquiries (
  id INT AUTO_INCREMENT PRIMARY KEY,
  full_name VARCHAR(255), email_address VARCHAR(255), phone_number VARCHAR(50),
  country VARCHAR(100), state_or_department VARCHAR(100), city VARCHAR(100),
  collaboration_type VARCHAR(100), honeypot_website_field VARCHAR(255) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) CHARACTER SET utf8mb4;
