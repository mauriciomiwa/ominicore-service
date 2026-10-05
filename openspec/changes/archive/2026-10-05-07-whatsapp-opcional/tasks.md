# Tasks

## 1. Cofre e boot
- [x] 1.1 `openbao`: leitura opcional de `kv/<prefixo>/whatsapp` (caminho inexistente = desligado; campo vazio = erro), `Segredos.WhatsAppLigado` e `CamposWebhooksMeta` exigidos só com o Canal ligado — `go test ./internal/adaptadores/saida/openbao/`
- [x] 1.2 `cmd/ominicore`: `canaisDesligados(seg)`, log "whatsapp desligado" no boot, Remetente do WhatsApp registrado só com o Canal ligado — `go build ./...`
- [x] 1.3 Teste T-63 (sem credenciais da Meta o processo inicia com o WhatsApp desligado)
- [x] 1.4 Teste T-64 (credenciais da Meta incompletas impedem o início, citando o campo)
- [x] 1.5 Rodar de novo T-13 e T-14 (REQ-6 modificado)

## 2. Canal desligado no pedido, no fallback e no envio
- [x] 2.1 `dominio`: `Destinatario.Desligados` em `Disponivel` e `ErrCanalIndisponivel` ("canal %s indisponível") — `go test ./internal/dominio/`
- [x] 2.2 `aplicacao`: `Desligados` injetado em `SendMessageUseCase`, `DespacharUseCase` e `ReceiveMessageUseCase`; `Pedir` devolve `ErrCanalIndisponivel` quando não sobra Canal e algum está desligado — `go test ./internal/aplicacao/`
- [x] 2.3 HTTP: `ErrCanalIndisponivel` → 422 "canal indisponível"; `api/openapi.yaml` documenta o motivo — teste de contrato
- [x] 2.4 Teste T-65 (pedido só por WhatsApp com o Canal desligado é recusado)
- [x] 2.5 Teste T-66 (Canal desligado no começo da ordem é pulado)
- [x] 2.6 Teste T-67 (fallback pula o Canal desligado)
- [x] 2.7 Teste T-68 (Tentativa pendente no Canal desligado segue a ordem)
- [x] 2.8 Rodar de novo T-15 a T-17, T-24 a T-27, T-32 e T-51 a T-53 (disponibilidade de Canal) com `go test -race ./...` e `make itest`

## 3. Webhook da Meta
- [x] 3.1 `webhooks`: `WhatsAppLigado` e rotas da Meta registradas só com o Canal ligado; `main` repassa o estado — `go test ./internal/adaptadores/entrada/webhooks/`
- [x] 3.2 Teste T-69 (webhook da Meta com o Canal desligado responde 404)
- [x] 3.3 Rodar de novo T-36, T-38 e T-40 com o Canal ligado

## 4. Deploy e documentação
- [x] 4.1 `deploy/aviva2/openbao/configurar-openbao.sh`: pergunta "Configurar WhatsApp agora? [s/N]" e `--so-whatsapp` (só `kv/<prefixo>/whatsapp` e `kv patch` dos campos da Meta, segredos pelo stdin) — `bash -n` e `shellcheck`
- [x] 4.2 `deploy/aviva2/README.md` (etapas 4 e 5b: Meta opcional, como ligar depois), `DOC/Infra/AMBIENTES.md` e `DOC/Infra/infraestrutura.md`; `deploy/local/` segue com o WhatsApp ligado (Meta falsa)
- [x] 4.3 `DOC/System/plano-de-testes.md` (T-63 a T-69), mapa (`tools/mapeia_projeto.py`) e `DOC/to-do.md` (itens da Meta deixam de bloquear o deploy)

## 5. Fechamento
- [x] 5.1 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 5.2 Auditar o diff com a skill sentinel
