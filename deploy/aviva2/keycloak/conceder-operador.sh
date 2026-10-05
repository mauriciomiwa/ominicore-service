#!/usr/bin/env bash
# Concede o papel de Operador (realm role ominicore-config, RN-11) ao service
# account de um client já existente: com ele, o client consulta e troca a
# configuração do WhatsApp em /v1/config/whatsapp. Cria o papel se faltar
# (idempotente). Nunca conceda a um Sistema de origem comum: quem tem o papel
# troca o número do WhatsApp de todos.
# Quem roda: o dono, no host, como root.
# Uso: sudo ./conceder-operador.sh hml|prod <client-id> [--revogar]
set -euo pipefail
cd "$(dirname "$0")"
USO="uso: $0 hml|prod <client-id> [--revogar]"
case "${1:-}" in hml|prod) ;; *) echo "$USO" >&2; exit 2 ;; esac
CLIENT="${2:?informe o client-id}"
[[ "$CLIENT" =~ ^[a-z][a-z0-9-]{2,40}$ ]] || { echo "client-id inválido" >&2; exit 2; }
ACAO="${3:-}"
case "$ACAO" in ""|--revogar) ;; *) echo "$USO" >&2; exit 2 ;; esac
# shellcheck source=/dev/null
. "../ambientes/$1.sh"
PAPEL=ominicore-config
read -r -s -p "senha do admin do Keycloak (realm master): " KC_ADMIN; echo
KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh)
"${KC[@]}" config credentials --server http://localhost:8080 --realm master --user admin --password "$KC_ADMIN" >/dev/null
unset KC_ADMIN
"${KC[@]}" get "roles/$PAPEL" -r "$REALM" >/dev/null 2>&1 ||
  "${KC[@]}" create roles -r "$REALM" -s name="$PAPEL" -s 'description=Operador: configura os Canais pela API (RN-11)' >/dev/null
CONTA="service-account-$CLIENT"
"${KC[@]}" get users -r "$REALM" -q username="$CONTA" --fields id --format csv --noquotes | grep -q . ||
  { echo "!! $CLIENT não existe ou não tem service account em $REALM" >&2; exit 1; }
if [ "$ACAO" = --revogar ]; then
  "${KC[@]}" remove-roles -r "$REALM" --uusername "$CONTA" --rolename "$PAPEL"
  echo ">> papel $PAPEL revogado de $CLIENT em $REALM (vale para tokens novos; os atuais expiram em até 5 min)"
else
  "${KC[@]}" add-roles -r "$REALM" --uusername "$CONTA" --rolename "$PAPEL"
  echo ">> papel $PAPEL concedido a $CLIENT em $REALM"
fi
