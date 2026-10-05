# Estado — ominicore

> Atualizado em 2026-10-05 por Mauricio Miwa + Claude (change 10 implementada)

## Em curso
- Changes `08-whatsapp-ultramsg` (31/32, falta 5.2) e `09-whatsapp-zapi` (26/27, falta 5.3) na `main` e em `origin`; arquivar 08 e depois 09 após os `curl` reais.
- Change `10-config-whatsapp-api`: 33/33 na branch `fatia/10-config-whatsapp-api` (planejamento commitado, implementação sem commit); auditada (aprovado com ressalvas). Arquivar depois da 08 e da 09.
- Próximo passo: commitar a 10; fechar 5.2/5.3 com instâncias reais; `/opsx:archive` 08 → 09 → 10.
- Depois: deploy em hml (skill ambientes): `configurar-openbao.sh hml --sem-segredos` (policy de escrita), `conceder-operador.sh hml <client>`, binário com a migration 00009.

## Decisões recentes
- WhatsApp por Meta, UltraMsg ou Z-API, um por ambiente (`provedor` no cofre; ADR 0004 com adendo da Z-API).
- RN-11 e termo Operador: configuração do WhatsApp pela API, sem reinício; app escreve no cofre só em `kv/<p>/whatsapp` (ADR 0005 a escrever, revisa a 0002); segredos da Meta migram para esse caminho com fallback em `webhooks` (REQ-28).

## Bloqueios e pendências
- Instância de trial do UltraMsg: criada pelo dono (5.2).
- `shellcheck` não roda local nem no CI (scripts só com `bash -n`).
- Oferecidos e não decididos: `DOC/util.md` e `DOC/migrations.md`.

## Arquivos que importam agora
- `openspec/changes/10-config-whatsapp-api/` — proposta, deltas, design, tarefas
- `internal/aplicacao/config_atual.go` e `configurar.go` — estado trocável, releitura e caso de uso
- `deploy/aviva2/README.md` — Operador, troca pela API, rollback com a Meta
