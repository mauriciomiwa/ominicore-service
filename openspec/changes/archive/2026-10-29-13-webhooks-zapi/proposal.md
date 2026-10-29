# Proposal

## Why

Com a Z-API, que é o Provedor do WhatsApp em produção desde 2026-10-08, a
Tentativa termina como "enviada" no aceite. O OminiCore não sabe se a mensagem
foi entregue ou lida. Também não percebe quando o número não tem WhatsApp ou
quando a Z-API recusa depois de aceitar (por exemplo, por shadow ban), e por
isso não passa ao próximo Canal da Ordem de Canais (RN-2).

A Z-API manda esses avisos, mas **sem assinatura**, e a RN-5 descartava todo
aviso sem assinatura válida. A sabatina de 2026-10-27 (`DOC/log.md`) revisou a
**RN-4** e a **RN-5**: nos Provedores que não assinam, vale como assinatura um
segredo na URL do aviso, trocado a cada gravação da configuração, junto com a
instância configurada.

## What Changes

- **Rota `POST /webhooks/zapi/<segredo>`**, pública como os demais
  `/webhooks/*`, só com a Z-API ativa:
  - aviso com o segredo e o `instanceId` configurados conta como Aviso de
    status válido (REQ-12, REQ-13), e o aviso autêntico recebe sempre
    `200 {"value": true}`;
  - segredo errado, outra instância ou Z-API inativa → 404, sem efeito.
- **"Ao enviar" (`DeliveryCallback`) com `error`:** a Tentativa falha.
  "Phone number does not exist" e "invalid phone number" contam como contato
  inválido; os demais erros, como recusado pelo Provedor. Nos dois casos, segue
  o fallback para o próximo Canal.
- **"Status" (`MessageStatusCallback`):** `RECEIVED` → entregue; `READ` e
  `PLAYED` → lida; `SENT` e `READ_BY_ME` são ignorados.
- **REQ-21 modificado:** só o UltraMsg segue sem Aviso de status;
  `/webhooks/whatsapp` continua 404 fora da Meta.
- **Cadastro automático:**
  - a cada gravação da Z-API pela API ou pelo Console, o OminiCore gera um
    segredo novo, guarda no cofre junto da configuração e cadastra as URLs de
    "Ao enviar" e de "Status" na própria Z-API;
  - falha no cadastro não desfaz a gravação, e a consulta da configuração
    informa o estado do cadastro, sem o segredo;
  - nova rota `POST /v1/config/whatsapp/webhooks`, só para o Operador, repete o
    cadastro com o segredo atual.
- **Console:** mostra o estado dos webhooks da Z-API e o botão "Cadastrar de
  novo".
- **ADR 0007:** aviso sem assinatura autenticado por segredo na URL.

## Capabilities

### New Capabilities

### Modified Capabilities

- `notificacoes`: os Avisos de status da Z-API (autenticação, mapeamento dos
  eventos e falha depois do aceite) e o REQ-21 modificado.
- `configuracao-canais`: o cadastro dos webhooks na gravação da Z-API e o
  recadastro. A capacidade nasce no archive da change 10.
- `console-operador`: o estado dos webhooks e o "Cadastrar de novo". A
  capacidade nasce no archive da change 11.

## Impact

- **Dependências:**
  - archive da 08 e da 09 (o REQ-21 é delta delas), da 10 e da 11 (as
    capacidades `configuracao-canais` e `console-operador`);
  - implementação **depois da change 12**, a dos defeitos de produção
    (decisão da sabatina).
- **Código:**
  - `internal/adaptadores/entrada/webhooks` (rota e conversão dos eventos);
  - `internal/adaptadores/saida/zapi` (cadastro das URLs);
  - `internal/dominio` (segredo do webhook entre os segredos da Z-API; estado
    do cadastro);
  - `internal/aplicacao/configurar.go` (gerar o segredo e cadastrar ao gravar;
    recadastrar);
  - `internal/adaptadores/entrada/http` (rota de recadastro e estado na
    consulta);
  - `web/console` (estado e botão).
- **Contrato:** `api/openapi.yaml` ganha `POST /v1/config/whatsapp/webhooks` e o
  campo de estado dos webhooks na consulta. A mudança é aditiva.
- **Borda:** nenhuma regra nova. `/webhooks/*` já é público e o Traefik não
  registra essas URLs (`accessLogs: false`).
- **A conferir com a instância real:** o método e o caminho da API da Z-API que
  cadastra cada webhook (a documentação não traz) e se o `messageId` dos avisos
  é o mesmo que a Z-API devolve no envio.
- **Fora do escopo:**
  - avisos do UltraMsg;
  - "Ao receber" (respostas, RN-6);
  - conexão e desconexão da instância;
  - presença do chat.
