# API `/api/v1`

Autenticação: `Authorization: Bearer <JWT Auth0>` em tudo, exceto onde indicado.

| Método | Rota | Quem | Descrição |
|---|---|---|---|
| GET | /health | público | saúde (fora de /api/v1) |
| POST | /partner-inquiries | público* | pedido de parceria |
| POST | /leads | público* | lead da landing |
| POST | /register/newcomer · hr · agency | token | cria conta |
| GET | /me | token | conta e papel |
| GET/PUT | /newcomers/me | newcomer | perfil, consentimento |
| GET | /properties/search · favoritos | newcomer | busca (endereço oculto) |
| GET | /hr/newcomers · POST convite | hr | acompanha equipe |
| GET | /agency/account · leads · notifications | agency | painel |
| POST | /agency/leads/:id/unlock | agency verificada | gasta 1 crédito |
| CRUD | /agency/properties | agency | imóveis |
| GET | /admin/partner-inquiries | admin | pedidos |
| POST | /admin/agencies/:id/verify · plan | admin | verificação e plano |

\* limite por IP e campo isca `website`. Erros: `{ "message": "...", "details": {campo: erro} }`; 402 = sem créditos, 429 = limite de requisições.
