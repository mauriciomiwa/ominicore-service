#!/usr/bin/env bash
# Configura o OpenBao do aviva2.0 para um ambiente do OminiCore (ADR 0002):
# mount KV v2 "kv" (se faltar), auth jwt próprio do realm, policy só-leitura em
# kv/<prefixo>/* e role presa ao azp ominicore-app. Depois pede os segredos
# (DSN, SendGrid, Twilio, webhooks e, se quiser, WhatsApp) sem ecoar e grava em
# kv/<prefixo>/*. Sem kv/<prefixo>/whatsapp o Canal WhatsApp fica desligado
# (REQ-20); --so-whatsapp liga o Canal depois, sem tocar nos outros segredos.
# Quem roda: o dono, no host, como root. Pede um token de admin do OpenBao.
# Uso: sudo ./configurar-openbao.sh hml|prod [--sem-segredos|--so-whatsapp]
set -euo pipefail
cd "$(dirname "$0")"
USO="uso: $0 hml|prod [--sem-segredos|--so-whatsapp]"
case "${1:-}" in hml|prod) ;; *) echo "$USO" >&2; exit 2 ;; esac
MODO="${2:-}"
case "$MODO" in ""|--sem-segredos|--so-whatsapp) ;; *) echo "$USO" >&2; exit 2 ;; esac
# shellcheck source=/dev/null
. "../ambientes/$1.sh"
read -r -s -p "token de admin do OpenBao: " BAO_TOKEN; echo
export BAO_TOKEN
BAO=(podman exec -i -e BAO_TOKEN -e BAO_ADDR=http://127.0.0.1:8200 openbao bao)

"${BAO[@]}" secrets list -format=json | grep -q '"kv/"' || "${BAO[@]}" secrets enable -path=kv kv-v2
"${BAO[@]}" auth list -format=json | grep -q "\"$AUTH_MOUNT/\"" || "${BAO[@]}" auth enable -path="$AUTH_MOUNT" jwt
# JWKS pelo backchannel da rede aviva; o iss dos tokens é o host público.
"${BAO[@]}" write "auth/$AUTH_MOUNT/config" \
  jwks_url="http://keycloak:8080/realms/$REALM/protocol/openid-connect/certs" \
  bound_issuer="https://auth.flipapi.com.br/realms/$REALM" >/dev/null
printf 'path "kv/data/%s/*" { capabilities = ["read"] }\n' "$PREFIXO" | "${BAO[@]}" policy write "$ROLE" - >/dev/null
"${BAO[@]}" write "auth/$AUTH_MOUNT/role/$ROLE" - >/dev/null <<JSON
{"role_type": "jwt", "user_claim": "azp", "bound_claims": {"azp": "ominicore-app"},
 "bound_audiences": ["openbao"], "token_policies": ["$ROLE"], "token_ttl": "1h"}
JSON
echo ">> auth $AUTH_MOUNT, policy e role $ROLE prontos"

# Chave do HMAC dos Opt-outs (REQ-18): gerada aqui uma única vez e NUNCA
# sobrescrita — trocá-la faz todos os Opt-outs pararem de valer.
if "${BAO[@]}" kv get "kv/$PREFIXO/optout" >/dev/null 2>&1; then
  echo ">> kv/$PREFIXO/optout já existe — mantido (não rotacionar)"
else
  openssl rand -hex 32 | tr -d '\n' | "${BAO[@]}" kv put "kv/$PREFIXO/optout" chave=- >/dev/null
  echo ">> kv/$PREFIXO/optout criado (chave permanente, valor não exibido)"
fi

ler() { local v; read -r -s -p "$1: " v; echo >&2; printf '%s' "$v"; }
existe() { "${BAO[@]}" kv get "kv/$PREFIXO/$1" >/dev/null 2>&1; }

# WhatsApp: credenciais da Cloud API + os dois segredos do webhook da Meta,
# sempre juntos (caminho incompleto impede o boot, REQ-6).
gravar_whatsapp() {
  local WA_TOKEN WA_ID WH_APP WH_VT
  WA_TOKEN=$(ler "WhatsApp token (System User)"); read -r -p "WhatsApp phone_number_id: " WA_ID
  WH_APP=$(ler "Meta App Secret"); WH_VT=$(ler "Meta verify token (o mesmo que você cadastrar no painel)")
  if [ -z "$WA_TOKEN" ] || [ -z "$WA_ID" ] || [ -z "$WH_APP" ] || [ -z "$WH_VT" ]; then
    echo "!! campo vazio: nada do WhatsApp foi gravado" >&2; return 1
  fi
  # um segredo por comando, sempre pelo stdin (nunca em argumento nem em JSON montado à mão)
  "${BAO[@]}" kv patch "kv/$PREFIXO/webhooks" meta_app_secret=- >/dev/null <<<"$WH_APP"
  "${BAO[@]}" kv patch "kv/$PREFIXO/webhooks" meta_verify_token=- >/dev/null <<<"$WH_VT"
  "${BAO[@]}" kv put "kv/$PREFIXO/whatsapp" token=- phone_number_id="$WA_ID" >/dev/null <<<"$WA_TOKEN"
  echo ">> WhatsApp ligado: kv/$PREFIXO/whatsapp e campos meta_* em kv/$PREFIXO/webhooks (reinicie api e worker)"
}

case "$MODO" in
--so-whatsapp)
  existe webhooks || { echo "!! kv/$PREFIXO/webhooks não existe: rode antes sem --so-whatsapp" >&2; exit 1; }
  gravar_whatsapp
  ;;
