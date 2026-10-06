# shellcheck shell=bash
# Login do kcadm no realm master do Keycloak do aviva2.0, para os scripts desta
# pasta. Tenta o usuário de bootstrap (KC_BOOTSTRAP_ADMIN_USERNAME em
# ../flip/infra/aviva2.0/keycloak/keycloak.env) com o podman secret
# kc-admin-password, lido como root; se o login falhar (o admin de bootstrap
# é provisório e pode ter sido trocado), pergunta usuário e senha no terminal,
# sem eco na senha. Define o array KC já autenticado.
KC_ADMIN_USUARIO="${KC_ADMIN_USUARIO:-bootstrap-admin}"
KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh)

login_kcadm() { # usuário senha
  "${KC[@]}" config credentials --server http://localhost:8080 --realm master \
    --user "$1" --password "$2" >/dev/null 2>&1
}

entrar_admin_keycloak() {
  local usuario=$KC_ADMIN_USUARIO senha
  senha=$(podman secret inspect --showsecret --format '{{.SecretData}}' kc-admin-password 2>/dev/null || true)
  if [ -n "$senha" ] && login_kcadm "$usuario" "$senha"; then
    return 0
  fi
  read -r -p "usuário admin do Keycloak (realm master) [$usuario]: " resposta
  usuario=${resposta:-$usuario}
  read -r -s -p "senha de $usuario: " senha; echo
  login_kcadm "$usuario" "$senha" || { echo "!! login no Keycloak recusado para $usuario" >&2; exit 1; }
}
