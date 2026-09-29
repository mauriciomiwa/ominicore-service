# Estado — ominicore

> Atualizado em 2026-09-29 por Mauricio Miwa + Claude (fatia 1 verificada)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- Change `01-enviar-email`: **30/30 tarefas**, verificada contra Postgres,
  Keycloak, OpenBao e Redis reais (Docker). Pronta para `/opsx:archive`.
- Commit `e70f224` na branch `fatia/01-enviar-email`; o que veio depois
  (contrato de ambientes, backchannel/prefixo, verificação) **ainda sem commit**.
- Próximo passo: commit, `/opsx:archive 01-enviar-email`, abrir a fatia 02
  (`02-sms-e-fallback`) com a sabatina.

## Decisões recentes
- hml e prod no mesmo host aviva2.0, separados por realm (`ominicore-hml`),
  prefixo do cofre (`kv/ominicore-hml/*`) e schema (`ominicore_hml`).
- Webhooks em `notify(-hml).flipapi.com.br/webhooks/*`; API só na rede `aviva`.
- Promove-se o mesmo binário de hml para prod; só o dono escreve segredo de prod.

## Bloqueios e pendências
- VM do Podman quebrada; o local roda no Docker (`COMPOSE="docker compose"`).
- No diff sem commit há arquivos alterados por ferramentas SDD
  (`mapa.html`, `mapa_html.py`, `log/historico.jsonl`, `.skills-sync.json`):
  conferir antes de commitar.
- Decisões em aberto no `to-do.md` (cadastro de Templates, consulta de status,
  chave repetida com pedido diferente, entrega "pelo menos uma vez").

## Arquivos que importam agora
- `../openspec/changes/01-enviar-email/` — change a arquivar
- `System/plano-de-testes.md` — status de cada T-n
- `Infra/AMBIENTES.md` — contrato de hml/prod
- `../deploy/local/bootstrap.sh` — cofre local com auth `jwt` igual a hml/prod
