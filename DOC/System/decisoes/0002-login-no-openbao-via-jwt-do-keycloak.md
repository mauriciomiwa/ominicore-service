# Login no OpenBao pelo JWT do Keycloak

Status: aceita (2026-09-28)

Sem Nomad (não há Workload Identity), o OminiCore autentica no OpenBao pelo
método `jwt` tendo o Keycloak da casa como emissor: obtém um token
`client_credentials` do client `ominicore-app` e faz login numa role presa ao
`azp`, com policy só de leitura em `kv/ominicore/*`. Assim o único segredo no
host é o client secret do Keycloak (podman secret), e revogar o acesso é
desabilitar o client — em vez de AppRole com secret-id a rotacionar à parte.
