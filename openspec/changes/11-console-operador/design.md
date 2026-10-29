# Design

## Context

O motivo está no `proposal.md`, em Why. Estado atual que molda a solução:

- A API de configuração (`/v1/config/whatsapp`, fatia 10) existe e exige o
  papel `ominicore-config`. Quem fez a troca vem do `azp`, que
  `keycloak.Validador.Identificar` devolve junto com os papéis
  (`internal/adaptadores/saida/keycloak/validador.go:69`).
- O Traefik só publica `Host(notify[-hml].flipapi.com.br) &&
  PathPrefix(/webhooks/)` (`deploy/aviva2/traefik/ominicore.yml.tmpl`). Os dois
  hostnames já passam pelo proxy da Cloudflare.
- O Keycloak é público em `https://auth.flipapi.com.br`, com um realm por
  Ambiente (`ominicore`, `ominicore-hml`). A API valida `iss` e `aud ∋
  ominicore`.
- Não há front no repositório. O precedente da casa é o `../choice`: Worker com
  Static Assets, Vue 3, Vite e bun, deploy pelo `wrangler`.

## Goals / Non-Goals

**Goals:**
- Um único Console estático, sem servidor próprio, que fala com as duas APIs
  (homologação e produção).
- A abertura nova na borda é a menor possível: um prefixo de rota, restrito
  pelo IP do escritório, e a API continua sendo a última barreira (JWT e
  papel).
- Toda a regra de acesso que dá para testar fica testada: no Go, no Vitest e
  numa verificação de fora do escritório, no CI.

**Non-Goals:**
- Sessão que sobrevive ao recarregar a página sem passar pelo Keycloak.
  Recarregar refaz o login; com a sessão SSO viva, isso não pede senha.
- Personalizar o tema do Keycloak, ter mais de um idioma, reaproveitar as
  páginas de demonstração do Materio.

## Decisions

### D1 — Exposição: rota pública só para `/v1/config/`, WAF por IP (ADR 0006)
O Traefik ganha o router `Host(notify[-hml]) && PathPrefix(/v1/config/)` para
a mesma API. Na zona `flipapi.com.br` da Cloudflare, uma regra de WAF bloqueia
`ominicore.flipapi.com.br` (inteiro) e `notify[-hml].flipapi.com.br/v1/*`
quando `ip.src` não está na lista do escritório. A lista fica numa *IP List* da
conta, para trocar o IP sem mexer na regra. `/v1/notificacoes` e `/v1/opt-outs`
continuam sem router: de fora, quem tenta recebe 403 da Cloudflare e, se
chegasse à origem, 404 do Traefik.

- *Alternativas:* Worker intermediário com Cloudflare Tunnel (decidido contra
  na sabatina: mais peças no host); Cloudflare Access na frente (atrito de CORS
  e cookie entre domínios); `ipAllowList` no Traefik com `ipStrategy.depth: 1`
  (funciona atrás da Cloudflare, como no `auth.flipapi.com.br`, mas põe o IP no
  repositório e não cobre o Console, que não passa pelo Traefik). O WAF cobre
  os dois hosts num lugar só.
- A ADR 0006 registra a decisão e revisa a 0003.

### D2 — Console: Worker só com Static Assets
`web/console/` é o Materio free, enxuto: layout, menu e tema. Saem as páginas
de demonstração, de login, de cadastro e o seletor de idioma. O build do Vite
sai em `web/console/dist`. O `wrangler.jsonc` na raiz de `web/console/` publica
um Worker `ominicore-console` **sem `main`** (só assets) com
`not_found_handling: single-page-application` e o domínio próprio
`ominicore.flipapi.com.br`. Um arquivo `_headers` define:
- `Content-Security-Policy`: `default-src 'self'`; `connect-src` com
  `auth.flipapi.com.br`, `notify.flipapi.com.br` e
  `notify-hml.flipapi.com.br`; `frame-ancestors 'none'`;
- `Referrer-Policy: no-referrer`;
- `X-Content-Type-Options: nosniff`.

- *Alternativas:* Cloudflare Pages (a Cloudflare concentra o estático novo nos
  Workers, e o `choice` já usa Workers); servir pelo Traefik do host (exigiria
  rota pública nova no host).
- **Gerenciador de pacotes:** bun, como no `choice`. A tarefa 2.1 confirma que
  o Materio builda com ele. Se não buildar, usamos o gerenciador do lockfile do
  template, e a troca fica registrada no `stack-tecnologico.md`.

