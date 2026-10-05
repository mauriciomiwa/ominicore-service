# Log — ominicore

Registro cronológico de decisões e eventos, mais recente no topo. Decisão sem
registro vira suposição de alguém daqui a seis meses.

Entradas antigas: `log/arquivo-AAAA.md` (`tools/arquiva_log.py`).

## 2026-10-05 — Fatia 10 (`10-config-whatsapp-api`) implementada: 33/33
- `dominio.ConfigWhatsApp` (campos e segredos por Provedor, validação única
  para boot, releitura e API); `openbao.Cofre` (ler, gravar, apagar, re-login
  em 403); `aplicacao.ConfigWhatsAppAtual` (retrato atômico lido por pedido,
  envio, aviso e webhooks; trocas serializadas para a releitura não aplicar
  leitura velha) e `ReleituraConfig`; `ConfigurarUseCase` + rotas
  `/v1/config/whatsapp` com o papel `ominicore-config`; `config_eventos`
  (migration 00009); `conceder-operador.sh`; policy com escrita só em
  `kv/<p>/whatsapp`; ADR 0005.
- Verde: `go test -race`, `make itest`, lint, contrato, `openspec validate
  --strict`, rastreabilidade. T-92 a T-105 passando. Sentinel: aprovado com
  ressalvas (sem `shellcheck`; deploy em hml ainda não feito).

## 2026-10-05 — Decisão: configuração do WhatsApp pela API (change `10-config-whatsapp-api`)
- Sabatina fechada com as recomendações (P1–P10): `GET`/`PUT`/`DELETE
  /v1/config/whatsapp` para Meta, UltraMsg e Z-API, só para o Operador (realm
  role `ominicore-config`; Sistema de origem sem o papel → 403); grava no
  OpenBao só em `kv/<p>/whatsapp` (ADR 0005, revisa a 0002); valida só os
  campos, sem testar no Provedor; segredos nunca saem (consulta diz só se
  estão configurados); vale sem reinício (releitura a cada 30 s; inválida
  mantém a anterior); `config_eventos` guarda quem, quando e qual Provedor.
- Segredos do webhook da Meta passam para `kv/<p>/whatsapp`, com `webhooks`
  como alternativa de leitura (REQ-28); rollback para binário antigo com a
  Meta pela API exige regravá-la pelo script (runbook).
- RN-11 e termo Operador no glossário; Fatia 10 no backlog. REQ-22 a REQ-28,
  T-92 a T-105.

## 2026-10-05 — Fatia 9 (`09-whatsapp-zapi`) implementada: 26/27
- `dominio.WhatsAppZAPI`, `EnviaTexto` e `ServirTexto` (forma + 4096, usado
  também pelo UltraMsg); cofre com `CamposWhatsApp[zapi]` e o opcional
  `client_token`; adaptador `saida/zapi` (JSON, telefone sem `+`, header
  `Client-Token`, tokens mascarados no motivo); `OMINICORE_ZAPI_URL`; scripts
  do cofre e runbook com a ordem cofre → reiniciar → ativar Client-Token.
- Verde: `go test -race`, `make itest`, lint, contrato, `openspec validate
  --strict`, rastreabilidade. T-82 a T-91 passando. Sentinel: aprovado com
  ressalvas (sem `shellcheck`; respostas reais da Z-API pendentes — tarefa 5.3).

## 2026-10-05 — Decisão: WhatsApp também pela Z-API (change `09-whatsapp-zapi`)
- Sabatina fechada com as recomendações: Z-API como terceiro Provedor, um por
  ambiente (`provedor=zapi`, `instance_id`, `token`, `client_token` opcional
  no header `Client-Token`); envia o texto da versão WhatsApp como o UltraMsg,
  com o mesmo limite de 4096 caracteres; sem Aviso de status (RN-5 intacta,
  Tentativa fica `enviada`, `/webhooks/whatsapp` → 404).
- Risco do gateway não oficial estendido à Z-API (adendo na ADR 0004).
- Token da instância vai na URL: erro de rede não leva o `err` ao motivo e a
  recusa mascara os tokens que a Z-API ecoar (T-84).
