# Regras do frontend

- Angular 21 standalone + Signals; componentes lazy por rota (`app.routes.ts`).
- Configuração em tempo de execução via `public/config.json` (mesmo build para qualquer ambiente). Nunca coloque segredos aqui: `clientId` e `domain` do SPA Auth0 são públicos por natureza.
- Chamadas HTTP somente por `core/api.service.ts`; erros viram texto seguro com `describeError`.
- Guardas em `core/guards.ts`: `roleGuard(...)` e `signedInGuard`. A guarda é conveniência; quem protege os dados é a API.
- Sem `innerHTML`/`bypassSecurityTrust*`. Todo texto vindo da API é interpolado.
- Identidade visual: Arial, fundo `#f5f5f5`, azul `#004a8f`, cartões brancos raio 8px/sombra suave, botão preto, destaque `#e6f2ff` — tokens em `src/styles.scss`.
- Formulários públicos têm campo isca (`website`) invisível; não remova.
