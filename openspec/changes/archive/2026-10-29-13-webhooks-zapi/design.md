# Design

## Context

O motivo está no `proposal.md`, em Why. Estado atual que molda a solução:

- **O mecanismo de Aviso de status já existe.** O webhook da Meta converte o
  payload em `dominio.AvisoStatus{IDProvedor, Status, Motivo, Categoria}` e
  chama `ReceiveMessageUseCase.Receber`
  (`internal/adaptadores/entrada/webhooks/whatsapp.go:76`,
  `internal/aplicacao/receber.go:32`). O `Receber` faz avançar o status
  (`StatusTentativa.Avanca`), aplica o fallback por `Router.DecidirAviso` e
  ignora id desconhecido (`AvisoDesconhecido`).
- **O id guardado no envio é o `messageId`.** A Tentativa da Z-API guarda o
  `messageId` (ou, na falta, o `id`) da resposta de envio como id do Provedor
  (`internal/adaptadores/saida/zapi/zapi.go:89-104`).
- **A configuração do WhatsApp mora em `kv/<prefixo>/whatsapp`.** Ela é gravada
  por `ConfigurarUseCase.Gravar` (`internal/aplicacao/configurar.go:34`), relida
  a cada 30 s por api e worker (REQ-26) e lida a cada requisição pelo
  `ConfigWhatsAppAtual`. A policy do OminiCore só escreve nesse caminho
  (ADR 0005).
- **A rota pública já existe e não é registrada.** O `/webhooks/*` é público,
  atrás da Cloudflare; o router do Traefik tem `accessLogs: false`
  (`deploy/aviva2/traefik/ominicore.yml.tmpl`). `OMINICORE_WEBHOOK_URL` é a
  origem pública (`https://notify.flipapi.com.br` em prod).
- **Formato dos avisos da Z-API** (developer.z-api.io/webhooks, lido em
  2026-10-27):
  - **"Ao enviar" (`DeliveryCallback`):** `phone`, `zaapId`, `messageId`,
    `instanceId`, `type` e `error` (só em falha);
  - **"Status" (`MessageStatusCallback`):** `status` e `ids[]`;
  - **sem assinatura;** a resposta esperada é `200 {"value": true}`;
  - existe API para cadastrar a URL de cada webhook (headers `instanceId` e
    `token`, corpo `{"value": url}`), sem método nem caminho documentados.

## Goals / Non-Goals

**Goals:**
- Reaproveitar o `Receber` e o `DecidirAviso` sem mudar o domínio do fallback:
  o novo adaptador só traduz o payload da Z-API em `AvisoStatus`.
- Segredo nunca em log, resposta, URL de resposta ou tela.
- O cadastro na Z-API não pode deixar a gravação da configuração frágil.

**Non-Goals:**
- Retentar o cadastro sozinho. A retentativa é o botão "Cadastrar de novo".
- Receber "Ao receber", conexão, desconexão ou presença, mesmo se a Z-API
  mandar para a mesma URL.

## Decisions

### D1 — Rota e autenticação (ADR 0007)
`POST /webhooks/zapi/{segredo}` registrado sempre. O handler decide a cada
requisição pelo `ConfigWhatsAppAtual`, como a rota da Meta já faz com
`soComMeta`:
- se o Provedor ativo não for a Z-API, responde **404**;
- se o segredo não bater, responde **404**; a comparação é com
  `subtle.ConstantTimeCompare`, contra `webhook_segredo` da configuração atual;
- se o `instanceId` do corpo for diferente do `instance_id` configurado,
  responde **404**.

Sempre 404, sem distinguir o motivo, para não confirmar a existência da rota.
O corpo é limitado a 64 KiB. Aviso autêntico recebe sempre `200 {"value": true}`,
mesmo com tipo desconhecido ou sem efeito. A ADR 0007 registra a decisão e a
revisão da RN-5.

- *Alternativas:* conferir o segredo em header (a Z-API não deixa configurar
  headers no webhook); proxy com Cloudflare Access (o webhook da Z-API não
  autentica); ficar sem webhook (a decisão anterior, que perde fallback e
  entrega).

### D2 — Tradução dos eventos
Função pura `avisosZAPI(corpo) []dominio.AvisoStatus`, testada em tabela:
- `type: "MessageStatusCallback"`: para cada id de `ids`, `RECEIVED` →
  `TentativaEntregue`; `READ` e `PLAYED` → `TentativaLida`; os demais status
  não geram aviso;
- `type: "DeliveryCallback"` **com** `error`: `IDProvedor = messageId`,
  `TentativaFalhou`, `Motivo = "zapi delivery: <error>"` (cortado em 200
  caracteres, sem o telefone). A categoria é:
  - `FalhaContatoInvalido`, se o erro contém "Phone number does not exist" ou
    "invalid phone number";
  - `FalhaRecusadoPeloProvedor`, nos demais;
- `DeliveryCallback` sem `error` e os demais `type` não geram aviso.

Cada aviso vai para o `Receber`. Erro de banco responde **500**, para a Z-API
tentar de novo, como no webhook da Meta.

### D3 — Segredo e estado no cofre
- `webhook_segredo` entra em `CamposOpcionaisWhatsApp[WhatsAppZAPI]` e em
  `segredosWhatsApp`. Com isso, nunca sai pela API (REQ-23) nem para log, e a
  consulta o mostra só como `configurado`.
