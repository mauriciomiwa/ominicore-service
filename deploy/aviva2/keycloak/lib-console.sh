# shellcheck shell=bash
# Funções do Console do Operador no Keycloak (RN-12, REQ-35), usadas por
# criar-console.sh e criar-operador.sh (hml/prod), pelo deploy/local e pelo
# teste T-118. Quem chama define o array KC com o kcadm.sh já autenticado no
# realm master, por exemplo: KC=(podman exec -i keycloak /opt/keycloak/bin/kcadm.sh).
# Só usa o que existe na imagem do Keycloak (bash e coreutils): roda no host
# ou dentro do container.

CONSOLE_CLIENT=ominicore-console
CONSOLE_FLUXO=navegador-otp
PAPEL_OPERADOR=ominicore-config

# garantir_papel <realm>: cria o realm role do Operador se faltar.
garantir_papel() {
  "${KC[@]}" get "roles/$PAPEL_OPERADOR" -r "$1" >/dev/null 2>&1 ||
    "${KC[@]}" create roles -r "$1" -s name="$PAPEL_OPERADOR" \
      -s 'description=Operador: configura os Canais pela API e pelo Console (RN-11, RN-12)' >/dev/null
}

# id_do_client <realm> <client-id>: id interno, vazio se não existe.
id_do_client() {
  "${KC[@]}" get clients -r "$1" -q clientId="$2" --fields id --format csv --noquotes
}

# configurar_fluxo_otp <realm>: copia o fluxo browser como navegador-otp e torna
# o OTP obrigatório (sem a condição "usuário já configurou"): quem não tem OTP
# é levado a configurá-lo. Idempotente. Imprime o id do fluxo.
configurar_fluxo_otp() {
  local realm=$1 id linha pai="" otp="" eid prov nivel subfluxo prio
  if ! "${KC[@]}" get "authentication/flows/$CONSOLE_FLUXO/executions" -r "$realm" >/dev/null 2>&1; then
    "${KC[@]}" create authentication/flows/browser/copy -r "$realm" -s newName="$CONSOLE_FLUXO" >/dev/null
  fi
  # execuções na ordem do fluxo: id,providerId,level,authenticationFlow,priority;
  # a pilha guarda o subfluxo aberto em cada nível para achar o pai de cada uma.
  # A prioridade vai junto na atualização: sem ela o Keycloak a zera e o OTP
  # passaria para antes do formulário de senha.
  local -a pilha=() condicoes=()
  local -A prioridade=()
  while IFS=, read -r eid prov nivel subfluxo prio; do
    prioridade[$eid]=$prio
    if [ "$subfluxo" = true ]; then pilha[nivel]=$eid; fi
    case "$prov" in
      auth-otp-form) otp=$eid; pai=${pilha[nivel - 1]:-} ;;
      conditional-user-configured) condicoes+=("$eid:${pilha[nivel - 1]:-}") ;;
    esac
  done < <("${KC[@]}" get "authentication/flows/$CONSOLE_FLUXO/executions" -r "$realm" \
    --fields id,providerId,level,authenticationFlow,priority --format csv --noquotes)
  [ -n "$otp" ] && [ -n "$pai" ] || { echo "!! fluxo $CONSOLE_FLUXO sem o formulário de OTP" >&2; return 1; }
  # só a condição do subfluxo do OTP (a do subfluxo de Organização fica)
  for linha in "${condicoes[@]}"; do
    if [ "${linha#*:}" = "$pai" ]; then atualizar_execucao "$realm" "${linha%%:*}" DISABLED "${prioridade[${linha%%:*}]}"; fi
  done
  atualizar_execucao "$realm" "$pai" REQUIRED "${prioridade[$pai]}"
  atualizar_execucao "$realm" "$otp" REQUIRED "${prioridade[$otp]}"
  "${KC[@]}" get authentication/flows -r "$realm" --fields id,alias --format csv --noquotes |
    while IFS=, read -r linha alias; do [ "$alias" = "$CONSOLE_FLUXO" ] && echo "$linha"; done
}

