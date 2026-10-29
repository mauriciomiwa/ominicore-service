# Tasks

## 1. Negócio e documentação de origem
- [x] 1.1 RN-11 em `DOC/Negocio/especificacao-negocio.md`, termo Operador no glossário e Fatia 10 em `DOC/System/backlog-mvp.md` (feitos na sabatina) — `python3 DOC/tools/checa_projeto.py .` sem furo
- [x] 1.2 ADR `DOC/System/decisoes/0005-escrita-no-cofre-pela-api.md` (revisa a 0002: escrita só em `kv/<p>/whatsapp`, papel `ominicore-config`) e entrada no `DOC/log.md` com as decisões P1–P10 da sabatina — arquivo presente e citado no log
- [x] 1.3 `DOC/to-do.md`: status da instância e QR, teste da credencial no Provedor, configuração de Email e SMS por API — linhas presentes

## 2. Domínio e estado trocável
- [x] 2.1 `dominio.ConfigWhatsApp` com `Validar()` e os segredos por Provedor; campos exigidos e opcionais movidos de `openbao` para o domínio, e `openbao.Carregar` passando a usá-los — `go test ./internal/dominio/ ./internal/adaptadores/saida/openbao/`
- [x] 2.2 `configwa.Atual` (retrato atômico, `Aplicar`, Remetente que delega) e porta `portas.EstadoCanais`; `SendMessageUseCase`, `DespacharUseCase` e `ReceiveMessageUseCase` leem o estado a cada chamada — `go test -race ./internal/aplicacao/ ./internal/adaptadores/entrada/worker/`
- [x] 2.3 Rodar de novo T-63 a T-91 (comportamento por Provedor e Canal desligado inalterado)

## 3. Cofre: escrita, releitura e segredos da Meta
- [x] 3.1 `openbao`: `GravarWhatsApp`, `ApagarWhatsApp`, `LerWhatsApp` e re-login em 403/expirado; segredos da Meta lidos de `whatsapp` com fallback em `webhooks` — `go test ./internal/adaptadores/saida/openbao/`
- [x] 3.2 Teste T-105 (segredos da Meta no caminho do WhatsApp, no antigo, ou em nenhum)
- [x] 3.3 Releitura periódica em api e worker (`OMINICORE_CONFIG_RELEITURA`, padrão 30 s); webhook da Meta decide a rota a cada requisição pelo `Atual()` — `go test -race ./cmd/ominicore/ ./internal/adaptadores/entrada/webhooks/`
- [x] 3.4 Teste T-101 (troca do UltraMsg pela Z-API sem reinício)
- [x] 3.5 Teste T-102 (configuração inválida na releitura mantém a anterior e loga sem credencial)
- [x] 3.6 Teste T-103 (rotas da Meta acompanham o Provedor ativo)
- [x] 3.7 Rodar de novo T-81 e T-91 (webhook 404 com Provedor de texto)

## 4. API de configuração
- [x] 4.1 `keycloak`: `Identidade{AZP, Papeis}` lendo `realm_access.roles`; middleware `exigirPapel("ominicore-config")` → 403 — `go test ./internal/adaptadores/saida/keycloak/ ./internal/adaptadores/entrada/http/`
- [x] 4.2 Migration `migrations/00009_config_eventos.sql` (tabela; `Down` remove) e `RegistrarEventoConfig` no repositório Postgres e no de memória — `make itest`
- [x] 4.3 `aplicacao.ConfigurarUseCase` (Consultar, Gravar, Apagar: valida, grava o cofre, aplica no processo, registra o evento; falha do evento → log de erro e 500) e rotas `GET`/`PUT`/`DELETE /v1/config/whatsapp`; `api/openapi.yaml` com as rotas, o esquema sem segredos e 401/403/422 — `go test ./internal/aplicacao/ ./internal/adaptadores/entrada/http/` e `make contrato`
- [x] 4.4 Teste T-92 (Operador grava a Z-API)
- [x] 4.5 Teste T-93 (campo exigido faltando → 422, cofre intacto)
- [x] 4.6 Teste T-94 (Provedor desconhecido → 422, cofre intacto)
- [x] 4.7 Teste T-95 (Meta exige e grava os segredos do webhook)
- [x] 4.8 Teste T-96 (consulta sem nenhum caractere dos segredos)
- [x] 4.9 Teste T-97 (consulta com o Canal desligado)
- [x] 4.10 Teste T-98 (desligar pela API; pedido só WhatsApp → 422 após a releitura)
- [x] 4.11 Teste T-99 (Sistema de origem sem o papel → 403, nada muda, sem evento)
- [x] 4.12 Teste T-100 (sem token ou token inválido → 401)
- [x] 4.13 Teste T-104 (gravar e apagar geram dois eventos com o Operador, sem segredo no evento nem no log)
- [x] 4.14 Rodar de novo T-1 a T-3 (REQ-1 inalterado com a leitura de papéis)

## 5. Deploy e operação
- [x] 5.1 `deploy/aviva2/openbao/configurar-openbao.sh` (policy com escrita só em `kv/data/<p>/whatsapp`) e `deploy/aviva2/keycloak/conceder-operador.sh` (cria o realm role `ominicore-config`, idempotente, e concede ao service account do client); `deploy/local/` com o papel no realm local e a policy do bao dev — `bash -n` (`shellcheck` não instalado localmente nem no CI: pendente)
- [x] 5.2 Runbook `deploy/aviva2/README.md` (conceder o papel, trocar o Provedor pela API com `curl`, rollback com Meta) e `DOC/Infra/AMBIENTES.md` (`OMINICORE_CONFIG_RELEITURA`, papel por ambiente) — seções presentes
- [x] 5.3 `DOC/System/plano-de-testes.md` (T-92 a T-105) e mapa (`python3 DOC/tools/mapeia_projeto.py .`, página `envio.md` ou nova área de configuração)
- [x] 5.4 `go test -race ./...`, `make itest`, `golangci-lint run` e `make contrato` passando

## 6. Fechamento
- [x] 6.1 Rodar `openspec validate --all --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 6.2 Auditar o diff com a skill sentinel
