# Tasks

## 1. Negócio e documentação de origem
- [x] 1.1 RN-12 em `DOC/Negocio/especificacao-negocio.md`, termos Console do Operador e Ambiente no glossário e Fatia 11 em `DOC/System/backlog-mvp.md` (feitos na sabatina) — `python3 DOC/tools/checa_projeto.py .` sem furo
- [x] 1.2 Conferir que `10-config-whatsapp-api` está arquivada e que `openspec/specs/configuracao-canais/spec.md` tem o REQ-27 com o texto do delta MODIFIED desta change — `openspec validate 11-console-operador --strict`
- [x] 1.3 ADR `DOC/System/decisoes/0006-config-publica-so-do-escritorio.md` (revisa a 0003: `/v1/config/` público só do IP do escritório, CORS só para o Console) e status da 0003 apontando para ela — arquivo presente e citado no `DOC/log.md`
- [x] 1.4 `DOC/to-do.md` com o que ficou fora (configuração por cliente da casa, Email/SMS/QR no Console, BFF com Tunnel, outros idiomas, tema do Keycloak) — linhas presentes

## 2. Console: base do Materio
- [x] 2.1 Trazer o Materio free (Vue, Vuetify, TS) para `web/console/` com o `LICENSE` MIT dele; confirmar que builda com bun (senão, usar o gerenciador do lockfile do template e registrar a troca) — `bun run build` gera `web/console/dist`
- [x] 2.2 Enxugar o template: tirar as páginas de demonstração, de login e de cadastro e o seletor de idioma; deixar o layout, o menu com "WhatsApp" e o tema — `bun run build` e `vue-tsc --noEmit` sem erro
- [x] 2.3 `vue-i18n` só com `pt-BR` (`src/locales/pt-BR.json`) e ESLint com `@intlify/vue-i18n/no-raw-text` e `vue/no-v-html` — `bun run lint` sem violação
- [x] 2.4 `wrangler.jsonc` (Worker `ominicore-console` só com assets, SPA, domínio próprio `ominicore.flipapi.com.br`) e `public/_headers` (CSP, `frame-ancestors 'none'`, `no-referrer`, `nosniff`) — `wrangler deploy --dry-run` sem erro
- [x] 2.5 Vitest + `@vue/test-utils` configurados, com um teste de fumaça do layout — `bun run test`

## 3. Console: login e Ambientes
- [x] 3.1 `src/ambientes.json` (homologação e produção: rótulo, `issuer`, `client_id`, base da API, `producao`) e a tela de escolha do Ambiente — `bun run test`
- [x] 3.2 Sessão com `oidc-client-ts`: um `UserManager` por Ambiente, PKCE S256, `userStore` em memória, Ambiente no `state`, callback e descarte na troca — `bun run test`
- [x] 3.3 Teste T-106 (escolher homologação leva ao realm hml com PKCE e nenhum token fica gravado no navegador)
- [x] 3.4 Teste T-107 (trocar de Ambiente descarta o token e exige novo login antes de chamar a outra API)
- [x] 3.5 Faixa de Ambiente em todo o layout, com destaque de alerta em produção, e diálogo de confirmação citando produção — `bun run test`
- [x] 3.6 Teste T-109 (produção destacada e confirmação antes de gravar; homologação sem destaque)

## 4. Console: tela do WhatsApp
- [x] 4.1 Cliente da API (`GET`/`PUT`/`DELETE /v1/config/whatsapp` com o token do Ambiente; 401 refaz o login; 403 e 422 viram estados da tela) — `bun run test`
- [x] 4.2 Tela do WhatsApp: estado ligado ou desligado, Provedor e campos não secretos, segredos só como "configurado", formulário por Provedor (Meta, UltraMsg, Z-API) com os segredos sempre vazios, botão de desligar com confirmação — `bun run test`
- [x] 4.3 Teste T-108 (conta sem o papel vê falta de permissão e nenhum formulário)
- [x] 4.4 Teste T-110 (consulta mostra o Provedor e os segredos só como configurados)
- [x] 4.5 Teste T-111 (recusa 422 citando o campo sem limpar o formulário; depois a confirmação, com os segredos vazios)
- [x] 4.6 Teste T-112 (desligar: cancelar não chama a API; confirmar chama e mostra desligado)

