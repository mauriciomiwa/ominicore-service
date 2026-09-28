# Estado — ominicore

> Atualizado em 2026-09-28 por Mauricio Miwa + Claude (fatia 1)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- Change `01-enviar-email`: código completo, **22/30 tarefas**. Unitários,
  lint, govulncheck, sqlc diff e contrato verdes.
- Falta verificar contra serviços reais: testes de integração (T-7, T-9, T-11 e
  versões Postgres de T-4/T-10/T-12), `compose` local e login `jwt` no OpenBao.
- Próximo passo: consertar o runtime de containers (VM do Podman) e rodar
  `make itest`; depois `make compose-up` seguindo `deploy/local/README.md`.

## Decisões recentes
- `golang-jwt/jwt/v5` para o JWT (precedente do aviva-back).
- Despacho segura a transação durante o envio (simples; "pelo menos uma vez").
- Imagens locais = versões do aviva2.0.

## Bloqueios e pendências
- VM `podman-machine-default` sem identidade SSH; Docker parado. Recriar a VM
  apaga a atual (nunca usada) — decisão do dono.
- Decisões em aberto no `to-do.md`.

## Arquivos que importam agora
- `../openspec/changes/01-enviar-email/tasks.md` — o que falta
- `System/plano-de-testes.md` — status de cada T-n
- `../internal/integracao/integracao_test.go` — testes que ainda não rodaram
- `../deploy/local/README.md` — ambiente local
