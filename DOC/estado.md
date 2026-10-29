# Estado — ominicore

> Atualizado em 2026-10-29 por Mauricio Miwa + Claude (changes 08 a 14 arquivadas; WAF corrigido)

## Em curso
- Nenhuma change ativa. Prod no ar: API `c3bcb0c` (fatia 14) e Console 57674d31; hml adiado (`log.md` 2026-10-06).
- `main` = `origin/main`, com tudo desta sessão. Branch de trabalho: `fatia/14-templates-wbem`.

## Decisões recentes
- WAF: o IP do escritório (`72.60.159.29`, túnel Hostinger) fica direto na regra "Block specific paths", porque o plano Free tem as 5 regras em uso (ADR 0006, `deploy/aviva2/README.md` C1/C2). Verificação diária pelo CI (`CONSOLE_PUBLICADO=true`).
- A 5.2 da 08 (UltraMsg real) foi adiada para o `to-do.md`: prod usa a Z-API.
- Templates `wbem-*` em prod; o Sistema de origem do aviva-back é o client `wbem` (realm `ominicore`).

## Bloqueios e pendências
- Email e SMS em prod com credenciais `pendente`; trocar a senha do role `aviva_back`.
- `publicar.sh`: `flock` e `.anterior` atômico; token de API da Cloudflare para wrangler e CI.
- RN-6 e RN-10 sem REQ: confirmar se ficam para depois.
- IP do túnel mudou? Trocar na regra da Cloudflare e no `keycloak-allowlist` (repositório `flip`).

## Arquivos que importam agora
- `DOC/to-do.md` — pendências de operação
- `deploy/aviva2/README.md` — runbook (Console C1–C8, Z-API, QUIC)
