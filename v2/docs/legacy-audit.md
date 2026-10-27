# Audit of the legacy PHP site (repo root, snapshot a3d95c5)
Scope: 33 PHP files + HTML, static review only (no running instance tested). Secret values are intentionally **not** reproduced here.

| # | Severity | Finding | Fix |
|---|---|---|---|
| 1 | **Critical** | A database password is hard-coded in `testconnection.php` and `db_connection.php` and reused as the HTTP-Basic password of `list_database.php` (a DB viewer). It lives in git history, so deleting the files is not enough. | **Rotate** the DB user password and the viewer password now; move secrets to environment/Secrets Manager; purge history (git filter-repo/BFG) if the repo is ever shared. |
| 1b | **Critical** | A Gmail SMTP app password is hard-coded in `partner_inquiry.php` (PHPMailer). | Revoke it in the Google account, create a new one, load from environment. |
| 2 | High | `list_database.php` and `testconnection.php` are diagnostic endpoints in the web root. | Remove from the server; block in nginx. |
| 3 | High | `display_errors=1` / `E_ALL` in login, register, partner_inquiry, default_profile, list_database: stack traces and SQL errors reach visitors. | Disable in production, log to file. |
| 4 | Medium | No CSRF tokens in any form handler. | Add tokens or SameSite cookies + origin checks. |
| 5 | Medium | Only 11 of 33 PHP files use prepared statements; the rest need manual review for SQL injection (grep found no obvious concatenation, which is not proof). | Review every query; prefer one data-access layer. |
| 6 | Low | Passwords use `password_hash` / `password_verify` (good). OAuth ids/secrets in `hybridauth_*.php` and `config_login.php` are placeholders (good) — keep real ones out of git. | Keep; migrate to Cognito in v2. |
| 7 | Info | nginx denies dotfiles and non-PHP files under `backend/profiles` (good). | Also deny `*.sql`, `composer.*`, diagnostic scripts. |

Migration plan: v2 uses Cognito/OIDC (no passwords stored) and a new schema; legacy users are imported without password hashes and sign in via reset/magic link.
