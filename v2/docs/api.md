# API reference (v0.1) — base URL `http://localhost:4000`
Auth: `Authorization: Bearer <jwt>`. `AUTH_MODE=dev` exposes `POST /auth/dev-login` (refused in production); `AUTH_MODE=oidc` verifies Cognito/OIDC JWTs via JWKS and reads the role from `custom:role` (only `newcomer|hr|agency` can be self-assigned; `admin` is set in the DB).

| Method & path | Role | Purpose |
|---|---|---|
| `GET /health` | – | liveness |
| `POST /auth/dev-login` `{email, role}` | – (dev) | token for local development |
| `GET /me` | any | current user |
| `PUT/GET /newcomer/profile` | newcomer | save profile (claims HR-created invite by e-mail), consent logged |
| `GET /newcomer/matches` | newcomer | scored listings (≥ 50) |
| `GET /properties`, `/properties/:id` | any | search: `city, district, min_price, max_price, type, min_bedrooms, furnished, available_before, limit, offset` |
| `GET /favorites`, `PUT/DELETE /favorites/:propertyId` | newcomer | favorites (unique) |
| `POST/GET /hr/company` | hr | company |
| `POST /hr/newcomers` | hr | add newcomer |
| `POST /hr/newcomers/import` (`text/csv`) | hr | bulk import (≤ 1000 rows) → `{created, errors[{row,error}]}` |
| `GET /hr/newcomers[?stage=]` | hr | dashboard data + stage summary |
| `PATCH /hr/newcomers/:id/stage` | hr | move through `invited → … → moved_in` |
| `POST/GET /agency` | agency | create / view agency (plan, credits) |
| `POST/GET /agency/properties`, `PATCH/DELETE /agency/properties/:id` | agency | listings CRUD (soft delete) |
| `GET /agency/leads[?min_score=]` | agency | **blurred** matched leads |
| `GET /agency/leads/:id` | agency | blurred, or full contact if unlocked and still consented |
| `POST /agency/leads/:id/unlock` | agency | uses 1 of 5 monthly credits (free plan) or none (active Pro/Enterprise) → `402` when exhausted |
| `GET /admin/agencies/pending`, `POST /admin/agencies/:id/verify` | admin | verification queue |
| `POST /webhooks/payments` | provider | HMAC-signed (`x-signature`), idempotent subscription updates |

## Privacy invariants (all covered by `apps/api/test/api.test.ts`)
1. Agencies never receive name, e-mail, phone or company name before unlocking — the response serializer (`blurred()`) has no such fields.
2. A lead can only be unlocked by an agency it was matched to (no id enumeration).
3. `allow_contact = false` blocks unlocking and hides previously unlocked contacts.
4. Every unlock/view is written to `contact_unlocks` + `audit_logs`.
