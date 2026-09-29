## 1. Esqueleto e ambiente local
- [x] 1.1 `go.mod` (Go 1.27), layout hexagonal de `DOC/System/arquitetura.md`, binário `ominicore` com subcomandos `api` e `worker`
- [x] 1.2 `.golangci.yml` com as verificações de `DOC/Padroes/stack-tecnologico.md` (inclui `depguard` para a fronteira do domínio) e job de CI em Go
- [x] 1.3 `deploy/local/compose.yaml` (`podman compose`): Postgres, Redis, OpenBao dev e Keycloak com realm `ominicore` de teste
- [x] 1.4 Migrations `goose` no schema `ominicore`: Notificação, Tentativa, Template, chave de idempotência; seed do Template `boas-vindas`
- [x] 1.5 `api/openapi.yaml` com `POST /v1/notificacoes` (header `Idempotency-Key`, respostas 202/400/401/422) e lint do contrato

## 2. Autenticação do Sistema de origem
- [x] 2.1 Adaptador Keycloak: validação do JWT pelo JWKS (cache no Redis), `iss`, `aud ∋ ominicore`, `exp`; Sistema de origem = `azp`
- [x] 2.2 Teste T-1 (pedido sem token é recusado)
- [x] 2.3 Teste T-2 (token de outro emissor ou público é recusado)
- [x] 2.4 Teste T-3 (token válido identifica o Sistema de origem)

## 3. Pedido de Notificação
- [x] 3.1 Caso de uso SendMessageUseCase + handler HTTP; Notificação e 1ª Tentativa gravadas na mesma transação
- [x] 3.2 Teste T-4 (pedido válido é aceito)
- [x] 3.3 Teste T-5 (Template inexistente é recusado)
- [x] 3.4 Teste T-6 (variável faltando é recusada)
- [x] 3.5 Chave de idempotência única por Sistema de origem com janela de 24 h
- [x] 3.6 Teste T-7 (reenvio com a mesma chave não duplica)
- [x] 3.7 Teste T-8 (pedido sem chave é recusado)
- [x] 3.8 Teste T-9 (a chave vale por Sistema de origem)

## 4. Despacho pelo Canal Email
- [x] 4.1 Worker com `FOR UPDATE SKIP LOCKED` (ADR 0001) e porta MessageSenderPort
- [x] 4.2 Adaptador SendGrid (com fake HTTP para os testes)
- [x] 4.3 Teste T-10 (Tentativa de Email é enviada)
- [x] 4.4 Teste T-11 (primeira Tentativa em até 5 segundos)
- [x] 4.5 Registro de falha do Provedor na Tentativa e na Notificação
- [x] 4.6 Teste T-12 (recusa do Provedor encerra a Notificação sem outro Canal)

## 5. Segredos
- [x] 5.1 Adaptador OpenBao: login `jwt` com token do client `ominicore-app` (ADR 0002), leitura de `kv/ominicore/*`, falha fechada no boot
- [x] 5.2 Teste T-13 (sem OpenBao o processo não inicia)
- [x] 5.3 Logger `slog` com redação de credencial, contato e conteúdo
- [x] 5.4 Teste T-14 (log sem credencial nem dado pessoal)

## 6. Fechamento
- [x] 6.1 Atualizar `DOC/System/plano-de-testes.md` com o arquivo e o status de T-1 a T-14
- [x] 6.2 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 6.3 Auditar o diff com a skill sentinel
