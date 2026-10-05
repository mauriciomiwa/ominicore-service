# Estado — ominicore

> Atualizado em 2026-10-05 por Mauricio Miwa + Claude (change 07 arquivada)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- Fatia 7 (`07-whatsapp-opcional`) entregue e arquivada; REQ-1 a REQ-20 vigentes,
  T-1 a T-69 passando. Nenhuma change ativa. Branch `fatia/07-whatsapp-opcional`
  ainda sem commit.
- Próximo passo: commitar a fatia 7 e levar para a `main`.
- Depois: deploy em hml pelo runbook `deploy/aviva2/README.md` (etapas do dono),
  respondendo N ao WhatsApp; ligar depois com `configurar-openbao.sh hml --so-whatsapp`.

## Decisões recentes
- WhatsApp desligado = ausência de `kv/<prefixo>/whatsapp`; recusa 422, pulo na
  ordem, webhook 404 (`log.md` 2026-10-05; `design.md` da change 07).
- MVP (fatias 1–6) entregue e commitado na `main`; sem remoto ainda.
- hml e prod no mesmo host aviva2.0, separados por realm, prefixo do cofre e
  schema; promove-se o mesmo binário (`Infra/AMBIENTES.md`).

## Bloqueios e pendências
- Conta Meta Business, Modelo `boas_vindas` aprovado e token: não bloqueiam mais
  o deploy depois da change 07 (`to-do.md`, Fatia 06).
- Runtime local: Colima (contexto `colima`); `make itest` segue o contexto.
- Decisões em aberto no `to-do.md` (chave repetida com pedido diferente,
  entrega "pelo menos uma vez", cadastro de Templates) — fora da change 07.

## Arquivos que importam agora
- `../internal/adaptadores/saida/openbao/segredos.go` — leitura opcional do WhatsApp
- `../deploy/aviva2/openbao/configurar-openbao.sh` — "[s/N]" e `--so-whatsapp` (sem shellcheck ainda)
- `../deploy/aviva2/README.md` — runbook do deploy
