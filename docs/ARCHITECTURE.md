# Arquitetura

```
Angular SPA ──(JWT Auth0)──▶ Fiber API ──pgx──▶ PostgreSQL 16
      │                          │
      └── login/signup ──▶ Auth0 ◀┘ (JWKS para validar RS256)
```

## Papéis
`newcomer` (recém-chegado), `hr` (RH de empresa, convida recém-chegados), `agency` (imobiliária), `admin`.

## Backend (`backend/`)
- `internal/core/domain`: entidades, erros, validação, matching.
- `internal/core/service`: registro, perfil, leads, imóveis, formulários públicos.
- `internal/adapters/repository`: pgx; `handler`: Fiber; `middleware`: Auth0, papéis, limite por IP, cabeçalhos de segurança.
- `cmd/api` (servidor), `cmd/migrate` (aplica `migrations/*.sql`, registra em `schema_migrations`).

## Fluxos principais
1. **Cadastro**: login Auth0 → `POST /register/{papel}` (somente com token). Convites do RH são vinculados por e-mail quando o recém-chegado se cadastra.
2. **Match**: ao salvar perfil com consentimento ou publicar imóvel, gera notificações para pares com score ≥ 50 (`notify.go`).
3. **Desbloqueio**: imobiliária verificada + opt-in + match ≥ 50 + crédito (ou plano Pro ativo) → contato revelado e registrado em `audit_logs`.

## Dados
Tabelas em `backend/migrations/001_init.sql`: users, consents, companies, hr_profiles, agencies, agency_profiles, newcomers, newcomer_contacts, properties, favorites, notifications, contact_unlocks, partner_inquiries, landing_leads, audit_logs.
