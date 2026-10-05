# Estado — ominicore

> Atualizado em 2026-10-05 por Mauricio Miwa + Claude (change 09 implementada)

## Em curso
- Change `08-whatsapp-ultramsg` (commitada): 31/32; falta a 5.2 (formato real da resposta do UltraMsg).
- Change `09-whatsapp-zapi`: 26/27 na branch `fatia/09-whatsapp-zapi`, implementação sem commit; auditada (aprovado com ressalvas). Falta a 5.3 (respostas reais da Z-API: sucesso e token inválido).
- Próximo passo: commitar a 09; com as instâncias de trial, fechar 5.2 (08) e 5.3 (09) com `curl`, ajustar `classificar` se divergir, `/opsx:archive` da 08 e depois da 09 (nessa ordem).
- Depois: deploy em hml (`configurar-openbao.sh hml --so-whatsapp`, skill ambientes).

## Decisões recentes
- WhatsApp por Meta, UltraMsg ou Z-API, um por ambiente (`provedor` no cofre); risco do gateway não oficial aceito (`System/decisoes/0004-…`, que ganha um adendo para a Z-API).
- Z-API: `provedor=zapi`, `instance_id`, `token`, `client_token` opcional (header `Client-Token`); envia o texto como o UltraMsg; sem Aviso de status (RN-5 intacta, Tentativa fica `enviada`); webhooks da Z-API vão para o `to-do.md`.
- Z-API põe o token na URL: erro de rede não pode levar a URL ao Motivo nem ao log (teste prova).

## Bloqueios e pendências
- Instância de trial do UltraMsg: criada pelo dono (5.2).
- `shellcheck` não roda local nem no CI (scripts só com `bash -n`).
- Oferecidos e não decididos: `DOC/util.md` e `DOC/migrations.md`.

## Arquivos que importam agora
- `openspec/changes/08-whatsapp-ultramsg/` — modelo para a 09
- `internal/adaptadores/saida/ultramsg/ultramsg.go` — base do futuro `saida/zapi`
- `internal/dominio/provedor_whatsapp.go` — `Servir` vira "Provedores de texto"
- Doc da Z-API: `developer.z-api.io/message/send-text.md`, `security/client-token.md`
