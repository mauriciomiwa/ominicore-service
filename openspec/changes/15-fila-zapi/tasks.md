# Tasks

## 1. Negócio e fatos da Z-API

- [ ] 1.1 RN-12 estendida em `DOC/Negocio/especificacao-negocio.md` (ver a conexão e a Fila do Provedor e retirar um item), Fatia 15 em `DOC/System/backlog-mvp.md` e decisões da sabatina de 2026-10-30 no `DOC/log.md` — verificar: `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [ ] 1.2 (dono) Script só de leitura em `deploy/aviva2/` (pede instância, token e client-token sem eco) que chama `GET /status` e `POST /queue` na instância real e mostra o status HTTP, as chaves da resposta e se algum `MessageId` da fila bate com o `messageId` de uma Tentativa enviada — verificar: métodos e caminhos confirmados e anotados no `design.md`; divergência volta para o design antes da seção 2
- [ ] 1.3 Confirmar com o aviva-back se ele lê a classificação da falha e aceita `removida_pelo_operador` — verificar: resposta anotada no `log.md`

## 2. Z-API e domínio

- [ ] 2.1 `dominio`: `ConexaoWhatsApp` (conectado, desconectado, celular sem internet), `FalhaRemovidaPeloOperador`, `ItemFila` e a máscara do telefone (DDI, DDD e 4 últimos) — verificar: `go test ./internal/dominio/`
- [ ] 2.2 `zapifake`: `POST /queue` com `pagingState`, `GET /status` e `DELETE /queue/{id}`, com fila e conexão configuráveis — verificar: `go vet ./internal/testes/...`
- [ ] 2.3 Pacote `zapi`: `Conexao`, `Listar` e `Retirar` da `FilaProvedorPort`, com timeout de 10 s, `Client-Token` e erro sem URL nem token — verificar: `go test ./internal/adaptadores/saida/zapi/`

## 3. Aplicação, banco e API (REQ-45, REQ-46, REQ-14)

- [ ] 3.1 Migration `00014_config_eventos_fila.sql` (coluna `item`, ação `fila_retirada`; down apaga as linhas e estreita o CHECK) e `EventoConfig.Item` — verificar: `go test ./internal/integracao/` aplica as migrations
- [ ] 3.2 Consulta `TentativasPorIDProvedor` no `queries.sql` e no repositório (Notificação, Template e Sistema de origem por `IDProvedor`) — verificar: `sqlc generate` sem diferença e `go build ./...`
- [ ] 3.3 `FilaUseCase` (Consultar e Retirar com a busca pelo cursor até 20 páginas e a falha pelo `ReceiveMessageUseCase`) e wiring em `cmd/ominicore/main.go` — verificar: `go build ./...`
- [ ] 3.4 Rotas `GET /v1/config/whatsapp/fila` e `DELETE /v1/config/whatsapp/fila/{id}` com `operador(...)`, preflight CORS e respostas 409, 404 e 502; `api/openapi.yaml` com as rotas e o valor novo do enum — verificar: `go test ./internal/adaptadores/entrada/http/`
- [ ] 3.5 Teste T-141 (fila com item do OminiCore e de fora, telefone mascarado, sem texto) — verificar: `go test -run T141 ./...`
- [ ] 3.6 Teste T-142 (35 itens, segunda página pelo cursor) — verificar: `go test -run T142 ./...`
- [ ] 3.7 Teste T-143 (UltraMsg ativo: recusa sem chamar a Z-API) — verificar: `go test -run T143 ./...`
- [ ] 3.8 Teste T-144 (Z-API 401 ou fora do ar: erro com status, sem token nem URL) — verificar: `go test -run T144 ./...`
- [ ] 3.9 Teste T-145 (sem papel 403, sem token 401, nada chamado nem registrado) — verificar: `go test -run T145 ./...`
- [ ] 3.10 Teste T-146 contra o Postgres (WhatsApp → SMS: Tentativa falhou com `removida_pelo_operador`, Tentativa de SMS criada, registro com `ana` e o id do item) — verificar: `go test -run T146 ./internal/integracao/`
- [ ] 3.11 Teste T-147 (item que já saiu: 404, nada muda, `DELETE` não chamado, nada registrado) — verificar: `go test -run T147 ./...`
- [ ] 3.12 Teste T-148 (item fora do OminiCore: retirado e registrado, nenhuma Tentativa muda) — verificar: `go test -run T148 ./...`
- [ ] 3.13 Teste T-149 (Aviso de entrega da Z-API depois da retirada é ignorado) — verificar: `go test -run T149 ./...`
- [ ] 3.14 Teste T-156 (consulta do `wbem` mostra a Tentativa com `removida_pelo_operador`, sem telefone nem motivo) — verificar: `go test -run T156 ./...`

## 4. Console (REQ-47, REQ-48)

- [ ] 4.1 `criarApi` com `consultarFila(cursor?)` e `retirarDaFila(id)`; componente `FilaZapi.vue` na tela do WhatsApp só com a Z-API ativa; textos em `locales/pt-BR.json` — verificar: `pnpm --dir web/console typecheck` e `lint`
- [ ] 4.2 Teste T-150 (conexão desconectada, item de Tentativa e item de fora, sem nova consulta sem "Atualizar") — verificar: `pnpm --dir web/console test -t T-150`
- [ ] 4.3 Teste T-151 ("Carregar mais" com o cursor e some quando não há mais) — verificar: `pnpm --dir web/console test -t T-151`
- [ ] 4.4 Teste T-152 (erro com status 401 e "Tentar de novo") — verificar: `pnpm --dir web/console test -t T-152`
- [ ] 4.5 Teste T-153 (UltraMsg ativo: sem conexão nem Fila, nenhuma consulta) — verificar: `pnpm --dir web/console test -t T-153`
- [ ] 4.6 Teste T-154 (confirmação cita produção; cancelar não chama; confirmar retira e recarrega) — verificar: `pnpm --dir web/console test -t T-154`
- [ ] 4.7 Teste T-155 (já saiu: aviso e recarga) — verificar: `pnpm --dir web/console test -t T-155`
- [ ] 4.8 T-141 a T-156 em `DOC/System/plano-de-testes.md` — verificar: `python3 DOC/tools/checa_projeto.py .` sem furo

## 5. Entrega

- [ ] 5.1 (dono) Publicar a API (migration `00014`) e o Console em prod, com o bloco de alvo da skill `ambientes` — verificar: a tela do WhatsApp em prod mostra a conexão e a Fila da instância real
- [ ] 5.2 (dono) Provar a retirada com um item real da fila em prod, com permissão do dono na hora — verificar: o item some da fila, a Tentativa fica `removida_pelo_operador` e `config_eventos` traz a retirada

## 6. Fechamento

- [ ] 6.1 Rodar `python3 DOC/tools/valida_rastreabilidade.py DOC` — verificar: sem furo
- [ ] 6.2 Auditar o diff com a skill sentinel e registrar com `sdd_hooks.py auditoria` — verificar: veredito registrado
