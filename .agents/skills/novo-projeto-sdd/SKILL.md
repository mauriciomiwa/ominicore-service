---
name: novo-projeto-sdd
description: >-
  Kickoff e manutenção de projeto sob SDD (RN → REQ → teste, DOC/ + OpenSpec,
  fatias como changes). Use SEMPRE para iniciar ou escopar projeto, levantar
  requisitos ou MVP, criar a DOC/, adotar OpenSpec, auditar projeto existente e
  suas garantias, ou mapear o código.
---

# Projeto sob SDD

Separa **Negócio** (o quê/por quê) de **Desenvolvimento** (o como), para cada
teste validar um critério de negócio. Não assuma nada: o que o repositório
responde você confirma; o resto você pergunta.

| Mora em | O quê |
|---|---|
| `DOC/Negocio/` | RN, critérios de aceite, glossário |
| `DOC/System/`, `Padroes/`, `Infra/` | arquitetura, decisões, backlog, plano de testes, padrões, ambientes |
| `DOC/estado.md` | onde o trabalho parou (skill `passagem`) |
| `openspec/specs/` | REQ vigentes + cenários T-n |
| `openspec/changes/` | cada fatia: proposal (RN), deltas, design, tasks |

Convenção (o `openspec/config.yaml` a injeta em todo artefato):

```
### Requirement: REQ-12 <nome curto>
Quando <evento>, o sistema SHALL <resposta verificável>.   ← EARS; SHALL/MUST em inglês
Origem: RN-3

#### Scenario: T-20 <nome>                                  ← cada T-n vira um teste que cita o ID
- **WHEN** … - **THEN** …
```

IDs REQ-n e T-n são únicos no repositório (próximo = maior existente + 1,
contando specs, changes, `archive/` e previstos do backlog); RN-n valem por
projeto; monorepo prefixa as capacidades (`loja-vendas/`).

## Qual é a situação?

| Situação | Faça |
|---|---|
| Projeto novo (greenfield) | o fluxo abaixo; detalhe de cada passo em `references/kickoff.md` |
| Repositório já tem código, `DOC/` ou `AGENTS.md` | `references/projeto-existente.md`: auditar → mostrar → blindar → mapear |
| Mapear o código (telas, rotas, endpoints, integrações) | `references/mapa-projeto.md` |
| Hooks, pre-commit, CI, sentinel, `.sentinelignore` | `references/garantias.md` |
| Estrutura da `DOC/`, portes, migrar EARS → OpenSpec | `references/estrutura-doc.md` |
| Glossário e decisões (ADR) | `references/glossario-e-decisoes.md` |
| `util.md` e `migrations.md` (colas fixas no Sentinel) | `references/util-e-migrations.md` |
| Formato de backlog, change e spec | `references/exemplos.md` |

## Como perguntar

Use a skill `sabatina`: rodadas com a resposta recomendada em cada pergunta,
fatos você busca, nada é criado antes do "fechado". `references/perguntas.md`
é o roteiro do que precisa ficar decidido (Rodada 1 negócio, Rodada 2 como),
não um questionário para despejar. Termo decidido entra no glossário na hora;
decisão que passa no critério de ADR vira `System/decisoes/NNNN-titulo.md`
(`references/glossario-e-decisoes.md`).

## Fluxo do kickoff

0. **Inventário antes de perguntar:** manifestos, CI, infra, `DOC/` existente
   (tabela em `references/kickoff.md`). Detectado é evidência a confirmar,
   nunca resposta; só vale para a Rodada 2.
1. **Negócio:** Rodada 1 → `Negocio/especificacao-negocio.md`, cada RN com
   critério de aceite verificável.
2. **Escopo/MVP:** dentro × fora; cada RN marcada MVP ou "depois".
3. **Desenvolvimento:** Rodada 2 → `Padroes/stack-tecnologico.md` (cite o
   arquivo que originou cada item; não confirmado = decisão em aberto).
4. **Estrutura:** `openspec init --tools claude,agents`, depois
   `scripts/scaffold.py <DOC/> --projeto <nome> --porte <nucleo|padrao|completo>`
   (porte derivado das respostas; proponha e confirme). Ele nunca sobrescreve;
   depois preencha os templates com as respostas.
5. **Arquitetura e fatias:** `System/arquitetura.md` e `System/backlog-mvp.md`
   com fatias verticais, cada uma com `Change: \`NN-slug\``.
6. **Changes:** abra só a fatia 1 e as independentes (`/opsx:propose`, após a
   sabatina). Delta `## ADDED Requirements` no formato acima; capacidade nova
   com `## Purpose` (50+ caracteres); `tasks.md` com uma tarefa por T-n,
   penúltima "rodar a validação", última "auditoria final do diff completo
   com a sentinel" — complementa a auditoria de cada tarefa, não a substitui.
7. **Agentes:** confira `AGENTS.md`, `CLAUDE.md` e `.agents/skills/` (o
   scaffold cria; `sync-skills.py` atualiza).
8. **Validar:** `openspec validate --all --strict` e
   `scripts/valida_rastreabilidade.py <DOC/>`. Furo que ficar vira item no
   `to-do.md`, nunca silêncio.

Pronto = RN com aceite e MVP fechado, stack com decisões em aberto explícitas,
`DOC/` + `openspec/` criados, fatias no backlog, change(s) iniciais validadas,
Dependabot cobrindo os manifestos, `checa_projeto.py` sem FURO e o `log.md`
com as decisões (lista completa em `references/kickoff.md`).

## Ciclo de cada mudança

`sabatina` → `/opsx:propose` (com RN de origem) → `/opsx:apply` (a `sentinel`
audita o diff de cada tarefa) → `openspec validate` + `valida_rastreabilidade.py` →
deploy pela `ambientes` → `/opsx:archive` → `log.md`, `plano-de-testes.md` e
`passagem` no fim da sessão.

Em mount que bloqueia deleção, sobrescreva em vez de apagar.
