# Estado — ominicore

> Atualizado em 2026-10-02 por Mauricio Miwa + Claude (fatia 4 arquivada)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- **MVP completo**: fatias 1 a 6 entregues e arquivadas; REQ-1 a REQ-19 vigentes,
  T-1 a T-62 passando (integração 30/30). Nenhuma change ativa.
- Tudo commitado na `main` (`74b0912`); as branches `fatia/*` ficam como
  histórico (todas contidas na `main`). Sem remoto ainda.
- Próximo: deploy em hml (runbook `deploy/aviva2/README.md`,
  etapas 1–5b do dono) e validação real dos Provedores (to-do).

## Decisões recentes
- Retentativa = nova linha em `tentativas` com `disponivel_em` futuro, na mesma
  transação da conclusão; sorteio do jitter injetado no router.
- WhatsApp pela Cloud API da Meta, só Modelo aprovado; ordem padrão inalterada.
- hml e prod no mesmo host aviva2.0, separados por realm (`ominicore-hml`),
  prefixo do cofre (`kv/ominicore-hml/*`) e schema (`ominicore_hml`).
- Webhooks em `notify(-hml).flipapi.com.br/webhooks/*`; API só na rede `aviva`.
- Promove-se o mesmo binário de hml para prod; só o dono escreve segredo de prod.

## Bloqueios e pendências
- Runtime local: Colima ativo (contexto `colima`); `make itest` segue o contexto.
  VM do Podman quebrada; com Docker Desktop de 2 GB o Keycloak fica sem memória.
- Decisões em aberto no `to-do.md` (cadastro de Templates, consulta de status,
  chave repetida com pedido diferente, entrega "pelo menos uma vez").

## Arquivos que importam agora
- `../deploy/aviva2/README.md` — runbook do deploy (quem roda cada etapa)
- `System/plano-de-testes.md` — status de cada T-n
- `Infra/AMBIENTES.md` — contrato de hml/prod
- `../deploy/local/bootstrap.sh` — cofre local com auth `jwt` igual a hml/prod
