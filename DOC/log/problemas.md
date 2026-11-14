# Problemas abertos

> Gerado em 2026-11-14 13:28 a partir da última execução de cada script. Não edite: rode os scripts de novo.

## checa_projeto-local — 2026-11-14 13:28

- [ ] **ALERTA** [estrutura] DOC/log.md com 27 KB (> 15 KB) — arquive as entradas mais antigas: python3 DOC/tools/arquiva_log.py DOC
- [ ] **ALERTA** [novo-projeto-sdd] T-n de changes ativas ainda sem teste (em curso): T-141, T-142, T-143, T-144, T-145, T-146, T-147, T-148, T-149, T-150, T-151, T-152
- [ ] **ALERTA** [sentinel] 23 arquivo(s) de código sem auditoria da sentinel no conteúdo atual: cmd/ominicore/main.go, internal/adaptadores/entrada/http/config_test.go, internal/adaptadores/entrada/http/servidor.go, internal/adaptadores/entrada/http/servidor_test.go, internal/adaptadores/entrada/http/webhooks_zapi_test.go, internal/adaptadores/entrada/webhooks/webhooks_test.go, internal/adaptadores/entrada/worker/worker_test.go, internal/adaptadores/saida/postgres/arquitetura_test.go …. Rode a skill sentinel sobre o diff e registre: python3 <DOC>/tools/sdd_hooks.py auditoria --veredito "aprovado|aprovado com ressalvas|corrigir antes de concluir".

## verifica-projeto — 2026-10-05 12:10

- [ ] **ALERTA** [Base SDD] sem DOC/util.md (comandos mais usados do projeto).
  - como corrigir: o util.md fica fixo no Sentinel ao lado do Git: é a cola dos comandos do dia a dia deste projeto, para não depender de histórico do terminal nem de memória.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/util.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
- [ ] **ALERTA** [Base SDD] o projeto tem migrations (migrations) e não tem DOC/migrations.md.
  - como corrigir: o migrations.md fica fixo no Sentinel: onde ficam, como criar, aplicar e ver o status em dev, e o lembrete de que hml e prod passam pela skill ambientes.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/migrations.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
