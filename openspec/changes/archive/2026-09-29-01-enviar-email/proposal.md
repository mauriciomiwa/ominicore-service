## Why
RN-1: os Sistemas de origem da casa precisam pedir Notificações num ponto só,
em vez de cada um integrar com o Provedor. RN-7: o reenvio do pedido (timeout
do chamador) não pode fazer o Destinatário receber duas vezes. RN-3: o primeiro
Canal é o Email (SendGrid). RN-9: as credenciais dos Provedores ficam só no
OpenBao desde o primeiro dia.

Esta é a Fatia 1 do `DOC/System/backlog-mvp.md`: prova o caminho inteiro —
autenticação, pedido, fila, Provedor — com um Canal só.

## What Changes
- Contrato OpenAPI de `POST /v1/notificacoes` (API-First).
- Autenticação do Sistema de origem por JWT do Keycloak (realm `ominicore`).
- Pedido de Notificação por Template + dados, com Chave de idempotência
  obrigatória (header `Idempotency-Key`) válida por 24 h por Sistema de origem.
- Worker que puxa Tentativas da fila no Postgres e envia pelo SendGrid.
- Falha do Provedor registrada na Tentativa; sem outro Canal, a Notificação
  termina `falhou` (retry e fallback chegam na Fatia 2).
- Credenciais lidas do OpenBao (login `jwt` via Keycloak); sem elas o processo
  não inicia.
- Templates desta fatia entram por migration (seed); cadastro por API é decisão
  em aberto (`DOC/to-do.md`).
- `podman compose` local com Postgres, Redis, OpenBao (dev) e Keycloak.

Fora do escopo: SMS, retry e fallback (Fatia 2); webhooks de status (Fatia 3);
Opt-out (Fatia 4); anonimização por prazo (Fatia 5); deploy no aviva2.0
(trilha de infraestrutura).

## Impact
- Novas capacidades `autenticacao-origem`, `notificacoes` e `segredos`.
- Novo módulo Go (`go.mod`), binário `ominicore` com subcomandos `api` e
  `worker`, migrations `goose` no schema `ominicore`.
- Dependências novas: `pgx/v5`, `sqlc`, `goose`, `go-redis/v9`,
  `openbao/api/v2`, `testcontainers-go`.
