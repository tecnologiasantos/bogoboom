# ADR-0001: MVP technology stack
Status: **Accepted (implemented in v0.1)**

## Decision
- Web: Next.js + Tailwind (SEO, SSR). API: Node.js + TypeScript (Fastify + zod + pg). DB: PostgreSQL (RDS) with PostGIS later.
- Auth: Amazon Cognito (email + Google, MFA). Files: S3 + CloudFront with pre-signed uploads. Email: SES. Payments: Stripe + MercadoPago. Search: Postgres FTS first, Algolia/OpenSearch when needed. Maps: Google Maps or MapLibre.
- Hosting: AWS (founder has AWS expertise); IaC with Terraform/CDK.

## Why
Matches the planning document's recommendations; relational model fits users/properties/subscriptions; avoids custom auth.

## Alternatives
Django + DRF (faster admin), Firebase (faster start, vendor lock-in), MongoDB (weaker for relational matching/billing).

## Consequences
Single language (TS) front-to-back; shared zod types; need migrations tooling and IaC discipline.
