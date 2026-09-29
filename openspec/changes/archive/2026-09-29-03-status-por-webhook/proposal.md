## Why
RN-5: hoje "enviada" quer dizer só "aceita pelo Provedor" — um bounce ou um
SMS não entregue fica "enviada" para sempre, e a meta de entrega final > 98%
(RN-2) não é medível nem cumprida. O Aviso de status de cada Provedor passa a
atualizar a Tentativa, a falha de entrega depois do aceite dispara o fallback,
e o Sistema de origem consulta o resultado.

É a Fatia 3 do `DOC/System/backlog-mvp.md`; depende das Fatias 2 e 6
(arquivadas). A validação real exige o deploy em hml (rota pública
`notify-hml.flipapi.com.br/webhooks/*`, ADR 0003); aqui tudo é provado contra
Provedores falsos.

## What Changes
- Webhooks públicos: `POST /webhooks/sendgrid` (Event Webhook, assinatura
  ECDSA), `POST /webhooks/twilio` (status callback, `X-Twilio-Signature`
  HMAC-SHA1 com o Auth Token), `GET` + `POST /webhooks/whatsapp` (verificação
  `hub.challenge` e `X-Hub-Signature-256` HMAC-SHA256 com o App Secret).
  Assinatura inválida = 401 sem efeito; aviso válido = 2xx sempre, inclusive
  repetido ou de `id_provedor` desconhecido.
- Status novos: Tentativa `entregue` e `lida` (só WhatsApp); Notificação
  `entregue`. O status só avança; `falhou` é final.
- Falha de entrega depois do aceite = Falha definitiva: nova Tentativa no
  próximo Canal com contato (REQ-9), salvo Notificação já entregue.
- `GET /v1/notificacoes/{id}` para o Sistema de origem dono (outro = 404):
  status, Tentativas (Canal, número, status, horários) e uma classificação
  genérica da falha — nunca contato nem texto do motivo.
- Twilio: `StatusCallback` em cada envio a partir de `OMINICORE_WEBHOOK_URL`.
- Credenciais de validação em `kv/<prefixo>/webhooks` (chave pública do
  SendGrid, Auth Token do Twilio, App Secret e verify token da Meta).

Fora do escopo: callback para o Sistema de origem, rastreio de abertura de
e-mail, cadastro das URLs nos painéis dos Provedores (runbook).

## Impact
- Capacidade `notificacoes`: REQ-12, REQ-13 e REQ-14 novos.
- Contrato: `GET /v1/notificacoes/{id}`; webhooks ficam fora do contrato
  público de envio (são dos Provedores) e documentados no design.
- Migration 00005 (aditiva): `atualizada_em` já existe; índice por
  `id_provedor`; coluna `categoria_falha`.
- Novo adaptador de entrada `entrada/webhooks`; OpenBao lê `<prefixo>/webhooks`.
