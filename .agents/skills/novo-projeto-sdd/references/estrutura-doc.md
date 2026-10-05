# Estrutura DOC/ — o quê e por quê

Árvore-alvo da documentação de um projeto novo. Quem **cria** a estrutura é
`scripts/scaffold.py`, a partir dos arquivos reais em `assets/templates/` —
este documento explica o propósito de cada peça (e serve de guia se for
preciso criar à mão). Em mounts que bloqueiam deleção, renomeie/sobrescreva em
vez de apagar.

Destino: **app único → `DOC/` na raiz do repositório; monorepo/múltiplos
componentes → `DOC/<projeto>/`** (a skill `conhecimento` procura nessa ordem).

```
DOC/<projeto>/
├── INDEX.md       (índice: mapa "onde achar cada coisa")
├── CLAUDE.md      (regra "repositório antes da web" + disciplina SDD)
├── log.md         (registro cronológico; começa com as perguntas de negócio e dev)
├── to-do.md       (pendências e próximos passos — lista viva)
├── brainstorm.md  (ideias e teorias, B-n — não é regra nem plano)
├── Negocio/       (o quê/por quê — especificacao-negocio.md)
├── System/        (o como — por aplicação: arquitetura.md, backlog-mvp.md, plano-de-testes.md)
├── Padroes/       (padrões transversais; stack-tecnologico.md)
├── Skill/         (skills do projeto para a IA)
├── Infra/         (infra, deploy, backend, rede)
├── DRAW/          (diagramas e assets visuais)
├── Suporte/       (materiais de referência; nunca segredos)
├── tools/         (scripts de manutenção da doc)
└── Garbage/       (inventário de descartáveis)

<raiz do repositório>/
├── .githooks/pre-commit            (checa_projeto.py --modo precommit)
├── .github/workflows/sdd-checks.yml (checagens no PR; a sentinel não roda no CI)
├── .claude/settings.json           (hooks: SessionStart, UserPromptSubmit, PreToolUse/Bash, Stop)
└── .sdd/                           (estado local dos hooks — no .gitignore)

<DOC>/tools/                        (checa_projeto.py, sdd_hooks.py, valida_rastreabilidade.py,
                                     dependabot_config.py, diff_ambientes.py, mapeia_projeto.py,
                                     mapa_html.py, registro_log.py, arquiva_log.py — o sync-skills.py
                                     os mantém em dia)

<raiz do repositório>/openspec/     (OpenSpec — o comportamento)
├── config.yaml    (contexto da casa injetado em todo artefato: pt-BR, DOC/, REQ/Origem/T-n)
├── specs/<capacidade>/spec.md      (REQ vigentes + cenários T-n)
└── changes/<NN-slug>/              (uma por fatia/mudança; archive/ = entregues)
```

A DOC/ guarda o porquê, os padrões e a operação; o OpenSpec guarda o
comportamento. Requisito técnico nunca é escrito nos dois lugares.

## Conteúdo mínimo de cada peça

- **INDEX.md** — tabela "se a pergunta é sobre X → vá para Y" + a árvore acima.
- **CLAUDE.md** — regra de ouro (repositório primeiro; web só p/ fatos do mundo
  atual), o link Negócio→System→Teste (SDD), e as convenções herdadas.
- **Negocio/especificacao-negocio.md** — visão, usuários/stakeholders, objetivo/
  métricas, escopo/MVP, regras de negócio, tabela de RN + critérios de aceite,
  restrições, glossário.
