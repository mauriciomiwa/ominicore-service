# Tasks

## 1. Negócio

- [x] 1.0 RN-15 em `DOC/Negocio/especificacao-negocio.md` (antecipada no propose, para a rastreabilidade)
- [x] 1.1 Fatias 17, 18 e 19 em `DOC/System/backlog-mvp.md` (a 15 depende da 18), decisões da sabatina de 2026-11-14 (P1 a P10) no `DOC/log.md` e `System/arquitetura.md` com a Organização — verificar: `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo

## 2. Domínio, banco e porta

- [x] 2.1 `dominio`: tipo `Organizacao` e erro `ErrSemOrganizacao`; `Notificacao.Organizacao` — verificar: `go test ./internal/dominio/`
- [x] 2.2 Migration `00013_organizacoes.sql` (tabelas, `wbem`, colunas com default `wbem` (compatível com o código anterior durante o deploy; sai na fatia 18), índice, RLS com `FORCE`; down reverte) — verificar: `go test ./internal/integracao/` aplica up e down
- [x] 2.3 `queries.sql`: `OrganizacaoDoSistema`, `InserirNotificacao` com `organizacao`, e `sqlc generate` — verificar: `sqlc diff` sem diferença e `go build ./...`
- [x] 2.4 Porta `OrganizacoesPort` e implementação no Postgres com cache de 60 s; helper `naOrganizacao(tx, org)` (`set_config(..., true)`) usado por `RegistrarNotificacao`, `ConsultarNotificacao` e pela busca por chave — verificar: `go test ./internal/adaptadores/saida/postgres/`

## 3. Aplicação e API (REQ-50, REQ-51, REQ-52)

- [x] 3.1 Middleware HTTP resolve a Organização do `azp` depois do `soSistemaDeOrigem`; 403 "Sistema de origem sem Organização" nas rotas de Notificação e Opt-out; casos de uso recebem a Organização; wiring em `cmd/ominicore/main.go`; fakes de teste com `org-a` e `org-b` — verificar: `go test ./internal/...`
- [x] 3.2 `api/openapi.yaml` com o 403 nas rotas de Notificação e Opt-out — verificar: `redocly lint api/openapi.yaml`
- [x] 3.3 Teste de arquitetura: nenhuma leitura de `notificacoes` feita pela API fora do `naOrganizacao` — verificar: `go test ./internal/adaptadores/saida/postgres/ -run Arquitetura`
- [x] 3.4 Teste T-159 (client sem Organização: 403 em pedido, consulta, registro e remoção de Opt-out; nada criado nem despachado) — verificar: `go test -run T159 ./...`
- [x] 3.5 Teste T-160 (client `wbem` aceito com o contrato de hoje; Notificação na Organização `wbem`) — verificar: `go test -run T160 ./...`
- [x] 3.6 Teste T-161 contra o Postgres (dados gravados na `00012`, sobe para a `00013`: tudo em `wbem`, status, Tentativas e conteúdo iguais; client `wbem` vinculado) — verificar: `go test -run T161 ./internal/integracao/`
- [x] 3.7 Teste T-162 contra o Postgres (down para a `00012`: Notificações, Tentativas e Templates iguais aos de antes) — verificar: `go test -run T162 ./internal/integracao/`
- [x] 3.8 Teste T-163 contra o Postgres (transação fixada em `org-b` não encontra Notificação da `wbem` por id, sem filtro de Sistema de origem) — verificar: `go test -run T163 ./internal/integracao/`
- [x] 3.9 Teste T-164 (`sistema-b` da `org-b` consulta Notificação do `wbem`: não encontrada) — verificar: `go test -run T164 ./...`
- [x] 3.10 Teste T-165 contra o Postgres (worker despacha pendentes da `wbem` e da `org-b`; Aviso de entrega marca a Tentativa da `org-b`) — verificar: `go test -run T165 ./internal/integracao/`

## 4. Change 15 e ambientes

- [x] 4.1 Renumerar na change `15-fila-zapi` a migration para `00014_config_eventos_fila.sql` (tasks e design) — verificar: `grep -rnE "0001[23]_" openspec/changes/15-fila-zapi` vazio
- [ ] 4.2 (dono, skill `ambientes`) Consulta só de leitura dos `sistema_origem` distintos dos últimos 30 dias em hml e prod; client além de `wbem` entra na migration antes do deploy — verificar: resultado anotado no `log.md`

## 5. Fechamento

- [x] 5.1 Checagem completa: `go test -race ./...`, `golangci-lint run`, `govulncheck ./...`, `openspec validate --all --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 5.2 Auditar o diff com a skill `sentinel` e registrar o veredito (`python3 DOC/tools/sdd_hooks.py auditoria`)