## 5. API: CORS, identidade e recusa do client do Console
- [x] 5.1 Conferir se a 443 do host aviva2.0 só aceita as faixas da Cloudflare (security group ou Traefik); se não, planejar o `ipAllowList` com as faixas da Cloudflare no router da 7.1 — resultado registrado em `DOC/Infra/infraestrutura.md`
- [x] 5.2 `keycloak.Identificar` devolve `Identidade{Client, Usuario, Papeis}` (`preferred_username`); realm JWT falso de `internal/testes/` emite `preferred_username` — `go test ./internal/adaptadores/saida/keycloak/ ./internal/adaptadores/entrada/http/`
- [x] 5.3 Middleware CORS só nas rotas `/v1/config/*` (`OMINICORE_CONSOLE_ORIGEM`; preflight 204 antes da autenticação; vazio = desligado) e `OPTIONS` em `api/openapi.yaml` — `go test ./internal/adaptadores/entrada/http/` e `make contrato`
- [x] 5.4 Teste T-116 (preflight da origem do Console → 204 com a origem, os métodos e o `Authorization`)
- [x] 5.5 Teste T-117 (outra origem, ou rota de envio, sem `Access-Control-Allow-Origin`)
- [x] 5.6 Recusa do client `OMINICORE_CONSOLE_CLIENT` (padrão `ominicore-console`) nas rotas de Notificação e de Opt-out → 403 — `go test ./internal/adaptadores/entrada/http/`
- [x] 5.7 Teste T-119 (token do Console no envio, na consulta e no Opt-out → 403, nada criado)
- [x] 5.8 Operador do evento = `Usuario` quando o client é o do Console, senão `Client` — `go test ./internal/aplicacao/ ./internal/adaptadores/entrada/http/`
- [x] 5.9 Teste T-120 (troca pelo Console registra `ana`, não o client)
- [x] 5.10 Rodar de novo T-1 a T-3, T-99, T-100 e T-104 (autenticação e registro inalterados para os clients de máquina) — `go test -race ./...`

## 6. Keycloak
- [x] 6.1 `deploy/aviva2/keycloak/criar-console.sh <amb>` (client público, standard flow, PKCE S256, redirect e web origin do Console, mapper de audiência `ominicore`, fluxo `navegador-otp` com OTP obrigatório ligado ao client por override), idempotente — `bash -n`
- [x] 6.2 `deploy/aviva2/keycloak/criar-operador.sh <amb> <usuario> <email>` (papel `ominicore-config`, `CONFIGURE_TOTP` e `UPDATE_PASSWORD`, senha temporária só no terminal), idempotente — `bash -n`
- [x] 6.3 `deploy/local/`: client `ominicore-console` (também com `http://localhost:5173/*`) e pessoa de teste no realm local; `OMINICORE_CONSOLE_ORIGEM` no compose — `podman compose up` (ou Docker) e login local pelo Console com `bun run dev`
- [x] 6.4 Teste T-118 (testcontainers com Keycloak 26.3: os dois scripts criam a pessoa com o papel e as ações obrigatórias e o override com OTP obrigatório) — `make itest`

## 7. Borda, deploy e operação
- [x] 7.1 Router `Host(${HOST_WEBHOOK}) && PathPrefix(/v1/config/)` em `deploy/aviva2/traefik/ominicore.yml.tmpl` (com `ipAllowList` da Cloudflare se a 5.1 pedir) e `OMINICORE_CONSOLE_ORIGEM`/`OMINICORE_CONSOLE_CLIENT` em `deploy/aviva2/ominicore.env.tmpl` — `bash -n deploy/aviva2/publicar.sh` e o arquivo gerado conferido
- [x] 7.2 `deploy/aviva2/verificar-exposicao.sh <amb>` (Console, `/v1/config/whatsapp`, `/v1/notificacoes` e `/v1/opt-outs` de fora do escritório → 403/404; sai com erro citando o T-n) — `bash -n`
- [x] 7.3 Testes T-113, T-114 e T-115 pelo `verificar-exposicao.sh`, rodando no CI de fora do escritório
- [x] 7.4 `.github/workflows/console.yml` (PR com filtro em `web/console/**`: lint, `vue-tsc`, Vitest, build; `workflow_dispatch` na `main`: `wrangler deploy` com `CLOUDFLARE_API_TOKEN` e depois o `verificar-exposicao.sh hml` e `prod`; `schedule` diário só da verificação) e Dependabot `npm` em `web/console` — workflow válido (`actionlint` ou o run do PR verde)
- [x] 7.5 Runbook em `deploy/aviva2/README.md`: IP List e regra de WAF na Cloudflare **antes** do router, `criar-console.sh`, `criar-operador.sh`, deploy do Console, troca do IP do escritório e rollback (tirar o router, `wrangler rollback`) — seções presentes
- [x] 7.6 `DOC/Infra/AMBIENTES.md` e `DOC/Infra/infraestrutura.md` (hostname do Console, WAF, client e variáveis por Ambiente) e `DOC/Padroes/stack-tecnologico.md` (linhas do front: Materio, Vue 3, Vuetify, Vite, TS, bun, `oidc-client-ts`, `vue-i18n`, Vitest, Workers Static Assets) — seções presentes
- [x] 7.7 `DOC/System/plano-de-testes.md` (T-106 a T-120, com o Vitest e a verificação de exposição como formas de teste) e mapa (`python3 DOC/tools/mapeia_projeto.py .`, área do Console) — linhas presentes
- [x] 7.8 `go test -race ./...`, `make itest`, `golangci-lint run`, `make contrato`, `bun run lint`, `bun run test` e `bun run build` passando

## 8. Fechamento
- [x] 8.1 Rodar `openspec validate --all --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 8.2 Auditar o diff com a skill sentinel
