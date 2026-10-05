# Proposal

## Why

Pela change 08, o Canal WhatsApp (RN-4) é enviado pela Meta ou pelo UltraMsg,
escolhido por ambiente. O cliente quer a Z-API como uma terceira opção, da
mesma família do UltraMsg: um gateway **não oficial** sobre o WhatsApp Web, que
envia texto livre (`POST https://api.z-api.io/instances/{instanceId}/token/{token}/send-text`,
JSON `phone`/`message`). Na sabatina de 2026-10-05 ficou decidido que ela
entra como mais um Provedor, um por ambiente, e que o risco aceito pelo
cliente na ADR 0004 vale também para ela.

A RN-4 já foi reescrita em `DOC/Negocio/especificacao-negocio.md`: o Canal
WhatsApp tem agora três Provedores (Meta, UltraMsg ou Z-API). UltraMsg e Z-API
enviam o texto da versão WhatsApp, com até 4096 caracteres. Credencial ou
instância recusada é Falha definitiva, e nenhum dos dois manda Aviso de status
assinado. A RN-5 não muda.

## What Changes

- **Z-API como Provedor do WhatsApp.** `kv/<prefixo>/whatsapp` aceita
  `provedor=zapi`, com os campos exigidos `instance_id` e `token` e o campo
  opcional `client_token` (token de segurança da conta). Como no UltraMsg, os
  segredos de webhook da Meta não são exigidos. Um campo exigido faltando
  impede o início.
- **Envio pela Z-API.** O sistema envia o texto da versão WhatsApp, preenchido
  com as variáveis, para o telefone do Destinatário, só com dígitos (sem o `+`
  do E.164). O `Client-Token` vai no header quando está configurado.
  Resultado:
  - 2xx → Tentativa enviada, com o `messageId` como id do Provedor;
  - 429 ou 5xx → Falha temporária;
  - campo `error`, ou qualquer outro status → Falha definitiva (token,
    instância ou Client-Token recusados);
  - texto preenchido com mais de 4096 caracteres → Falha definitiva, sem
    chamar o Provedor.
- **O token não vaza.** A Z-API exige o token da instância na URL. Por isso,
  nenhum erro da chamada pode levar a URL ao motivo da Tentativa nem ao log.
- **A mesma forma do UltraMsg.** A Z-API envia o texto da versão WhatsApp. O
  pedido com o Canal WhatsApp na ordem e uma versão sem texto é recusado com o
  motivo "versão whatsapp sem texto". Trocar UltraMsg por Z-API não invalida as
  Tentativas pendentes.
- **Sem Aviso de status da Z-API.** Os webhooks *Ao enviar* e *Status* da Z-API
  não têm assinatura documentada. Com a Z-API ativa, a Tentativa aceita fica
  `enviada` e `/webhooks/whatsapp` responde 404.
- **Script do cofre e bootstrap local.** O `configurar-openbao.sh` oferece a
  Z-API, inclusive no `--so-whatsapp`. O `bootstrap.sh` aceita
  `WHATSAPP_PROVEDOR=zapi`.

## Capabilities

### New Capabilities

Nenhuma.

### Modified Capabilities

- `segredos`: o REQ-6 aceita a Z-API como Provedor do WhatsApp, com os campos
  exigidos dela.
- `notificacoes`: o REQ-10 (envio), o REQ-11 (versão WhatsApp) e o REQ-21
  (Provedor sem Aviso de status) passam a incluir a Z-API.

## Impact

- **Dependência:** os deltas MODIFIED partem dos textos da change
  `08-whatsapp-ultramsg`. A 08 precisa ser arquivada **antes** da 09.
- **Negócio e DOC:**
  - RN-4 e glossário ("Provedor") já atualizados na sabatina;
  - adendo na ADR `0004-whatsapp-por-gateway-nao-oficial` (risco estendido à
    Z-API) e entrada no `DOC/log.md`;
  - Z-API em `DOC/Padroes/fontes-externas.md`;
  - Fatia 9 em `DOC/System/backlog-mvp.md` (já incluída);
  - webhooks da Z-API com segredo na URL vão para o `DOC/to-do.md`, porque
    exigem revisar a RN-5.
- **Código:**
  - novo `internal/adaptadores/saida/zapi` e um falso da Z-API para os testes;
  - `internal/dominio` (Provedor `zapi`, forma "texto" compartilhada com o
    UltraMsg);
  - `internal/adaptadores/saida/openbao` (campos da Z-API);
  - `cmd/ominicore` (escolha do Remetente e `OMINICORE_ZAPI_URL`).
- **Contrato:** `api/openapi.yaml` cita a Z-API junto do UltraMsg: o motivo
  "versão whatsapp sem texto" e a Tentativa que fica `enviada`. Os códigos de
  resposta não mudam.
- **Banco:** nenhuma mudança. A versão WhatsApp do `boas-vindas` já tem texto
  (migration 00008).
- **Deploy:** `deploy/aviva2/openbao/configurar-openbao.sh`,
  `deploy/local/bootstrap.sh`, o runbook `deploy/aviva2/README.md` e
  `DOC/Infra/AMBIENTES.md`. Hml e prod usam instâncias e números Z-API
  distintos.
- **Fora do escopo:**
  - webhooks *Ao enviar* e *Status* da Z-API, e com eles a detecção de número
    sem WhatsApp, que só chega por eles;
  - `delayMessage`/`delayTyping` (ficam os padrões da Z-API);
  - mídia, botões e grupos;
  - respostas do Destinatário (RN-6).
