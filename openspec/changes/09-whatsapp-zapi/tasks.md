# Tasks

## 1. Negócio e documentação de origem
- [x] 1.1 RN-4 em `DOC/Negocio/especificacao-negocio.md` com a Z-API, glossário ("Provedor") e a Fatia 9 em `DOC/System/backlog-mvp.md` (feitos na sabatina) — `python3 DOC/tools/checa_projeto.py .` sem furo
- [ ] 1.2 Adendo na ADR `DOC/System/decisoes/0004-whatsapp-por-gateway-nao-oficial.md` (risco aceito estendido à Z-API, número dedicado por ambiente) e entrada no `DOC/log.md` com as decisões P1–P6 da sabatina — adendo presente e citado no log
- [ ] 1.3 `DOC/Padroes/fontes-externas.md` ganha a Z-API (`send-text`, `Client-Token`, token na URL, limite próprio de 4096); `DOC/to-do.md` ganha "webhooks da Z-API com segredo na URL (revisar RN-5)" — linhas presentes

## 2. Domínio e cofre
- [ ] 2.1 `dominio`: `WhatsAppZAPI = "zapi"`, `EnviaTexto()`, `Servir` e o motivo "versão whatsapp sem texto" por `EnviaTexto`; `LimiteTextoWhatsApp` com a checagem compartilhada, e o UltraMsg passando a usá-la — `go test ./internal/dominio/ ./internal/adaptadores/saida/ultramsg/`
- [ ] 2.2 `openbao`: `CamposWhatsApp[zapi] = {instance_id, token}`, com `client_token` opcional — `go test ./internal/adaptadores/saida/openbao/`
- [ ] 2.3 Teste T-82 (WhatsApp pela Z-API inicia sem os segredos da Meta nem `client_token`; o log cita o Provedor sem a credencial)
- [ ] 2.4 Teste T-83 (credenciais da Z-API incompletas impedem o início, citando o campo)
- [ ] 2.5 Rodar de novo T-13, T-14, T-63, T-64 e T-70 a T-72 (REQ-6 modificado; T-72 com Provedor fora de meta/ultramsg/zapi)

## 3. Envio pela Z-API
- [ ] 3.1 Falso `internal/testes/zapifake` (`send-text` com instância e token no caminho, registra o header `Client-Token` e o corpo, respostas programáveis) — usado pelos testes 3.3 a 3.8
- [ ] 3.2 Adaptador `internal/adaptadores/saida/zapi` (JSON `phone`/`message`, telefone sem `+`, `Client-Token` só se configurado, timeout 10 s, classificação do design, erro de rede sem o `err`) e `OMINICORE_ZAPI_URL` no `cmd/ominicore`; `remetentes` liga a Z-API por `ProvedorWhatsApp` — `go test ./internal/adaptadores/saida/zapi/` e `go build ./...`
- [ ] 3.3 Teste T-84 (erro de rede e recusa da Z-API: motivo e log sem o token da instância nem o `client_token`)
- [ ] 3.4 Teste T-85 (Tentativa enviada como texto pela Z-API para o telefone sem `+`, sem `Client-Token`, com o `messageId` como id)
- [ ] 3.5 Teste T-86 (com `client_token` no cofre, toda chamada leva o header `Client-Token`)
- [ ] 3.6 Teste T-87 (recusa da Z-API, inclusive `{"error":"null not allowed"}`, passa ao próximo Canal)
- [ ] 3.7 Teste T-88 (Z-API com 429, 5xx ou sem resposta é Falha temporária com Retentativa)
- [ ] 3.8 Teste T-89 (texto acima de 4096 caracteres não chega à Z-API e segue a ordem)
- [ ] 3.9 Rodar de novo T-28 a T-31 (Meta) e T-73 a T-77 (UltraMsg; T-77 também trocando UltraMsg → Meta com a versão só de texto)

## 4. Pedido e webhook
- [ ] 4.1 Teste T-90 (versão sem texto é recusada com a Z-API ativa, motivo "versão whatsapp sem texto"); `api/openapi.yaml` cita a Z-API junto do UltraMsg no motivo e na Tentativa que fica `enviada` — `go test ./internal/aplicacao/ ./internal/adaptadores/entrada/http/`
- [ ] 4.2 Teste T-91 (webhook da Meta com a Z-API ativa responde 404 e nenhuma Tentativa muda) — `go test ./cmd/ominicore/ ./internal/adaptadores/entrada/webhooks/`
- [ ] 4.3 Rodar de novo T-32, T-33, T-78 a T-81

## 5. Deploy e operação
- [ ] 5.1 `deploy/aviva2/openbao/configurar-openbao.sh` (opção `zapi` na pergunta e no `--so-whatsapp`; `instance_id`, `token` e o opcional `client_token` pelo stdin; segredos da Meta só com `meta`) e `deploy/local/bootstrap.sh` (`WHATSAPP_PROVEDOR=zapi` com `ZAPI_INSTANCE_ID`, `ZAPI_TOKEN`, `ZAPI_CLIENT_TOKEN`) — `bash -n` nos dois
- [ ] 5.2 Runbook `deploy/aviva2/README.md` (instância e número por ambiente, ordem cofre → reiniciar → ativar Client-Token no painel, rotação do token, monitorar instância desconectada, rollback pelo cofre) e `DOC/Infra/AMBIENTES.md` — seções presentes
- [ ] 5.3 Conferir com uma instância real da Z-API a resposta de sucesso e a de token inválido; ajustar o teste de classificação se divergir e registrar o formato em `DOC/Padroes/fontes-externas.md`
- [ ] 5.4 `DOC/System/plano-de-testes.md` (T-82 a T-91) e mapa (`python3 DOC/tools/mapeia_projeto.py`) — gerados sem furo
- [ ] 5.5 `go test -race ./...` e `make itest` passando

## 6. Fechamento
- [ ] 6.1 Rodar `openspec validate --all --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [ ] 6.2 Auditar o diff com a skill sentinel
