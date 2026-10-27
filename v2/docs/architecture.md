# Architecture
```mermaid
flowchart LR
  N[Newcomer] & H[HR] & A[Agency] --> W[Next.js web]
  W -->|JWT| API[Fastify API · TypeScript]
  API --> DB[(PostgreSQL)]
  API -. later .-> S3[(S3 media/docs)]
  API -. later .-> SES[SES e-mail]
  PAY[Stripe / MercadoPago] -->|HMAC webhook| API
  COG[Cognito / OIDC] -. JWKS .-> API
```
Monorepo: `apps/api` (Fastify, zod, pg, jose, Vitest) · `apps/web` (Next.js App Router) · `db/migrations` (plain SQL, tracked in `schema_migrations`).
Key rule: **the API is the only place that decides what a caller may see** (see `docs/api.md` → privacy invariants).
