# AGENTS.md — regras para alterar o Bogoboom

1. Backend em `backend/`, frontend em `frontend/bogoboom-frontend/`. Leia `backend/backend-rules.md` e `frontend/frontend-rules.md` antes de mexer.
2. **Nunca** commite segredos, `.env`, chaves, senhas ou tokens. Nenhum segredo tem valor padrão no código.
3. Regra de negócio fica em `backend/internal/core` (somente stdlib, testável sem rede). Adaptadores (HTTP, Postgres) não decidem regra.
4. Toda rota nova em `/api/v1` exige autenticação e guarda de papel; o teste em `cmd/api/main_test.go` falha se uma rota privada responder sem token.
5. Alterou o schema? Crie uma nova migration `backend/migrations/NNN_nome.sql`; nunca edite uma já aplicada.
6. Dados pessoais: endereço do imóvel nunca vai ao recém-chegado; contato do recém-chegado só após consentimento + desbloqueio auditado.
7. Antes de abrir PR: `gofmt -l`, `go vet ./...`, `go test -race ./...`, `npm run build`, `npm test`.
8. Atualize `docs/` quando mudar comportamento (matching, API, segurança).
