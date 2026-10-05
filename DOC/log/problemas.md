# Problemas abertos

> Gerado em 2026-10-05 11:04 a partir da última execução de cada script. Não edite: rode os scripts de novo.

## checa_projeto-local — 2026-10-05 11:04

- [ ] **ALERTA** [estrutura] DOC/estado.md com 33 linhas (> 25) — é retrato, não histórico; reescreva com a skill passagem.

## verifica-projeto — 2026-10-05 10:58

- [ ] **ALERTA** [Base SDD] DOC/estado.md com 33 linhas (> 25) — é retrato, não histórico; reescreva com a skill passagem.
  - como corrigir: o estado.md é o retrato de onde o trabalho parou (≤ 25 linhas); o hook de início o entrega ao chat novo no lugar de reler log e código.
  - `no Claude (projeto aberto ou pasta conectada): "faça a passagem de bastão com a skill passagem"`
- [ ] **ALERTA** [Base SDD] sem DOC/util.md (comandos mais usados do projeto).
  - como corrigir: o util.md fica fixo no Sentinel ao lado do Git: é a cola dos comandos do dia a dia deste projeto, para não depender de histórico do terminal nem de memória.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/util.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
- [ ] **ALERTA** [Base SDD] o projeto tem migrations (migrations) e não tem DOC/migrations.md.
  - como corrigir: o migrations.md fica fixo no Sentinel: onde ficam, como criar, aplicar e ver o status em dev, e o lembrete de que hml e prod passam pela skill ambientes.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/migrations.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