### D3 — Login: `oidc-client-ts`, um `UserManager` por Ambiente
`web/console/src/ambientes.json` (versionado, sem segredo) lista os dois
Ambientes: rótulo, `issuer`, `client_id: ominicore-console`, base da API e
`producao: true|false`. Ao escolher o Ambiente, o Console cria o `UserManager`
daquele `issuer`:
- Authorization Code + PKCE S256;
- `userStore`: `InMemoryWebStorage`, o token só em memória (REQ-29);
- `stateStore`: o `sessionStorage` padrão, que guarda só `state` e
  `code_verifier` até o callback e é limpo depois.

Trocar de Ambiente descarta o `UserManager` anterior e faz `signinRedirect` no
novo (T-107). O Ambiente escolhido vai na `redirect_uri`
(`/callback/<ambiente>`) para o callback saber em que realm concluir. Os
endpoints do realm vão fixos em `metadata`, sem buscar o discovery. A renovação usa o refresh token em memória.

- *Alternativa:* `keycloak-js` (adaptador oficial). Ficamos com o
  `oidc-client-ts` porque ele é agnóstico e deixa escolher onde cada coisa é
  guardada; o `keycloak-js` guarda o estado do callback do jeito dele e assume
  uma instância por página.

### D4 — CORS na API, só em `/v1/config/`
Um middleware novo em `internal/adaptadores/entrada/http` envolve só as rotas
de configuração:
- com `Origin` igual a `OMINICORE_CONSOLE_ORIGEM`, devolve
  `Access-Control-Allow-Origin` com essa origem e `Vary: Origin`;
- no preflight (`OPTIONS`), responde 204 **antes** da autenticação, com
  `Allow-Methods: GET, PUT, DELETE`, `Allow-Headers: Authorization,
  Content-Type` e `Max-Age: 600`.

O token vai no `Authorization`, não em cookie, então não há
`Allow-Credentials`. Com a variável vazia (local sem Console, testes antigos),
o middleware não faz nada. As rotas de envio e de Opt-out não passam por ele
(T-117).

### D5 — Identidade: client e usuário
O port `Identificar` passa a devolver uma `Identidade{Client, Usuario,
Papeis}`; o `Usuario` vem de `preferred_username`. Com `OMINICORE_CONSOLE_CLIENT`
(padrão `ominicore-console`) igual ao `Client`:
- as rotas de envio e de Opt-out respondem 403 (REQ-36);
- a configuração registra o `Usuario` como Operador (REQ-27 / T-120).

Com outro client, nada muda: o Operador continua sendo o `azp`.

- *Alternativa:* registrar sempre o `preferred_username`. Para um client de
  máquina viria `service-account-<client>`, o que mudaria o T-104 sem ganho.

### D6 — Keycloak: client público e Operador com OTP
- `deploy/aviva2/keycloak/criar-console.sh <amb>` cria ou atualiza o client
  `ominicore-console`: público, só o *standard flow*, PKCE S256 obrigatório,
  `redirectUris`/`webOrigins` só para `https://ominicore.flipapi.com.br`,
  mapper de audiência `ominicore`. Também cria o fluxo `navegador-otp` (cópia
  do *browser* com OTP **obrigatório**) e o liga ao client por
  *authentication flow override*. Assim o OTP vale no Console sem mudar o
  login dos outros clients do realm.
- `deploy/aviva2/keycloak/criar-operador.sh <amb> <usuario> <email>` cria a
  pessoa com o papel `ominicore-config`, as ações obrigatórias
  `CONFIGURE_TOTP` e `UPDATE_PASSWORD` e uma senha temporária mostrada uma
  única vez no terminal (nunca em arquivo ou log).
- `deploy/local/` ganha o mesmo client (com o `redirectUri` também para
  `http://localhost:5173/*`) e uma pessoa de teste.

### D7 — Testes
- **Go:** T-116, T-117, T-119 e T-120 em `internal/adaptadores/entrada/http`,
  com o realm JWT falso (`internal/testes/`), que ganha `preferred_username`.
- **Integração (testcontainers):** T-118 sobe o Keycloak 26.3, roda
  `criar-console.sh` e `criar-operador.sh` contra ele e confere usuário, papel,
  ações obrigatórias e o override do fluxo. Pula sozinho sem runtime de
  containers, como os demais.
