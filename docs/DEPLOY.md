# Deploy (AWS ou Cloudflare)

- **Frontend (estático)**: Cloudflare Pages ou S3 + CloudFront. Gere com `npm run build` (saída `dist/.../browser`); edite `config.json` por ambiente.
- **API Go**: precisa de contêiner/VM — AWS (ECS Fargate, App Runner ou EC2) é o caminho direto. Cloudflare Workers **não** roda este binário Go/Fiber.
- **Banco**: RDS PostgreSQL 16 (ou outro Postgres gerenciado), com backup e SSL (`DB_SSLMODE=require`).
- Defina `APP_ENV=production`, `FRONTEND_URL`, `AUTH0_*`, `TRUSTED_PROXIES` (IP do balanceador) e rode `/app/migrate` antes da API.
- Recomendação: Cloudflare na frente (DNS/WAF/cache) + API e banco na AWS.
