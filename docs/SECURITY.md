# Segurança

Implementado: JWT RS256 via JWKS; usuário por `sub`; guardas de papel; validação/sanitização de todas as entradas; SQL parametrizado; limite de taxa por IP real (proxy confiável via `TRUSTED_PROXIES`); corpo máx. 64 KB; CORS só para `FRONTEND_URL`; cabeçalhos de segurança (CSP/HSTS/nosniff/frame); honeypot; auditoria de desbloqueios e ações admin; contêiner Go distroless não-root; sem segredos no código; CodeQL, Dependabot e govulncheck.

## Antes de publicar (ações do dono)
1. **Revogar** o token do GitHub que foi colado no início do projeto.
2. **Trocar** a senha do banco, a chave Gemini e a senha de app do Gmail que existiam no histórico antigo do Git (considere-os vazados).
3. Ativar no GitHub: proteção da branch `main`, secret scanning, push protection, Dependabot alerts, 2FA obrigatório.
4. Auth0: ver `AUTH0-SETUP.md` (e-mail verificado, rotação de refresh token).
5. Revisão jurídica da política de privacidade e termos.
6. Habilitar HTTPS/HSTS no provedor e backups do Postgres.
