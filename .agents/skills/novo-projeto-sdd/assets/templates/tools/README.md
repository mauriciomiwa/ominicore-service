# tools/ — scripts das garantias SDD

Gerados pela skill `novo-projeto-sdd` (não edite aqui: o `sync-skills.py`
substitui pela versão da fonte).

- `checa_projeto.py` — checagem completa (pre-commit, CI e hook Stop).
- `sdd_hooks.py` — hooks do Claude e registro da auditoria da `sentinel`.
- `valida_rastreabilidade.py` — cruza RN → REQ → teste e aponta furos.
  Rode a cada requisito novo, não só no kickoff.
- `dependabot_config.py` — gera e confere o `.github/dependabot.yml`.
- `diff_ambientes.py` — compara chaves de config entre hml e prod.
- `mapeia_projeto.py` e `mapa_html.py` — inventário do código e vista do mapa.
- `registro_log.py` e `arquiva_log.py` — histórico dos scripts e arquivo do `log.md`.
