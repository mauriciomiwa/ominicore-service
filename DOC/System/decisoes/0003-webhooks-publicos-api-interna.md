# Webhooks públicos, API de envio só na rede interna

Status: aceita (2026-09-28); revisada pela 0006 (`/v1/config/*` público só do escritório)

O OminiCore é interno, mas SendGrid e Twilio precisam alcançar os webhooks de
status pela internet. Só `/webhooks/*` ganha rota pública no Traefik (atrás da
Cloudflare), e cada requisição tem a assinatura do Provedor validada
(Twilio `X-Twilio-Signature`, SendGrid Event Webhook ECDSA) antes de qualquer
efeito. A API de envio (`/v1/*`) não tem rota pública: só é alcançável pela rede
`aviva` do Podman, e ainda exige JWT do Keycloak (zero trust).
