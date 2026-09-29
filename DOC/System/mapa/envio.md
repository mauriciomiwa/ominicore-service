> Revisado por: Mauricio Miwa, 2026-09-28

# Envio

> Fatos do inventário de 2026-09-28 (`inventario.md`) e do código da change
> `01-enviar-email`; deduções marcadas com (deduzido). Revisado por: —, —.

## Para que serve
Um Sistema de origem autenticado pede uma Notificação por Template; o worker a
envia pelos Canais Email (SendGrid) e SMS (Twilio) na Ordem de Canais, com
Retentativa de Falha temporária e fallback para o próximo Canal, sem duplicar
em reenvio (RN-1, RN-2, RN-3, RN-7).

## Telas / entradas
| Rota | Tela | Arquivo |
|---|---|---|
| `POST /v1/notificacoes` | — (API; contrato em `api/openapi.yaml`) | `internal/adaptadores/entrada/http/servidor.go:34` |
| `GET /healthz` | — (healthcheck do quadlet) | `internal/adaptadores/entrada/http/servidor.go:33` |
| worker (fila no Postgres) | — | `internal/adaptadores/entrada/worker/worker.go` |

## Fluxo principal
```mermaid
sequenceDiagram
  participant S as Sistema de origem
  participant A as ominicore api
  participant K as Keycloak (JWKS)
  participant P as Postgres
  participant W as ominicore worker
  participant G as SendGrid
  S->>A: POST /v1/notificacoes (Bearer + Idempotency-Key)
  A->>K: JWKS (cache no Redis, 10 min)
  A->>P: ReivindicarChave + InserirNotificacao + InserirTentativa (1 transação)
  A-->>S: 202 {id, status}
  W->>P: ReservarTentativa (FOR UPDATE SKIP LOCKED)
  W->>G: POST /v3/mail/send (ou Twilio, Canal SMS)
  W->>P: ConcluirTentativa + próxima Tentativa (Retentativa ou próximo Canal) + status (commit)
```

## Chamadas ao back
| Método | Endpoint | Chamado em | Para quê |
|---|---|---|---|
| POST | SendGrid `/v3/mail/send` | `internal/adaptadores/saida/sendgrid/sendgrid.go` | enviar a Tentativa de Email |
| POST | Twilio `/2010-04-01/Accounts/{sid}/Messages.json` | `internal/adaptadores/saida/twilio/twilio.go` | enviar a Tentativa de SMS |
| GET | Keycloak `<realm>/protocol/openid-connect/certs` | `internal/adaptadores/saida/keycloak/validador.go` | JWKS para validar o token |
| POST | Keycloak `<realm>/protocol/openid-connect/token` | `internal/adaptadores/saida/keycloak/token.go:17` | token do próprio OminiCore para o OpenBao |
| — | OpenBao `auth/jwt/login`, `kv/data/ominicore/*` | `internal/adaptadores/saida/openbao/segredos.go` | credenciais no boot (REQ-6) |

## Estado e dados
Tabelas (schema `ominicore`, `migrations/00001_notificacoes.sql`): `templates`,
`template_conteudos`, `notificacoes`, `idempotencia` (chave por Sistema de
origem, `expira_em` = 24 h), `tentativas` (a fila; Retentativa = nova linha
com `disponivel_em` no futuro). Seed: Template `boas-vindas`, Email → SMS
(`migrations/00002_*`, `00003_sms.sql`); `notificacoes.destinatario_telefone` (E.164).

## Regras de negócio encontradas (a confirmar)
Nenhuma além das RN da especificação.
- (deduzido) `canais` repetido no pedido é recusado com 400
  (`entrada/http/servidor.go`, `validarFormato`) — formato, não regra; confirmar.

## Specs
Specs vigentes `openspec/specs/autenticacao-origem/`, `notificacoes/` e
`segredos/` (REQ-1 a REQ-6); change ativa `openspec/changes/02-sms-e-fallback/`
(REQ-2 modificado, REQ-7 a REQ-9).
