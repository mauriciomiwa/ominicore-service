# Estado — ominicore

> Atualizado em 2026-10-05 por Mauricio Miwa + Claude (change 11 implementada, sem commit)

## Em curso
- Change `11-console-operador`: 43/45 na branch `fatia/11-console-operador` (sem commit); sentinel: aprovado com ressalvas.
- Faltam: 1.2 (archive da 10 antes do archive da 11: o REQ-27 MODIFIED exige `configuracao-canais` vigente) e 7.3 (T-113 a T-115 só rodam depois do deploy).
- Próximo passo: commitar a 11; fechar 5.2 (08) e 5.3 (09) com instâncias reais; `/opsx:archive` 08 → 09 → 10 → 11.
- Deploy (skill ambientes): runbook `deploy/aviva2/README.md` seção Console (C1–C8). WAF antes do `publicar.sh`.

## Decisões recentes
- Console: Materio free 2.3.0 em `web/console/` (Vite 6, Vuetify 3.13, `oidc-client-ts`, `vue-i18n` pt-BR), Worker só com assets em `ominicore.flipapi.com.br` (ADR 0006).
- `/v1/config/*` público só do IP do escritório (WAF) + JWT + CORS do Console; token do Console recusado no envio/Opt-out; troca registra a pessoa.
- OTP obrigatório pelo fluxo `navegador-otp` (override no client); a atualização das execuções preserva a `priority`.

## Bloqueios e pendências
- Dono: IP fixo do escritório na IP List da Cloudflare e no `keycloak-allowlist` do repositório `flip`; secrets `CLOUDFLARE_API_TOKEN`/`CLOUDFLARE_ACCOUNT_ID` no environment `console`.
- 6 CVEs altas só em ferramentas de build do Console, sem correção publicada (mitigação no `log.md`).
- Instância de trial do UltraMsg (5.2 da 08); `shellcheck` não roda local nem no CI.

## Arquivos que importam agora
- `openspec/changes/11-console-operador/` — proposta, deltas, design, tarefas
- `web/console/src/console/` — Ambientes, sessão OIDC, cliente da API
- `deploy/aviva2/keycloak/lib-console.sh` — client do Console, OTP e Operadores
- `deploy/aviva2/README.md` — runbook do Console (C1–C8)
