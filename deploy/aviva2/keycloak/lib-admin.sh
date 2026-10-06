# shellcheck shell=bash
# Login do kcadm no realm master do Keycloak do aviva2.0, para os scripts desta
# pasta. O admin da casa é o usuário de bootstrap (KC_BOOTSTRAP_ADMIN_USERNAME
# em ../flip/infra/aviva2.0/keycloak/keycloak.env); a senha é o podman secret
# kc-admin-password, lido aqui como root, sem passar pelo terminal. Sem o
# secret, pede a senha (sem eco). Define o array KC já autenticado.
KC_ADMIN_USUARIO="${KC_ADMIN_USUARIO:-bootstrap-admin}"
KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh)

entrar_admin_keycloak() {
  local senha
  senha=$(podman secret inspect --showsecret --format '{{.SecretData}}' kc-admin-password 2>/dev/null || true)
  if [ -z "$senha" ]; then
    read -r -s -p "senha do admin do Keycloak ($KC_ADMIN_USUARIO, realm master): " senha; echo
  fi
  "${KC[@]}" config credentials --server http://localhost:8080 --realm master \
    --user "$KC_ADMIN_USUARIO" --password "$senha" >/dev/null
}
