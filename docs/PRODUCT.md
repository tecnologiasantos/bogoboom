# Bogoboom: visão do produto

Fonte: documentos de negócio da pasta do projeto no Google Drive (descrição do negócio, estrutura do MVP, sugestões de produto, convite às imobiliárias). O deployment antigo (PHP, MariaDB, nginx) não é referência para nada aqui.

## A ideia

Bogoboom é um portal que liga **recém-chegados**, **equipes de RH** e **imobiliárias verificadas** para resolver a mudança para uma cidade nova com uma moradia de longo prazo, em vez de semanas de hotel. Lança em **Bogotá, Colômbia**, e foi desenhado para repetir o modelo em qualquer cidade.

A ideia nasceu da experiência de se mudar para Bogotá: achar casa segura, entender bairros e contratos, equilibrar qualidade e orçamento. Empresas que contratam de outras cidades sofrem o mesmo problema em escala (BPO, tecnologia, terceirização e educação).

## Para quem

| Papel | Acesso | O que faz | Dinheiro |
|---|---|---|---|
| Recém-chegado | Grátis, sempre | Perfil (cidade do contrato, data de chegada, preferências), busca com filtros, favoritos, envio de documentos opcional | Nenhum |
| RH | Grátis | Conta da empresa, cadastra recém-chegados (manual ou CSV), vê imóveis compatíveis, acompanha o progresso da moradia | Ferramentas premium no futuro |
| Imobiliária | Gratuito para listar, pago para falar com leads | Perfil, imóveis com fotos, vê necessidades de moradia (borradas), assina para liberar contato, mensagens e downloads | Planos Free, Pro e Enterprise |

## Regras que definem o produto

1. **Contato protegido no servidor.** Imobiliária sem assinatura ativa vê a necessidade (cidade, tipo, datas, faixa de preço), mas nome e contato ficam borrados. Esconder só na tela não basta: a API não pode entregar o dado.
2. **Quem pode ser contatado é escolha do usuário.** O recém-chegado e a empresa podem desligar o contato de imobiliárias.
3. **Mensagens e downloads de documentos só para imobiliária com assinatura ativa.** Todo download é registrado.
4. **Recém-chegado nunca escreve direto para a imobiliária sem ela estar assinante.** O interesse gera uma notificação.
5. **Combinação por regras primeiro** (cidade, tipo, datas, preço, mobiliado). Aprendizado de máquina só depois, com dados reais.

## Fluxo central: da necessidade ao contato

1. O RH cadastra um recém-chegado que chega em 15 de maio a Bogotá e precisa de um quarto mobiliado.
2. O sistema registra a necessidade e gera uma notificação para imobiliárias com imóveis na cidade.
3. A imobiliária vê a notificação no painel.
4. Sem assinatura: vê os dados borrados e um convite para assinar. Com assinatura: vê o perfil completo, escreve ao RH ou ao recém-chegado e baixa os documentos anexados.

## Modelo de negócio

- **Imobiliárias, começo:** contatos gratuitos por mês (a sugestão é 5) para criar confiança, depois assinatura.
- **Planos:** Free (lista sem limite, poucos créditos de contato), Pro (contato ilimitado, análises, destaque) e Enterprise (equipe, API, integração com CRM). Preços por mercado ainda não definidos.
- **Depois da tração:** imóveis em destaque, pacotes corporativos de mudança para RH, afiliados (aluguel de móveis, internet, banco), dados agregados de mercado.

## Parcerias previstas

Imobiliárias (selo de parceiro, eventos e webinars, expansão), empresas e RH (marca empregadora, programa de indicação), mídia e conteúdo (podcasts, guias, newsletters), e alianças futuras com bancos, escritórios de imigração e direito, escolas de idiomas e empresas de mudança.

## Roadmap

| Marco | Objetivo |
|---|---|
| V3 · Landing e crescimento | Site no ar, lista de espera, SEO, parceiros e piloto |
| V3 · Fundação | Hospedagem, fluxos dos três públicos, segurança e modelo de dados |
| V3 · MVP Bogotá | Valor central para os três papéis, com contato protegido e combinação por regras |
| V3 · Monetização e confiança | Assinaturas, verificação, avaliações, mensagens e conformidade (Lei 1581 e LGPD) |
| V3 · Escala | Assistente MIRA, integração com a vaga MIRA/MRE-AI, várias cidades, API e dados de mercado |

Estimativa original do documento do Drive para o MVP: 8 a 12 semanas com 1 a 3 pessoas.

## Ainda em aberto

Hospedagem da plataforma (a stack já está decidida: ver `docs/adr/0001-stack-e-fluxo-de-trabalho.md`), preços por mercado, entidade legal para a América Latina, processo operacional de verificação de imobiliárias e incentivo de entrada para as primeiras agências.