# atualizar_execucao <realm> <id> <requisito> <prioridade>
atualizar_execucao() {
  "${KC[@]}" update "authentication/flows/$CONSOLE_FLUXO/executions" -r "$1" \
    -b "{\"id\":\"$2\",\"requirement\":\"$3\",\"priority\":${4:-0}}" >/dev/null
}

# configurar_console <realm> <origem> [origem-extra]: client público do Console
# (standard flow, PKCE S256, redirect e web origin só das origens dadas,
# audiência ominicore) com o fluxo navegador-otp como override do browser.
configurar_console() {
  local realm=$1 origem=$2 extra=${3:-} fluxo id redirects origens
  garantir_papel "$realm"
  fluxo=$(configurar_fluxo_otp "$realm")
  redirects="\"$origem/*\""
  origens="\"$origem\""
  if [ -n "$extra" ]; then
    redirects="$redirects,\"$extra/*\""
    origens="$origens,\"$extra\""
  fi
  local json="{\"clientId\":\"$CONSOLE_CLIENT\",\"name\":\"Console do Operador\",\"enabled\":true,
 \"protocol\":\"openid-connect\",\"publicClient\":true,\"standardFlowEnabled\":true,
 \"implicitFlowEnabled\":false,\"directAccessGrantsEnabled\":false,\"serviceAccountsEnabled\":false,
 \"redirectUris\":[$redirects],\"webOrigins\":[$origens],
 \"attributes\":{\"pkce.code.challenge.method\":\"S256\",\"post.logout.redirect.uris\":\"$origem/*\"},
 \"authenticationFlowBindingOverrides\":{\"browser\":\"$fluxo\"}}"
  id=$(id_do_client "$realm" "$CONSOLE_CLIENT")
  if [ -z "$id" ]; then
    "${KC[@]}" create clients -r "$realm" -f - >/dev/null <<<"$json"
    id=$(id_do_client "$realm" "$CONSOLE_CLIENT")
  else
    "${KC[@]}" update "clients/$id" -r "$realm" -f - <<<"$json"
  fi
  "${KC[@]}" get "clients/$id/protocol-mappers/models" -r "$realm" --fields name --format csv --noquotes |
    grep -qx audience-ominicore ||
    "${KC[@]}" create "clients/$id/protocol-mappers/models" -r "$realm" -f - >/dev/null <<'JSON'
{"name": "audience-ominicore", "protocol": "openid-connect", "protocolMapper": "oidc-audience-mapper",
 "config": {"included.custom.audience": "ominicore", "access.token.claim": "true", "id.token.claim": "false"}}
JSON
  echo ">> client $CONSOLE_CLIENT pronto em $realm (origem $origem; OTP obrigatório pelo fluxo $CONSOLE_FLUXO)"
}

# criar_operador <realm> <usuario> <email>: pessoa com o papel do Operador e as
# ações obrigatórias de configurar o OTP e trocar a senha. Mostra a senha
# temporária UMA vez no terminal (nunca em arquivo ou log). Se a pessoa já
# existe, só garante o papel: senha e OTP ficam como estão.
criar_operador() {
  local realm=$1 usuario=$2 email=$3 id senha
  garantir_papel "$realm"
  id=$("${KC[@]}" get users -r "$realm" -q username="$usuario" -q exact=true --fields id --format csv --noquotes)
  if [ -z "$id" ]; then
    "${KC[@]}" create users -r "$realm" -s username="$usuario" -s email="$email" -s emailVerified=true \
      -s enabled=true -s 'requiredActions=["CONFIGURE_TOTP","UPDATE_PASSWORD"]' >/dev/null
    senha=$(head -c 18 /dev/urandom | base64 | tr -d '/+=')
    "${KC[@]}" set-password -r "$realm" --username "$usuario" --new-password "$senha" --temporary
    echo ">> $usuario criado em $realm. Senha temporária (entregue por canal seguro; troca no primeiro login):"
    echo "$senha"
  else
    echo ">> $usuario já existia em $realm: papel garantido; senha e OTP intactos (o fluxo do Console exige OTP)"
  fi
  "${KC[@]}" add-roles -r "$realm" --uusername "$usuario" --rolename "$PAPEL_OPERADOR"
}
