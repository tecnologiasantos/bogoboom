# Match score

Implementado em `backend/internal/core/domain/matching.go` (com testes).

| Critério | Pontos |
|---|---|
| Orçamento | 35 |
| Disponibilidade (data) | 20 |
| Quartos | 15 |
| Mobiliado | 15 |
| Bairro preferido | 15 |

Regras duras (score 0): cidade diferente ou preço acima de 125% do orçamento.
Limite para notificar/desbloquear: `NotifyThreshold = 50`.
Cidades são comparadas normalizadas (sem acento/caixa) em `city_key`.
