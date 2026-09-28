# Arquitetura — ominicore

Hexagonal (ports and adapters): o domínio não conhece HTTP, banco nem
Provedor; adaptadores implementam as portas. Diagrama de origem:
`gemini-code-1790619852596.html` (kickoff 2026-09-28).

## Componentes

```
cmd/ominicore/            main: subcomandos `api` e `worker` (mesmo binário)
internal/
  dominio/                Notificação, Tentativa, Template, Canal, Opt-out;
                          NotificationRouter (Ordem de Canais, retry, fallback)
  aplicacao/              casos de uso (portas de entrada):
                          SendMessageUseCase, ReceiveMessageUseCase
  portas/                 portas de saída: MessageSenderPort (um por Canal),
                          MessageRepositoryPort, SecretsPort, RateLimiterPort
  adaptadores/
    entrada/http/         API REST de envio (/v1/*) e webhooks (/webhooks/*)
    entrada/worker/       laço que puxa Tentativas da fila no Postgres
    saida/postgres/       repositório + fila (pgx/sqlc, SKIP LOCKED)
    saida/redis/          rate limit por Provedor, cache de Template e JWKS
    saida/sendgrid/       Canal Email
    saida/twilio/         Canal SMS
    saida/openbao/        leitura de segredos (login jwt via Keycloak)
    saida/keycloak/       validação de JWT (JWKS)
api/openapi.yaml          contrato (API-First)
migrations/               goose
deploy/                   podman compose (local), quadlets e publicar.sh (aviva2.0)
```

| Componente | Responsabilidade |
|---|---|
| `ominicore api` | autentica o Sistema de origem, valida o pedido, grava a Notificação e a 1ª Tentativa na mesma transação; recebe webhooks dos Provedores |
| `ominicore worker` | puxa Tentativas prontas, envia pelo Provedor do Canal, decide retry/fallback pelo NotificationRouter |
| Postgres (`ominicore`) | Notificações, Tentativas (a fila), Templates, Opt-outs, chaves de idempotência |
| Redis | rate limit por Provedor, cache de Template e JWKS — perder o Redis degrada, não perde dado |
| OpenBao | credenciais SendGrid/Twilio e senha do banco (`kv/ominicore/*`) |
| Keycloak | emite os tokens dos Sistemas de origem e do próprio OminiCore (login no Bao) |

## Fluxos-chave

**Envio (RN-1, RN-7):** Sistema de origem → `POST /v1/notificacoes` com JWT e
`Idempotency-Key` → API valida token (JWKS), Template e variáveis → numa
transação grava Notificação + Tentativa `pronta` (chave única por Sistema de
origem + chave de idempotência em 24 h) → `202` com o id.

**Despacho (RN-2, RN-3, RN-8):** worker faz `SELECT … FOR UPDATE SKIP LOCKED`
nas Tentativas prontas → pula Canal com Opt-out → envia pelo Provedor →
sucesso: Tentativa `enviada`; Falha temporária: nova Tentativa no mesmo Canal
com espera 5 s / 30 s / 2 min (até 3); Falha definitiva ou esgotou: próximo
Canal da Ordem; sem Canal: Notificação `falhou`.

**Status (RN-5):** Provedor → `POST /webhooks/{provedor}` (público) →
assinatura validada → status atualiza a Tentativa.

**Retenção (RN-9):** tarefa periódica do worker anonimiza conteúdo e contato de
Notificações com mais de 30 dias (prazo configurável).

## Decisões
- `decisoes/0001-fila-de-despacho-no-postgres.md`
- `decisoes/0002-login-no-openbao-via-jwt-do-keycloak.md`
- `decisoes/0003-webhooks-publicos-api-interna.md`