- **Vitest + `@vue/test-utils`:** T-106 a T-112, com o `UserManager` e o
  `fetch` falsos. O nome do teste cita o T-n.
- **Verificação de fora do escritório:**
  `deploy/aviva2/verificar-exposicao.sh <amb>` (T-113 a T-115) faz `curl` no
  Console e nas rotas e falha se algo responder diferente de 403/404. Roda no
  GitHub Actions, que fica fora do escritório: depois de cada deploy e uma vez
  por dia (`schedule`), para pegar regra de WAF desligada à mão.

### D8 — i18n e CI do Console
- `vue-i18n` com um único locale, `pt-BR`. As strings ficam em
  `web/console/src/locales/pt-BR.json` e nenhum texto fica no template. O ESLint
  do template ganha `@intlify/vue-i18n/no-raw-text`.
- `.github/workflows/console.yml`, com filtro de caminho `web/console/**`:
  `lint`, `vue-tsc`, `vitest` e `build` em todo PR. O deploy pelo `wrangler`
  é por `workflow_dispatch` na `main`, com o secret `CLOUDFLARE_API_TOKEN` (do
  dono), seguido do `verificar-exposicao.sh`.
- O Dependabot ganha o ecossistema `npm` em `web/console`.

## Risks / Trade-offs

- **[Risco] Alguém alcança a origem sem passar pela Cloudflare** (IP do host
  direto) e pula o WAF → mitigação já existente: o security group só deixa a
  Cloudflare chegar à 443 (`../flip/infra/aviva2.0/traefik/traefik.yml:4`,
  tarefa 5.1). A API continua exigindo JWT e papel de qualquer forma.
- **[Dependência] O Keycloak público também é restrito por IP:**
  `auth.flipapi.com.br` tem `ipAllowList` no Traefik da casa
  (`../flip/infra/aviva2.0/traefik/dynamic/keycloak.yml`). Sem o IP do
  escritório nessa lista, o login pelo Console é recusado → o runbook inclui o
  ajuste nesse arquivo (repositório `flip`) antes do primeiro login.
- **[Risco] Regra de WAF desligada ou editada à mão** → mitigação: a
  verificação diária do CI (D7) falha e avisa.
- **[Risco] O IP do escritório muda** (troca de operadora) → mitigação: a IP
  List da Cloudflare é o único lugar a trocar, e o runbook diz como. Até a
  troca, o Console fica inacessível, o que é uma falha segura.
- **[Trade-off] Um Console para os dois Ambientes:** um deploy do front com
  defeito afeta também a configuração de produção → mitigação: deploy manual
  depois do CI verde. Para voltar, `wrangler rollback`.
- **[Risco] XSS no Console vaza o token em memória** → mitigação: CSP restrita
  (D2), sem HTML dinâmico (`v-html` proibido no lint) e token de vida curta (5
  min, padrão do realm).
- **[Trade-off] Recarregar a página refaz o login** (D3): aceito. Com a sessão
  SSO viva, é só um redirecionamento.

## Migration Plan

1. Arquivar a 08, a 09 e a 10 (pré-requisito).
2. Publicar a API com `OMINICORE_CONSOLE_ORIGEM` e `OMINICORE_CONSOLE_CLIENT`
   em hml. Sem a rota no Traefik, nada fica exposto ainda.
3. IP do escritório no `keycloak-allowlist` de
   `../flip/infra/aviva2.0/traefik/dynamic/keycloak.yml` (repositório `flip`).
   Em hml: `criar-console.sh hml` e `criar-operador.sh hml <usuario> <email>`
   para cada pessoa.
4. Cloudflare (dono): a IP List do escritório e a regra de WAF **antes** do
   router novo, para não haver janela aberta.
5. Router `/v1/config/` no Traefik de hml, deploy do Console e
   `verificar-exposicao.sh hml`. Teste do escritório: entrar, consultar e
   gravar.
6. Repetir os passos 2 a 5 em prod, com a skill `ambientes`.

**Rollback:** tirar o router de `/v1/config/` do Traefik fecha a configuração
de fora na hora. O Console pode sair do ar com `wrangler delete` ou voltar para
a versão anterior com `wrangler rollback`. Esvaziar `OMINICORE_CONSOLE_ORIGEM`
desliga o CORS. Nenhum dado muda de formato. O `config_eventos` só passa a
receber nomes de usuário.

## Open Questions

- O IP fixo do escritório: o dono informa no passo 4. Ele não muda o plano e
  nunca entra no repositório.