- O estado do cadastro fica no mesmo caminho, em dois campos não secretos:
  `webhooks_estado` (`cadastrados` | `falhou`) e `webhooks_em` (RFC 3339). A
  consulta já devolve campos públicos (`Publicos()`), então o Console os lê sem
  rota nova.
- Os três campos são sempre do OminiCore: se vierem no corpo do `PUT`, são
  descartados antes de gerar o segredo (teste na tarefa 3.2).
- O `Validar()` não exige esses campos. Uma configuração gravada antes desta
  change continua válida e fica sem webhook até a próxima gravação ou o
  "Cadastrar de novo".

### D4 — Gravar e recadastrar
No `ConfigurarUseCase`:
- **`Gravar` com a Z-API:**
  1. gera o segredo (`crypto/rand`, 32 bytes, base64url sem padding);
  2. grava a configuração com `webhook_segredo` e `webhooks_estado=falhou`;
  3. aplica no processo e registra o evento, como hoje;
  4. chama a porta nova `CadastroWebhooksPort.Cadastrar(ctx, cfg, url)`;
  5. regrava só `webhooks_estado` e `webhooks_em` com o resultado.

  A resposta do `PUT` traz o estado. A falha no passo 4 vai para o log (sem
  segredo nem token) e **não** muda o código de resposta (200), porque a
  configuração valeu.
- **`Recadastrar`** (`POST /v1/config/whatsapp/webhooks`, papel de Operador):
  - sem a Z-API ativa → **409** `"recadastro só vale para a Z-API"`;
  - senão, repete os passos 4 e 5 com o segredo atual;
  - responde 200 com o estado; registra em `config_eventos` com a ação
    `webhooks`.
- **URL cadastrada:** `OMINICORE_WEBHOOK_URL + "/webhooks/zapi/" + segredo`, a
  mesma para "Ao enviar" e para "Status" (o `type` do corpo diferencia). Sem
  `OMINICORE_WEBHOOK_URL` (local), o cadastro não é tentado e o estado fica
  `falhou`, com log "sem URL pública".

### D5 — Adaptador de cadastro na Z-API
- `internal/adaptadores/saida/zapi` ganha `Webhooks.Cadastrar`, com duas
  chamadas: "Ao enviar" e "Status".
- Elas levam os headers `Client-Token` (quando configurado) e o token da
  instância, e o corpo `{"value": url}`.
- Timeout de 10 s por chamada, sem retentativa. 2xx sem `error` = sucesso.
- O **método e o caminho** ficam numa constante só (hipótese:
  `PUT /instances/{id}/token/{token}/update-webhook-delivery` e
  `.../update-webhook-message-status`). A tarefa 2.1 confirma com a instância
  real antes de qualquer outro código.
- O motivo de erro é mascarado com o `mascarar` que já existe (token na URL).

### D6 — Console
A tela do WhatsApp mostra, com a Z-API ativa:
- "Webhooks cadastrados na Z-API em <data>" ou o aviso "Webhooks não
  cadastrados";
- o botão **"Cadastrar de novo"**, que chama `POST /v1/config/whatsapp/webhooks`
  e atualiza o estado.

O segredo aparece só como "configurado", na lista de segredos que já existe.
Todos os textos entram no `pt-BR.json`.

## Risks / Trade-offs

- **[Risco] A URL vaza (log de terceiros, print, logs da Cloudflare) e alguém
  forja avisos** → mitigação:
  - o segredo é trocado a cada gravação;
  - a instância precisa bater;
  - o pior caso é marcar uma Tentativa como entregue, lida ou falhou, o que
    afeta status e fallback, mas não vaza dado pessoal e não envia mensagem a
    outro número.

  A ADR 0007 registra o risco aceito.
- **[Risco] O caminho da API de cadastro diverge do suposto** → mitigação: a
  tarefa 2.1 confere com a instância real antes de escrever o adaptador; se
  não houver API por tipo, a alternativa é o "Atualizar todos os webhooks" sem
  `notifySentByMe`.
- **[Risco] O `messageId` do aviso não é o mesmo da resposta de envio** → o
  aviso cai como desconhecido (200, nada muda). A tarefa 2.1 confere; se
  divergir, o envio passa a guardar o id que vem nos avisos.
- **[Trade-off] A regravação invalida o segredo na hora:** avisos em trânsito
  com o segredo antigo recebem 404 e se perdem. Aceito: a gravação é rara e a
  Tentativa só deixa de avançar.
- **[Risco] Aviso duplicado ou fora de ordem** → já coberto: `Avanca` só deixa
  avançar e falhou é final (REQ-12).

## Migration Plan

1. Arquivar a 08, a 09, a 10 e a 11; implementar depois da change 12.
2. Publicar a API em prod (`publicar.sh prod`) e o Console (`wrangler deploy`).
   A configuração da Z-API já gravada fica sem webhook até a próxima ação.
3. No Console, "Cadastrar de novo", ou regravar a Z-API. Conferir na Z-API
   (painel › Webhooks) que as duas URLs foram preenchidas.
4. Mandar uma Notificação real e ver a Tentativa passar a entregue ou lida.

**Rollback:** voltar o binário (`publicar.sh prod --voltar`). O REQ-21 antigo
volta a valer, e a rota `/webhooks/zapi` some (404). As URLs na Z-API podem
ficar; os avisos recebem 404 e são ignorados. Limpar no painel da Z-API, se
quiser.

## Open Questions

- O caminho exato e o `messageId` dos avisos: respondidos pela tarefa 2.1, sem
  mudar a abordagem (D5 e Risks já preveem a variação).
