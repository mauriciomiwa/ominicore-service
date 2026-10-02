# Log arquivado — 2026

Entradas antigas do `log.md`, mais recente no topo.

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
