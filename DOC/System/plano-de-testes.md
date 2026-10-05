# Plano de testes — ominicore

Um teste por cenário marcado com T-n nas specs do OpenSpec
(`#### Scenario: T-n …` em `openspec/specs/` ou nos deltas de
`openspec/changes/`). A spec diz *o que* verificar e a qual REQ o T-n
pertence; aqui fica só *como* e *em que pé está* — sem repetir o REQ, para não
haver dois lugares a manter. Tipo: **Exemplo** (caso concreto) ou
**Propriedade** (invariante). Status: pendente / escrito (não rodado) /
passando.

T-1 a T-62 vigentes desde 2026-10-02 e T-63 a T-69 desde 2026-10-05 (`openspec/specs/{autenticacao-origem,notificacoes,segredos}`).

Nome do teste em Go cita o T-n: `func TestT7_ReenvioComAMesmaChaveNaoDuplica(t *testing.T)`.
Unitários com fakes (`internal/testes/`: repositório em memória, realm JWT,
SendGrid via `httptest`); integração em `internal/integracao/` com
`testcontainers-go` (Postgres real) — pula sozinho sem runtime de containers;
rode com `make itest` (Docker Desktop, Colima ou Podman: usa o socket do contexto ativo). Última execução: 2026-10-02, 30/30 contra Postgres 17 (Colima). Provedores reais nunca entram em teste.

| Teste | Tipo | Onde (arquivo do teste) | Status |
|---|---|---|---|
| T-1 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-2 | Exemplo | `internal/adaptadores/saida/keycloak/validador_test.go` | passando |
| T-3 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-4 | Exemplo | `internal/aplicacao/aplicacao_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-5 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-6 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-7 | Exemplo | `internal/integracao/integracao_test.go` | passando |
| T-8 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-9 | Exemplo | `internal/integracao/integracao_test.go` | passando |
| T-10 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-11 | Exemplo | `internal/integracao/integracao_test.go` | passando |
| T-12 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-13 | Exemplo | `cmd/ominicore/main_test.go` | passando |
| T-14 | Propriedade | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-15 | Exemplo | `internal/aplicacao/aplicacao_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-16 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-17 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-18 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-19 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-20 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` · `internal/adaptadores/saida/twilio/twilio_test.go` | passando |
| T-21 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-22 | Propriedade | `internal/dominio/dominio_test.go` | passando |
| T-23 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres, relógio adiantado) | passando (domínio e Postgres) |
| T-24 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-25 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-26 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-27 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-28 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-29 | Exemplo | `internal/adaptadores/saida/whatsapp/whatsapp_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-30 | Exemplo | `internal/adaptadores/saida/whatsapp/whatsapp_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-31 | Exemplo | `internal/adaptadores/saida/whatsapp/whatsapp_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-32 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-33 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-34 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-35 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-36 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-37 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-38 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-39 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-40 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-41 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` · `internal/integracao/integracao_test.go` (Postgres; + 8 avisos simultâneos = 1 fallback) | passando (fake e Postgres) |
| T-42 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-43 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-44 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-45 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-46 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-47 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-48 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-49 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-50 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-51 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-52 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-53 | Exemplo | `internal/aplicacao/aplicacao_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-54 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-55 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-56 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-57 | Exemplo | `internal/aplicacao/aplicacao_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-58 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-59 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-60 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-61 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-62 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-63 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` · `cmd/ominicore/main_test.go` | passando |
| T-64 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` | passando |
| T-65 | Exemplo | `internal/aplicacao/desligado_test.go` · `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-66 | Exemplo | `internal/aplicacao/desligado_test.go` | passando |
| T-67 | Exemplo | `internal/aplicacao/desligado_test.go` | passando |
| T-68 | Exemplo | `internal/aplicacao/desligado_test.go` | passando |
| T-69 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
