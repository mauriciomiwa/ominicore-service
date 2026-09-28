# ominicore — Índice da documentação

Mapa "se a pergunta é sobre X → vá para Y". Mantenha-o atualizado ao criar ou
mover qualquer doc — a regra "repositório primeiro" depende dele.

| Se a pergunta é sobre… | Vá para |
|---|---|
| onde o trabalho parou, próximo passo (leia primeiro) | `estado.md` |
| o quê / por quê, RN, critérios de aceite | `Negocio/especificacao-negocio.md` |
| o que significa um termo do domínio (e qual palavra evitar) | `Negocio/glossario.md` |
| por que uma escolha de arquitetura foi feita (decisões difíceis de reverter) | `System/decisoes/` |
| componentes, layout hexagonal e fluxos | `System/arquitetura.md` |
| stack, versões e lint | `Padroes/stack-tecnologico.md` |
| princípios que a sentinel audita | `Padroes/goldrules.md` |
| docs oficiais de cada tecnologia | `Padroes/fontes-externas.md` |
| como e onde cada T-n é testado | `System/plano-de-testes.md` |
| política do Dependabot, SLA de CVE | `Infra/seguranca-dependencias.md` |
| comportamento vigente: REQ e cenários de teste (T-n) | `../openspec/specs/` (OpenSpec) |
| mudanças em curso: proposal, deltas, tarefas | `../openspec/changes/` (OpenSpec) |
| por que uma mudança entregue foi feita | `../openspec/changes/archive/` |
| fatias do MVP e ordem de entrega (fatia = change) | `System/backlog-mvp.md` |
| como o código está organizado: telas, endpoints, integrações | `System/mapa/visao-geral.md` (fatos em `System/mapa/inventario.md`, gerado por `tools/mapeia_projeto.py`) |
| problemas abertos e histórico das execuções dos scripts SDD | `log/problemas.md` e `log/historico.jsonl` (gerados pelos scripts e pelo openmd) |
| decisões e histórico | `log.md` |
| pendências e lembretes | `to-do.md` |
<!-- acrescente linhas conforme os docs forem criados -->

## Quando precisar, crie

Artefatos padrão ainda não criados neste porte — o lugar reservado é esta
lista, não um arquivo vazio. Crie quando a primeira informação real chegar:

- `Padroes/i18n-string-files.md` — convenção i18n: nenhum texto hardcoded
- `System/i18n/strings.pt-BR.json` — arquivo de strings chave-valor
- `DRAW/README.md` — diagramas (fonte editável + export)
- `Suporte/README.md` — materiais de referência (nunca segredos)
- `Garbage/README.md` — inventário de descartáveis
