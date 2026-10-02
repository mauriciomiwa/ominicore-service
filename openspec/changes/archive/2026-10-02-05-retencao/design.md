## Context
Dado pessoal fica em `notificacoes.destinatario_email`, `destinatario_telefone`,
`dados` e em `tentativas.motivo`. O Opt-out já é só HMAC (REQ-18) e não entra
na retenção. O worker já roda em laço contínuo.

## Decisions
- **Rodada no worker:** `worker.Retencao` com `time.Ticker` de 1 h (mais uma
  rodada no boot), lado a lado com o laço de despacho, sem componente novo. O
  relógio é injetado (`agora func() time.Time`) para os testes.
- **SQL em lote:** uma query `AnonimizarLote` faz, numa transação, `UPDATE` de
  até 1.000 Notificações elegíveis — `anonimizada_em IS NULL`,
  `criada_em < agora - prazo`, status final e `NOT EXISTS` Tentativa pendente —
  com `FOR UPDATE SKIP LOCKED` (não briga com a fila) e o `UPDATE` do `motivo`
  das Tentativas delas. Repete enquanto o lote vier cheio.
- **Pendentes além do prazo:** query `ContarPendentesVencidas` só para o alerta
  no log (número, sem dado pessoal).
- **Migration 00007 (aditiva):** `notificacoes.anonimizada_em timestamptz NULL`
  e índice parcial `(criada_em) WHERE anonimizada_em IS NULL`.
- **Prazo:** `OMINICORE_RETENCAO_DIAS` (padrão 30, mínimo 1), validado no boot.

## Risks / Trade-offs
- O texto livre de `dados` é apagado inteiro (não dá para saber qual variável é
  pessoal) — a consulta não perde nada, porque nunca o expôs.
- O backup do Aurora (2 dias) ainda guarda dado anterior à anonimização por até
  2 dias — aceitável e menor que o prazo; anotado em `Infra/infraestrutura.md`.
