# Tasks

## 1. Negócio

- [x] 1.1 RN-14 em `DOC/Negocio/especificacao-negocio.md`, Fatia 16 no `DOC/System/backlog-mvp.md`, decisão no `DOC/log.md` e renumeração da `15-fila-zapi` no `DOC/to-do.md` — verificar: `python3 DOC/tools/checa_projeto.py .` sem furo

## 2. Modelos (REQ-49)

- [x] 2.1 Migration `00012_modelos_regua_wbem.sql` (up e down) com os 16 Templates — verificar: `go test ./internal/integracao/` aplica as migrations no Postgres
- [x] 2.2 Testes T-157 e T-158 em `internal/integracao/` — verificar: `go test -run 'T157|T158' ./internal/integracao/`
- [x] 2.3 T-157 e T-158 em `DOC/System/plano-de-testes.md` — verificar: `checa_projeto.py` sem furo

## 3. Entrega

- [ ] 3.1 (dono) Publicar a API (migration `00012`) nos schemas de hml e prod — verificar: `POST /v1/notificacoes` com `wbem-open-finance-confirmado` para o celular do dono, de dentro da rede `aviva`, dá 202 e a mensagem chega

## 4. Fechamento

- [x] 4.1 Auditar o diff com a skill sentinel e registrar com `sdd_hooks.py auditoria` — verificar: veredito registrado
