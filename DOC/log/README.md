# log/ — execuções dos scripts SDD

Gerado pelos scripts das skills (verifica-projeto, checa_projeto, mapeia_projeto,
mapa_html, sync-skills) e pelo painel do openmd.

| Arquivo | O que é |
|---|---|
| `problemas.md` | o que está aberto agora: última execução de cada script, com FALTA/ALERTA, erros e como corrigir |
| `historico.jsonl` | uma linha por execução (data, script, resultado, faltas, alertas) — dá a evolução do projeto |
| `execucoes/` | detalhe de cada execução em JSON (fora do git) |

Para tratar: abra `problemas.md`, rode o comando de cada item (ou peça ao Claude:
"trate os problemas de DOC/.../log/problemas.md") e rode a verificação de novo.
