# Perguntas de Kickoff — Negócio → Desenvolvimento

Duas rodadas, nesta ordem: **Negócio** (o quê/por quê) e, só com o negócio
fechado, **Desenvolvimento** (o como). Não assuma nada — pergunte. Respostas de
negócio alimentam `Negocio/especificacao-negocio.md`; as de desenvolvimento,
`Padroes/stack-tecnologico.md` e `System/arquitetura.md`.

## Rodada 1 — Negócio

Comece pela 1 (destrava as demais).

1. **O que é** o produto/feature e qual **problema de negócio** resolve? Para **quem**?
2. Quem são os **usuários** (quem opera) e os **stakeholders** (quem decide / é impactado)?
3. Qual o **objetivo de negócio** e como mediremos **sucesso** (métrica observável)?
4. **Situação atual:** o que existe hoje (processo manual, planilha, sistema a
   **substituir** ou **integrar**)? Greenfield ou brownfield? Há dados a migrar?
5. Quais as **regras de negócio / políticas** que o software *sempre* ou *nunca* pode fazer?
6. **Critérios de aceite:** como saberemos, de forma **verificável**, que faz o
   que deveria? (cada um vira teste — coração do SDD)
7. **Restrições:** prazo, orçamento, compliance/regulatório (ex.: LGPD), dados
   sensíveis/PII envolvidos?
8. **Volume/escala esperada** (ordem de grandeza): usuários, registros,
   operações/dia? Tendência de crescimento?
9. **Escopo / MVP:** o que está **dentro** e **fora**? Qual a **menor versão útil**?

### Fechamento de escopo (decisões explícitas)

Quando o domínio tiver variações grandes, feche o MVP decidindo:

- Atender **todas as variações** já, ou começar **genérico/uma só**?
- Funcionalidades acessórias (relatórios, painéis) no MVP ou **depois**?
- **Multi-cliente (multi-tenant)** desde o início?
- O que define a **transação central** no MVP? Corte para o mínimo que entrega valor.

> Marque cada RN como **MVP** ou **depois**. O que não é MVP vai para
> "Fora (próximas etapas)" no escopo.

## Rodada 2 — Desenvolvimento (só após o Negócio fechado)

Esta rodada **parte do inventário do Passo 0**: o que o repositório já revelou
vira confirmação ("achei X no arquivo Y — confirma?"), não pergunta aberta.
Pergunte do zero só o que não deu para detectar.

### Base e stack

1. **Relação com projetos existentes:** do zero, reaproveita **convenções** de um
   existente, ou **faz parte** dele?
2. **Linguagem/stack principal** do backend e do cliente (não assuma — pergunte).
3. **Persistência/backend:** qual banco e qual servidor/API? (reaproveitar padrão existente?)
4. **Build/ferramentas:** gerenciador de pacotes, build, CI.

### Decisões técnicas que o MVP exige

5. **Autenticação/autorização:** como loga e como o **papel** controla acesso? (ex.: JWT)
6. **Multi-tenant:** como isolar dados por cliente/loja? (coluna tenant + filtro, schema, RLS…)
7. **Exposição/segurança:** o serviço é **interno** ou **público**? Como se chega
   nele (VPN/túnel, internet aberta)? Qual o grau de zero trust proporcional?
8. **Infra-alvo de deploy:** onde roda (Nomad, cloud, on-prem, container)?
   E os dados: qual banco (motor e versão) e onde fica o de **hml** e o de **prod** (provedor, instância, nome do banco, backup, quem acessa, path da credencial no cofre)? Vai para `Infra/infraestrutura.md`.
   Reaproveita a esteira de deploy de um projeto existente?
9. **Integrações externas:** quais sistemas/APIs **consome** e quais **expõe**?
   Há contrato (API-First) a definir antes?
10. **Requisitos não-funcionais:** desempenho/latência/disponibilidade que o
    negócio exige — ligue à resposta de volume/escala (Negócio, pergunta 8).

### Condicionais — pergunte só quando a natureza do projeto indicar

- **Cliente opera sem rede?** → offline e **sincronização**: como os dados sobem
  depois sem duplicar? (padrão **outbox + upsert idempotente** por UUID)
- **Registros nascem no cliente?** → **identidade**: UUID gerado no cliente é a
  chave da idempotência? Timestamps de negócio vs de servidor?
- **Edição concorrente / multi-dispositivo?** → estratégia de resolução de conflito.

> Marque o que ficou **em aberto** explicitamente em
> `Padroes/stack-tecnologico.md` (ex.: versões a fixar) para não virar suposição.
