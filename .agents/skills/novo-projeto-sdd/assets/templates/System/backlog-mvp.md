# Backlog do MVP — {{PROJETO}}

Fatias verticais ordenadas por dependência: cada uma entrega valor fim-a-fim
(UI → API → banco) e tem "pronto" = seus testes passando. "Fazer todo o
banco" ou "fazer todas as telas" não é fatia — é camada.

**Cada fatia é uma change do OpenSpec** (`openspec/changes/<change>/`): o
proposal diz o porquê (RN), os deltas em `specs/` trazem os REQ e cenários
(T-n), o `tasks.md` é o passo a passo. Este arquivo é só a ordem e o mapa —
o detalhe mora na change. Status: change ativa = em curso; movida para
`openspec/changes/archive/` = entregue.

### Fatia 1 — <nome>
Change: `01-<slug>`
Valor: <o que o usuário passa a conseguir fazer>
Cobre: RN-1 (REQ-1)
Depende de: —
Pronto quando: T-1 passando e change arquivada.
