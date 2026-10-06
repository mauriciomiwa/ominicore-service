# Área — Console do Operador

> Interpretação sobre o inventário de 2026-10-05 (`inventario.md`). Fatia 11, RN-12.

## O que faz
Página web estática em `ominicore.flipapi.com.br` (Cloudflare Workers Static
Assets), aberta só do escritório. O Operador escolhe o Ambiente (homologação
ou produção), entra no realm dele com conta pessoal e OTP e consulta, troca ou
desliga o WhatsApp pela API daquele Ambiente, sem nunca ver segredos.

## Onde está
| Peça | Arquivo |
|---|---|
| Ambientes (realm, client, API; sem segredo) | `web/console/src/console/ambientes.ts`, `ambientes-publicados.json`, `ambientes-locais.json` (só no `bun run dev`) |
| Sessão OIDC (PKCE, token em memória, um `UserManager` por Ambiente) | `web/console/src/console/sessao.ts` |
| Cliente da API de configuração | `web/console/src/console/api.ts` |
| Telas | `web/console/src/pages/{index,callback,whatsapp,erro}.vue` |
| Destaque de produção e confirmação | `web/console/src/components/{FaixaAmbiente,ChipAmbiente,DialogoConfirmar}.vue` |
| Rotas e guarda de sessão | `web/console/src/plugins/router/` |
| Textos (só pt-BR) | `web/console/src/locales/pt-BR.json` |
| Hospedagem e CSP | `web/console/wrangler.jsonc`, `web/console/public/_headers` |
| Base do template (Materio free, MIT) | `web/console/src/@layouts`, `src/@core`, `LICENSE-materio` |

## Lado da API
- CORS só em `/v1/config/*` e só para `OMINICORE_CONSOLE_ORIGEM`:
  `internal/adaptadores/entrada/http/cors.go`.
- Token do client `OMINICORE_CONSOLE_CLIENT` recusado no envio e no Opt-out;
  nas trocas, o Operador registrado é a pessoa (`preferred_username`):
  `internal/adaptadores/entrada/http/servidor.go` (`soSistemaDeOrigem`, `quem`).

## Keycloak e borda
- Client público e fluxo `navegador-otp`: `deploy/aviva2/keycloak/lib-console.sh`
  (`criar-console.sh`, `criar-operador.sh`; também usado pelo `deploy/local/bootstrap.sh`).
- Rota `/v1/config/` no Traefik: `deploy/aviva2/traefik/ominicore.yml.tmpl`.
- WAF só do escritório: runbook em `deploy/aviva2/README.md` (C1–C8); verificação
  de fora: `deploy/aviva2/verificar-exposicao.sh` (CI `console.yml`).

## Pontos de atenção
- Recarregar a página perde o token (em memória) e volta à escolha do Ambiente;
  com a sessão SSO viva, o Keycloak devolve sem pedir senha.
- O `auth.flipapi.com.br` tem `ipAllowList` própria no Traefik da casa: o IP do
  escritório precisa estar lá também.
- O Materio free está fixado na 2.3.0 (01/2025), com ESLint 8 legado.
