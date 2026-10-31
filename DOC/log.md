# Log — ominicore

Registro cronológico de decisões e eventos, mais recente no topo. Decisão sem
registro vira suposição de alguém daqui a seis meses.

Entradas antigas: `log/arquivo-AAAA.md` (`tools/arquiva_log.py`).

## 2026-10-31 — Modelos da régua do wbem (fatia 16)
- **Dono (31/10, na sessão do aviva-back):** criar os modelos de WhatsApp da apresentação v9 do aviva-back, só os modelos (os disparos ficam no aviva-back, quando cada evento existir). Texto com o link escrito no lugar do botão; o PDF das Mensagens 7, 8 e 11 vai como link.
- **Change `16-modelos-regua-wbem`** (RN-14, REQ-49, T-157 e T-158): migration `00012_modelos_regua_wbem.sql` com 16 Templates. O link inicial da v9 é um Template novo (`wbem-financiamento-na-clinica`): reescrever o `wbem-link-tomador` quebraria o pedido de hoje do aviva-back, que só manda `link`.
- **Numeração:** a `15-fila-zapi` (ainda não publicada) planejava a `00012`; passa para a `00013`, porque o goose não aplica migration mais antiga que a última.

## 2026-10-29 — Changes 08 a 14 arquivadas
- Arquivadas em cadeia: 08, 09, 10, 11, 12, 13 e 14. A 5.2 da 08 (formato real
  da resposta do UltraMsg) foi arquivada em aberto, por decisão do dono: prod
  usa a Z-API. O item foi para o `to-do.md`, como condição para ligar o UltraMsg.
- O archive deixou o REQ-6, o REQ-10 e o REQ-11 acima de 500 caracteres. A
  change editorial `ajuste-texto-whatsapp` (arquivada) reescreveu os três, sem
  mudar regra nem cenário. Resultado: `openspec validate --all --strict` com 5/5.

## 2026-10-29 — WAF do Console estava desligado; corrigido
- A verificação de exposição pelo CI (primeira vez, de fora do escritório) deu
  200 no Console e 401 em `/v1/config/whatsapp`. Na Cloudflare, a lista
  `escritorio` e a regra C2 não existiam: a zona `flipapi.com.br` está no plano
  Free, com as 5 regras personalizadas em uso. O que protegia a configuração era
  só o JWT, o papel e o OTP, o risco já previsto na ADR 0006.
- Correção, feita pelo dono no painel: o IP `72.60.159.29` entrou direto na regra
  "Block specific paths", com um `or`, sem lista. CI depois: T-113 a T-115 com
  403 (fecha a 7.3 da change 11). Variável `CONSOLE_PUBLICADO=true` criada para
  a verificação diária.
- O ajuste da IA no painel foi barrado pelo modo automático (recurso de produção
  compartilhado); a IA leu a configuração pela API e o dono aplicou.

## 2026-10-29 — Fatia 14 em produção; aviso real da Z-API confere
- API `c3bcb0c` (`HOST=aviva publicar.sh prod`, pelo dono): goose na versão 11,
  `wbem-link-tomador` e `wbem-otp-tomador` no schema `ominicore`.
- Envio real com o Sistema de origem `wbem` (client já existente no realm
  `ominicore`) para o celular do dono: 202, notificação `83e8a763…`, Tentativa
  de WhatsApp `entregue` pelo aviso da Z-API — o `messageId` do aviso é o da
  resposta de envio. Fecha a 3.1 da change 14 e a 7.4 da change 13.
- Do host, `auth.flipapi.com.br` dá 403 (WAF da Cloudflare): token e chamadas de
  teste saem pela rede `aviva` (`keycloak:8080`, `ominicore-api:8080`, podman rootful).

## 2026-10-29 — Change `14-templates-wbem` (Templates do aviva-back)
- Pedido do aviva-back (change `whatsapp-pelo-ominicore` de lá): link de acesso
  e código de acesso do tomador pelo WhatsApp. Decisões do dono na sabatina do
  aviva-back: dois Templates de texto curto, só WhatsApp, por migration (o
  cadastro por API segue em aberto). RN-13, REQ-44, T-139 e T-140.
- Migration `00011_templates_wbem.sql` (commit `1a123da`, antes da change:
  furo apontado no `to-do.md` e fechado aqui). Falta publicar (tarefa 3.1).

## 2026-10-29 — Change 13 em produção
- API `4f1d1b8` (`HOST=aviva publicar.sh prod`, por Claude com o "pode" do dono):
  healthy; `.anterior` = `ac762e0` íntegro; `/webhooks/zapi/<errado>` → 404.
- Console publicado (versão 57674d31) com o estado dos webhooks e "Cadastrar
  de novo". Falta a 7.4: cadastrar pelo Console e conferir um aviso real.
- Achado no primeiro "Cadastrar de novo" (17:00): os webhooks foram
  cadastrados na Z-API, mas o registro em `config_eventos` falhou (o CHECK só
  aceitava gravada/apagada) e a API respondeu 500. Correção: migration 00010.

