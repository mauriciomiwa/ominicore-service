# Estado — ominicore

> Atualizado em 2026-09-29 por Mauricio Miwa + Claude (fatia 1 verificada)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- Fatia 1 entregue e arquivada (REQ-1 a REQ-6 vigentes).
- Fatia 2 `02-sms-e-fallback`: **24/24 tarefas**, T-15 a T-27 passando (unitários
  e 14/14 integração no Postgres). Branch `fatia/02-sms-e-fallback`, **sem commit**.
- Próximo passo: commit, `/opsx:archive 02-sms-e-fallback`; depois a trilha de
  infraestrutura (deploy no aviva2.0, precisa estar pronta antes da fatia 3).

## Decisões recentes
- Retentativa = nova linha em `tentativas` com `disponivel_em` futuro, na mesma
  transação da conclusão; sorteio do jitter injetado no router.
- Twilio por API Key + Messaging Service; regra de classificação comum em
  `saida/provedor`.
- hml e prod no mesmo host aviva2.0, separados por realm (`ominicore-hml`),
  prefixo do cofre (`kv/ominicore-hml/*`) e schema (`ominicore_hml`).
- Webhooks em `notify(-hml).flipapi.com.br/webhooks/*`; API só na rede `aviva`.
- Promove-se o mesmo binário de hml para prod; só o dono escreve segredo de prod.

## Bloqueios e pendências
- VM do Podman quebrada; o local roda no Docker (`COMPOSE="docker compose"`).
- Decisões em aberto no `to-do.md` (cadastro de Templates, consulta de status,
  chave repetida com pedido diferente, entrega "pelo menos uma vez").

## Arquivos que importam agora
- `../openspec/changes/02-sms-e-fallback/` — change a arquivar
- `System/plano-de-testes.md` — status de cada T-n
- `Infra/AMBIENTES.md` — contrato de hml/prod
- `../deploy/local/bootstrap.sh` — cofre local com auth `jwt` igual a hml/prod
