# Log — ominicore

Registro cronológico de decisões e eventos, mais recente no topo. Decisão sem
registro vira suposição de alguém daqui a seis meses.

Entradas antigas: `log/arquivo-AAAA.md` (`tools/arquiva_log.py`).

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

## 2026-09-29 — Fatia 6 (`06-whatsapp`) implementada: 13/13
- Canal WhatsApp pela Cloud API da Meta (Graph `v26.0`, conferida na doc em
  2026-09-29); migration 00004 (Modelo aprovado em `template_conteudos`);
  `Variaveis()` inclui os parâmetros do Modelo aprovado; contato = telefone.
- Códigos confirmados na doc da Meta: 131026/132001 definitiva, 130429/131000
  temporária; 131049 (marketing por usuário, 24 h) cai na regra comum (4xx =
  definitiva), porque 24 h não cabem na Retentativa.
- Verde: `go test -race`, `make itest` 17/17, lint, `sqlc diff`, contrato,
  `openspec validate --strict`, rastreabilidade. T-28 a T-33 passando.
- Sentinel: aprovado com ressalvas — `to` vai em E.164 com `+` (aceito pela
  Meta segundo a doc, não validado contra conta real); validação real depende
  da conta Meta Business (to-do).

## 2026-09-29 — Sabatina do WhatsApp fechada; RN-4 entra no MVP (fatia 06)
- Decidido (recomendações aceitas): Cloud API da Meta direto (não Twilio);
  só Modelo aprovado, sem texto livre (depende de RN-6); contato = telefone
  E.164; 131026/132001 definitiva, 130429/131000 temporária; versão WhatsApp do
  `boas-vindas`, ordem padrão inalterada (Email → SMS); versão da Graph API
  fixada no adaptador; validação real depois da conta Meta (não bloqueia).
- Glossário: **Modelo aprovado**. Backlog: fatia 06 com REQ-10/REQ-11; REQ
  previstos das fatias 3–5 renumerados (REQ-12 a REQ-15).
- Change `06-whatsapp` aberta (T-28 a T-33).

## 2026-09-29 — Trilha de infraestrutura: artefatos do aviva2.0 preparados
- `deploy/aviva2/`: quadlets (template api/worker), env sem segredo, rota do
  Traefik só `/webhooks/*`, SQL do schema/role, scripts do realm (kcadm, sem
  reiniciar o Keycloak) e do OpenBao (auth `jwt` por realm), `publicar.sh
  hml|prod` (sem padrão; prod pede confirmação e árvore limpa; `--voltar`).
- Lacunas de código achadas ao desenhar o deploy e corrigidas (config, sem
  regra nova): `OMINICORE_DB_SCHEMA` (hml = `ominicore_hml`, validado contra
  injeção), `OMINICORE_BAO_AUTH_MOUNT` (um mount `jwt` aceita um só emissor),
  `OMINICORE_REDIS_PASSWORD` (Redis do aviva2.0 com `requirepass`). Testes:
  isolamento de schemas no Postgres real, login no mount do ambiente.
- Nada executado no host: as etapas com credencial de admin são do dono.

## 2026-09-29 — Entrega: change `02-sms-e-fallback` arquivada
- Arquivada em `openspec/changes/archive/2026-09-29-02-sms-e-fallback/`; spec
  `notificacoes` sincronizada: REQ-2 modificado, REQ-7 (SMS), REQ-8
  (Retentativa), REQ-9 (fallback). Cobre RN-2 e RN-3 por completo e RN-1
  (ordem do pedido). T-15 a T-27 passando.

## 2026-09-29 — Fatia 2 (`02-sms-e-fallback`) implementada: 24/24
- Router novo no domínio (Retentativa 5 s/30 s/2 min ±20%, sorteio injetado;
  fallback ao próximo Canal com contato); Retentativa e próximo Canal gravados
  na transação da conclusão (sem schema novo além de `destinatario_telefone`).
- Twilio por API Key + Messaging Service; classificação 2xx/429/5xx/4xx movida
  para `saida/provedor` e reusada pelo SendGrid. Segredo `kv/<prefixo>/twilio`
  passa a ser obrigatório no boot (REQ-6).
- Pedido: `canais` opcional (vazio ou repetido = 400; sem versão = 422),
  `telefone` E.164. Contrato sem `enum` em `canais`, para bater com o 422.
- Verde: `go test -race`, `make itest` 14/14 (Postgres 17), lint, `sqlc diff`,
  contrato, `openspec validate --strict`, rastreabilidade; boot local no Docker
  com migration 00003 aplicada. T-15 a T-27 passando.
- Sentinel: aprovado com ressalvas — entrega "pelo menos uma vez" segue; sem
  rate limit por Provedor para rajada de Retentativas; `motivo` do Twilio pode
  citar o telefone (fatia 5 anonimiza).

## 2026-09-29 — Sabatina da fatia 02 fechada; change `02-sms-e-fallback` aberta
- Decidido (recomendações aceitas): `canais` opcional no pedido (Canal sem
  versão no Template = 422); Canal sem contato é pulado, 422 só se nenhum tem;
  telefone E.164; 3 Retentativas (4 Tentativas) com 5 s/30 s/2 min ±20%;
  Twilio por API Key + Messaging Service (`kv/<prefixo>/twilio`); 21211/21610
  e demais 4xx = Falha definitiva; `boas-vindas` ganha SMS, ordem Email → SMS.
- RN-2 corrigida: o critério dizia "até 3 Tentativas"; vale 3 Retentativas.
  Glossário: termo **Retentativa**.

## 2026-09-29 — Entrega: change `01-enviar-email` arquivada
- Arquivada em `openspec/changes/archive/2026-09-29-01-enviar-email/`; specs
  vigentes criadas: `autenticacao-origem` (REQ-1), `notificacoes` (REQ-2 a
  REQ-5), `segredos` (REQ-6). Cobre RN-1, RN-3 (Email), RN-7 e RN-9
  (credenciais); RN-2 parcial (falha encerra; retry/fallback na fatia 02).
- T-1 a T-14 passando (`System/plano-de-testes.md`).

