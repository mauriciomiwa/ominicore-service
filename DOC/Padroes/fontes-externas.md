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
| SendGrid | https://www.twilio.com/docs/sendgrid · eventos: https://www.twilio.com/docs/sendgrid/for-developers/tracking-events/event · assinatura: https://www.twilio.com/docs/sendgrid/for-developers/tracking-events/getting-started-event-webhook-security-features | ECDSA/SHA-256 sobre timestamp + corpo cru; formato da chave e relação `sg_message_id` × `X-Message-Id` não documentados na página (suposição: prefixo antes do 1º `.`, validar em hml) |
| Webhooks da Meta | https://developers.facebook.com/docs/graph-api/webhooks/getting-started · status: https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks/components | `hub.challenge`; `X-Hub-Signature-256: sha256=<hex>` (App Secret); reenvio por 36 h |
| WhatsApp Cloud API (Meta) | https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages · erros: https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes | versão fixada `v26.0` (conferida 2026-09-29; https://developers.facebook.com/docs/graph-api/changelog/versions) |
| UltraMsg (WhatsApp, gateway não oficial) | https://docs.ultramsg.com/api/post/messages/chat | `POST https://api.ultramsg.com/{instance_id}/messages/chat`, form `token`/`to`/`body` (até 4096 caracteres); instância desconectada enfileira; sem exemplo de resposta nem assinatura de webhook na doc (conferida 2026-10-05) — formato da resposta a confirmar com a instância de trial |
| Twilio SMS | https://www.twilio.com/docs/messaging · status: https://www.twilio.com/docs/messaging/guides/track-outbound-message-status · assinatura: https://www.twilio.com/docs/usage/webhooks/webhooks-security | `X-Twilio-Signature` = HMAC-SHA1 (Auth Token, não API Key) sobre URL + parâmetros ordenados |
| testcontainers-go | https://golang.testcontainers.org | |
| golangci-lint | https://golangci-lint.run | |
| Podman / quadlet | https://docs.podman.io · `man podman-systemd.unit` | |
| OpenAPI | https://spec.openapis.org/oas/v3.1.0 | |
| Infra da casa (aviva2.0) | `../flip/infra/aviva2.0/` e `../flip/aviva-back/deploy/aviva2/` | molde de quadlet, Traefik, Keycloak e OpenBao |
