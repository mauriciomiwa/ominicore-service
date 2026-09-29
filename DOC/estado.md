# Estado — ominicore

> Atualizado em 2026-09-29 por Mauricio Miwa + Claude (fatia 1 verificada)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- Fatias 1, 2, 3 e 6 **entregues e arquivadas**: REQ-1 a REQ-14 vigentes, T-1 a
  T-44 passando. Nenhuma change ativa.
- Branch `fatia/03-status-por-webhook` (a partir de `934d770`): implementação e
  arquivamento da fatia 3 **sem commit**.
- Próximos: fatias 4 (Opt-out, RN-8) e 5 (retenção, RN-9) fecham o MVP; o
  deploy em hml (runbook, etapas 1–5b do dono) valida os webhooks de verdade.

## Decisões recentes
- Retentativa = nova linha em `tentativas` com `disponivel_em` futuro, na mesma
  transação da conclusão; sorteio do jitter injetado no router.
- WhatsApp pela Cloud API da Meta, só Modelo aprovado; ordem padrão inalterada.
- hml e prod no mesmo host aviva2.0, separados por realm (`ominicore-hml`),
  prefixo do cofre (`kv/ominicore-hml/*`) e schema (`ominicore_hml`).
- Webhooks em `notify(-hml).flipapi.com.br/webhooks/*`; API só na rede `aviva`.
- Promove-se o mesmo binário de hml para prod; só o dono escreve segredo de prod.

## Bloqueios e pendências
- VM do Podman quebrada; o local roda no Docker (`COMPOSE="docker compose"`), mas
  a VM do Docker tem 2 GB e o Supabase de outro projeto deixa o Keycloak sem memória.
- Decisões em aberto no `to-do.md` (cadastro de Templates, consulta de status,
  chave repetida com pedido diferente, entrega "pelo menos uma vez").

## Arquivos que importam agora
- `../deploy/aviva2/README.md` — runbook do deploy (quem roda cada etapa)
- `System/plano-de-testes.md` — status de cada T-n
- `Infra/AMBIENTES.md` — contrato de hml/prod
- `../deploy/local/bootstrap.sh` — cofre local com auth `jwt` igual a hml/prod
