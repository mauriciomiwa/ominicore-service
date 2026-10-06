#!/usr/bin/env bash
# Cria o realm do OminiCore no Keycloak do aviva2.0 pela API de admin (kcadm),
# SEM reiniciar o Keycloak (o --import-realm só roda no boot e derrubaria o
# login do aviva-back). Cria o client ominicore-app (login no OpenBao, ADR 0002)
# e grava o secret dele direto no podman secret ominicore<sufixo>-kc-secret.
# Quem roda: o dono, no host, como root. Pede a senha de admin do Keycloak.
# Uso: sudo ./criar-realm.sh hml|prod
set -euo pipefail
cd "$(dirname "$0")"
case "${1:-}" in hml|prod) ;; *) echo "uso: $0 hml|prod" >&2; exit 2 ;; esac
# shellcheck source=/dev/null
. "../ambientes/$1.sh"

# shellcheck source=lib-admin.sh
. ./lib-admin.sh
entrar_admin_keycloak

if "${KC[@]}" get "realms/$REALM" >/dev/null 2>&1; then
  echo ">> realm $REALM já existe — nada criado"
else
  "${KC[@]}" create realms -f - <<JSON
{"realm": "$REALM", "displayName": "OminiCore ($AMBIENTE)", "enabled": true, "accessTokenLifespan": 300}
JSON
  echo ">> realm $REALM criado"
fi

if [ -z "$("${KC[@]}" get clients -r "$REALM" -q clientId=ominicore-app --fields id --format csv --noquotes)" ]; then
  "${KC[@]}" create clients -r "$REALM" -f - <<'JSON'
{"clientId": "ominicore-app", "name": "OminiCore (login no OpenBao)", "enabled": true,
 "protocol": "openid-connect", "publicClient": false, "clientAuthenticatorType": "client-secret",
 "serviceAccountsEnabled": true, "standardFlowEnabled": false, "directAccessGrantsEnabled": false,
 "protocolMappers": [{"name": "audience-openbao", "protocol": "openid-connect", "protocolMapper": "oidc-audience-mapper",
   "config": {"included.custom.audience": "openbao", "access.token.claim": "true", "id.token.claim": "false"}}]}
JSON
  echo ">> client ominicore-app criado"
fi

ID=$("${KC[@]}" get clients -r "$REALM" -q clientId=ominicore-app --fields id --format csv --noquotes)
if podman secret exists "ominicore${SUFIXO}-kc-secret"; then
  echo ">> podman secret ominicore${SUFIXO}-kc-secret já existe — mantido"
else
  "${KC[@]}" get "clients/$ID/client-secret" -r "$REALM" --fields value --format csv --noquotes \
    | tr -d '\n' | podman secret create "ominicore${SUFIXO}-kc-secret" - >/dev/null
  echo ">> podman secret ominicore${SUFIXO}-kc-secret criado (valor não exibido)"
fi