"")
  DSN=$(ler "DSN do Postgres ($SCHEMA, com sslrootcert=/etc/ominicore/rds-ca.pem)")
  "${BAO[@]}" kv put "kv/$PREFIXO/postgres" dsn=- >/dev/null <<<"$DSN"; unset DSN
  SG_KEY=$(ler "SendGrid api_key"); read -r -p "SendGrid remetente (e-mail verificado): " SG_DE
  "${BAO[@]}" kv put "kv/$PREFIXO/sendgrid" api_key=- remetente="$SG_DE" >/dev/null <<<"$SG_KEY"; unset SG_KEY
  read -r -p "Twilio account_sid: " TW_AC; read -r -p "Twilio api_key_sid: " TW_SK
  TW_SEC=$(ler "Twilio api_key_secret"); read -r -p "Twilio messaging_service_sid: " TW_MG
  "${BAO[@]}" kv put "kv/$PREFIXO/twilio" account_sid="$TW_AC" api_key_sid="$TW_SK" \
    api_key_secret=- messaging_service_sid="$TW_MG" >/dev/null <<<"$TW_SEC"; unset TW_SEC
  echo "Webhooks (validação dos Avisos de status):"
  read -r -p "SendGrid: chave pública de verificação (base64, do painel): " WH_SG
  WH_TW=$(ler "Twilio Auth Token (Console > Account)")
  # patch preserva os campos meta_* de um WhatsApp já ligado
  if existe webhooks; then
    "${BAO[@]}" kv patch "kv/$PREFIXO/webhooks" sendgrid_chave_publica="$WH_SG" twilio_auth_token=- >/dev/null <<<"$WH_TW"
  else
    "${BAO[@]}" kv put "kv/$PREFIXO/webhooks" sendgrid_chave_publica="$WH_SG" twilio_auth_token=- >/dev/null <<<"$WH_TW"
  fi
  unset WH_TW
  echo ">> segredos gravados em kv/$PREFIXO/{postgres,sendgrid,twilio,webhooks}"
  read -r -p "Configurar WhatsApp agora? [s/N] " WA_AGORA
  case "$WA_AGORA" in
  s|S) gravar_whatsapp ;;
  *) if existe whatsapp; then echo ">> kv/$PREFIXO/whatsapp já existe — mantido (WhatsApp segue ligado)"
     else echo ">> WhatsApp desligado; para ligar depois: $0 $1 --so-whatsapp"; fi ;;
  esac
  ;;
esac
unset BAO_TOKEN
