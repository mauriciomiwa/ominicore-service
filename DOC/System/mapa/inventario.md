# Inventário do código — ominicore

> GERADO por `tools/mapeia_projeto.py` em 2026-10-05T13:36:49. Não edite: rode o script de novo.
> A interpretação (o que cada área faz, fluxos, RN candidatas) está nas outras páginas desta pasta.

## Stack

| Manifesto | Linguagem | Frameworks | Deps |
|---|---|---|---|
| `go.mod` | go | — | — |

## Módulos

```mermaid
graph LR
  cmd["cmd (3)"]
  internal["internal (66)"]
  migrations["migrations (1)"]
```

| Módulo | Arquivos | Importa |
|---|---|---|
| `internal` | 66 | — |
| `cmd` | 3 | — |
| `migrations` | 1 | — |

## Endpoints (19)

| Método | Caminho | Onde |
|---|---|---|
| GET | `/healthz` | `internal/adaptadores/entrada/http/servidor.go`:42 |
| DELETE | `/v1/config/whatsapp` | `api/openapi.yaml` (contrato) |
| DELETE | `/v1/config/whatsapp` | `internal/adaptadores/entrada/http/servidor.go`:50 |
| GET | `/v1/config/whatsapp` | `api/openapi.yaml` (contrato) |
| GET | `/v1/config/whatsapp` | `internal/adaptadores/entrada/http/servidor.go`:48 |
| PUT | `/v1/config/whatsapp` | `api/openapi.yaml` (contrato) |
| PUT | `/v1/config/whatsapp` | `internal/adaptadores/entrada/http/servidor.go`:49 |
| POST | `/v1/notificacoes` | `api/openapi.yaml` (contrato) |
| POST | `/v1/notificacoes` | `internal/adaptadores/entrada/http/servidor.go`:43 |
| GET | `/v1/notificacoes/{id}` | `api/openapi.yaml` (contrato) |
| GET | `/v1/notificacoes/{id}` | `internal/adaptadores/entrada/http/servidor.go`:44 |
| DELETE | `/v1/opt-outs` | `api/openapi.yaml` (contrato) |
| DELETE | `/v1/opt-outs` | `internal/adaptadores/entrada/http/servidor.go`:46 |
| POST | `/v1/opt-outs` | `api/openapi.yaml` (contrato) |
| POST | `/v1/opt-outs` | `internal/adaptadores/entrada/http/servidor.go`:45 |
| POST | `/webhooks/sendgrid` | `internal/adaptadores/entrada/webhooks/webhooks.go`:61 |
| POST | `/webhooks/twilio` | `internal/adaptadores/entrada/webhooks/webhooks.go`:62 |
| GET | `/webhooks/whatsapp` | `internal/adaptadores/entrada/webhooks/webhooks.go`:63 |
| POST | `/webhooks/whatsapp` | `internal/adaptadores/entrada/webhooks/webhooks.go`:64 |

## Chamadas HTTP de saída (2)

| Método | URL | Onde |
|---|---|---|
| POST | `/protocol/openid-connect/token` | `internal/adaptadores/saida/keycloak/token.go:17` |
| POST | `/v3/mail/send` | `internal/adaptadores/saida/sendgrid/sendgrid.go:78` |

## Variáveis de ambiente lidas no código (18)

| Variável | Declarada em .env* | Onde é lida |
|---|---|---|
| `COMPOSE` | **não** | `deploy/local/bootstrap.sh` |
| `HOST` | **não** | `deploy/aviva2/publicar.sh` |
| `KEY` | **não** | `deploy/aviva2/publicar.sh` |
| `SENDGRID_API_KEY` | **não** | `deploy/local/bootstrap.sh` |
| `SENDGRID_REMETENTE` | **não** | `deploy/local/bootstrap.sh` |
| `SUFIXO` | **não** | `deploy/aviva2/keycloak/criar-realm.sh`, `deploy/aviva2/publicar.sh` |
| `TWILIO_ACCOUNT_SID` | **não** | `deploy/local/bootstrap.sh` |
| `TWILIO_API_KEY_SECRET` | **não** | `deploy/local/bootstrap.sh` |
| `TWILIO_API_KEY_SID` | **não** | `deploy/local/bootstrap.sh` |
| `TWILIO_MESSAGING_SERVICE_SID` | **não** | `deploy/local/bootstrap.sh` |
| `ULTRAMSG_INSTANCE_ID` | **não** | `deploy/local/bootstrap.sh` |
| `ULTRAMSG_TOKEN` | **não** | `deploy/local/bootstrap.sh` |
| `WHATSAPP_PHONE_NUMBER_ID` | **não** | `deploy/local/bootstrap.sh` |
| `WHATSAPP_PROVEDOR` | **não** | `deploy/local/bootstrap.sh` |
| `WHATSAPP_TOKEN` | **não** | `deploy/local/bootstrap.sh` |
| `ZAPI_CLIENT_TOKEN` | **não** | `deploy/local/bootstrap.sh` |
| `ZAPI_INSTANCE_ID` | **não** | `deploy/local/bootstrap.sh` |
| `ZAPI_TOKEN` | **não** | `deploy/local/bootstrap.sh` |

## Dados

- Migrations: `migrations`
- Schemas/modelos: —
- Arquivos .sql: 11

## Infra

- docker: —
- compose: `deploy/local/compose.yaml`
- nomad: —
- ci: `.github/workflows/dependabot-automerge.yml`, `.github/workflows/go.yml`, `.github/workflows/sdd-checks.yml`
