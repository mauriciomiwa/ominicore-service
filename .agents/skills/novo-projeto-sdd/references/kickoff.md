# Kickoff de projeto novo — passo a passo

Detalhe do fluxo resumido no SKILL.md. Leia a seção do passo em que está.

## Passo 0 — Inventário do ambiente (antes da primeira pergunta)

Perguntar o que já está escrito no repositório queima a paciência do usuário, e
ele passa a responder raso justamente onde importa. Faça uma varredura rasa
antes: raiz e até dois níveis, só manifestos e arquivos de configuração, sem ler
código-fonte.

| Onde olhar | O que responde |
|---|---|
| remote do git, nome da pasta, `README` | nome e contexto do projeto |
| manifesto (`package.json`, `pom.xml`, `build.gradle*`, `go.mod`, `pyproject.toml`, `*.csproj`, `composer.json`, `pubspec.yaml`) + lockfile | linguagem, gerenciador, versões reais (dev 2 e 4) |
| `.github/workflows`, `.gitlab-ci.yml`, `Jenkinsfile` | build e CI (dev 4) |
| `Dockerfile`, `*.hcl`, `compose*.yml`, manifests k8s, `*.tf` | infra-alvo de deploy (dev 8) |
| migrations, `*.sql`, config de conexão | persistência (dev 3) |
| `.env.example`, config de auth | integrações e esquema de auth (dev 5 e 9) |
| `DOC/`, `AGENTS.md`, `CLAUDE.md` existentes | se isto é kickoff mesmo — ver saída antecipada |
| pastas ou repos irmãos com `DOC/Padroes/` | convenções a referenciar em vez de duplicar (dev 1) |

Pasta vazia é resposta legítima: diga em uma linha que é greenfield e siga para
as perguntas.

**Saída antecipada — projeto existente:** se o repositório já tem código, `DOC/`
ou `AGENTS.md`, isto não é greenfield. Rode primeiro a auditoria completa
(`references/projeto-existente.md`) e mostre o resultado; depois pergunte
se o pedido é (a) blindar o projeto — instalar o que falta para as skills
rodarem —, (b) escopar um módulo novo dentro dele — siga o fluxo,
reaproveitando os `Padroes/` de lá — ou (c) uma dúvida sobre o que já está
documentado, que é da skill `conhecimento`.

### O que fazer com o inventário

O detectado é evidência para confirmar, nunca resposta:

- **Só vale para a Rodada 2 (o como).** A Rodada 1 continua integralmente
  perguntada — nenhum arquivo de build revela objetivo de negócio.
- **Detectado não é desejado.** O que está lá pode ser exatamente o que vai ser
  substituído (Negócio, pergunta 4). Pergunte na forma: "achei Spring Boot 3.2
  no `pom.xml` — é a base para seguir ou é o legado a substituir?"
- **Projeto irmão com `Padroes/` e `Skill/`:** referencie os arquivos dele em vez
  de copiar; registre só o que é específico deste projeto. Duplicata vira
  divergência na primeira mudança.

Abra a Rodada 2 com o resumo do que achou, cada item com o arquivo ao lado, e
peça confirmação em bloco; pergunte aberto só o que não deu para detectar. Em
`Padroes/stack-tecnologico.md`, cite o arquivo que originou cada item — depois
ninguém precisa adivinhar se aquilo foi decisão ou chute. O que você detectou e
o usuário não confirmou entra como **decisão em aberto**, não como fato. Registre
o inventário no `log.md`: ele é insumo das decisões.

## Proporcionalidade do scaffold

O mesmo scaffold para tudo produz documentação morta num projeto pequeno — e
documento vazio não é neutro: a skill `conhecimento` promete "repositório
primeiro" e passa a apontar o índice para arquivo oco, ensinando o time (e a
IA) a desconfiar da `DOC/`. Calibre o que criar pelo porte:

- **Núcleo (qualquer porte):** `INDEX.md`, `estado.md`, `log.md`, `to-do.md`, `brainstorm.md`,
  `Negocio/especificacao-negocio.md`, `Negocio/glossario.md`,
  `System/backlog-mvp.md`, `AGENTS.md` + `CLAUDE.md` e o `openspec/` na raiz
  (`config.yaml`, `specs/`, `changes/`).
  É o mínimo que mantém a promessa RN → REQ → teste.
- **Padrão** (produto com usuários, mais de um dev ou vida esperada > 3 meses):
  núcleo + `Padroes/stack-tecnologico.md`, `System/plano-de-testes.md`,
  `System/arquitetura.md`, `System/decisoes/` (README com o critério),
  `Padroes/goldrules.md`, `Padroes/fontes-externas.md`,
  `Infra/seguranca-dependencias.md` + workflow de auto-merge do Dependabot.
- **Completo** (multi-app, multi-tenant, compliance/PII): tudo — i18n/
  string-files, sub-estrutura
  `System/<app>/`.

