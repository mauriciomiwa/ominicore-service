#!/usr/bin/env bash
# Grava no OpenBao LOCAL (modo dev) os segredos que o OminiCore lê no boot.
# Só para desenvolvimento: valores de exemplo. Migrations: `ominicore migrate`.
# Para enviar e-mail de verdade, troque api_key/remetente por uma conta
# SendGrid de teste — nunca versione a chave real.
set -euo pipefail
cd "$(dirname "$0")"

# COMPOSE="docker compose" para usar o Docker no lugar do Podman
read -r -a COMPOSE_CMD <<<"${COMPOSE:-podman compose}"
BAO=("${COMPOSE_CMD[@]}" -f compose.yaml exec -T -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN=dev-root openbao bao)

echo ">> esperando o OpenBao"
until "${BAO[@]}" status >/dev/null 2>&1; do sleep 1; done

echo ">> segredos em kv/ominicore/*"
# o modo dev já monta um KV v2 em secret/; o OminiCore usa o mount kv/ como hml/prod
"${BAO[@]}" secrets enable -path=kv kv-v2 >/dev/null 2>&1 || true
"${BAO[@]}" kv put kv/ominicore/postgres \
  dsn="postgres://ominicore:ominicore-local@127.0.0.1:5432/ominicore?sslmode=disable" >/dev/null
"${BAO[@]}" kv put kv/ominicore/sendgrid \
  api_key="${SENDGRID_API_KEY:-SG.local-invalida}" remetente="${SENDGRID_REMETENTE:-no-reply@localhost}" >/dev/null
# Twilio: API Key + Messaging Service (fatia 2). Valores inválidos por padrão.
"${BAO[@]}" kv put kv/ominicore/twilio \
  account_sid="${TWILIO_ACCOUNT_SID:-AClocal}" api_key_sid="${TWILIO_API_KEY_SID:-SKlocal}" \
  api_key_secret="${TWILIO_API_KEY_SECRET:-local-invalida}" \
  messaging_service_sid="${TWILIO_MESSAGING_SERVICE_SID:-MGlocal}" >/dev/null
# WhatsApp (fatias 6, 8 e 9): Meta por padrão, com token de System User +
# phone_number_id inválidos; WHATSAPP_PROVEDOR=ultramsg ou zapi grava instância
# + token (na Z-API, ZAPI_CLIENT_TOKEN opcional).
if [ "${WHATSAPP_PROVEDOR:-meta}" = ultramsg ]; then
  "${BAO[@]}" kv put kv/ominicore/whatsapp provedor=ultramsg \
    instance_id="${ULTRAMSG_INSTANCE_ID:-instance0}" token="${ULTRAMSG_TOKEN:-local-invalido}" >/dev/null
elif [ "${WHATSAPP_PROVEDOR:-meta}" = zapi ]; then
  ZAPI_CAMPOS=(provedor=zapi instance_id="${ZAPI_INSTANCE_ID:-instancia0}" token="${ZAPI_TOKEN:-local-invalido}")
  if [ -n "${ZAPI_CLIENT_TOKEN:-}" ]; then ZAPI_CAMPOS+=(client_token="$ZAPI_CLIENT_TOKEN"); fi
  "${BAO[@]}" kv put kv/ominicore/whatsapp "${ZAPI_CAMPOS[@]}" >/dev/null
else
  "${BAO[@]}" kv put kv/ominicore/whatsapp provedor=meta \
    token="${WHATSAPP_TOKEN:-local-invalido}" phone_number_id="${WHATSAPP_PHONE_NUMBER_ID:-000000000}" >/dev/null
fi

# Webhooks: chave pública ECDSA de verdade (o boot recusa chave inválida) e
# segredos de exemplo; localmente os Provedores não alcançam o OminiCore.
SG_PUB=$(openssl ecparam -name prime256v1 -genkey -noout 2>/dev/null | openssl ec -pubout 2>/dev/null | grep -v PUBLIC | tr -d '\n')
"${BAO[@]}" kv put kv/ominicore/webhooks sendgrid_chave_publica="$SG_PUB" \
  twilio_auth_token="local-invalido" meta_app_secret="local-invalido" meta_verify_token="local-verify" >/dev/null

# Opt-out: chave do HMAC dos contatos (só local; em hml/prod é permanente).
"${BAO[@]}" kv get kv/ominicore/optout >/dev/null 2>&1 || \
  "${BAO[@]}" kv put kv/ominicore/optout chave="$(openssl rand -hex 32)" >/dev/null

echo ">> auth jwt com o Keycloak local (ADR 0002 — o mesmo desenho de hml/prod)"
"${BAO[@]}" auth enable jwt >/dev/null 2>&1 || true
# JWKS pelo backchannel (rede do compose); o iss dos tokens é o host público localhost:8081
# o OpenBao valida o jwks_url ao gravar: espera o Keycloak terminar de subir
for tentativa in $(seq 1 60); do
  "${BAO[@]}" write auth/jwt/config \
    jwks_url=http://keycloak:8080/realms/ominicore/protocol/openid-connect/certs \
    bound_issuer=http://localhost:8081/realms/ominicore >/dev/null 2>&1 && break
  [ "$tentativa" = 60 ] && { echo "Keycloak não respondeu o JWKS em 120 s" >&2; exit 1; }
  sleep 2
done
# leitura de tudo; escrita só na configuração do WhatsApp (ADR 0005, RN-11)
printf 'path "kv/data/ominicore/*" { capabilities = ["read"] }\npath "kv/data/ominicore/whatsapp" { capabilities = ["create", "update", "read", "delete"] }\n' |
  "${BAO[@]}" policy write ominicore - >/dev/null
# bound_claims é um mapa: só entra como JSON pelo stdin
"${BAO[@]}" write auth/jwt/role/ominicore - >/dev/null <<'JSON'
{"role_type": "jwt", "user_claim": "azp", "bound_claims": {"azp": "ominicore-app"},
 "bound_audiences": ["openbao"], "token_policies": ["ominicore"], "token_ttl": "1h"}
JSON

echo ">> Console do Operador no Keycloak local (RN-12): client, OTP obrigatório e a pessoa operador.local"
# a mesma biblioteca de hml/prod, com o kcadm do container
# shellcheck source=../aviva2/keycloak/lib-console.sh
. ../aviva2/keycloak/lib-console.sh
KC=("${COMPOSE_CMD[@]}" -f compose.yaml exec -T keycloak /opt/keycloak/bin/kcadm.sh)
"${KC[@]}" config credentials --server http://localhost:8080 --realm master --user admin --password admin-local >/dev/null
configurar_console ominicore http://localhost:5173
# a senha temporária sai uma vez; no primeiro login a pessoa troca a senha e configura o OTP
criar_operador ominicore operador.local operador@localhost

echo ">> pronto. Variáveis para rodar o binário no host:"
cat <<'EOF'
export OMINICORE_BAO_ADDR=http://127.0.0.1:8200
export OMINICORE_BAO_TOKEN=dev-root          # só local; hml/prod usam JWT (ADR 0002)
export OMINICORE_KC_ISSUER=http://localhost:8081/realms/ominicore
export OMINICORE_REDIS_ADDR=127.0.0.1:6380
export OMINICORE_CONSOLE_ORIGEM=http://localhost:5173   # Console em `bun run dev` (web/console)
# para testar o login jwt (caminho de hml/prod), troque o token por:
#   unset OMINICORE_BAO_TOKEN; export OMINICORE_KC_CLIENT_SECRET=ominicore-app-secret
EOF