## 2026-10-29 — Change `13-webhooks-zapi` implementada (31/35; falta conferir a Z-API real e publicar)
- Rota `POST /webhooks/zapi/{segredo}` (segredo em tempo constante + `instanceId`;
  404 sem efeito; aviso autêntico → 200 `{"value":true}`): "Ao enviar" com
  erro → falhou e fallback; Status RECEIVED/READ/PLAYED → entregue/lida.
- Gravar a Z-API gera segredo novo e cadastra as duas URLs na Z-API
  (`zapi.Webhooks`, caminho em constante — a conferir); `POST
  /v1/config/whatsapp/webhooks` recadastra; Console mostra o estado e
  "Cadastrar de novo". ADR 0007.
- T-121 a T-136 passando (T-128 no Postgres escrito, `make itest` pendente);
  sentinel: aprovado com ressalvas.

## 2026-10-28 — Decisão: origem fixa do Console é o túnel na Hostinger
- O escritório não sai mais pelo `177.69.185.193`. O dono usa uma máquina
  própria e exclusiva na Hostinger só como túnel para ter IP fixo:
  `72.60.159.29` (já liberado no `keycloak-allowlist` do host; falta no
  repositório `flip`). Esse IP passa a ser a lista `escritorio` da Cloudflare,
  e `ominicore.flipapi.com.br` e `notify.flipapi.com.br` passam pelo túnel.
- Achado: a regra de WAF não bloqueava — de `186.204.15.26` (fora da lista) o
  Console abriu (200) e `/v1/config/*` chegou à API (401). A conferir e ativar
  pelo dono; reteste pelo celular no 4G (403 esperado).

## 2026-10-27 — Change 12 em produção
- `a9a3bf9` publicado pelo dono (`HOST=aviva publicar.sh prod`, túnel SSM na
  porta 2222 — o SSH público não aceita IP fora do escritório). Conferido no
  host: `RestartUSec=10s` e `StartLimitIntervalUSec=0` em api e worker.
- Console publicado (`wrangler deploy`, versão 76eed4ae) com "Tentar de novo".
- Achado: healthcheck sempre "unhealthy" — `HealthCmd` de texto vira
  `CMD-SHELL` e a imagem distroless não tem `/bin/sh`. Corrigido em `45a39b3`
  (forma de array, validada no gerador do quadlet em dry-run); publicado em
  `ac762e0` (por Claude, com o "pode" do dono): api e worker "healthy".
- Achado: dois `publicar.sh prod` sobrepostos cortaram o
  `/opt/ominicore/ominicore.anterior` (segfault; o `--voltar` derrubaria a API).
  Recriado pelo dono a partir do binário atual; melhoria do script no `to-do.md`.
- Achado: de fora do escritório (186.204.15.26) o Console abriu (200) e
  `/v1/config/*` chegou à API (401): o WAF não bloqueou — a confirmar com o dono
  se esse IP está na lista `escritorio`.

## 2026-10-27 — Change `12-defeitos-prod` implementada (11/12; falta publicar e arquivar)
- Desligar o WhatsApp: o KV v2 só marca a versão como apagada (soft delete), e
  o `LerWhatsApp` a lia como "incompleta: sem o campo token" (o erro de prod).
  Agora versão apagada ou sem dados = Canal desligado. O `baofake` passou a
  simular o soft delete como o real (antes apagava de verdade, por isso o
  defeito passou); o T-98 ganhou teste no cofre e de integração com o OpenBao
  2.6.3 real — `make itest` passando (rodado pelo dono no Docker Desktop).
- Quadlet: `RestartSec=10s` e `StartLimitIntervalSec=0` — religa para sempre.
- Console: API sem resposta (status 0) ou 5xx mostra a mensagem e "Tentar de
  novo"; carregamento termina em `finally` (REQ-43, T-137 e T-138).

## 2026-10-27 — Decisão: Avisos de status da Z-API (sabatina fechada; change 13)
- Eventos: "Ao enviar" (DeliveryCallback) com `error` → Tentativa falhou
  ("Phone number does not exist"/"invalid phone number" = contato inválido; o
  resto = recusado pelo Provedor) e fallback; "Status" (MessageStatusCallback)
  RECEIVED → entregue, READ/PLAYED → lida; SENT e READ_BY_ME ignorados. "Ao
  receber" (RN-6), conexão e presença fora.
- Autenticação: segredo de 32 bytes na URL (`/webhooks/zapi/<segredo>`) +
  `instanceId` igual ao configurado; senão 404 sem efeito. Aviso autêntico →
  200 `{"value": true}` sempre (desconhecido, repetido ou atrasado só vai para
  o log). RN-4 e RN-5 revisadas na especificação.
