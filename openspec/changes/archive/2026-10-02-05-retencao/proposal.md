## Why
RN-9 (LGPD): contato, nome e outros dados das variáveis e o motivo devolvido
pelo Provedor ficam guardados para sempre hoje. Depois do prazo de retenção
(padrão 30 dias) a Notificação finalizada tem de ficar sem dado pessoal
legível, mantendo o que mede a entrega (status, Canais, classificação da falha,
horários). É a última fatia do MVP (Fatia 5 do `DOC/System/backlog-mvp.md`).

## What Changes
- Tarefa periódica no worker (a cada 1 h, lotes de 1.000 por transação) que
  anonimiza Notificações **finalizadas** com `criada_em` além do prazo:
  `destinatario_email` e `destinatario_telefone` vazios, `dados` = `{}`,
  `motivo` das Tentativas vazio, `anonimizada_em` preenchido. Idempotente.
- Notificação ainda pendente além do prazo não é anonimizada; vira alerta no
  log (id) e entra numa rodada seguinte, depois de finalizada.
- Prazo global `OMINICORE_RETENCAO_DIAS` (padrão 30), igual em hml e prod.
- `GET /v1/notificacoes/{id}` passa a devolver `anonimizada_em` (REQ-14).
- Opt-out (só HMAC) não é tocado; Aviso de status atrasado continua aplicado.

Fora do escopo: prazo por Template ou por Sistema de origem; apagar as linhas;
retenção de backups do banco (to-do de Dados).

## Impact
- Capacidade `notificacoes`: REQ-19 novo; REQ-14 modificado.
- Migration 00007 (aditiva): `notificacoes.anonimizada_em` e índice para a busca
  por criada_em das não anonimizadas.
- Contrato: `anonimizada_em` na consulta.
