## ADDED Requirements

### Requirement: REQ-28 Segredos do webhook da Meta no caminho do WhatsApp
Onde o Provedor do WhatsApp for a Meta, o sistema SHALL ler o App Secret e o
verify token do webhook do caminho do WhatsApp no OpenBao e, na falta deles
ali, do caminho dos webhooks, e SHALL impedir o início quando não estiverem em
nenhum dos dois.
Origem: RN-11, RN-9

#### Scenario: T-105 Segredos da Meta no caminho do WhatsApp ou no antigo
- **WHEN** o processo inicia com a Meta ativa e o App Secret e o verify token só no caminho do WhatsApp, ou só no caminho dos webhooks, ou em nenhum dos dois
- **THEN** nos dois primeiros casos o processo inicia com as rotas da Meta validando por esses segredos, e no terceiro encerra com erro citando o campo faltante
