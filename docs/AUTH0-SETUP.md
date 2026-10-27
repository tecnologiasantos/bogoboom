# Configuração do Auth0

1. **API**: Applications → APIs → Create. Identifier = `AUTH0_AUDIENCE` (ex.: `https://api.bogoboom.local`), algoritmo RS256.
2. **Aplicação SPA**: Allowed Callback/Logout/Web Origins = URL do frontend. Habilite *Refresh Token Rotation*.
3. **Action (Login/Post-Login)** para incluir o e-mail no access token (claim `email`) e `email_verified`.
4. Exija e-mail verificado.
5. Preencha `frontend/bogoboom-frontend/public/config.json` (`domain`, `clientId`, `audience`, `apiUrl`) e as variáveis `AUTH0_DOMAIN`/`AUTH0_AUDIENCE` do backend.
6. Admin: defina o papel `admin` diretamente no banco (`UPDATE users SET role='admin' WHERE ...`) — não há rota pública para isso.
