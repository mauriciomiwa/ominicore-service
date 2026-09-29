# Stack tecnológico — ominicore

Fonte de verdade das **versões** = manifesto/lockfile do repositório (`go.mod`,
quando existir), não este doc. Cite ao lado de cada item o arquivo que o
originou (inventário do kickoff) — assim ninguém precisa adivinhar se foi
decisão ou chute. "Sabatina" = confirmado pelo usuário no kickoff de 2026-09-28.

## Stack

| Camada | Tecnologia | Origem |
|---|---|---|
| Linguagem | Go 1.27 (`go 1.27.0` no `go.mod`; local `go1.27.1 darwin/arm64`); módulo `ominicore` (trocar pelo caminho do GitHub quando o repo existir) | `go.mod` |
| Arquitetura | Hexagonal (ports and adapters): domínio sem dependência de infra | sabatina + diagrama `gemini-code-1790619852596.html` |
| HTTP | `net/http` da stdlib (roteamento com método e parâmetro, Go 1.22+) | sabatina (P14) |
| Contrato | OpenAPI 3.1 escrito antes do código (`api/openapi.yaml`) | sabatina (P14) |
| Banco | PostgreSQL — schema `ominicore` com role próprio no Aurora `db-zillionvf-prod` (`new_aviva`, sa-east-1) | sabatina (P12c) + `flip/infra/aviva2.0/keycloak/keycloak.env` |
| Acesso a banco | `pgx/v5` + `sqlc` (SQL tipado, sem ORM) | sabatina (P14) |
| Migrations | `goose` | sabatina (P14) |
| Fila de despacho | tabela no Postgres com `FOR UPDATE SKIP LOCKED` (ADR 0001) | sabatina (P11) |
| Cache / rate limit | Redis (`go-redis/v9`): rate limit por Provedor, cache de Template e de JWKS | sabatina (P11) + `flip/infra/aviva2.0/quadlet/redis.container` |
| Segredos | OpenBao, KV v2 em `kv/ominicore/*`; login pelo método `jwt` com o Keycloak (ADR 0002); client `github.com/openbao/openbao/api/v2` | sabatina (P13) + `flip/infra/aviva2.0/openbao/config.hcl` |
| Validação de JWT | `github.com/golang-jwt/jwt/v5` (mesma do aviva-back); JWKS lido e guardado no Redis pelo próprio adaptador | `go.mod` + `../flip/aviva-back/go.mod` |
| Autenticação | Keycloak, realm `ominicore` / `ominicore-hml`; Sistema de origem = client confidencial com `client_credentials`; JWT validado pelo JWKS (`iss`, `aud ∋ ominicore`, `exp`); Sistema de origem = `azp` | sabatina (P10) + `flip/infra/aviva2.0/keycloak/gerar-realms.py` |
| Provedores | SendGrid (Email), Twilio (SMS: API Key + Messaging Service, API de Messages form-encoded); regra comum de classificação em `internal/adaptadores/saida/provedor`; WhatsApp/Meta depois | sabatina (P6) + sabatina da fatia 02 (P4) |
| Logs / métricas | `log/slog` JSON no stdout; `/metrics` Prometheus | sabatina (P15) |
| Testes | `testing` + `testcontainers-go` (Postgres, Redis, OpenBao) na integração | sabatina (P14) |
| Execução | um binário, dois subcomandos: `ominicore api` e `ominicore worker` | sabatina (P12) |
| Local | `podman compose` com Postgres, Redis, OpenBao (dev) e Keycloak | sabatina (P12) |
| Deploy | quadlets Podman no host aviva2.0 (EC2 arm64, us-east-1), binário `CGO_ENABLED=0 GOARCH=arm64` montado em `distroless/static`, atrás do Traefik | sabatina (P12a) + `flip/aviva-back/deploy/aviva2/publicar.sh` |
| Exposição | só `/webhooks/*` público (Traefik + Cloudflare, assinatura do Provedor validada); API de envio só na rede `aviva` (ADR 0003) | sabatina (P12b) |
| Repositório / CI | GitHub + GitHub Actions + Dependabot | sabatina (P14) |

## Verificação automática

| Verificação | Ferramenta (desta stack) | Onde roda |
|---|---|---|
| Formatação | `gofumpt` + `goimports` (via `golangci-lint fmt`; `module-path: ominicore`) | `.golangci.yml`, CI |
| Lint geral | `golangci-lint` (`govet`, `staticcheck`, `errcheck`, `gosec`) | pre-commit, CI |
| Tamanho e complexidade (Clean Code) | `golangci-lint`: `funlen` (60 linhas), `gocyclo` (15), `gocognit` | CI |
| Duplicação (DRY) | `golangci-lint`: `dupl` | CI |
| Fronteira hexagonal (domínio não importa infra) | `golangci-lint`: `depguard` | CI |
| Vulnerabilidades | `govulncheck` | CI |
| Código SQL gerado em dia | `sqlc diff` | CI |
| Contrato | `redocly lint api/openapi.yaml` (`make contrato`) | CI (`.github/workflows/go.yml`) |

## Decisões em aberto
- **Cadastro de Templates no MVP:** seed por migration (proposto na fatia 1) ou
  API de administração? — ver `to-do.md`.
- **Latência Aurora sa-east-1 ↔ host us-east-1** (~120 ms por ida e volta):
  aceita no MVP; revisitar quando o Aurora migrar para a Virgínia (pendência já
  registrada no `keycloak.env`).
- Versões exatas: `go.mod` (pgx v5.11, goose v3.28, go-redis v9.22, openbao api v2.7, golang-jwt v5.3, testcontainers v0.44).
- `/metrics` Prometheus e rate limit por Provedor no Redis: previstos na stack, ainda sem REQ — entram com a fatia que os exigir.
