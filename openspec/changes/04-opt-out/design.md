## Context
Canal sem contato já é pulado por `Destinatario.PrimeiroComContato` no pedido,
no fallback (REQ-9) e na falha tardia (REQ-13). O Opt-out entra no mesmo ponto.

## Decisions
- **Domínio:** `Destinatario.OptOut map[Canal]bool`; `Disponivel(c)` = tem
  contato e não está em Opt-out; `PrimeiroComContato` usa `Disponivel`. O
  router não muda. Categoria nova `FalhaOptOut`.
- **Dados (migration 00006, aditiva):** `opt_outs(canal, contato_hmac bytea,
  origem, registrado_em, PRIMARY KEY (canal, contato_hmac))` — o estado atual —
  e `opt_out_eventos(canal, contato_hmac, acao, por, em)` — o histórico
  (registro/remoção, quem). Sem coluna de contato em claro.
- **HMAC:** `saida/optout` (ou no repositório) calcula HMAC-SHA256(chave,
  `canal|contato_normalizado`); chave de `kv/<prefixo>/optout` (`chave`).
  Normalização: e-mail `strings.ToLower(strings.TrimSpace)`; telefone já E.164.
- **Carga:** `ReservarTentativa`, `ReservarPorIDProvedor` e o pedido carregam o
  Opt-out dos contatos da Notificação (até 3 HMACs, uma query). O worker, antes
  de chamar o Provedor, pula Canal em Opt-out com `FalhaDefinitiva` +
  `FalhaOptOut` (T-53).
- **Sinais do Provedor:** o handler do SendGrid trata `unsubscribe` e
  `spamreport` usando o `email` do evento; o Twilio devolve 21610 como
  `FalhaDefinitiva` + `FalhaOptOut` e o despacho registra o Opt-out com o
  telefone da Notificação (origem `twilio`).
- **API:** `POST`/`DELETE /v1/opt-outs` com `{canal, contato}`; validação de
  formato no handler (e-mail, E.164, Canal conhecido); 204 em ambos.

## Risks / Trade-offs
- HMAC com chave: trocar a chave invalida todos os Opt-outs — a chave é
  permanente e a rotação exigiria recalcular a partir de contatos que não
  guardamos (fica no runbook como "não rotacionar").
- O Opt-out automático do SendGrid depende do evento trazer `email` (a doc
  mostra o campo no payload); validar em hml.
