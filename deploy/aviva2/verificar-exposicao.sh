#!/usr/bin/env bash
# Verifica, DE FORA do escritório, que o Console do Operador e a configuração
# não respondem e que o resto da API segue sem rota pública (RN-12, REQ-33):
#   T-113  https://$HOST_CONSOLE                    -> 403 da Cloudflare (WAF)
#   T-114  https://$HOST_WEBHOOK/v1/config/whatsapp -> 403 da Cloudflare (WAF)
#   T-115  https://$HOST_WEBHOOK/v1/notificacoes e /v1/opt-outs -> 403 ou 404
# Roda no GitHub Actions (console.yml), que nunca está no IP do escritório:
# depois de cada deploy do Console e uma vez por dia. Rodar do escritório
# falha de propósito: lá o WAF deixa passar.
# Uso: ./verificar-exposicao.sh hml|prod
set -euo pipefail
cd "$(dirname "$0")"
case "${1:-}" in hml|prod) ;; *) echo "uso: $0 hml|prod" >&2; exit 2 ;; esac
# shellcheck source=/dev/null
. "ambientes/$1.sh"

FALHAS=0
# status <url> [método]: código HTTP e, separado por espaço, o cabeçalho server
status() {
  curl -sS -o /dev/null -X "${2:-GET}" --max-time 20 -w '%{http_code} %header{server}' "$1" || echo "000 -"
}

bloqueado_pela_cloudflare() { # T-n url
  local r
  r=$(status "$2")
  if [[ "$r" == "403 cloudflare" ]]; then
    echo "ok   $1 $2 -> $r"
  else
    echo "FURO $1 $2 -> $r (esperado 403 do WAF da Cloudflare)"
    FALHAS=$((FALHAS + 1))
  fi
}

sem_rota() { # T-n url método
  local r
  r=$(status "$2" "$3")
  if [[ "${r%% *}" == 403 || "${r%% *}" == 404 ]]; then
    echo "ok   $1 $3 $2 -> $r"
  else
    echo "FURO $1 $3 $2 -> $r (esperado 403 ou 404: a API de envio não é pública)"
    FALHAS=$((FALHAS + 1))
  fi
}

echo ">> $AMBIENTE, de $(curl -sS --max-time 10 https://api.ipify.org 2>/dev/null || echo 'IP desconhecido')"
bloqueado_pela_cloudflare T-113 "https://$HOST_CONSOLE/"
bloqueado_pela_cloudflare T-113 "https://$HOST_CONSOLE/callback/$AMBIENTE"
bloqueado_pela_cloudflare T-114 "https://$HOST_WEBHOOK/v1/config/whatsapp"
sem_rota T-115 "https://$HOST_WEBHOOK/v1/notificacoes" POST
sem_rota T-115 "https://$HOST_WEBHOOK/v1/notificacoes/00000000-0000-0000-0000-000000000000" GET
sem_rota T-115 "https://$HOST_WEBHOOK/v1/opt-outs" POST
sem_rota T-115 "https://$HOST_WEBHOOK/v1/opt-outs" DELETE

[ "$FALHAS" -eq 0 ] || { echo "!! $FALHAS furo(s) de exposição em $AMBIENTE" >&2; exit 1; }
echo ">> exposição de $AMBIENTE conforme a ADR 0006"
