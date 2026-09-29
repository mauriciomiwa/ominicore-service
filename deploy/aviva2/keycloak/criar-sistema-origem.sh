#!/usr/bin/env bash
# Cria um Sistema de origem (client confidencial com service account e público
# "ominicore") no realm do ambiente e mostra o client secret UMA vez, para você
# entregar ao time do sistema por canal seguro.
# Quem roda: o dono, no host, como root. Uso: sudo ./criar-sistema-origem.sh hml|prod <client-id>
set -euo pipefail
cd "$(dirname "$0")"
case "${1:-}" in hml|prod) ;; *) echo "uso: $0 hml|prod <client-id>" >&2; exit 2 ;; esac
CLIENT="${2:?informe o client-id do Sistema de origem}"
[[ "$CLIENT" =~ ^[a-z][a-z0-9-]{2,40}$ ]] || { echo "client-id inválido" >&2; exit 2; }
# shellcheck source=/dev/null
. "../ambientes/$1.sh"
read -r -s -p "senha do admin do Keycloak (realm master): " KC_ADMIN; echo
KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh)
"${KC[@]}" config credentials --server http://localhost:8080 --realm master --user admin --password "$KC_ADMIN" >/dev/null
unset KC_ADMIN
"${KC[@]}" create clients -r "$REALM" -f - <<JSON
{"clientId": "$CLIENT", "enabled": true, "protocol": "openid-connect", "publicClient": false,
 "clientAuthenticatorType": "client-secret", "serviceAccountsEnabled": true,
 "standardFlowEnabled": false, "directAccessGrantsEnabled": false,
 "protocolMappers": [{"name": "audience-ominicore", "protocol": "openid-connect", "protocolMapper": "oidc-audience-mapper",
   "config": {"included.custom.audience": "ominicore", "access.token.claim": "true", "id.token.claim": "false"}}]}
JSON
ID=$("${KC[@]}" get clients -r "$REALM" -q clientId="$CLIENT" --fields id --format csv --noquotes)
echo ">> $CLIENT criado em $REALM. Client secret (entregue por canal seguro; não fica salvo aqui):"
"${KC[@]}" get "clients/$ID/client-secret" -r "$REALM" --fields value --format csv --noquotes
