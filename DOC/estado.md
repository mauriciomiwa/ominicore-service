# Estado — ominicore

> Atualizado em 2026-09-29 por Mauricio Miwa + Claude (fatia 1 verificada)

Retrato de onde o trabalho parou, reescrito a cada fim de sessão (skill
`passagem`, no máximo 40 linhas). É o que o próximo chat lê primeiro.

## Em curso
- Fatia 1 **entregue e arquivada** (`openspec/changes/archive/2026-09-29-01-enviar-email/`);
  specs vigentes `autenticacao-origem`, `notificacoes`, `segredos` (REQ-1 a REQ-6,
  T-1 a T-14 passando). Nenhuma change ativa.
- Branch `fatia/01-enviar-email` (commits até `df297b2`; o arquivamento ainda sem commit).
- Próximo passo: sabatina da fatia 02 (`02-sms-e-fallback`: SMS/Twilio, retry
  3× e fallback pela Ordem de Canais — RN-2, RN-3), depois `/opsx:propose`.

## Decisões recentes
- hml e prod no mesmo host aviva2.0, separados por realm (`ominicore-hml`),
  prefixo do cofre (`kv/ominicore-hml/*`) e schema (`ominicore_hml`).
- Webhooks em `notify(-hml).flipapi.com.br/webhooks/*`; API só na rede `aviva`.
- Promove-se o mesmo binário de hml para prod; só o dono escreve segredo de prod.

## Bloqueios e pendências
- VM do Podman quebrada; o local roda no Docker (`COMPOSE="docker compose"`).
- Decisões em aberto no `to-do.md` (cadastro de Templates, consulta de status,
  chave repetida com pedido diferente, entrega "pelo menos uma vez").

## Arquivos que importam agora
- `../openspec/specs/` — comportamento vigente (REQ-1 a REQ-6)
- `System/plano-de-testes.md` — status de cada T-n
- `Infra/AMBIENTES.md` — contrato de hml/prod
- `../deploy/local/bootstrap.sh` — cofre local com auth `jwt` igual a hml/prod