- Segredo novo a cada gravação da Z-API; o OminiCore cadastra as duas URLs na
  Z-API pela API dela (caminho a conferir com a instância real); falha no
  cadastro não desfaz a gravação, e o Console mostra o estado e "Cadastrar de
  novo". ADR 0007 a escrever na change.
- Ordem: change 12 (defeitos de produção) antes; esta é a 13. UltraMsg fica fora
  (to-do).

## 2026-10-27 — Change `ajuste-texto-notificacoes` arquivada (só redação)
- O OpenSpec 1.15.0 passou o aviso de texto de requisito acima de 500
  caracteres de INFO para WARNING, e o `validate --all --strict` (CI,
  pre-commit, hook Stop) reprovava `spec/notificacoes`.
- REQ-2 (pedido por Template) e REQ-14 (consulta) reescritos mais enxutos (493
  e 480 caracteres), com as mesmas obrigações e os mesmos cenários (T-4 a T-6,
  T-15 a T-18, T-43, T-44, T-56), sem REQ novo. Separar em REQ novos não deu: o
  OpenSpec 1.15 recusa MODIFIED que larga cenários.
- `openspec validate --all --strict`: 7/7.

## 2026-10-06 — Decisão: primeiro deploy direto em prod, hml adiado
- O dono decidiu publicar o OminiCore e o Console direto em prod
  (`notify.flipapi.com.br`, realm `ominicore`, cofre `kv/ominicore`, schema
  `ominicore`), sem passar por hml, para testar o Console; ainda não há Sistema
  de origem usando o OminiCore. Contraria o "hml primeiro" do contrato, agora
  anotado em `Infra/AMBIENTES.md` §1. Recusada a alternativa de usar o nome
  `notify` com os dados de hml e trocar depois (mistura de Ambientes).
- O Console publicado lista só Produção (`ambientes-publicados.json`) até o
  hml existir. O wrangler usa *route* sobre o registro DNS `ominicore` criado
  no painel, em vez de `custom_domain`.

## 2026-10-05 — Fatia 11 (`11-console-operador`) implementada: 42/45
- API: `portas.Identidade` (client, pessoa, papéis); CORS só em `/v1/config/*`
  para `OMINICORE_CONSOLE_ORIGEM`; token de `OMINICORE_CONSOLE_CLIENT` recusado
  no envio e no Opt-out; troca pelo Console registra o `preferred_username`.
- Keycloak: `lib-console.sh` (client público PKCE S256 + fluxo `navegador-otp`
  com OTP obrigatório por override; a atualização das execuções preserva a
  `priority`, senão o OTP vai para antes da senha — achado no login local) e
  `criar-console.sh`/`criar-operador.sh`; o Keycloak 26 pede nome e sobrenome
  no primeiro login.
- Console `web/console`: Materio free 2.3.0 enxuto, Vite 6, Vuetify 3.13,
  `oidc-client-ts`, `vue-i18n` pt-BR, CSP em `_headers`, Worker só com assets.
- Borda: rota `/v1/config/` no Traefik, `verificar-exposicao.sh` (T-113 a T-115,
  CI de fora do escritório), runbook C1–C8, ADR 0006. Fatos: a 443 do host só
  aceita a Cloudflare; o `auth.flipapi.com.br` tem `ipAllowList` própria.
- CVE (política em `Infra/seguranca-dependencias.md`): `bun audit` acusa 6
  altas em `minimatch`, `braces` e `extract-zip`, todas em ferramentas de build
  e lint do Console (ESLint 8, sass/chokidar, `@iconify/tools`), sem correção
  compatível publicada (`braces` 3.0.3 e `extract-zip` 2.0.1 são as últimas).
  Mitigação: uso não exposto (rodam só no build local e no CI, sobre arquivos
  do próprio repositório); nada disso vai no bundle publicado. Revisitar na
  migração para o ESLint 9 ou quando o Dependabot abrir PR (dono: Mauricio Miwa).
- Pendentes: 1.2 (archive da 10), 7.3 (T-113 a T-115 só depois do deploy).

## 2026-10-05 — Sabatina do Console do Operador fechada; change `11-console-operador` proposta
- Respostas do dono: só WhatsApp; template Materio free (Vuexy é pago);
  `/v1/config/*` público em `notify[-hml].flipapi.com.br` (opção a, não BFF),
  WAF só do IP fixo do escritório + JWT + CORS para
  `https://ominicore.flipapi.com.br`; client público `ominicore-console`
  (PKCE, token em memória); seletor escolhe o Ambiente (hml/prod), não
  cliente; conta por pessoa com OTP; só pt-BR. Termos novos no glossário:
  Console do Operador, Ambiente. A abrir: RN-12, ADR 0006 (revisa a 0003),
  change. Configuração por cliente (W Bem, Lara) fica fora.
- Na proposta: token do client do Console recusado no envio e no Opt-out
  (REQ-36); `config_eventos` registra o usuário da pessoa nas trocas pelo
  Console (REQ-27 modificado); OTP obrigatório por override de fluxo no client.

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

