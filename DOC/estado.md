# Estado — ominicore

> Atualizado em 2026-10-05 por Mauricio Miwa + Claude (change 10 proposta)

## Em curso
- Changes `08-whatsapp-ultramsg` (31/32, falta 5.2) e `09-whatsapp-zapi` (26/27, falta 5.3) na `main` e em `origin`; arquivar 08 e depois 09 após os `curl` reais.
- Change `10-config-whatsapp-api` proposta (1/32 tarefas, sem commit): API `GET/PUT/DELETE /v1/config/whatsapp` só para o Operador (realm role `ominicore-config`), grava em `kv/<p>/whatsapp`, releitura a quente a cada 30 s, `config_eventos`. T-92 a T-105.
- Próximo passo: revisar os artefatos da 10 e `/opsx:apply 10-config-whatsapp-api` (ideal depois de arquivar 08 e 09).

## Decisões recentes
- WhatsApp por Meta, UltraMsg ou Z-API, um por ambiente (`provedor` no cofre; ADR 0004 com adendo da Z-API).
- RN-11 e termo Operador: configuração do WhatsApp pela API, sem reinício; app escreve no cofre só em `kv/<p>/whatsapp` (ADR 0005 a escrever, revisa a 0002); segredos da Meta migram para esse caminho com fallback em `webhooks` (REQ-28).

## Bloqueios e pendências
- Instância de trial do UltraMsg: criada pelo dono (5.2).
- `shellcheck` não roda local nem no CI (scripts só com `bash -n`).
- Oferecidos e não decididos: `DOC/util.md` e `DOC/migrations.md`.

## Arquivos que importam agora
- `openspec/changes/10-config-whatsapp-api/` — proposta, deltas, design, tarefas
- `cmd/ominicore/main.go` — estado do Canal hoje congelado no boot (vira `configwa.Atual`)
- `internal/adaptadores/saida/openbao/segredos.go` — leitura; ganha escrita e releitura
