# System/ — o como

Arquitetura, backlog fatiado, plano de testes, ADRs. Tudo aqui deriva de
`Negocio/` — nunca o contrário. Com mais de um componente, crie a
sub-estrutura `System/<app>/`.

Os **requisitos técnicos** (REQ + cenários de teste) não moram aqui: vivem no
OpenSpec, na raiz do repositório — `openspec/specs/` (comportamento vigente) e
`openspec/changes/` (mudanças em curso). Formato e regras de ID em
`openspec/config.yaml`.