- **openspec/** — REQ (frase EARS com SHALL, `Origem: RN-n`, cenários
  `T-n`) em `specs/` (vigente) e nos deltas de `changes/` (planejado).
- **System/arquitetura.md** — componentes, fluxos-chave, decisões.
- **System/backlog-mvp.md** — fatias verticais ordenadas por dependência, cada
  uma com sua `Change:` do OpenSpec, mapeando RN/REQ e com "pronto" = testes
  passando e change arquivada.
- **Padroes/stack-tecnologico.md** — stack + fonte de verdade das versões +
  decisões em aberto.

## Artefatos por porte

Crie conforme o porte decidido no kickoff (seção "Proporcionalidade do
scaffold" da skill). Padrão e Completo somam ao nível anterior. O que não for
criado entra no `INDEX.md` numa seção "Quando precisar, crie", com o caminho
futuro — o lugar reservado é a linha no índice, não um arquivo vazio.

**Núcleo (sempre):**

- `INDEX.md`, `log.md`, `to-do.md` — mapa, cronologia e pendências (o `to-do.md`
  é lista viva, distinta do `backlog-mvp.md`, que tem as fatias planejadas).
  Entradas do `log.md` com mais de 30 dias vão para `log/arquivo-AAAA.md`
  (`tools/arquiva_log.py`), para o log do dia a dia continuar curto.
- `estado.md` — retrato de onde o trabalho parou (≤ 25 linhas), reescrito a
  cada fim de sessão pela skill `passagem` e entregue ao chat novo pelo hook
  de início. É o que a IA lê primeiro.
- `Negocio/especificacao-negocio.md` — RN + critérios de aceite.
- `Negocio/glossario.md` — a linguagem do domínio (termo, definição curta,
  palavras a evitar), atualizada durante a conversa. Ver
  `references/glossario-e-decisoes.md`.
- `System/decisoes/` — criado quando a primeira decisão passar no critério
  dos três pontos (a partir do porte padrão, o scaffold já cria o README).
- `openspec/` na raiz — `config.yaml` (criado ou mesclado pelo scaffold),
  `specs/` e `changes/`. Comandos `/opsx` via `openspec init --tools …`.
- `System/backlog-mvp.md` — fatias verticais, uma change por fatia.
- `AGENTS.md` + `CLAUDE.md` na raiz — seção "Instruções para agentes de IA".

**Padrão (produto com usuários, mais de um dev ou vida > 3 meses):**

- `Padroes/stack-tecnologico.md` — stack + fonte de verdade das versões +
  decisões em aberto.
- `System/plano-de-testes.md` — um teste por cenário T-n das specs, com tipo
  (Exemplo/Propriedade), arquivo e status. A spec diz o que verificar; o plano,
  como e em que pé está.
- `System/arquitetura.md` — componentes, fluxos-chave, decisões.
- `Padroes/goldrules.md` — princípios de engenharia (Clean Code/SOLID,
  API-First, modularidade/DDD, CI-CD, Trunk-Based, IaC, pirâmide de testes,
  shift-left, observabilidade, design for failure, zero trust, scan de
  dependências, IA como assistente). Transversal, por isso mora em `Padroes/`;
  é a base que a skill `sentinel` audita.
- `Padroes/fontes-externas.md` — doc **vivo** das fontes externas confiáveis
  (doc oficial por tecnologia). O time adiciona ao longo do projeto; a IA
  consulta **antes da web** e usa a busca mais econômica em tokens (manifesto
  para versões, deep-link oficial para API, uma query precisa). Aponte para ele
  no `CLAUDE.md`.
- `Infra/seguranca-dependencias.md` — política do Dependabot: recursos ligados
  no repo, cadência (semanal, minor+patch agrupados, major à parte), regra de
  merge (patch e minor de dev/segurança em auto-merge com CI obrigatório
  verde; minor de produção e major com humano), SLA por severidade e regra de
  `ignore`. Acompanha `.github/workflows/dependabot-automerge.yml` na raiz,
  **fail-closed**: sem check obrigatório na branch, não mergeia.

O `.github/dependabot.yml` em si não é doc e sai em **qualquer porte**: é
gerado dos manifestos por `tools/dependabot_config.py`, que também confere
cobertura (`--check`) quando aparece manifesto novo.

**Completo (multi-app, multi-tenant, compliance/PII):**

- `Padroes/i18n-string-files.md` (nenhum texto hardcoded) + template
  `System/i18n/strings.pt-BR.json` chave-valor (move para o app quando o
  código começar).
- Sub-estrutura `System/<app>/` — seção no fim deste arquivo.

> Regra-mãe: artefato sem dono e sem informação real não entra no scaffold —
> nasce quando a primeira informação chegar. Documento vazio não é neutro: a
> skill `conhecimento` passa a apontar o índice para arquivo oco.

## Instruções para agentes de IA — criar SEMPRE na raiz do repositório

Objetivo: qualquer LLM (Claude, Codex, Windsurf/Cascade, Devin, Cursor…) que
abrir o repositório é obrigada a seguir a documentação do projeto, sem depender
de skill instalada.

1. **`AGENTS.md`** (raiz) — padrão lido por Codex, Windsurf, Devin, Cursor e
   outros. O conteúdo é o de `assets/templates/_raiz/AGENTS.md` (o scaffold o
   instala; não copie daqui, para não haver duas versões): fonte de verdade
   (`DOC/` + `openspec/`), disciplina SDD, `Padroes/`, a tabela "Skills do
   projeto" com o gatilho de cada skill e a seção "Garantias" (auditoria da
   `sentinel` antes do commit, `checa_projeto.py`, pre-commit, CI e hooks).

2. **`CLAUDE.md`** (raiz) — para Claude Code/Cowork. Conteúdo: uma linha
   apontando para o `AGENTS.md` ("Siga as instruções de `AGENTS.md`") + o que
   for específico do Claude. Não duplique o conteúdo — um único lugar a manter.

3. **`.agents/skills/`** (raiz) — o scaffold copia todas as skills do
   repositório de skills (nomad-worker só em projeto da DKM com job Nomad) e o
   `sync-skills.py` as mantém em dia. Skills específicas do projeto
   (`DOC/<projeto>/Skill/`) também vão para cá. Codex, Windsurf e Devin leem
   esse diretório automaticamente; assim as skills viajam com o repositório
   sem instalação.

> Ajuste os caminhos do template ao projeto real (ex.: `DOC/INDEX.md` se não
> houver subpasta por projeto) e mantenha `AGENTS.md` curto — detalhe fica na
> `DOC/`, ele só obriga a consulta.

## Sub-estrutura por aplicação (quando houver mais de um componente)

```
System/<app>/
├── arquitetura/
├── manual/
└── operacao/
```

Os requisitos de cada app ficam no OpenSpec, em capacidades prefixadas pelo
app/projeto (`openspec/specs/<projeto>-<capacidade>/`); o validador filtra
por esse prefixo quando a DOC/ é `DOC/<projeto>/`.

## Migrar do EARS para o OpenSpec (projeto existente)

Para projetos criados antes da adoção, com `System/requisitos-ears.md`. Faça
num único PR/commit, **sem** change do OpenSpec — é a linha de base do que já
existe, não uma mudança de comportamento:

1. `openspec init --tools claude,agents` na raiz e rode o `scaffold.py` de
   novo com o mesmo porte — ele não sobrescreve nada, mescla o
   `openspec/config.yaml` e avisa se `tools/valida_rastreabilidade.py` está
   desatualizado: copie já a versão nova (o comando vem no aviso). Enquanto
   houver REQ só no `requisitos-ears.md`, o validador acusa "migração
   incompleta".
2. Agrupe os REQ por capacidade (domínio/bounded context) e escreva
   **direto** em `openspec/specs/<capacidade>/spec.md` — é comportamento que
   já existe, então é spec vigente, não delta. Cada linha da tabela vira:
   título `### Requirement: REQ-n <nome>` (mesmo ID), a frase EARS trocando
   "deve" por "SHALL", `Origem: RN-n`, e o teste `T-n` como
   `#### Scenario: T-n …` com WHEN/THEN. Cada spec começa com `## Purpose`.
3. REQ ainda não implementado não é vigente: vai para uma change
   (`openspec/changes/<NN-slug>/`) como `## ADDED Requirements`.
4. `openspec validate --all --strict` e `tools/valida_rastreabilidade.py`
   sem furo.
5. Substitua o conteúdo de `requisitos-ears.md` por um ponteiro para
   `openspec/specs/` (ou apague-o), atualize `INDEX.md`, `AGENTS.md` e o
   `backlog-mvp.md` (linha `Change:` por fatia) e registre no `log.md`.

