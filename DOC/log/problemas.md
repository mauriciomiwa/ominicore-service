# Problemas abertos

> Gerado em 2026-10-27 19:45 a partir da última execução de cada script. Não edite: rode os scripts de novo.

## checa_projeto-local — 2026-10-27 19:45

- [ ] **ALERTA** [estrutura] DOC/log.md com 19 KB (> 15 KB) — arquive as entradas mais antigas: python3 DOC/tools/arquiva_log.py DOC
- [ ] **ALERTA** [novo-projeto-sdd] T-n de changes ativas ainda sem teste (em curso): T-114, T-121, T-122, T-123, T-124, T-125, T-126, T-127, T-128, T-129, T-130, T-131

## verifica-projeto — 2026-10-05 12:10

- [ ] **ALERTA** [Base SDD] sem DOC/util.md (comandos mais usados do projeto).
  - como corrigir: o util.md fica fixo no Sentinel ao lado do Git: é a cola dos comandos do dia a dia deste projeto, para não depender de histórico do terminal nem de memória.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/util.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
- [ ] **ALERTA** [Base SDD] o projeto tem migrations (migrations) e não tem DOC/migrations.md.
  - como corrigir: o migrations.md fica fixo no Sentinel: onde ficam, como criar, aplicar e ver o status em dev, e o lembrete de que hml e prod passam pela skill ambientes.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/migrations.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
