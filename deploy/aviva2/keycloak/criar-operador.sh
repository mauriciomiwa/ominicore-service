#!/usr/bin/env bash
# Cria uma pessoa Operadora (RN-12) no realm do ambiente: papel
# ominicore-config, configurar o OTP e trocar a senha no primeiro login. Mostra
# a senha temporária UMA vez, para entregar por canal seguro. Uma conta por
# pessoa: a troca registra quem fez (REQ-27). Para tirar o acesso, desabilite
# a pessoa no Keycloak. Idempotente.
# Quem roda: o dono, no host, como root.
# Uso: sudo ./criar-operador.sh hml|prod <usuario> <email>
set -euo pipefail
cd "$(dirname "$0")"
case "${1:-}" in hml|prod) ;; *) echo "uso: $0 hml|prod <usuario> <email>" >&2; exit 2 ;; esac
USUARIO="${2:?informe o usuário}"
EMAIL="${3:?informe o e-mail}"
[[ "$USUARIO" =~ ^[a-z][a-z0-9.-]{1,40}$ ]] || { echo "usuário inválido" >&2; exit 2; }
[[ "$EMAIL" =~ ^[^@[:space:]]+@[^@[:space:]]+\.[a-z]{2,}$ ]] || { echo "e-mail inválido" >&2; exit 2; }
# shellcheck source=/dev/null
. "../ambientes/$1.sh"
# shellcheck source=lib-console.sh
. ./lib-console.sh
# shellcheck source=lib-admin.sh
. ./lib-admin.sh
entrar_admin_keycloak
criar_operador "$REALM" "$USUARIO" "$EMAIL"
