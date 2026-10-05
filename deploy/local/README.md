# Ambiente local

Postgres, Redis, OpenBao (modo dev) e Keycloak (`start-dev`) via
`podman compose` (ou Docker: `COMPOSE="docker compose"` antes do `make` e do
`bootstrap.sh`); o binário roda no host. Só desenvolvimento — hml e prod
seguem a skill `ambientes` (`DOC/Infra/`).

```bash
make compose-up                      # sobe os 4 serviços e grava os segredos no OpenBao
eval "$(./deploy/local/bootstrap.sh | grep '^export')"   # idempotente
go run ./cmd/ominicore migrate
go run ./cmd/ominicore api           # :8080
go run ./cmd/ominicore worker        # em outro terminal
```

Pedir uma Notificação com o Sistema de origem de teste (`sistema-local`):

```bash
TOKEN=$(curl -s -d grant_type=client_credentials -d client_id=sistema-local \
  -d client_secret=sistema-local-secret \
  http://localhost:8081/realms/ominicore/protocol/openid-connect/token | jq -r .access_token)
curl -i -X POST http://localhost:8080/v1/notificacoes \
  -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: teste-1" \
  -d '{"template_id":"boas-vindas","destinatario":{"email":"voce@exemplo.com"},"dados":{"nome":"Você"}}'
```

Sem chave SendGrid real, a Tentativa termina `falhou` com `sendgrid 401` e o
fallback tenta o SMS (se o Destinatário tiver `telefone`), que falha com
`twilio 401`. Para enviar de verdade: `SENDGRID_API_KEY=... SENDGRID_REMETENTE=...`
e/ou `TWILIO_ACCOUNT_SID=... TWILIO_API_KEY_SID=... TWILIO_API_KEY_SECRET=...
TWILIO_MESSAGING_SERVICE_SID=...` antes do `./deploy/local/bootstrap.sh`.

## Variáveis

| Variável | Padrão | Uso |
|---|---|---|
| `OMINICORE_ADDR` | `0.0.0.0:8080` | endereço da API |
| `OMINICORE_BAO_ADDR` | — | OpenBao (obrigatória) |
| `OMINICORE_BAO_TOKEN` | — | **só local**; sem ela, login `jwt` (ADR 0002) |
| `OMINICORE_BAO_MOUNT` | `kv` | mount KV v2 |
| `OMINICORE_BAO_ROLE` | `ominicore` | role do auth `jwt` |
| `OMINICORE_BAO_PREFIXO` | `ominicore` | prefixo dos segredos (`kv/<prefixo>/*`); hml = `ominicore-hml` |
| `OMINICORE_KC_ISSUER` | — | realm do Keycloak (`iss` esperado) |
| `OMINICORE_KC_URL` | — (usa o emissor) | backchannel do Keycloak para JWKS e token (`http://keycloak:8080` no aviva2.0) |
| `OMINICORE_KC_AUDIENCE` | `ominicore` | público exigido no token |
| `OMINICORE_KC_CLIENT_ID` / `_SECRET` | `ominicore-app` / — | token do próprio OminiCore para o OpenBao |
| `OMINICORE_REDIS_ADDR` | `redis:6379` | cache do JWKS |
| `OMINICORE_REDIS_PASSWORD` | — | senha do Redis (podman secret `redis-password` no aviva2.0) |
| `OMINICORE_DB_SCHEMA` | `ominicore` | schema do Postgres; hml = `ominicore_hml` |
| `OMINICORE_BAO_AUTH_MOUNT` | `jwt` | mount do auth `jwt`; hml = `jwt-ominicore-hml`, prod = `jwt-ominicore` |
| `OMINICORE_SENDGRID_URL` | `https://api.sendgrid.com` | trocar só em teste |
| `OMINICORE_TWILIO_URL` | `https://api.twilio.com` | trocar só em teste |
| `OMINICORE_WEBHOOK_URL` | — | origem pública dos webhooks (`https://notify-hml.flipapi.com.br`); vazia no local: SMS sem `StatusCallback` e webhook do Twilio recusado |
| `OMINICORE_RETENCAO_DIAS` | `30` | prazo de retenção: Notificação finalizada é anonimizada depois disso (RN-9) |
| `OMINICORE_WHATSAPP_URL` | `https://graph.facebook.com` | trocar só em teste (versão da Graph API fixada no adaptador) |

Segredos no OpenBao: `kv/<prefixo>/postgres` (`dsn`), `kv/<prefixo>/sendgrid`
(`api_key`, `remetente`), `kv/<prefixo>/twilio` (`account_sid`, `api_key_sid`,
`api_key_secret`, `messaging_service_sid`), `kv/<prefixo>/whatsapp` (`token`,
`phone_number_id`; opcional — sem ele o Canal WhatsApp fica desligado) e
`kv/<prefixo>/webhooks` (`sendgrid_chave_publica`, `twilio_auth_token` e, com o
WhatsApp ligado, `meta_app_secret`, `meta_verify_token`) e `kv/<prefixo>/optout`
(`chave` do HMAC dos Opt-outs — permanente, ver `deploy/aviva2/README.md`). Endereços de hml e prod: `DOC/Infra/AMBIENTES.md`.
