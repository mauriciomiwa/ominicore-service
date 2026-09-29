#!/usr/bin/env bash
# Publica o OminiCore no aviva2.0 (i-044194deaab7ca807, us-east-1, t4g/arm64),
# no molde do aviva-back (flip/aviva-back/deploy/aviva2/publicar.sh). Idempotente.
#
# Uso (da raiz do repo):  ./deploy/aviva2/publicar.sh hml|prod [--voltar]
#   sem ambiente padrão: o alvo é sempre explícito (skill ambientes).
#   --voltar  troca o binário pelo anterior (/opt/ominicore<sufixo>/ominicore.anterior).
#   HOST=ubuntu@44.209.101.158  KEY=~/.ssh/aviva2.0.pem  (padrões)
#
# Pré-requisitos no host (uma vez por ambiente; ver deploy/aviva2/README.md):
#   schema/role no Aurora, realm + podman secret ominicore<sufixo>-kc-secret
#   (keycloak/criar-realm.sh) e cofre (openbao/configurar-openbao.sh).
set -euo pipefail
cd "$(dirname "$0")/../.."
case "${1:-}" in hml|prod) ;; *) echo "uso: $0 hml|prod [--voltar]" >&2; exit 2 ;; esac
# shellcheck source=/dev/null
. "deploy/aviva2/ambientes/$1.sh"
HOST=${HOST:-ubuntu@44.209.101.158}
KEY=${KEY:-$HOME/.ssh/aviva2.0.pem}
SSH=(ssh -i "$KEY" -o ConnectTimeout=15 "$HOST")
NOME="ominicore${SUFIXO}"

if [ "$AMBIENTE" = prod ]; then
  read -r -p "Publicar em PRODUÇÃO ($HOST, realm $REALM, schema $SCHEMA)? Digite 'prod': " ok
  [ "$ok" = prod ] || { echo "cancelado"; exit 1; }
fi

if [ "${2:-}" = --voltar ]; then
  echo ">> rollback $NOME: binário anterior"
  "${SSH[@]}" "set -e
    test -f /opt/$NOME/ominicore.anterior
    sudo mv /opt/$NOME/ominicore /opt/$NOME/ominicore.revertido
    sudo mv /opt/$NOME/ominicore.anterior /opt/$NOME/ominicore
    sudo systemctl restart $NOME-api.service $NOME-worker.service
    sleep 3; sudo podman logs --tail 5 $NOME-api; sudo podman logs --tail 5 $NOME-worker"
  exit 0
fi

echo ">> make check"
make check >/dev/null
SHA=$(git rev-parse --short HEAD)$(git diff --quiet || echo -dirty)
[ "$AMBIENTE" = prod ] && [[ "$SHA" == *-dirty ]] && { echo "prod só com árvore limpa (commit $SHA)" >&2; exit 1; }

OUT="dist/aviva2-$AMBIENTE"; rm -rf "$OUT"; mkdir -p "$OUT"
echo ">> build linux/arm64 $SHA"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X main.version=$SHA" \
  -o "$OUT/ominicore" ./cmd/ominicore

renderizar() { # template destino [PAPEL]
  PAPEL="${3:-}" AMBIENTE="$AMBIENTE" SUFIXO="$SUFIXO" REALM="$REALM" PREFIXO="$PREFIXO" SCHEMA="$SCHEMA" \
  AUTH_MOUNT="$AUTH_MOUNT" ROLE="$ROLE" HOST_WEBHOOK="$HOST_WEBHOOK" \
    perl -pe 's/\$\{(\w+)\}/exists $ENV{$1} ? $ENV{$1} : die "variável $1 sem valor\n"/ge' "$1" > "$2"
}
renderizar deploy/aviva2/ominicore.env.tmpl "$OUT/ominicore.env"
renderizar deploy/aviva2/quadlet/ominicore.container.tmpl "$OUT/$NOME-api.container" api
renderizar deploy/aviva2/quadlet/ominicore.container.tmpl "$OUT/$NOME-worker.container" worker
renderizar deploy/aviva2/traefik/ominicore.yml.tmpl "$OUT/$NOME.yml"

echo ">> pré-requisitos no host"
"${SSH[@]}" "sudo podman secret exists $NOME-kc-secret && sudo podman secret exists redis-password" \
  || { echo "falta podman secret $NOME-kc-secret (keycloak/criar-realm.sh $AMBIENTE)" >&2; exit 1; }

echo ">> enviando $NOME ($SHA)"
"${SSH[@]}" "rm -rf /tmp/$NOME && mkdir -p /tmp/$NOME"
scp -i "$KEY" -q "$OUT"/* "$HOST:/tmp/$NOME/"

echo ">> instalando, migrando e reiniciando"
"${SSH[@]}" "set -e
  cd /tmp/$NOME
  sudo install -d -m 755 /opt/$NOME /etc/$NOME
  [ -f /opt/$NOME/ominicore ] && sudo cp -p /opt/$NOME/ominicore /opt/$NOME/ominicore.anterior
  sudo install -m 755 ominicore /opt/$NOME/ominicore
  sudo install -m 644 ominicore.env /etc/$NOME/ominicore.env
  sudo install -m 644 $NOME-api.container $NOME-worker.container /etc/containers/systemd/
  sudo install -m 644 $NOME.yml /etc/traefik/dynamic/
  sudo systemctl daemon-reload
  # migration antes do binário novo atender (só aditiva, ver AMBIENTES.md §6)
  sudo podman run --rm --network aviva --env-file /etc/$NOME/ominicore.env \
    --secret $NOME-kc-secret,type=env,target=OMINICORE_KC_CLIENT_SECRET \
    -v /opt/$NOME/ominicore:/ominicore:ro,Z -v /etc/keycloak/rds-global-bundle.pem:/etc/ominicore/rds-ca.pem:ro,Z \
    gcr.io/distroless/static-debian12:nonroot /ominicore migrate
  sudo systemctl restart $NOME-api.service $NOME-worker.service
  rm -rf /tmp/$NOME
  sleep 4
  sudo podman logs --tail 5 $NOME-api; sudo podman logs --tail 5 $NOME-worker"

echo ">> conferência: os dois no ar"
"${SSH[@]}" "sudo podman logs $NOME-api 2>&1 | grep -q 'api no ar' && sudo podman logs $NOME-worker 2>&1 | grep -q 'worker no ar'" \
  && echo "ok: $NOME $SHA" || { echo "FALHOU: veja os logs acima; rollback: $0 $AMBIENTE --voltar" >&2; exit 1; }
