# Modelo de dados (conceitual)

Ponto de partida: o esquema do documento de negócio do Drive, com correções. É conceitual: o banco e a stack serão decididos na issue de decisão.

## Entidades

| Entidade | Papel | Guarda |
|---|---|---|
| Usuário | todos | e-mail, papel (`newcomer`, `hr`, `agency`), status, verificado |
| Empresa | RH | nome, setor, endereço, permite contato de imobiliárias |
| Perfil de RH | RH | usuário, empresa, nome, telefone |
| Recém-chegado | recém-chegado e RH | cidade do contrato, data de chegada, preferências, permite contato, quem cadastrou (ele mesmo ou o RH) |
| Imobiliária | imobiliária | nome, endereço, site, documento, descrição |
| Perfil de imobiliária | imobiliária | usuário e imobiliária (várias pessoas por imobiliária) |
| Imóvel | imobiliária | título, descrição, preço, tipo (`apartment`, `room`, `house`, `studio`), mobiliado, localização (latitude e longitude), fotos, disponível desde |
| Favorito | recém-chegado | recém-chegado e imóvel |
| Assinatura | imobiliária | plano, início, fim, status (`active`, `expired`, `cancelled`) e créditos de contato |
| Notificação | imobiliária | cidade, preferências, tipo (`new_search`, `update`, `interest_shown`), lida |
| Mensagem | assinante | remetente, destinatário, conteúdo, lida |
| Arquivo enviado | todos | dono, tipo, descrição, localização no armazenamento |
| Download | imobiliária assinante | quem baixou, qual arquivo, quando |
| Registro de auditoria | sistema | quem fez o quê em qual registro |

## Correções em relação ao esquema antigo

- **Nenhuma senha em tabelas de perfil.** O esquema antigo repetia `password` e `confirm_password` em cada perfil. Senha só existe em um lugar (o provedor de identidade ou uma única tabela com hash).
- **Um único vocabulário de papéis:** `newcomer`, `hr`, `agency`. O esquema antigo misturava singular e plural (`newcomers`, `companies`, `agencies`).
- **Contato borrado é regra do servidor**, decidida pela assinatura ativa, nunca só pela tela.
- **Preferências flexíveis** em campo estruturado (JSON), para filtros novos sem migração.
- **Créditos de contato** na assinatura, para o modelo "alguns contatos grátis por mês".
- **Latitude e longitude** em imóveis e empresas, para busca por proximidade.
- **Dados pessoais mínimos** e consentimento registrado (Lei 1581 da Colômbia e LGPD).
