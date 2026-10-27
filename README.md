# Bogoboom

Plataforma que conecta **recém-chegados a Bogotá** (expatriados em transferência de trabalho), **RH das empresas** e **imobiliárias**: o recém-chegado informa o que precisa, a imobiliária vê leads com *match score* e desbloqueia o contato (com consentimento) para oferecer imóveis.

A identidade visual vem da landing original (`index.html`): Arial, fundo `#f5f5f5`, azul `#004a8f`, cartões brancos com raio 8px.

## Stack (mesma do projeto miratech)

| Camada | Tecnologia |
|---|---|
| API | Go 1.25 + Fiber v3 + pgx v5, arquitetura limpa (`internal/{core,adapters,config}`) |
| Frontend | Angular 21 (standalone, Signals, SCSS) |
| Auth | Auth0 (JWT RS256 via JWKS) |
| Banco | PostgreSQL 16 |
| Infra | Docker Compose (local); AWS ou Cloudflare a decidir — ver `docs/DEPLOY.md` |

## Começar (desenvolvedor local)

```bash
cp .env.example .env            # preencha; não há segredos padrão
# Auth0: ver docs/AUTH0-SETUP.md (crie API + SPA e ajuste frontend/.../public/config.json)
cd devops-compose && docker compose --env-file ../.env up --build
# Web http://localhost:4200  ·  API http://localhost:8080/api/v1/health
```

Sem Docker:

```bash
# backend (precisa de Postgres e das variáveis DB_*, AUTH0_*)
cd backend && go run ./cmd/migrate && go run ./cmd/api
# frontend
cd frontend/bogoboom-frontend && npm ci && npm start
```

Testes: `cd backend && go test -race ./...` (defina `BOGOBOOM_TEST_DB` para o teste de integração) e `cd frontend/bogoboom-frontend && npm test`.

## Mapa do repositório

- `backend/` API Go — leia `backend/backend-rules.md`
- `frontend/bogoboom-frontend/` SPA Angular — leia `frontend/frontend-rules.md`
- `devops-compose/` stack local
- `docs/` arquitetura, matching, API, segurança, deploy, Auth0, roadmap
- `AGENTS.md` regras para quem (ou qual agente) altera o código

## Documentos

[ARCHITECTURE](docs/ARCHITECTURE.md) · [MATCHING](docs/MATCHING.md) · [API](docs/API.md) · [SECURITY](docs/SECURITY.md) · [AUTH0-SETUP](docs/AUTH0-SETUP.md) · [DEPLOY](docs/DEPLOY.md) · [ROADMAP](docs/ROADMAP.md)

> A política de privacidade do site é um **rascunho** e precisa de revisão jurídica antes da publicação (LGPD/Lei 1581 da Colômbia).
