# Estado — ominicore

> Atualizado em 2026-10-05 por Mauricio Miwa + Claude (change 08 em curso)

## Em curso
- Change `08-whatsapp-ultramsg`: 31/32 em `tasks.md`; falta a 5.2 (formato real da resposta do UltraMsg).
- Branch `fatia/08-whatsapp-ultramsg` commitada (fatia e ferramentas SDD em commits separados); ainda não levada à `main`.
- Próximo passo: colar 2 respostas do `curl` em `.../messages/chat` (sucesso e token errado), ajustar `classificar` em `internal/adaptadores/saida/ultramsg/ultramsg.go` se divergir, `/opsx:archive` e merge na `main`.
- Depois: deploy em hml com `configurar-openbao.sh hml --so-whatsapp` → `ultramsg` (skill ambientes).

## Decisões recentes
- WhatsApp por Meta ou UltraMsg, um por ambiente (`provedor` no cofre; ausente = meta); risco do gateway não oficial aceito pelo cliente (`System/decisoes/0004-…`, `log.md` 2026-10-05).
- UltraMsg sem Aviso de status: Tentativa fica `enviada`, `/webhooks/whatsapp` → 404 (REQ-21).

## Bloqueios e pendências
- Instância de trial do UltraMsg: criada pelo dono (5.2).
- `shellcheck` não roda local nem no CI (scripts só com `bash -n`).
- Oferecidos e não decididos: `DOC/util.md` e `DOC/migrations.md`.

## Arquivos que importam agora
- `openspec/changes/08-whatsapp-ultramsg/` — proposta, deltas, design, tarefas
- `internal/adaptadores/saida/ultramsg/ultramsg.go` — envio e classificação
- `deploy/aviva2/README.md` — ligar/trocar o Provedor e rollback
