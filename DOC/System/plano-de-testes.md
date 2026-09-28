# Plano de testes — ominicore

Um teste por cenário marcado com T-n nas specs do OpenSpec
(`#### Scenario: T-n …` em `openspec/specs/` ou nos deltas de
`openspec/changes/`). A spec diz *o que* verificar e a qual REQ o T-n
pertence; aqui fica só *como* e *em que pé está* — sem repetir o REQ, para não
haver dois lugares a manter. Tipo: **Exemplo** (caso concreto) ou
**Propriedade** (invariante). Status: pendente / escrito (não rodado) /
passando.

Nome do teste em Go cita o T-n: `func TestT7_ReenvioComAMesmaChaveNaoDuplica(t *testing.T)`.
Unitários com fakes (`internal/testes/`: repositório em memória, realm JWT,
SendGrid via `httptest`); integração em `internal/integracao/` com
`testcontainers-go` (Postgres real) — pula sozinho sem runtime de containers;
rode com `make itest` (Podman). Provedores reais nunca entram em teste.

| Teste | Tipo | Onde (arquivo do teste) | Status |
|---|---|---|---|
| T-1 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-2 | Exemplo | `internal/adaptadores/saida/keycloak/validador_test.go` | passando |
| T-3 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-4 | Exemplo | `internal/aplicacao/aplicacao_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake) · escrito, não rodado (Postgres) |
| T-5 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-6 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-7 | Exemplo | `internal/integracao/integracao_test.go` | escrito, não rodado |
| T-8 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-9 | Exemplo | `internal/integracao/integracao_test.go` | escrito, não rodado |
| T-10 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake) · escrito, não rodado (Postgres) |
| T-11 | Exemplo | `internal/integracao/integracao_test.go` | escrito, não rodado |
| T-12 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio) · escrito, não rodado (Postgres) |
| T-13 | Exemplo | `cmd/ominicore/main_test.go` | passando |
| T-14 | Propriedade | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
