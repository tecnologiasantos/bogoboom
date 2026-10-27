# Bogoboom (site PHP legado) – deploy seguro

Este PR remove segredos do código, corrige injeção de SQL, XSS, CSRF, upload inseguro e endpoints de depuração.
**Nenhum segredo vive no repositório**: tudo vem de variáveis de ambiente ou de `/etc/bogoboom/env`.

## 0. ANTES de publicar (obrigatório)
Estes segredos estiveram no histórico do Git e devem ser considerados vazados. Troque-os:
1. Senha do banco de dados (MySQL/MariaDB) → crie usuário novo com permissões mínimas (SELECT/INSERT/UPDATE nas tabelas do app).
2. Senha de app do Gmail (SMTP) → revogue em myaccount.google.com/apppasswords e crie outra (ou use AWS SES).
3. Chave da API Gemini → apague e crie outra no Google AI Studio, com restrição de uso.
4. O token do GitHub colado no chat → revogue em GitHub › Settings › Developer settings.
5. Opcional: reescrever o histórico (`git filter-repo`) – só ajuda se o repositório nunca foi clonado por terceiros; a rotação é o que realmente protege.

## 1. Arquivo de ambiente `/etc/bogoboom/env` (chmod 640, dono root:www-data)
```
APP_ENV=production
APP_BASE_URL=https://bogoboom.com
APP_ROOT=/var/www/html/bogoboom_project
ALLOWED_HOSTS=bogoboom.com,www.bogoboom.com
UPLOADS_DIR=/var/lib/bogoboom/uploads      # fora do web root; chmod 750 www-data
RATE_LIMIT_DIR=/var/lib/bogoboom/ratelimit
TRUST_PROXY=0        # 1 somente atrás de Cloudflare/ALB (usa CF-Connecting-IP / X-Forwarded-For)
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=bogoboom_app
DB_PASS=...
DB_NAME=bogoboom
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=...
SMTP_PASS=...
SMTP_FROM=no-reply@bogoboom.com
NOTIFY_TO=contato@bogoboom.com
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-pro
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
FACEBOOK_CLIENT_ID=
FACEBOOK_CLIENT_SECRET=
LINKEDIN_CLIENT_ID=
LINKEDIN_CLIENT_SECRET=
```
O login social só entra em contas **já existentes** com e-mail verificado pelo provedor; nunca cria contas.

## 2. Passos
```
composer update --no-dev --optimize-autoloader   # o composer.lock precisa ser regenerado (phpmailer foi adicionado)
mariadb bogoboom < docs/legacy-schema.sql         # esquema INFERIDO do código; compare com o banco real antes
# copie "nginx config file" para /etc/nginx/sites-available/bogoboom e ative; certbot para o TLS
```
Só os endpoints listados no nginx são executáveis; `lib/`, `vendor/`, `*.sql`, `composer.*` retornam 403/404.

## 3. AWS ou Cloudflare?
- **Este site é PHP + MySQL**: o Cloudflare *não executa PHP/MySQL*. Opções:
  - **AWS (recomendado para este código)**: Lightsail ou EC2 (Ubuntu + nginx + PHP-FPM) + RDS MySQL (ou MariaDB na própria instância no começo), Security Group só 80/443, SES para e-mail, Secrets Manager/SSM Parameter Store para as variáveis acima.
  - **Cloudflare na frente da AWS**: DNS + proxy + WAF + rate limiting; defina `TRUST_PROXY=1`, modo SSL "Full (strict)" e libere no Security Group apenas os IPs da Cloudflare. Nada muda no código.
  - **Cloudflare puro**: só serve o novo frontend (`v2/web`, Next.js em Pages/Workers) com a API v2 em outro lugar (PR #1).
- Em qualquer caso: HTTPS obrigatório, backups automáticos do banco e `fail2ban` no SSH.

## 4. O que mudou (resumo)
| Problema | Correção |
|---|---|
| Credenciais do banco/SMTP/Gemini no código | `lib/env.php` (somente ambiente) |
| `list_database.php`, `testconnection.php`, `verify_user.php`, `login_report.php`, `config*.php` expostos | removidos |
| SQL injection | consultas preparadas em tudo |
| XSS | saída sempre via `h()`; cabeçalhos de segurança |
| CSRF | token de sessão + checagem de Origin (`csrf.php` + `csrf.js` para páginas estáticas) |
| Escalada de privilégio no cadastro (`role=admin`) | lista branca de papéis |
| Alterar status de usuários sem login (GET) | POST, só admin, CSRF |
| Senhas/login | mínimo 10 caracteres, hash seguro, tempo constante, limite 5/15 min por conta e 20/15 min por IP, sessão regenerada |
| Upload | PDF/DOC/DOCX por MIME real (finfo), 5 MB, nome aleatório, 0640, fora do web root |
| Chatbot (custo/abuso) | limites 8/min e 150/dia por IP, origem validada |
| Dados pessoais no Google Tag Manager (parcerias) | removidos do `dataLayer` |
| jQuery do CDN sem SRI | SRI adicionado |

## 5. Pendências conhecidas (decisão de produto/privacidade)
- A lista do `login.php` mostra e-mail/telefone de recém-chegados a agências. Considere o modelo da v2 (leads borrados + desbloqueio auditado).
- `allow_agency_contact` das empresas ainda não filtra nada na listagem.
- Login social só vincula contas existentes; criar contas via OAuth exige decisão de produto.
- Não há 2FA nem recuperação de senha por e-mail.
- Testado com PHP 8.3 + MariaDB (cadastro, login, CSRF, limite de tentativas, upload malicioso, papel admin). Login social e envio de e-mail **não** foram testados (sem rede/credenciais).