- Webhooks *Ao enviar*/*Status* da Z-API ficam no `to-do.md` (exigem revisar a
  RN-5); a 09 só arquiva depois da 08 (deltas MODIFIED partem dela).
- RN-4 com três Provedores, glossário, Fatia 9 no backlog. REQ-6/10/11/21
  modificados, T-82 a T-91.

## 2026-10-05 — Fatia 8 (`08-whatsapp-ultramsg`) implementada: 31/32
- `dominio.ProvedorWhatsApp` (`meta`/`ultramsg`, `Servir`); cofre lê `provedor`
  (ausente = meta) e exige os campos de cada um; adaptador `saida/ultramsg`
  (form, 4096 runas, classificação tolerante, `sent:"false"` = definitiva);
  recusa 422 pela forma no pedido; rotas da Meta só com `provedor=meta`
  (`MetaLigada`); migration 00008 (texto do `boas-vindas`).
- Verde: `go test -race`, `make itest` 31/31, lint, contrato, `openspec validate
  --strict`, rastreabilidade. T-70 a T-81 passando. Sentinel: aprovado com
  ressalvas (sem `shellcheck`; formato real da resposta do UltraMsg pendente —
  tarefa 5.2, depende da instância de trial).

## 2026-10-05 — Decisão: WhatsApp pelo UltraMsg (change `08-whatsapp-ultramsg`)
- Sabatina fechada: Meta e UltraMsg convivem, um por ambiente, pelo campo
  `provedor` em `kv/<prefixo>/whatsapp` (ausente = `meta`); UltraMsg envia o
  texto da versão WhatsApp (até 4096); versão sem a forma do Provedor ativo →
  422; sem Aviso de status do UltraMsg (Tentativa fica enviada, webhook 404);
  Opt-out de WhatsApp só pela API; instância e número distintos em hml e prod.
- O cliente aceitou o risco do gateway não oficial (ADR 0004).
- Formato da resposta do UltraMsg não documentado: classificação tolerante,
  conferir com a instância de trial antes do deploy.
- RN-4 reescrita, glossário ("Provedor"), Fatia 8 no backlog. REQ-6/10/11
  modificados, REQ-21, T-70 a T-81.

## 2026-10-05 — Entrega: change `07-whatsapp-opcional` arquivada
- Arquivada em `openspec/changes/archive/2026-10-05-07-whatsapp-opcional/`; specs
  sincronizadas: `segredos` REQ-6 modificado (T-63, T-64) e `notificacoes` REQ-20
  (Canal desligado, T-65 a T-69). Cobre RN-4 e RN-9. T-63 a T-69 passando.
- O deploy em hml não depende mais da conta Meta.

## 2026-10-05 — Fatia 7 (`07-whatsapp-opcional`) implementada: 21/21
- Sem `kv/<prefixo>/whatsapp` o Canal fica desligado (`Segredos.WhatsAppLigado`;
  campos `meta_*` exigidos só com ele ligado); `Destinatario.Desligados` entra em
  `Disponivel` (pedido, fallback e falha tardia); recusa `CanalIndisponivel` →
  422; rotas da Meta só registradas com o Canal ligado; worker sem Remetente.
- `configurar-openbao.sh`: "[s/N]", `--so-whatsapp` e `kv patch` em `webhooks`
  (preserva `meta_*`). Runbook, AMBIENTES, infraestrutura e mapa atualizados.
- Verde: `go test -race`, `make itest` 30/30, lint, contrato, `openspec validate
  --strict`, rastreabilidade. T-63 a T-69 passando. Sentinel: aprovado com
  ressalvas (sem `shellcheck` local; script não rodado contra OpenBao real).

## 2026-10-05 — Decisão: WhatsApp opcional para o core ir ao ar (change `07-whatsapp-opcional`)
- Sabatina fechada: sem `kv/<prefixo>/whatsapp` o Canal fica desligado (caminho
  incompleto segue erro de boot); pedido só com Canais desligados → 422 "canal
  whatsapp indisponível", Canal desligado pulado na ordem e no fallback;
  `/webhooks/whatsapp` → 404; `configurar-openbao.sh` com "[s/N]" e `--so-whatsapp`.
- Escopo só isto; decisões abertas do `to-do.md` ficam para depois do tráfego real.
- RN-4 ajustada, termo "Canal desligado" no glossário, Fatia 7 no backlog.
  Change proposta: REQ-6 modificado, REQ-20, T-63 a T-69.

## 2026-10-02 — Entrega: change `05-retencao` arquivada — MVP completo
- Arquivada em `openspec/changes/archive/2026-10-02-05-retencao/`; spec
  `notificacoes` sincronizada: REQ-14 modificado (`anonimizada_em`) e REQ-19
  (anonimização depois do prazo). Cobre RN-9. T-56 a T-62 passando.
- **Todas as fatias do backlog do MVP entregues** (1 a 6): RN-1 a RN-5, RN-7,
  RN-8 e RN-9 cobertas por REQ-1 a REQ-19 e T-1 a T-62. Ficam "depois": RN-6
  (respostas do Destinatário) e RN-10 (janela de horário e agendamento).
- Falta para produção: deploy em hml pelo runbook (etapas do dono) e validação
  real dos Provedores (to-do).

## 2026-10-02 — Fatia 5 (`05-retencao`) implementada: 15/15
- Migration 00007 (`anonimizada_em`, índice parcial); `AnonimizarLote` numa
  query (CTE com `FOR UPDATE OF n SKIP LOCKED`, limpa `motivo` das Tentativas e
  contato/`dados` da Notificação); rodada no worker no boot e a cada 1 h, lotes
  de 1.000 até esvaziar; pendentes vencidas viram alerta (só contagem).
  `OMINICORE_RETENCAO_DIAS` (padrão 30) validado no boot do worker.
  `anonimizada_em` na consulta e no contrato.
- `make itest` passa a usar o socket do contexto ativo do docker CLI: o
  runtime trocou para o Colima e o testcontainers não lê contextos.
- Verde: `go test -race`, `make itest` 30/30 (Postgres 17 no Colima), lint,
  `sqlc diff`, contrato, `openspec validate --strict`, rastreabilidade.
  T-56 a T-62 passando. Sentinel: aprovado.

## 2026-10-02 — Sabatina da fatia 05 fechada; change `05-retencao` aberta
- Decidido (recomendações aceitas): anonimizar (não apagar) — contato, `dados`
  e `motivo` saem; status, Canais, categoria e horários ficam, com
  `anonimizada_em`; prazo global `OMINICORE_RETENCAO_DIAS` (padrão 30); só
  Notificação finalizada (pendente vira alerta no log); rodada no worker a cada
  1 h em lotes de 1.000; `GET` continua e mostra `anonimizada_em` (REQ-14
  modificado); Aviso de status atrasado continua aplicado.
- RN-9 detalhada; backlog da fatia 5 atualizado. Change com REQ-19, T-56 a T-62.

## 2026-10-02 — `Infra/infraestrutura.md` criado (checagem nova das ferramentas SDD)
- Stack, onde roda, bancos e recursos de hml e prod, a partir de `AMBIENTES.md`,
  da stack e do runbook do Aurora no `flip` (PostgreSQL 15.15; backup com 2 dias
  de retenção). Risco anotado no `to-do.md`: 2 dias é pouco para os Opt-outs.

## 2026-10-02 — Entrega: change `04-opt-out` arquivada
- Arquivada em `openspec/changes/archive/2026-10-02-04-opt-out/`; spec
  `notificacoes` sincronizada: REQ-14 modificado (categoria `opt_out`), REQ-15
  (API), REQ-16 (sinais do Provedor), REQ-17 (Canal pulado), REQ-18 (só HMAC).
  Cobre RN-8. T-45 a T-55 passando.

## 2026-09-30 — Fatia 4 (`04-opt-out`) implementada: 21/21
- `Destinatario.Disponivel` (tem contato e não está em Opt-out) no mesmo ponto
  que já pulava Canal sem contato: pedido, fallback e falha tardia respeitam o
  Opt-out sem mudar o router; o worker confere de novo antes de chamar o
  Provedor (categoria `opt_out`).
- `POST`/`DELETE /v1/opt-outs` (contato no corpo); SendGrid `unsubscribe`/
  `spamreport` e Twilio 21610 registram Opt-out; Twilio 21211 passa a marcar
  `contato_invalido`. Migration 00006 com `opt_outs` (PK `(contato_hmac, canal)`,
  corrigida na revisão para a busca por hmac usar o índice) e `opt_out_eventos`.
- HMAC-SHA256(chave, canal|contato normalizado); chave de `kv/<prefixo>/optout`
  com mínimo de 32 caracteres no boot; `configurar-openbao.sh` cria uma vez e
  nunca sobrescreve; runbook: "não rotacionar". T-55 prova que outra chave não
  enxerga o Opt-out.
- Verde: `go test -race`, `make itest` 24/24, lint, `sqlc diff`, contrato,
  `openspec validate --strict`, rastreabilidade. T-45 a T-55 passando.
- Sentinel: aprovado com ressalvas — qualquer Sistema de origem pode remover o
  Opt-out de outro (decidido; fica no histórico); o evento do SendGrid precisa
  trazer `email` (validar em hml).

## 2026-09-29 — Sabatina da fatia 04 fechada; change `04-opt-out` aberta
- Decidido (recomendações aceitas): Opt-out por API e pelos sinais do Provedor
  (SendGrid unsubscribe/spamreport, Twilio 21610); vale para toda a casa, por
  contato + Canal; só a API remove, com registro de quem/quando; contato só em
  HMAC-SHA256 (chave em `kv/<prefixo>/optout`), sem expirar; Canal em Opt-out
  pulado no pedido (422 se nenhum), no fallback e no envio (categoria `opt_out`).
- Menores, pela recomendação: `POST`/`DELETE /v1/opt-outs` com contato no corpo
  (nunca na URL); sem `GET` no MVP; REQ-14 modificado (categoria `opt_out`).
- RN-8 detalhada. Backlog: corrigido o "Pronto quando" trocado entre as fatias
  4 e 5 (erro da renumeração da fatia 3); fatia 4 = REQ-15 a REQ-18, fatia 5 = REQ-19.
- Change aberta com T-45 a T-55.

## 2026-09-29 — Entrega: change `03-status-por-webhook` arquivada
- Arquivada em `openspec/changes/archive/2026-09-29-03-status-por-webhook/`;
  spec `notificacoes` ganha REQ-12 (Aviso de status), REQ-13 (falha depois do
  aceite) e REQ-14 (consulta). Cobre RN-5 e fecha a meta de entrega final da
  RN-2. T-34 a T-44 passando. Validação real depende do deploy em hml.

## 2026-09-29 — Fatia 3 (`03-status-por-webhook`) implementada: 21/21
- Webhooks `/webhooks/{sendgrid,twilio,whatsapp}` com assinatura validada sobre
  o corpo cru (ECDSA SendGrid, HMAC-SHA1 Twilio sobre `OMINICORE_WEBHOOK_URL`,
  HMAC-SHA256 Meta), comparação em tempo constante; `GET` de verificação da
  Meta só devolve o desafio numérico. Status só avança; falha tardia vira
  Falha definitiva e fallback, com `FOR UPDATE` em Tentativa + Notificação (8
  avisos simultâneos = 1 fallback, provado no Postgres). `GET
  /v1/notificacoes/{id}` só para o dono, com `categoria_falha` e sem motivo.
- Formatos conferidos na doc oficial; suposição a validar em hml: chave pública
  do SendGrid (aceita PEM ou base64 DER) e `sg_message_id` = `X-Message-Id` + `.sufixo`.
- Traefik: access log desligado na rota de webhooks (o `hub.verify_token` vai
  na query).
- Verde: `go test -race`, `make itest` 21/21, lint, `sqlc diff`, contrato,
  `openspec validate --strict`, rastreabilidade; boot local com o segredo
  `webhooks`. T-34 a T-44 passando.
- Sentinel: aprovado com ressalvas — sem proteção de replay além da
  idempotência (status só avança); Auth Token do Twilio no cofre; ambiente
  local: Keycloak morto por falta de memória (VM do Docker com 2 GB, dividida
  com o Supabase de outro projeto).

## 2026-09-29 — Sabatina da fatia 03 fechada; change `03-status-por-webhook` aberta
- Decidido (recomendações aceitas): falha de entrega depois do aceite =
  Falha definitiva e fallback, salvo Notificação já entregue; status só avança
  (enviada → entregue → lida; falhou final); consulta `GET /v1/notificacoes/{id}`
  só para o dono (callback depois); credenciais de validação em
  `kv/<prefixo>/webhooks` (inclui Auth Token do Twilio); aviso válido sempre
  2xx; rotas `/webhooks/{sendgrid,twilio,whatsapp}`; `OMINICORE_WEBHOOK_URL`
  para o `StatusCallback`; o `GET` expõe só a classificação da falha.
- RN-5 detalhada; glossário: **Aviso de status**; to-do: decisão de consulta ×
  callback resolvida. Backlog: fatia 3 com REQ-12 a REQ-14; fatias 4 e 5 com
  REQ-15 e REQ-16. Change aberta com T-34 a T-44.

## 2026-09-29 — Entrega: change `06-whatsapp` arquivada
- Arquivada em `openspec/changes/archive/2026-09-29-06-whatsapp/`; spec
  `notificacoes` ganha REQ-10 (envio pelo Modelo aprovado) e REQ-11 (versão
  WhatsApp do Template). Cobre RN-4. T-28 a T-33 passando. Validação contra a
  Meta real segue pendente da conta (to-do).

