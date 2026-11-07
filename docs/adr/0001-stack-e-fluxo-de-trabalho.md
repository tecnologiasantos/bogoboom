# ADR 0001 · Stack e fluxo de trabalho

Status: aceita (decisão do responsável pelo projeto, 2026-11-07). Issue: #71.

## Decisão

- **Linguagens e stack da plataforma:** as mesmas do MIRA (miratech). Backend em Go com Fiber e PostgreSQL; frontend em Angular. Arquitetura hexagonal no backend e contrato JSON em `camelCase` entre frontend e backend.
- **Fluxo de trabalho:** o mesmo do MIRA. Toda mudança parte de uma issue; cada issue tem spec e plan próprios, escritos antes de codar; um PR por issue, com `Closes #n`; testes de tela com Playwright.
- **Hospedagem:** a landing continua estática no GitHub Pages. A hospedagem da plataforma é uma decisão futura e fica fora desta ADR. O modelo do MIRA (Docker Swarm atrás de proxy reverso, deploy por script) é a referência provável.

## Motivo

Uma equipe pequena mantém dois produtos. Reaproveitar linguagem, padrões e processo reduz custo de manutenção e de entrada de novas pessoas.

## Consequências

- A v2 em Next.js e Node e a proposta de Go + Angular do PR #61 deixam de ser alternativas em aberto.
- Definições que dependem de hospedagem (domínio da API, armazenamento de arquivos, backups) esperam a decisão futura.
- Segredos nunca entram no git, em nenhuma camada.

## Fluxo por issue

| Documento | Caminho | Conteúdo |
|---|---|---|
| Spec (o quê e por quê) | `docs/superpowers/specs/AAAA-MM-DD-<branch>-design.md` | Problema, critérios de aceite, abordagem, fora de escopo, como testar |
| Plan (como) | `docs/superpowers/plans/AAAA-MM-DD-<branch>.md` | Passos pequenos com teste primeiro, arquivos afetados, validação |

Os dois começam com `Issue: #<n>` e o nome da branch, são o primeiro commit da branch e são linkados na seção `## Spec / plan` da issue.
