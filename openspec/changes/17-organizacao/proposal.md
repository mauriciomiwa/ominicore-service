# Proposal

## Why

O OminiCore nasceu para uma casa só (kickoff de 2026-09-28: "não
multi-tenant"). A sabatina de 2026-11-14 decidiu que ele passa a atender
várias **Organizações** (empresas ou marcas do grupo), cada uma com Provedores,
credenciais, Templates e Operadores próprios. Esta primeira fatia dá a base: toda
Notificação pertence à Organização do Sistema de origem que a pediu, e o banco
impede que uma Organização leia a Notificação de outra. Origem: **RN-15**
(nova), que estende a RN-1 e a RN-5.

## What Changes

- **RN-15 nova** em `DOC/Negocio/especificacao-negocio.md`: "Cada Sistema de
  origem pertence a uma só Organização; a Notificação fica com a Organização de
  quem a pediu e nenhuma Organização enxerga a de outra. Sistema de origem sem
  Organização é recusado. Opt-out e retenção valem para todas." Fatias 17, 18 e
  19 no `DOC/System/backlog-mvp.md`; decisões da sabatina no `DOC/log.md`.
- **Cadastro por migration:** Organizações e o vínculo client → Organização ficam
  no banco e mudam só por migration, feita pelo dono. A primeira Organização é
  `wbem`, e o client `wbem` (aviva-back) pertence a ela.
- **Dados existentes:** a migration coloca na Organização `wbem` todas as
  Notificações e Templates já gravados. É aditiva: o down remove as colunas e as
  tabelas novas, sem perder nada do que já existia.
- **Recusa sem Organização (BREAKING para client não cadastrado):** um token
  válido de um client sem Organização recebe 403 ao pedir ou consultar uma
  Notificação e ao registrar ou remover um Opt-out, e nada é criado. Em prod o
  único Sistema de origem é o `wbem`, que já fica cadastrado. O client do
  Console continua com o 403 do REQ-36.
- **Barreira no banco:** além do filtro por Sistema de origem que já existe, o
  Postgres recusa entregar a uma requisição da API uma Notificação de outra
  Organização (row-level security).
- **Sem mudança visível para o `wbem`:** contrato da API, Templates, Canais,
  credenciais, Opt-out, retenção, worker e Avisos de status ficam como estão.
  Templates e credenciais por Organização ficam para a fatia 18; Operadores e
  Console por Organização, para a 19. A change 15 (fila da Z-API) espera a 18.

## Capabilities

### New Capabilities
- `organizacoes`: a quem cada Sistema de origem e cada Notificação pertencem e
  o isolamento entre Organizações. As fatias 18 e 19 acrescentam Templates,
  credenciais e Operadores.

### Modified Capabilities
(nenhuma: o REQ-1 e o REQ-14 continuam valendo como estão; o 403 para client
sem Organização é requisito novo em `organizacoes`)

## Impact

- **Banco:** migration `00013_organizacoes.sql` com as tabelas `organizacoes` e
  `sistemas_origem`, a coluna `organizacao` em `notificacoes` e `templates`,
  preenchida com `wbem`, e uma política de row-level security em
  `notificacoes`. A change 15 renumera a sua migration para a próxima livre.
- **Código:** a autenticação resolve a Organização do client, o envio grava a
  Organização na Notificação e as consultas da API rodam com a Organização
  fixada na transação. O worker, a retenção e os Avisos de status continuam
  sem filtro de Organização.
- **Ambientes:** a migration roda em hml e prod pelo `publicar.sh` (skill
  `ambientes`). Antes de publicar, confirmar que nenhum outro client chama a
  API em hml ou prod.
- **Contrato:** `api/openapi.yaml` documenta o 403 "Sistema de origem sem
  Organização". Não há campo novo nas respostas.
