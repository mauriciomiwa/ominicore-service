## Context
Cada Tentativa aceita guarda `id_provedor` (SendGrid `X-Message-Id`, Twilio
`sid`, Meta `wamid`). O fallback (REQ-9) já cria a próxima Tentativa na
transação da conclusão. A rota pública só existe para `/webhooks/*` (ADR 0003).

## Decisions
- **Adaptador de entrada `entrada/webhooks`**, um handler por Provedor, que
  valida a assinatura **antes** de ler o corpo como dado e traduz o aviso para
  um `AvisoStatus{Provedor, IDProvedor, Status (entregue|lida|falhou), Motivo}`
  do domínio. Caso de uso `ReceiveMessageUseCase` (o nome do diagrama) aplica.
- **Assinaturas:**
  - SendGrid Event Webhook: ECDSA P-256 sobre `timestamp + corpo`, headers
    `X-Twilio-Email-Event-Webhook-Signature` / `-Timestamp`, chave pública
    (base64 DER) do painel. Correlação: `sg_message_id` começa pelo
    `X-Message-Id` gravado (prefixo até o primeiro `.`).
  - Twilio: `X-Twilio-Signature` = base64(HMAC-SHA1(Auth Token, URL completa +
    parâmetros POST ordenados)). A URL é `OMINICORE_WEBHOOK_URL + /webhooks/twilio`
    (atrás do Traefik o host da requisição não é confiável para isso).
  - Meta: `X-Hub-Signature-256` = `sha256=` + hex(HMAC-SHA256(App Secret,
    corpo)); `GET` de verificação compara `hub.verify_token` e devolve
    `hub.challenge`.
  - Comparações em tempo constante (`hmac.Equal`).
- **Máquina de status monotônica no domínio:** ordem enviada(1) < entregue(2)
  < lida(3); falhou é final e só entra a partir de enviada. Aviso que não
  avança é ignorado (idempotência).
- **Falha tardia:** o repositório, numa transação com `FOR UPDATE` na
  Notificação, marca a Tentativa falhou e, se a Notificação não estiver
  entregue, pede ao `NotificationRouter` a próxima Tentativa (mesmo caminho do
  REQ-9, sem Retentativa: Falha definitiva).
- **Classificação da falha** (`categoria_falha`, migration 00005) gravada
  quando a Tentativa falha: `contato_invalido` (131026, 21211, bounce),
  `retentativas_esgotadas`, `recusado_pelo_provedor` (demais). O `GET` só expõe
  isso, nunca `motivo`.
- **Consulta** `GET /v1/notificacoes/{id}`: filtra por `sistema_origem` do
  token na própria query (outro dono = 0 linhas = 404).

## Risks / Trade-offs
- Auth Token do Twilio no OpenBao: credencial forte; só leitura pela policy do
  ambiente (ADR 0002).
- SendGrid manda eventos em lote; um lote com assinatura válida e evento
  desconhecido responde 2xx inteiro.
- Cadastro das URLs de webhook nos painéis (SendGrid, Meta) e do verify token
  é do dono, no runbook do deploy.