O porte não é pergunta ao usuário — deriva de respostas que as Rodadas 1 e 2
já colhem (volume/escala, multi-tenant, compliance, nº de componentes).
Proponha o porte com a justificativa e peça confirmação, como no Passo 0.

O que ficar de fora não some: o `INDEX.md` ganha uma seção "Quando precisar,
crie" listando cada artefato não criado e seu caminho futuro — o lugar
reservado é uma linha no índice, não um arquivo vazio. Regra-mãe: artefato sem
dono e sem informação real não entra no scaffold; nasce quando a primeira
informação chegar.

## Fluxo detalhado (faça nesta ordem)

1. **Perguntas de negócio.** Faça a Rodada 1 de `references/perguntas.md`.
   Comece pela visão/problema (destrava o resto). Consolide em
   `Negocio/especificacao-negocio.md`, cada requisito de negócio (RN-1, RN-2…)
   com um **critério de aceite verificável**.
2. **Fechar escopo/MVP.** Defina explicitamente o que está dentro e fora, e qual
   é a menor versão útil (MVP). Marque cada RN como MVP ou "depois".
3. **Perguntas de desenvolvimento.** Faça a Rodada 2 de `references/perguntas.md`
   (base/stack, auth, multi-tenant, exposição/segurança, infra-alvo,
   integrações, NFRs; condicionais só quando o projeto indicar). Registre em
   `Padroes/stack-tecnologico.md`.
4. **Inicializar o OpenSpec e gerar a estrutura `DOC/`.** Primeiro o
   OpenSpec, que é dono dos arquivos que gera: na raiz do repositório,
   `openspec init --tools claude,agents` (acrescente `cursor`, `codex`… conforme
   as ferramentas do time; requer Node 20.19+ e
   `npm install -g @fission-ai/openspec@latest`). Sem como instalar agora,
   siga — o scaffold cria o `openspec/` mínimo e avisa; registre no `to-do.md`.
   Depois o scaffold. Destino da DOC/ segue a convenção: app único →
   `DOC/` na raiz; monorepo/múltiplos componentes → `DOC/<projeto>/`. Rode
   `scripts/scaffold.py <destino da DOC/> --projeto <nome>
   --porte <nucleo|padrao|completo>` (porte: seção "Proporcionalidade do scaffold" acima). O script monta a árvore a partir de
   `assets/templates/`, escreve o "Quando precisar, crie" no `INDEX.md`, põe
   `AGENTS.md`/`CLAUDE.md` na raiz, cria ou **mescla** o
   `openspec/config.yaml` com o contexto da casa (pt-BR, ponteiro para a DOC/,
   convenção REQ/Origem/T-n; nunca remove o que o `init` gravou), instala as
   garantias (`references/garantias.md`: hooks do Claude, pre-commit, CI), gera o
   `.github/dependabot.yml` a partir
   dos manifestos (qualquer porte — Dependabot é o padrão da casa; greenfield
   só avisa) e **não sobrescreve nada existente** — é
   seguro re-rodar para subir de porte. Depois **preencha** os templates com o
   que o kickoff respondeu: o script monta a moldura, o conteúdo vem das
   respostas. Sem como executar o script, crie à mão copiando
   `assets/templates/`; o quê/porquê de cada peça está em
   `references/estrutura-doc.md`.
5. **Arquitetura + fatiamento.** Escreva `System/arquitetura.md` (o "como") e
   `System/backlog-mvp.md` quebrando o MVP em **fatias verticais** (cada fatia
   entrega valor fim-a-fim e tem "pronto" = seus testes passando). Cada fatia
   ganha um nome de change em kebab-case com número de ordem (`01-registrar-venda`)
   na linha `Change:`. Formato: `references/exemplos.md`.
