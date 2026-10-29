# Tasks

## 1. Pré-requisitos e documentação de origem
- [x] 1.1 RN-4 e RN-5 revisadas em `DOC/Negocio/especificacao-negocio.md`, Fatia 13 em `DOC/System/backlog-mvp.md`, decisões no `DOC/log.md` (feitos na sabatina) — `python3 DOC/tools/checa_projeto.py .` sem furo
- [x] 1.2 Conferir que a change 12 está implementada e que a 08, a 09, a 10 e a 11 estão arquivadas (o REQ-21 e as capacidades `configuracao-canais` e `console-operador` vigentes) — `openspec validate 13-webhooks-zapi --strict` sem o aviso de "Archive would refuse"
- [x] 1.3 ADR `DOC/System/decisoes/0007-aviso-sem-assinatura-por-segredo-na-url.md` (revisa a RN-5 nos gateways sem assinatura; risco aceito; alternativas) e status da 0003 apontando para ela — arquivo presente e citado no `DOC/log.md`

## 2. Fatos da Z-API (antes do código)
- [x] 2.1 Com a instância real de prod, conferir que existem as rotas que cadastram as URLs de "Ao enviar" e de "Status" e registrar em `DOC/Padroes/fontes-externas.md` — feito em 2026-10-29: `update-webhook-delivery` e `update-webhook-message-status` respondem 405 a GET (existem), caminho inexistente responde 200

## 3. Domínio e tradução dos avisos
- [x] 3.1 `dominio`: `webhook_segredo` em `CamposOpcionaisWhatsApp[WhatsAppZAPI]` e em `segredosWhatsApp`; `webhooks_estado` e `webhooks_em` como campos públicos opcionais da Z-API — `go test ./internal/dominio/`
- [x] 3.2 Campos `webhook_segredo`, `webhooks_estado` e `webhooks_em` vindos no corpo do `PUT` são descartados — `go test ./internal/aplicacao/`
- [x] 3.3 `webhooks.avisosZAPI` (função pura: `MessageStatusCallback` e `DeliveryCallback` → `[]dominio.AvisoStatus`, categoria pelo texto do erro, motivo sem telefone, até 200 caracteres) — testes em tabela em `go test ./internal/adaptadores/entrada/webhooks/`

## 4. Rota `/webhooks/zapi/{segredo}`
- [x] 4.1 Handler: Z-API ativa, segredo em tempo constante, `instanceId` igual ao configurado, senão 404; corpo até 64 KiB; aviso autêntico → `Receber` e 200 `{"value": true}`; erro de banco → 500; nada do segredo em log — `go test ./internal/adaptadores/entrada/webhooks/`
- [x] 4.2 Teste T-121 (segredo e instância certos: 200 e Tentativa entregue)
- [x] 4.3 Teste T-122 (segredo errado, inclusive o anterior a uma regravação: 404, nada muda)
- [x] 4.4 Teste T-123 (outra instância: 404, nada muda)
- [x] 4.5 Teste T-124 (UltraMsg, Meta ou WhatsApp desligado: 404)
- [x] 4.6 Teste T-125 (mensagem desconhecida ou status repetido: 200, nada muda)
- [x] 4.7 Teste T-126 (`READ` deixa a Tentativa lida)
- [x] 4.8 Teste T-127 (`SENT` não muda nada, 200)
- [x] 4.9 Teste T-128 ("Phone number does not exist" → contato inválido e Tentativa pendente no SMS), também em `internal/integracao/` contra o Postgres
- [x] 4.10 Teste T-129 (shadow ban → recusado pelo Provedor)
- [x] 4.11 Teste T-130 ("Ao enviar" sem `error` não muda nada, 200)
- [x] 4.12 REQ-21 modificado: rodar de novo T-81 e T-91 (`/webhooks/whatsapp` segue 404 com UltraMsg e Z-API) — `go test ./internal/adaptadores/entrada/webhooks/`

## 5. Cadastro e recadastro
- [x] 5.1 Porta `portas.CadastroWebhooksPort` e adaptador `zapi.Webhooks.Cadastrar` (método e caminho da 2.1, `Client-Token`, timeout de 10 s, motivo mascarado) testado contra o `zapifake` (que passa a registrar as URLs cadastradas) — `go test ./internal/adaptadores/saida/zapi/`
- [x] 5.2 `ConfigurarUseCase.Gravar` com a Z-API: segredo novo (`crypto/rand`, 32 bytes, base64url), grava, aplica, registra o evento, cadastra, regrava o estado; sem `OMINICORE_WEBHOOK_URL` não tenta e fica `falhou` — `go test ./internal/aplicacao/`
- [x] 5.3 `ConfigurarUseCase.Recadastrar` e rota `POST /v1/config/whatsapp/webhooks` (Operador; 409 sem Z-API; evento `webhooks` em `config_eventos`); `api/openapi.yaml` com a rota e o estado na consulta — `go test ./internal/adaptadores/entrada/http/` e `make contrato`
- [x] 5.4 Teste T-131 (gravar cadastra as duas URLs com o mesmo segredo novo; o anterior recebe 404; consulta mostra cadastrados)
- [x] 5.5 Teste T-132 (falha no cadastro: Z-API gravada e ativa, estado `falhou` na resposta e na consulta, log sem segredo nem token)
- [x] 5.6 Teste T-133 (consulta mostra `webhook_segredo` só como configurado, sem nenhum caractere do valor)
- [x] 5.7 Teste T-134 (recadastro com o segredo atual, que continua valendo)
- [x] 5.8 Teste T-135 (recadastro com UltraMsg: 409, nada chamado)
- [x] 5.9 Rodar de novo T-92 a T-105 (configuração do WhatsApp inalterada para Meta e UltraMsg) — `go test -race ./...`
- [x] 5.10 Migration `00010_config_eventos_webhooks.sql` (o CHECK de `config_eventos.acao` aceita `webhooks`; achado em prod em 2026-10-29: o recadastro funcionava, mas o registro falhava e a API respondia 500) e T-104 no Postgres com a ação `webhooks` — `make itest`

## 6. Console
- [x] 6.1 Cliente da API (`recadastrarWebhooks`) e tela do WhatsApp com o estado dos webhooks (Z-API ativa) e o botão "Cadastrar de novo"; textos em `pt-BR.json` — `bun run lint` e `bun run typecheck`
- [x] 6.2 Teste T-136 (não cadastrados → "Cadastrar de novo" → cadastrados; sem segredo nem URL na tela) — `bun run test`

## 7. Operação e documentação
- [x] 7.1 `deploy/aviva2/README.md` (seção Z-API: os webhooks são cadastrados pelo Console; conferir no painel; rollback) e `DOC/Infra/AMBIENTES.md` (`webhook_segredo` em `kv/<p>/whatsapp`) — seções presentes
- [x] 7.2 `DOC/System/plano-de-testes.md` (T-121 a T-136) e mapa (`python3 DOC/tools/mapeia_projeto.py .`, rota nova) — linhas presentes
- [x] 7.3 `go test -race ./...`, `make itest`, `golangci-lint run`, `make contrato`, `bun run lint`, `bun run test` e `bun run build` passando
- [x] 7.4 Depois do deploy: "Cadastrar de novo" no Console mostra cadastrados e o painel da Z-API mostra as duas URLs (confirma o `PUT`); uma Notificação real por WhatsApp passa a entregue ou lida (confirma que o `messageId` dos avisos é o da resposta de envio; se não for, o aviso cai como desconhecido e o envio passa a guardar o id certo)

## 8. Fechamento
- [x] 8.1 Rodar `openspec validate --all --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 8.2 Auditar o diff com a skill sentinel
