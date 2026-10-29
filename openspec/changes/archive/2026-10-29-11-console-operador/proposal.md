# Proposal

## Why

Desde a fatia 10 (RN-11), o Operador troca o Provedor do WhatsApp pela API,
mas a API só é alcançável pela rede interna `aviva` (ADR 0003) e o Operador
precisa montar a requisição e o token na mão. A nova **RN-12** (sabatina de
2026-10-05) dá a ele uma página web, o **Console do Operador**. O acesso é só
do escritório, com conta pessoal e segundo fator, e a página escolhe o
Ambiente (homologação ou produção). Para isso, a rota de configuração precisa
ser alcançável de fora, sem abrir o resto da API.

## What Changes

- **Console do Operador** em `https://ominicore.flipapi.com.br`, servido como
  arquivos estáticos pela Cloudflare. Usa o template Materio free (Vue 3 +
  Vuetify, MIT) e todos os textos são pt-BR. Com ele, o Operador:
  - escolhe o Ambiente (homologação ou produção) e entra no Keycloak daquele
    Ambiente com a própria conta e segundo fator. O login usa Authorization
    Code + PKCE, sem segredo no navegador, e o token fica só em memória;
  - consulta, grava ou desliga o WhatsApp pela API da fatia 10, sem ver
    segredo nenhum;
  - vê produção sempre destacada; trocar de Ambiente exige novo login.
- **Rota pública só para `/v1/config/*`** em `notify.flipapi.com.br` e
  `notify-hml.flipapi.com.br`, atrás da Cloudflare. Uma regra de WAF libera o
  Console e essa rota só para o IP fixo do escritório. A API continua exigindo
  o JWT com o papel `ominicore-config`. `/v1/notificacoes` e `/v1/opt-outs`
  seguem sem rota pública. A ADR 0006 registra isso e revisa a 0003.
- **CORS** na API, só nas rotas `/v1/config/*` e só para a origem do Console.
- **A conta do Console não pede Notificações nem mexe em Opt-out**: um token do
  client do Console é recusado nessas rotas, mesmo que alguém alcance a rede
  interna.
- **O registro da troca identifica a pessoa** (o nome de usuário no Keycloak)
  quando a troca vem do Console; vinda de um client com o papel, continua
  registrando o client. **BREAKING** só para quem lê `config_eventos`: o campo
  do Operador passa a trazer o usuário nas trocas feitas pelo Console.
- **Keycloak e infra:**
  - client público `ominicore-console` nos realms `ominicore` e
    `ominicore-hml`;
  - um script cria o Operador pessoa com o papel `ominicore-config` e exige que
    ele configure o OTP no primeiro login;
  - rota no Traefik, regra de WAF e deploy do Console pelo `wrangler` no GitHub
    Actions.

## Capabilities

### New Capabilities

- `console-operador`: o Console web do Operador. Cobre a escolha do Ambiente,
  o login pessoal com segundo fator, a configuração do WhatsApp sem exibir
  segredos, o destaque de produção e o acesso de fora restrito ao escritório e
  à rota de configuração, com CORS só para o Console.

### Modified Capabilities

- `autenticacao-origem`: requisito novo. O token do client do Console não vale
  como Sistema de origem para pedir Notificações nem para mexer em Opt-out.
- `configuracao-canais`: o REQ-27 (registro das trocas) passa a identificar a
  pessoa quando a troca vem do Console. Essa capacidade nasce no archive da
  change 10, que precisa estar arquivada antes desta.

## Impact

- **Dependência:** `10-config-whatsapp-api` arquivada (e, antes dela, a 08 e a
  09).
- **Negócio e DOC:** já feitos na sabatina: a RN-12, os termos Console do
  Operador e Ambiente no glossário e a Fatia 11 no backlog. Ficam para a
  implementação:
  - a ADR 0006 (revisa a 0003);
  - `Padroes/stack-tecnologico.md` (front);
  - `Infra/infraestrutura.md` e `Infra/AMBIENTES.md` (DNS, WAF, Console);
  - o `plano-de-testes.md`;
  - o `to-do.md` com o que ficou de fora.
- **Código Go:**
  - `internal/adaptadores/entrada/http`: CORS em `/v1/config/*` e recusa do
    client do Console nas rotas de envio e de Opt-out;
  - `internal/adaptadores/saida/keycloak`: nome de usuário no token;
  - `internal/aplicacao`: quem fez a troca.
  - Variáveis novas: `OMINICORE_CONSOLE_ORIGEM` e `OMINICORE_CONSOLE_CLIENT`.
- **Código novo:** `web/console/` (Materio, `vue-i18n` pt-BR, cliente OIDC) e
  `wrangler.jsonc`, com testes de unidade (Vitest).
- **Contrato:** `api/openapi.yaml` documenta o `OPTIONS` (preflight) de
  `/v1/config/whatsapp`, sem mudar as respostas existentes.
- **Deploy:**
  - `deploy/aviva2/traefik/ominicore.yml.tmpl` (rota `/v1/config/`);
  - `deploy/aviva2/keycloak/` (client do Console, script de Operador);
  - `deploy/local/` (client e usuário no realm local);
  - regra de WAF e DNS na Cloudflare (dono, com o IP do escritório);
  - workflow `.github/workflows/console.yml` (build, teste, deploy e
    verificação de fora do escritório).
- **Fora do escopo:**
  - configuração por cliente da casa (W Bem, Lara);
  - Email, SMS, status e QR no Console;
  - Worker intermediário com Cloudflare Tunnel;
  - outros idiomas;
  - tema do Keycloak com a cara do Console.
