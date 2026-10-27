# Regras do backend

- Camadas: `core/domain` (entidades, validação, matching) → `core/service` (casos de uso) → `adapters/{handler,repository,middleware}` → `cmd/*`.
- `core` importa apenas a stdlib. Dependências apontam para dentro.
- Erros de domínio são sentinelas (`internal/core/domain/errors.go`); `handler/respond.go` mapeia para HTTP (402 sem créditos, 403, 404, 409, 422).
- Validação e sanitização de entrada em `domain/validate.go`; nunca confie no corpo da requisição.
- Usuário resolvido **apenas** pelo `sub` do Auth0 (nunca por e-mail do corpo). E-mail é normalizado em minúsculas.
- Rotas: grupos com middleware por prefixo (`/hr`, `/agency`, `/admin`); rotas de recém-chegado usam guarda por rota (Fiber aplica middleware de `Group` às rotas registradas depois).
- Crédito: gasto atômico no repositório (`agency_repository.Unlock`); reset mensal preguiçoso (5 grátis/mês).
- SQL só parametrizado (pgx). Transações via `inTx`.
- Configuração só por variáveis de ambiente; em `APP_ENV=production` a validação exige Auth0 e CORS explícito.
- Testes: unitários no core (rápidos), `main_test.go` para segurança das rotas, integração com `BOGOBOOM_TEST_DB`.
