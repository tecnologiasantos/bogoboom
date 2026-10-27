# Bogoboom — Developer Handoff

Status date: 2026-10-27 · Branch: `rewrite-go-angular` · PR #61 (CI green, not yet merged to `main`)

Purpose: give the local developer everything needed to review, run, fix and extend the project. Read `README.md`, `AGENTS.md`, `backend/backend-rules.md` and `frontend/frontend-rules.md` first.

## 1. What this is
Bogoboom connects **newcomers** relocating to Bogotá, **HR teams** and **real-estate agencies**. Newcomers state needs and opt in to contact; agencies see blurred leads with a match score and unlock a contact (verified agency + opt-in + score ≥ 50 + credit or Pro plan). Every unlock is audited.

## 2. Stack
Go 1.25 · Fiber v3 (rc.3) · pgx v5 · PostgreSQL 16 · Angular 21 (standalone, Signals, SCSS) · Auth0 (RS256/JWKS) · Docker Compose. Same patterns as the sibling project *miratech*.

## 3. Run it
```bash
cp .env.example .env                       # fill values; no secret has a default
cd devops-compose && docker compose --env-file ../.env up --build
# Web http://localhost:4200 · API http://localhost:8080/api/v1  · health /health
```
Needs an Auth0 tenant first: `docs/AUTH0-SETUP.md`, then edit `frontend/bogoboom-frontend/public/config.json`.

Tests:
```bash
cd backend && go vet ./... && go test -race ./...        # BOGOBOOM_TEST_DB=postgres://... enables the integration test
cd frontend/bogoboom-frontend && npm ci && npm run build && npm test
```

## 4. Repository map
| Path | Content |
|---|---|
| `backend/internal/core/domain` | entities, errors, validation, match scoring (stdlib only) |
| `backend/internal/core/service` | use cases (registration, leads/unlock, properties, public forms) |
| `backend/internal/adapters/{handler,repository,middleware}` | Fiber handlers, pgx repos, auth/role/limiter/headers |
| `backend/cmd/{api,migrate}` | server and SQL migration runner |
| `backend/migrations` | `001_init.sql`, `seed.sql` (dev only) |
| `frontend/bogoboom-frontend/src/app` | `core/` (api, session, guards), `features/` (landing, register, newcomer, hr, agency, admin, privacy, partnership) |
| `devops-compose` | local stack |
| `docs` | architecture, matching, API, security, Auth0, deploy, roadmap |

## 5. Verified vs. NOT verified
**Verified (CI):** build, vet, gofmt, migrations, `go test -race` incl. full-flow integration test on real PostgreSQL 16, route-security tests (every private route → 401 without token), frontend build and unit tests, `npm audit` (0 vulnerabilities).

**Not verified — please test first:**
1. Real login/signup against a live Auth0 tenant (token claims, `email` claim Action, refresh-token rotation).
2. `docker compose up` end to end, and the production Dockerfiles.
3. Cross-browser/mobile UI review of every Angular page; accessibility pass.
4. Load/abuse behavior of the rate limiter behind a real proxy (`TRUSTED_PROXIES`, `X-Real-IP`).
5. Frontend test coverage is minimal (API service only). Add component and guard tests.

## 6. Known gaps / suggested work (priority order)
1. **Auth0 end-to-end test** and fix any claim/role mismatch found.
2. **Admin bootstrap:** no public way to become admin by design; set `users.role='admin'` via SQL. Consider a CLI command.
3. **E-mails:** none are sent (invites, notifications). Add a provider (SES or similar) behind a port in `core`.
4. **Payments:** Pro plan is set manually by admin (`PUT /admin/agencies/:id/plan`).
5. **Admin UI:** agency actions require pasting the agency UUID; add an agencies list endpoint and table.
6. **Photos/uploads** for listings (none yet; plan S3 + presigned URLs).
7. **i18n:** UI is English only; target ES / PT / EN.
8. **Frontend bundle** is ~509 kB initial (budget warning raised to 600 kB); consider trimming.
9. Make `govulncheck` and `npm audit` blocking in CI once the baseline is clean.
10. Privacy policy/terms are a **draft**; legal review required (LGPD, Colombian Law 1581).

## 7. Design decisions to keep
- Business rules only in `core` (stdlib, no network in tests).
- User resolved by Auth0 `sub` only; e-mails normalized to lowercase.
- Fiber `Group` middleware applies to routes registered after it: newcomer routes use a per-route guard; `hr/agency/admin` use prefixed groups. Add routes in the right place and extend `cmd/api/main_test.go`.
- Credits spent atomically in `agency_repository.Unlock`; monthly reset is lazy (5 free/month).
- Property street address is never returned to newcomers.
- Public forms carry a `website` honeypot; do not remove.
- New schema change = new migration file; never edit applied ones.
- Runtime frontend config via `/config.json` so one build serves any environment.

## 8. Security checklist before any public deploy
- [ ] Rotate DB password, Gemini key and Gmail app password that existed in the **old** git history; revoke the GitHub token shared early in the project (treat all as compromised).
- [ ] Enable on GitHub: `main` branch protection, secret scanning + push protection, Dependabot alerts, required 2FA.
- [ ] `APP_ENV=production`, explicit `FRONTEND_URL`, `AUTH0_*`, `TRUSTED_PROXIES` set to the load-balancer IP.
- [ ] `DB_SSLMODE=require`, backups enabled, secrets in a secret manager (never in the image or repo).
- [ ] HTTPS + HSTS at the edge; WAF/rate limits in front.
- [ ] Legal review of privacy policy and terms.

## 9. Deployment notes (AWS vs Cloudflare)
Frontend is static: Cloudflare Pages or S3 + CloudFront. The Go API needs a container/VM (AWS ECS Fargate, App Runner or EC2); it does not run on Cloudflare Workers. Postgres on RDS. Details: `docs/DEPLOY.md`.

## 10. Working agreement
Branch from `main`, small PRs, CI must be green, update `docs/` with behavior changes, never commit secrets. Questions about product rules: see `docs/MATCHING.md` and `docs/API.md`; for anything else contact support@tecnologiasantos.com.
