# {{PROJETO}} — Índice da documentação

Mapa "se a pergunta é sobre X → vá para Y". Mantenha-o atualizado ao criar ou
mover qualquer doc — a regra "repositório primeiro" depende dele.

| Se a pergunta é sobre… | Vá para |
|---|---|
| onde o trabalho parou, próximo passo (leia primeiro) | `estado.md` |
| o quê / por quê, RN, critérios de aceite | `Negocio/especificacao-negocio.md` |
| stack, onde roda, onde está o banco de hml e de prod | `Infra/infraestrutura.md` |
| o que significa um termo do domínio (e qual palavra evitar) | `Negocio/glossario.md` |
| por que uma escolha de arquitetura foi feita (decisões difíceis de reverter) | `System/decisoes/` (quando houver) |
| comportamento vigente: REQ e cenários de teste (T-n) | `{{OPENSPEC_REL}}/specs/` (OpenSpec) |
| mudanças em curso: proposal, deltas, tarefas | `{{OPENSPEC_REL}}/changes/` (OpenSpec) |
| por que uma mudança entregue foi feita | `{{OPENSPEC_REL}}/changes/archive/` |
| fatias do MVP e ordem de entrega (fatia = change) | `System/backlog-mvp.md` |
| como o código está organizado: telas, endpoints, integrações | `System/mapa/visao-geral.md` (fatos em `System/mapa/inventario.md`, gerado por `tools/mapeia_projeto.py`) |
| problemas abertos e histórico das execuções dos scripts SDD | `log/problemas.md` e `log/historico.jsonl` (gerados pelos scripts e pelo openmd) |
| decisões e histórico | `log.md` |
| pendências e lembretes | `to-do.md` |
| ideias e teorias (não é regra nem plano de implantação) | `brainstorm.md` |
<!-- acrescente linhas conforme os docs forem criados -->

## Quando precisar, crie

Artefatos padrão ainda não criados neste porte — o lugar reservado é esta
lista, não um arquivo vazio. Crie quando a primeira informação real chegar:

<!-- QUANDO-PRECISAR-CRIE -->
