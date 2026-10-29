# shellcheck shell=bash
# Login do kcadm no realm master do Keycloak do aviva2.0, para os scripts desta
# pasta, na ordem:
#   1. client de serviço ominicore-deploy (master, papel create-realm: cria
#      realms e administra só os que criou), com o secret no podman secret
#      ominicore-kc-deploy. É o caminho normal: o admin humano tem OTP, e o
#      login direto do kcadm não passa por OTP;
#   2. usuário de bootstrap com o podman secret kc-admin-password;
#   3. usuário e senha perguntados no terminal (senha sem eco).
# Define o array KC já autenticado.
KC_DEPLOY_CLIENT="${KC_DEPLOY_CLIENT:-ominicore-deploy}"
KC_ADMIN_USUARIO="${KC_ADMIN_USUARIO:-bootstrap-admin}"
KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh)

secret_do_host() { podman secret inspect --showsecret --format '{{.SecretData}}' "$1" 2>/dev/null || true; }

login_kcadm() { # usuário senha
  "${KC[@]}" config credentials --server http://localhost:8080 --realm master \
    --user "$1" --password "$2" >/dev/null 2>&1
}

entrar_admin_keycloak() {
  local usuario=$KC_ADMIN_USUARIO senha segredo
  segredo=$(secret_do_host ominicore-kc-deploy)
  if [ -n "$segredo" ] && "${KC[@]}" config credentials --server http://localhost:8080 --realm master \
    --client "$KC_DEPLOY_CLIENT" --secret "$segredo" >/dev/null 2>&1; then
    return 0
  fi
  senha=$(secret_do_host kc-admin-password)
  if [ -n "$senha" ] && login_kcadm "$usuario" "$senha"; then
    return 0
  fi
  read -r -p "usuário admin do Keycloak (realm master, sem OTP) [$usuario]: " resposta
  usuario=${resposta:-$usuario}
  read -r -s -p "senha de $usuario: " senha; echo
  login_kcadm "$usuario" "$senha" || {
    echo "!! login no Keycloak recusado para $usuario (usuário com OTP não serve: crie o client ominicore-deploy, deploy/aviva2/README.md)" >&2
    exit 1
  }
}
