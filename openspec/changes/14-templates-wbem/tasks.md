# Tasks

## 1. Negócio

- [x] 1.1 RN-13 em `DOC/Negocio/especificacao-negocio.md`, Fatia 14 em `DOC/System/backlog-mvp.md` e decisão no `DOC/log.md` — verificar: `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo

## 2. Templates (REQ-44)

- [x] 2.1 Migration `00011_templates_wbem.sql` (up e down) com `wbem-link-tomador` e `wbem-otp-tomador`, só a versão `whatsapp` em texto — verificar: `go test ./internal/integracao/` aplica as migrations no Postgres
- [x] 2.2 Teste T-139 em `internal/integracao/` contra o Postgres: os dois Templates preenchem o texto com `link` e `codigo`, Ordem de Canais só WhatsApp, texto servido à Z-API e ao UltraMsg — verificar: `go test -run T139 ./internal/integracao/`
- [x] 2.3 Teste T-140 em `internal/aplicacao/` (ou integração): sem `codigo` ou pedindo SMS, 422 e nada registrado — verificar: `go test -run T140 ./...`
- [x] 2.4 T-139 e T-140 em `DOC/System/plano-de-testes.md` — verificar: `python3 DOC/tools/checa_projeto.py .` sem furo

## 3. Entrega

- [x] 3.1 (dono) Publicar a API no host (a migration roda no boot ou pelo runbook) nos schemas de hml e prod — verificar: `POST /v1/notificacoes` com `wbem-otp-tomador` para o celular do dono, de dentro da rede `aviva`, dá 202 e a mensagem chega

## 4. Fechamento

- [x] 4.1 Rodar `python3 DOC/tools/valida_rastreabilidade.py DOC` — verificar: sem furo
- [x] 4.2 Auditar o diff com a skill sentinel e registrar com `sdd_hooks.py auditoria` — verificar: veredito registrado
