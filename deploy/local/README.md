# Ambiente local

Postgres, Redis, OpenBao (modo dev) e Keycloak (`start-dev`) via
`podman compose`; o binário roda no host. Só desenvolvimento — hml e prod
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

Sem chave SendGrid real, a Tentativa termina `falhou` com `sendgrid 401`.
Para enviar de verdade: `SENDGRID_API_KEY=... SENDGRID_REMETENTE=... ./deploy/local/bootstrap.sh`.

## Variáveis

| Variável | Padrão | Uso |
|---|---|---|
| `OMINICORE_ADDR` | `0.0.0.0:8080` | endereço da API |
| `OMINICORE_BAO_ADDR` | — | OpenBao (obrigatória) |
| `OMINICORE_BAO_TOKEN` | — | **só local**; sem ela, login `jwt` (ADR 0002) |
| `OMINICORE_BAO_MOUNT` | `kv` | mount KV v2 |
| `OMINICORE_BAO_ROLE` | `ominicore` | role do auth `jwt` |
| `OMINICORE_KC_ISSUER` | — | realm do Keycloak (`iss` esperado) |
| `OMINICORE_KC_AUDIENCE` | `ominicore` | público exigido no token |
| `OMINICORE_KC_CLIENT_ID` / `_SECRET` | `ominicore-app` / — | token do próprio OminiCore para o OpenBao |
| `OMINICORE_REDIS_ADDR` | `redis:6379` | cache do JWKS |
| `OMINICORE_SENDGRID_URL` | `https://api.sendgrid.com` | trocar só em teste |

Segredos no OpenBao: `kv/ominicore/postgres` (`dsn`) e `kv/ominicore/sendgrid`
(`api_key`, `remetente`).
