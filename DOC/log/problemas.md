# Problemas abertos

> Gerado em 2026-10-31 19:42 a partir da última execução de cada script. Não edite: rode os scripts de novo.

## checa_projeto-local — 2026-10-31 19:42

- [ ] **ALERTA** [estrutura] DOC/log.md com 26 KB (> 15 KB) — arquive as entradas mais antigas: python3 DOC/tools/arquiva_log.py DOC
- [ ] **ALERTA** [sentinel] 2 arquivo(s) de código sem auditoria da sentinel no conteúdo atual: internal/integracao/integracao_test.go, migrations/00012_modelos_regua_wbem.sql. Rode a skill sentinel sobre o diff e registre: python3 <DOC>/tools/sdd_hooks.py auditoria --veredito "aprovado|aprovado com ressalvas|corrigir antes de concluir".
