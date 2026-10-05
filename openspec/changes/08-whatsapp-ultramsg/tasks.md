# Tasks

## 1. Negócio e documentação de origem
- [x] 1.1 RN-4 reescrita em `DOC/Negocio/especificacao-negocio.md` (texto da proposta); glossário: "Provedor" cita Meta e UltraMsg para o WhatsApp — `python3 DOC/tools/checa_projeto.py .` sem furo
- [x] 1.2 ADR `DOC/System/decisoes/0004-whatsapp-por-gateway-nao-oficial.md` (risco aceito pelo cliente, número dedicado por ambiente, uso só transacional) e entrada no `DOC/log.md` — arquivo presente e citado no log
- [x] 1.3 `DOC/Padroes/fontes-externas.md` ganha o UltraMsg (endpoint `messages/chat`, limite de 4096, formato da resposta "a confirmar") e `DOC/System/backlog-mvp.md` a Fatia 8 (`08-whatsapp-ultramsg`, RN-4, REQ-6/10/11 modificados, REQ-21, T-70 a T-81)

## 2. Cofre e boot
- [x] 2.1 `dominio`: tipo `ProvedorWhatsApp` (`meta`, `ultramsg`) com `Servir(ConteudoCanal) error` e `ErrVersaoWhatsAppSemForma` — `go test ./internal/dominio/`
- [x] 2.2 `openbao`: campo `provedor` em `kv/<prefixo>/whatsapp` (ausente = `meta`), campos exigidos por Provedor, `CamposWebhooksMeta` só com `meta`, `Segredos.ProvedorWhatsApp` — `go test ./internal/adaptadores/saida/openbao/`
- [x] 2.3 Teste T-70 (WhatsApp pelo UltraMsg inicia sem os segredos da Meta; log cita o Provedor sem a credencial)
- [x] 2.4 Teste T-71 (credenciais do UltraMsg incompletas impedem o início, citando o campo)
- [x] 2.5 Teste T-72 (Provedor do WhatsApp desconhecido impede o início)
- [x] 2.6 Rodar de novo T-13, T-14, T-63 e T-64 (REQ-6 modificado; T-64 também com `provedor=meta` explícito)

## 3. Envio pelo UltraMsg
- [x] 3.1 Adaptador `internal/adaptadores/saida/ultramsg` (form `token`/`to`/`body`, timeout 10 s, limite de 4096 runas, classificação tolerante do design, token fora da URL e do motivo) e `OMINICORE_ULTRAMSG_URL` no `cmd/ominicore`; `remetentes` escolhe Meta ou UltraMsg por `ProvedorWhatsApp` — `go test ./internal/adaptadores/saida/ultramsg/` e `go build ./...`
- [x] 3.2 Migration `migrations/00008_whatsapp_texto.sql` (corpo da versão WhatsApp do `boas-vindas` = texto do SMS; `Down` volta a `''`) — `make itest`
- [x] 3.3 Teste T-73 (Tentativa de WhatsApp enviada como texto pelo UltraMsg, com o id do Provedor)
- [x] 3.4 Teste T-74 (recusa do UltraMsg passa ao próximo Canal)
- [x] 3.5 Teste T-75 (UltraMsg com 429, 5xx ou sem resposta é Falha temporária com Retentativa)
- [x] 3.6 Teste T-76 (texto acima de 4096 caracteres não chega ao UltraMsg e segue a ordem)
- [x] 3.7 Teste T-77 (Tentativa pendente sem a forma do Provedor ativo segue a ordem)
- [x] 3.8 Rodar de novo T-28 a T-31 com a Meta ativa

## 4. Pedido e webhook
- [x] 4.1 `aplicacao`: `SendMessageUseCase.ProvedorWhatsApp` e `Pedir` chama `Servir` quando o WhatsApp está ligado e na ordem; HTTP: `ErrVersaoWhatsAppSemForma` → 422 com o motivo; `api/openapi.yaml` documenta os motivos e que, com o UltraMsg, a Tentativa fica `enviada` — `go test ./internal/aplicacao/ ./internal/adaptadores/entrada/http/`
- [x] 4.2 Teste T-78 (versão sem texto é recusada com o UltraMsg ativo)
- [x] 4.3 Teste T-79 (versão sem Modelo aprovado é recusada com a Meta ativa)
- [x] 4.4 Teste T-80 (variável do texto do WhatsApp é exigida)
- [x] 4.5 Rodar de novo T-32, T-33 e T-65 a T-68
- [x] 4.6 `webhooks`: `WhatsAppLigado` → `MetaLigada`; `main` passa Canal ligado e Provedor `meta` — `go test ./internal/adaptadores/entrada/webhooks/`
- [x] 4.7 Teste T-81 (webhook da Meta com o UltraMsg ativo responde 404)
- [x] 4.8 Rodar de novo T-36, T-38, T-40 e T-69

## 5. Deploy e operação
- [x] 5.1 `deploy/aviva2/openbao/configurar-openbao.sh` (pergunta Meta ou UltraMsg; `--so-whatsapp` grava `provedor`, `instance_id` e `token` pelo stdin; segredos da Meta só com `meta`) e `deploy/local/bootstrap.sh` (`WHATSAPP_PROVEDOR=ultramsg`) — `bash -n` (`shellcheck` não instalado localmente nem no CI: pendente)
- [ ] 5.2 Conferir o formato real da resposta do UltraMsg com a instância de trial (sucesso e token inválido) e ajustar o teste de classificação se divergir; registrar o formato em `DOC/Padroes/fontes-externas.md`
- [x] 5.3 Runbook `deploy/aviva2/README.md` (instância e número por ambiente, monitorar a instância desconectada, rollback pelo cofre), `DOC/Infra/AMBIENTES.md` e `DOC/Infra/infraestrutura.md`
- [x] 5.4 `DOC/System/plano-de-testes.md` (T-70 a T-81) e mapa (`tools/mapeia_projeto.py`)
- [x] 5.5 `go test -race ./...` e `make itest` passando

## 6. Fechamento
- [x] 6.1 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 6.2 Auditar o diff com a skill sentinel
