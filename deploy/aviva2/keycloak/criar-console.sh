#!/usr/bin/env bash
# Cria ou atualiza o client público ominicore-console (Console do Operador,
# RN-12) no realm do ambiente: Authorization Code + PKCE S256, sem segredo,
# redirect e web origin só de https://$HOST_CONSOLE, audiência ominicore e o
# fluxo navegador-otp (OTP obrigatório) como override do login. Idempotente.
# Quem roda: o dono, no host, como root. Uso: sudo ./criar-console.sh hml|prod
set -euo pipefail
cd "$(dirname "$0")"
case "${1:-}" in hml|prod) ;; *) echo "uso: $0 hml|prod" >&2; exit 2 ;; esac
# shellcheck source=/dev/null
. "../ambientes/$1.sh"
# shellcheck source=lib-console.sh
. ./lib-console.sh
read -r -s -p "senha do admin do Keycloak (realm master): " KC_ADMIN; echo
KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh)
"${KC[@]}" config credentials --server http://localhost:8080 --realm master --user admin --password "$KC_ADMIN" >/dev/null
unset KC_ADMIN
configurar_console "$REALM" "https://$HOST_CONSOLE"
