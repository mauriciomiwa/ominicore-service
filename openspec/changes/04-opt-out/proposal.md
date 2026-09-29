## Why
RN-8: quem pede "não quero mais SMS" não pode continuar recebendo — nem de
outro Sistema de origem da casa. Hoje o OminiCore não conhece Opt-out: só os
Provedores bloqueiam, e cada um do seu jeito. A regra é **nunca** enviar por um
Canal em que o Destinatário fez Opt-out, pulando-o como já se pula Canal sem
contato (RN-2).

É a Fatia 4 do `DOC/System/backlog-mvp.md`; depende das Fatias 2 e 3
(arquivadas).

## What Changes
- `POST /v1/opt-outs` e `DELETE /v1/opt-outs` (corpo `{canal, contato}`,
  contato nunca na URL), idempotentes (204); a remoção grava quem e quando.
  Sem `GET` no MVP (evita enumerar contatos).
- Opt-out automático pelos sinais do Provedor: eventos `unsubscribe` e
  `spamreport` do SendGrid (Event Webhook da Fatia 3) e o código 21610 do
  Twilio no envio (STOP).
- Opt-out vale para toda a casa, por contato + Canal (SMS e WhatsApp são Canais
  distintos para o mesmo telefone).
- Canal em Opt-out é pulado no pedido (nenhum disponível = 422), no fallback e
  no envio: o worker confere de novo antes de chamar o Provedor e, sem Canal
  restante, encerra com a categoria nova `opt_out`.
- Contato guardado só como HMAC-SHA256 do contato normalizado (e-mail em
  minúsculas, telefone E.164), chave em `kv/<prefixo>/optout`; o registro não
  expira (sobrevive à retenção da Fatia 5).
- Consulta (REQ-14) passa a expor a categoria `opt_out`.

Fora do escopo: `GET` de Opt-out; STOP recebido como mensagem (RN-6);
sincronizar a lista de supressão dos Provedores.

## Impact
- Capacidade `notificacoes`: REQ-15 a REQ-18 novos; REQ-14 modificado.
- Contrato: `/v1/opt-outs` (POST, DELETE); enum de `categoria_falha`.
- Migration 00006 (aditiva): tabelas `opt_outs` e `opt_out_eventos`.
- Segredo novo `kv/<prefixo>/optout` (campo `chave`).
