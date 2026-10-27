# Proposal

## Why

O OminiCore está em produção desde 2026-10-06, e o `DOC/to-do.md` (seção
"Prod — operação", 2026-10-27) registra três defeitos:

1. **"Desligar o WhatsApp" quebra a configuração.** O OminiCore apaga a versão
   no KV v2 (*soft delete*) e depois lê essa versão apagada como uma
   configuração incompleta. A consulta passa a dar 503 e o worker fica preso na
   configuração anterior. Isso contraria o REQ-24 (RN-11); em 27/10 foi
   recuperado com `kv undelete`.
2. **A API e o worker não voltam sozinhos.** Ficaram parados de 16/10 a 27/10:
   subiram com o Keycloak fora, o systemd reiniciou em sequência, estourou o
   limite padrão e desistiu.
3. **O Console não avisa quando a API não responde.** Ele fica carregando para
   sempre, e o Operador não sabe que o problema é a API (RN-12).

## What Changes

- **Desligar (REQ-24, sem mudar a spec):** o *soft delete* continua (dá para
  recuperar), mas a leitura passa a tratar a versão apagada como Canal
  desligado. O T-98 ganha um teste contra o OpenBao real (testcontainers), que
  é o que deixou o defeito passar.
- **Reinício (deploy):** o quadlet ganha `RestartSec=10s` e
  `StartLimitIntervalSec=0`. O systemd tenta a cada 10 s, sem desistir.
- **Console (REQ-43 novo):** quando a API não responde ou responde erro do
  servidor (5xx), mostra que a API do Ambiente não respondeu e um botão
  "Tentar de novo", em vez de carregar para sempre.

## Capabilities

### New Capabilities

### Modified Capabilities

- `console-operador`: o REQ-43 (API sem resposta ou com erro do servidor). A
  capacidade nasce no archive da change 11.

## Impact

- **Código:**
  - `internal/adaptadores/saida/openbao/segredos.go` (`LerWhatsApp`);
  - `internal/integracao/` (OpenBao real no T-98);
  - `web/console/src/console/api.ts` e `src/pages/whatsapp.vue`.
- **Deploy:** `deploy/aviva2/quadlet/ominicore.container.tmpl`, que vale no
  próximo `publicar.sh prod`.
- **Dependência:** arquivar a 11 antes desta, porque a capacidade
  `console-operador` só nasce lá.
- **Fora do escopo:**
  - alerta quando os serviços caem (monitoramento);
  - apagar o client `teste-zapi` (operação, já no `to-do`).
