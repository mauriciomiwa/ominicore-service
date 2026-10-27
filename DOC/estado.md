# Estado — ominicore

> Atualizado em 2026-10-27 por Mauricio Miwa + Claude (changes 12 e 13 propostas; spec notificacoes ajustada)

## Em curso
- Prod no ar desde 2026-10-06 (`notify.flipapi.com.br`, Console em `ominicore.flipapi.com.br`); hml adiado (`log.md` 2026-10-06).
- Change `13-webhooks-zapi` proposta (1/39): Avisos de status da Z-API (RN-4/RN-5 revisadas, REQ-37 a REQ-42, REQ-21 modificado, T-121 a T-136). Implementar **depois** da change 12.
- Change `12-defeitos-prod` implementada 11/12 na branch `fatia/12-defeitos-prod` (sem commit; sentinel com ressalvas; `make itest` passando, inclusive o T-98 no OpenBao real). Falta 1.1 (archive 08→11).
- Em prod: `a9a3bf9` (API, quadlet) e Console 76eed4ae. Próximo passo: `HOST=aviva ./deploy/aviva2/publicar.sh prod` para o healthcheck (`45a39b3`, tarefa 3.3) e confirmar o WAF; o archive da 12 e da 13 exige arquivar antes 08 → 09 → 10 → 11 (a 08 falta 5.2; a 11 falta 1.2 e 7.3).
- Sem commit: change `ajuste-texto-notificacoes` (arquivada), change 13, DOC desta sessão e as mudanças de 27/10 de outra sessão (`to-do.md`, `fontes-externas.md`, tasks da 09).

## Decisões recentes
- Webhooks da Z-API: segredo na URL + `instanceId`, 404 sem efeito; cadastro automático na Z-API a cada gravação; botão "Cadastrar de novo" (`log.md` 2026-10-27; ADR 0007 a escrever).
- REQ-2 e REQ-14 reescritos abaixo de 500 caracteres (OpenSpec 1.15 reprova no `--strict`).

## Bloqueios e pendências
- Caminho da API de cadastro de webhook da Z-API não documentado: tarefa 2.1 da 13, com a instância real.
- WAF do Console: confirmar de fora do escritório (4G) que dá 403.
- Email e SMS em prod com credenciais `pendente`; trocar a senha do role `aviva_back` (esteve no cofre do OminiCore).

## Arquivos que importam agora
- `openspec/changes/13-webhooks-zapi/` — proposta, deltas, design, tarefas
- `DOC/to-do.md` — seção "Prod — operação (2026-10-27)" (escopo da change 12)
- `internal/adaptadores/entrada/webhooks/whatsapp.go` — modelo do adaptador de avisos
