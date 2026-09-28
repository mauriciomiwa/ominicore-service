#!/usr/bin/env bash
# Grava no OpenBao LOCAL (modo dev) os segredos que o OminiCore lê no boot.
# Só para desenvolvimento: valores de exemplo. Migrations: `ominicore migrate`.
# Para enviar e-mail de verdade, troque api_key/remetente por uma conta
# SendGrid de teste — nunca versione a chave real.
set -euo pipefail
cd "$(dirname "$0")"

BAO=(podman compose -f compose.yaml exec -T -e BAO_ADDR=http://127.0.0.1:8200 -e BAO_TOKEN=dev-root openbao bao)

echo ">> esperando o OpenBao"
until "${BAO[@]}" status >/dev/null 2>&1; do sleep 1; done

echo ">> segredos em kv/ominicore/*"
# o modo dev já monta um KV v2 em secret/; o OminiCore usa o mount kv/ como hml/prod
"${BAO[@]}" secrets enable -path=kv kv-v2 >/dev/null 2>&1 || true
"${BAO[@]}" kv put kv/ominicore/postgres \
  dsn="postgres://ominicore:ominicore-local@127.0.0.1:5432/ominicore?sslmode=disable" >/dev/null
"${BAO[@]}" kv put kv/ominicore/sendgrid \
  api_key="${SENDGRID_API_KEY:-SG.local-invalida}" remetente="${SENDGRID_REMETENTE:-no-reply@localhost}" >/dev/null

echo ">> pronto. Variáveis para rodar o binário no host:"
cat <<'EOF'
export OMINICORE_BAO_ADDR=http://127.0.0.1:8200
export OMINICORE_BAO_TOKEN=dev-root          # só local; hml/prod usam JWT (ADR 0002)
export OMINICORE_KC_ISSUER=http://localhost:8081/realms/ominicore
export OMINICORE_REDIS_ADDR=127.0.0.1:6379
EOF
