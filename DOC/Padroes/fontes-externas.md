# Fontes externas confiáveis

Doc **vivo**: adicione a fonte oficial de cada tecnologia adotada. A IA
consulta esta lista **antes** de busca genérica na web e usa a forma mais
econômica em tokens: manifesto para versões, deep-link oficial para
API/comportamento, uma query precisa.

| Tecnologia | Fonte oficial | Observação |
|---|---|---|
| Go | https://go.dev/doc/ · https://pkg.go.dev/std | release notes em https://go.dev/doc/devel/release |
| pgx | https://pkg.go.dev/github.com/jackc/pgx/v5 | |
| sqlc | https://docs.sqlc.dev | |
| goose | https://pressly.github.io/goose/ | |
| go-redis | https://redis.uptrace.dev · https://pkg.go.dev/github.com/redis/go-redis/v9 | |
| OpenBao | https://openbao.org/docs/ | auth `jwt`: https://openbao.org/docs/auth/jwt/ · KV v2: https://openbao.org/docs/secrets/kv/kv-v2/ |
| Keycloak | https://www.keycloak.org/documentation | service accounts / `client_credentials` |
| PostgreSQL | https://www.postgresql.org/docs/current/ | `SKIP LOCKED`: seção SELECT, "The Locking Clause" |
| SendGrid | https://www.twilio.com/docs/sendgrid | Event Webhook com assinatura ECDSA |
| WhatsApp Cloud API (Meta) | https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages · erros: https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes | versão fixada `v26.0` (conferida 2026-09-29; https://developers.facebook.com/docs/graph-api/changelog/versions) |
| Twilio SMS | https://www.twilio.com/docs/messaging | validação `X-Twilio-Signature` |
| testcontainers-go | https://golang.testcontainers.org | |
| golangci-lint | https://golangci-lint.run | |
| Podman / quadlet | https://docs.podman.io · `man podman-systemd.unit` | |
| OpenAPI | https://spec.openapis.org/oas/v3.1.0 | |
| Infra da casa (aviva2.0) | `../flip/infra/aviva2.0/` e `../flip/aviva-back/deploy/aviva2/` | molde de quadlet, Traefik, Keycloak e OpenBao |
