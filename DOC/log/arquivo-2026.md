# Log arquivado — 2026

Entradas antigas do `log.md`, mais recente no topo.

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

## 2026-09-29 — Fatia 1 verificada contra serviços reais (30/30)
- Docker ligado: `make itest` 7/7 contra Postgres 17 (T-4, T-7 com 10 reenvios
  concorrentes, T-9, T-10, T-11 ~0,65 s, T-12, janela de 24 h).
- Ambiente local de ponta a ponta (`COMPOSE="docker compose" make compose-up`):
  token real do Keycloak → 202; mesma chave → mesmo id + `Idempotent-Replayed`;
  sem token → 401; worker enviou (SendGrid falso local, para não chamar a API
  real) e gravou `enviada`.
- Login `jwt` no OpenBao via Keycloak (ADR 0002) provado: `ominicore-app` entra
  e lê `kv/ominicore/*`; `sistema-local` é recusado (`aud` ≠ `openbao`).
- Achados só visíveis rodando, corrigidos: arquivo de realm precisa do nome do
  realm (`ominicore-realm.json`); `bound_claims` só entra como JSON; OpenBao
  valida o JWKS ao gravar (retry no bootstrap); Redis local na 6380 (6379 é de
  outro projeto); `make itest` prefere o Docker; audience mapper `openbao` no
  client `ominicore-app` (hml/prod devem espelhar).

## 2026-09-29 — Contrato de ambientes (`Infra/AMBIENTES.md`) fechado
- Dono aceitou as recomendações: realms `ominicore`/`ominicore-hml` em
  `auth.flipapi.com.br`; webhooks em `notify(-hml).flipapi.com.br/webhooks/*`;
  API interna `ominicore-api(-hml):8080`; schema `ominicore_hml` no mesmo
  Aurora; cofre `kv/ominicore-hml/*` × `kv/ominicore/*`; só o dono escreve
  segredo de prod; hml só com dado sintético; promove-se o mesmo binário.
- Código: `OMINICORE_KC_URL` (JWKS e token pelo backchannel `http://keycloak:8080`,
  porque o host público tem allowlist) e `OMINICORE_BAO_PREFIXO` (isola hml de
  prod no mesmo OpenBao), com testes.

## 2026-09-28 — Fatia 1 (`01-enviar-email`) implementada, verificação parcial
- Código hexagonal em Go 1.27 (módulo `ominicore`): domínio, casos de uso
  SendMessage/Despachar, adaptadores HTTP, worker, Postgres (pgx + sqlc + goose,
  fila `SKIP LOCKED`), SendGrid, Keycloak (JWT/JWKS com cache Redis e releitura
  limitada a 1/min), OpenBao (falha fechada), logger `slog` com redação.
- Dependência nova além da sabatina: `golang-jwt/jwt/v5` (a mesma do aviva-back).
- Verde: `go test -race ./...`, `golangci-lint` (0 issues), `govulncheck`,
  `sqlc diff`, `redocly lint`, `openspec validate --strict`, rastreabilidade.
  Passando: T-1 a T-6, T-8, T-10, T-12, T-13, T-14 (unitários/fakes).
- **Não verificado:** T-7, T-9, T-11 e as versões Postgres de T-4/T-10/T-12
  estão escritos mas pulam — sem runtime de containers (VM do Podman quebrada,
  Docker parado). `compose` local e login `jwt` no OpenBao nunca rodaram.
  Tarefas correspondentes ficaram desmarcadas no `tasks.md`.
- Imagens locais fixadas nas versões do aviva2.0 (Keycloak 26.3, OpenBao 2.6.3,
  Redis 7.4). `Infra/AMBIENTES.md` criado (só local preenchido).

## 2026-09-28 — Kickoff
- Estrutura DOC/ criada (scaffold, porte: padrao — produto com vários
  Sistemas de origem, vida longa e dados pessoais; não multi-tenant nem
  multi-app). `openspec init --tools claude,agents` (OpenSpec 1.13.2).
- **Inventário:** pasta greenfield (vazia). Local: `go1.27.1 darwin/arm64`,
  `openspec`, `docker`; CLI `bao` ausente. Infra da casa em
  `../flip/infra/aviva2.0/`: host EC2 arm64 com quadlets Podman (Traefik,
  Keycloak, OpenBao raft + unseal KMS, Redis) na rede `aviva`; Keycloak e
  aviva-back usam o Aurora `db-zillionvf-prod` (sa-east-1), um schema por app.
  Molde de deploy: `../flip/aviva-back/deploy/aviva2/publicar.sh`.
- **Negócio (Rodada 1):** serviço interno de notificações; fallback pela Ordem
  de Canais do pedido/Template, retry só em Falha temporária; webhooks do MVP =
  só status; ~100 mil/dia, entrega > 98%, 1ª Tentativa < 5 s; retenção de 30
  dias configurável; credenciais só no OpenBao; MVP com Email + SMS (WhatsApp
  depois); Templates no OminiCore; Chave de idempotência obrigatória (24 h);
  Opt-out no OminiCore; janela de horário/agendamento depois. RN-1 a RN-10.
- **Desenvolvimento (Rodada 2):** Go 1.27 hexagonal; auth dos Sistemas de
  origem pelo Keycloak existente (realm `ominicore`, `client_credentials`, JWT
  por JWKS); entrada só REST, fila no Postgres (ADR 0001), Redis para rate
  limit e cache; Podman — quadlets no aviva2.0 com binário compilado, local
  `podman compose` (Nomad descartado); só `/webhooks/*` público (ADR 0003);
  schema `ominicore` no Aurora `new_aviva`; login no OpenBao via JWT do
  Keycloak (ADR 0002); `slog` + Prometheus; retry 3× (5 s/30 s/2 min);
  `net/http`, `pgx`+`sqlc`, `goose`, `go-redis`, `testcontainers-go`,
  OpenAPI primeiro, GitHub Actions + Dependabot.
- Backlog: 5 fatias + trilha de infraestrutura. Aberta só a change
  `01-enviar-email` (REQ-1 a REQ-6, T-1 a T-14); as demais dependem dela.
