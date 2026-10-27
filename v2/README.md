# Bogoboom 🏡

> **Location:** this app lives in `v2/` of the `bogoboom` repository; run all commands from this folder. The legacy PHP site at the repo root is untouched.

> Connecting newcomers, HR teams and real estate agencies for a smooth landing in a new city. Launching in **Bogotá, Colombia**, built to scale. A Tecnologia Santos × TechBR Solutions project, sister product of the MIRA / MRE-AI job portal.

| Role | Access | Does | Money |
|---|---|---|---|
| **Newcomer** | Free | profile, search/filter, favorites, matches | – |
| **HR** | Free | company, add/import newcomers, track relocation | premium tools later |
| **Agency** | Freemium | list properties, receive matched leads (contact hidden), unlock with 5 free credits/month or Pro | Pro / Enterprise |

## What works today (v0.1)
- PostgreSQL schema v1 + migration runner
- REST API: auth (dev / OIDC), newcomer, property search, HR (CSV import, stage tracking), agency (listings, leads, credits), admin verification, signed payment webhook
- **Server-side contact privacy** + rule-based matching, covered by 18 integration/unit tests
- Next.js web: landing, login, newcomer, HR dashboard, agency leads
- CI (typecheck, tests on Postgres, build, gitleaks, CodeQL), Dependabot, issue/PR templates, 48 seeded roadmap issues

## Quick start
```bash
cp .env.example .env
docker compose up -d db
npm install
npm run migrate && npm run seed      # demo data for Bogotá
npm run dev:api                      # http://localhost:4000
npm run dev:web                      # http://localhost:3000
```
Demo logins (dev mode, no password): `hr@demo.test` (HR), `agency@demo.test` (agency), `newcomer@demo.test` (newcomer).

Tests need a throw-away database (**it is wiped**): `TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/bogoboom_test npm test`

## Layout
`apps/api` · `apps/web` · `db/migrations` · `docs/` (api, matching, architecture, ADRs) · `.github/` (CI + seed data for issues) · `scripts/seed_github.sh`

## GitHub setup (once)
1. Push to `main`. 2. Actions → **Seed labels, milestones & issues** → Run. 3. Enable branch protection, secret scanning/push protection, Dependabot alerts. See [ROADMAP.md](ROADMAP.md).

## Not production-ready yet
Auth is dev-only until Cognito is wired, no e-mail sending, no S3 photos, no payments UI. Tracked in the issues.
