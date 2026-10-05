# Problemas abertos

> Gerado em 2026-10-05 08:47 a partir da última execução de cada script. Não edite: rode os scripts de novo.

## verifica-projeto — 2026-09-30 18:45

- [ ] **FALTA** [Skills no projeto] 5 arquivo(s) de skill desatualizado(s) (ex.: conhecimento/SKILL.md).
  - `python3 ~/Mauricio/skills/sync-skills.py ~/Mauricio/Project/ominicore`
- [ ] **FALTA** [Scripts] DOC/tools/ desatualizado: checa_projeto.py, sdd_hooks.py, mapeia_projeto.py, mapa_html.py, registro_log.py.
  - `python3 ~/Mauricio/skills/sync-skills.py ~/Mauricio/Project/ominicore`
- [ ] **ALERTA** [Base SDD] sem DOC/util.md (comandos mais usados do projeto).
  - como corrigir: o util.md fica fixo no Sentinel ao lado do Git: é a cola dos comandos do dia a dia deste projeto, para não depender de histórico do terminal nem de memória.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/util.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
- [ ] **ALERTA** [Base SDD] o projeto tem migrations (migrations) e não tem DOC/migrations.md.
  - como corrigir: o migrations.md fica fixo no Sentinel: onde ficam, como criar, aplicar e ver o status em dev, e o lembrete de que hml e prod passam pela skill ambientes.
  - `no Claude (projeto aberto ou pasta conectada): "crie o DOC/migrations.md seguindo o guia util-e-migrations.md da skill novo-projeto-sdd"`
- [ ] **ALERTA** [Ambientes] sem DOC/Infra/infraestrutura.md: stack, onde roda e onde está o banco de hml e de prod não estão documentados.
  - como corrigir: o Claude cria o documento a partir do template, preenche o que o repositório mostra (manifestos, Docker/Compose/Nomad, migrations) e pergunta o resto — hosts e paths do cofre, nunca senhas.
  - `no Claude: "documente a infraestrutura (stack, onde roda, bancos de hml e prod) com a skill ambientes"`