6. **Abrir a(s) primeira(s) change(s) (requisitos técnicos).** No kickoff,
   abra só a fatia 1 e as que não dependem de nenhuma outra; as demais ficam
   no backlog com `Change:`, RN e REQ previstos, e cada change é aberta quando
   a fatia começar. Abrir tudo de uma vez envelhece design e tarefas, e uma
   fatia não pode modificar (MODIFIED) um REQ de outra que ainda não foi
   arquivada. Para cada change aberta, crie `openspec/changes/<NN-slug>/` com:
   - `proposal.md` — porquê citando as RN; escopo dentro/fora;
   - `specs/<capacidade>/spec.md` — delta `## ADDED Requirements` traduzindo
     cada RN em um ou mais REQ no formato da convenção acima (frase EARS com
     SHALL, `Origem: RN-n`, ao menos um `#### Scenario: T-n`). Capacidade nova
     leva `## Purpose` no delta (uma ou duas frases, 50+ caracteres);
   - `tasks.md` — uma tarefa por T-n, penúltima "rodar a validação", última
     "auditoria final do diff completo com a sentinel" (complementa a auditoria
     de cada tarefa no `/opsx:apply`). `design.md` só quando houver decisão técnica
     que valha registrar.
   Com os comandos instalados, `/opsx:propose <fatia>` gera tudo isso já no
   formato (o `config.yaml` injeta a convenção) — revise RN/REQ/T-n antes de
   aceitar. Sem eles, crie os arquivos à mão. IDs REQ-n e T-n são únicos no
   repositório inteiro e sequenciais — o próximo é o maior já usado nas specs,
   nas changes (inclusive `archive/`) e nos previstos do backlog, mais 1; RN-n
   valem por projeto. As `openspec/specs/` ficam **vazias até o primeiro
   `/opsx:archive`** — correto: spec vigente descreve o que já existe, e no
   greenfield nada existe ainda. Monorepo: um `openspec/` na raiz e
   capacidades prefixadas pelo projeto (`loja-vendas/`). Formato completo:
   `references/exemplos.md`.
7. **Instruções para agentes de IA.** O scaffold (passo 4) já criou o
   `AGENTS.md` na raiz (obriga qualquer LLM — Codex, Windsurf, Devin, Cursor —
   a consultar `DOC/` antes de agir), o `CLAUDE.md` apontando para ele, e
   copiou para `.agents/skills/` todas as skills do repositório de skills
   (nomad-worker só em projeto da DKM com job Nomad) quando as fontes estavam
   acessíveis (ele avisa quando não estão). As skills `openspec-*` que o
   `openspec init` pôs em `.agents/skills/` convivem ali — são do OpenSpec (o
   `openspec update` as regenera); não edite, customize pelo `config.yaml`. A
   cópia repetida em `.claude/skills/` sai com o `sync-skills.py` (o `/opsx:*`
   continua). Aqui, confira que tudo
   existe e acrescente as skills específicas do projeto. Para atualizar as
   cópias depois, rode o `sync-skills.py` do repositório de skills. Template e
   detalhes: seção "Instruções para agentes de IA" em
   `references/estrutura-doc.md`.
8. **Validar.** Duas checagens complementares:
   `openspec validate --all --strict` (formato: seções, SHALL/MUST, cenários) e
   `scripts/valida_rastreabilidade.py <caminho da DOC/>` (a cadeia): ele lê as
   specs vigentes e os deltas das changes ativas e aponta REQ sem ID, sem RN
   ou sem cenário T-n, RN inexistente, ID reaproveitado, T-n em dois REQ,
   fatia citando REQ ou change fantasma, T-n ausente do plano. Cruzar isso
   "no olho" é onde a leitura falha — o script é determinístico. Furos que você
   não corrigir agora viram itens no `to-do.md`, nunca silêncio. Se o script
   reportar que não conseguiu parsear (formato fora do padrão), faça a
   checagem manual e diga isso no relato. O scaffold já o copiou para
   `<DOC>/tools/` — a validação vale a cada change, muito depois do kickoff.

## "Pronto" do kickoff

- `Negocio/especificacao-negocio.md` com RN + critérios de aceite e MVP fechado.
- `Padroes/stack-tecnologico.md` com a stack e decisões em aberto explícitas.
- Estrutura `DOC/` criada (INDEX, CLAUDE, subdiretórios).
- `openspec/` inicializado (`config.yaml` com o bloco `sdd-rastreabilidade`;
  comandos `/opsx` instalados ou pendência no `to-do.md`).
- `System/arquitetura.md` e `System/backlog-mvp.md` (fatias verticais ordenadas,
  cada uma com sua `Change:`).
- Change(s) da fatia 1 (e das fatias independentes) abertas com REQ
  rastreáveis (RN → REQ → T-n) e `openspec validate --all --strict` passando;
  demais fatias no backlog com `Change:`, RN e REQ previstos.
- `AGENTS.md` + `CLAUDE.md` na raiz e `.agents/skills/` com as skills do projeto.
- `.github/dependabot.yml` cobrindo todos os manifestos
  (`tools/dependabot_config.py --check` com rc 0) — ou, em greenfield, item no
  `to-do.md` para gerar quando o primeiro manifesto existir. Do porte padrão
  em diante: `Infra/seguranca-dependencias.md` com SLA e donos preenchidos e
  os pré-requisitos do repositório (alerts, security updates, auto-merge,
  checks obrigatórios) listados no `to-do.md` até serem ligados.
- Validação de rastreabilidade rodada (passo 8), sem furo aberto — ou com os
  furos registrados no `to-do.md`.
- Garantias instaladas e `python3 <DOC>/tools/checa_projeto.py .` sem FURO;
  job `checagens` do `sdd-checks` obrigatório na branch (ou pendência no `to-do.md`).
- `log.md` registrando todas as decisões.
