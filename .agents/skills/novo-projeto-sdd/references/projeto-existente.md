# Projeto existente — auditar, blindar e mapear

Quando o repositório já tem código, `DOC/` ou `AGENTS.md`, isto não é greenfield: a primeira entrega é o diagnóstico, não perguntas.

## Modo projeto existente — auditar o diretório todo e blindar

Quando o repositório já existe, a primeira entrega é o diagnóstico, não
perguntas:

1. **Auditar.** Da pasta desta skill, rode
   `python3 scripts/checa_projeto.py <raiz-do-repo>`. Ele varre o repositório
   inteiro e reporta, por skill, FURO / ALERTA / OK:
   - *estrutura* — `DOC/INDEX.md`, `AGENTS.md` citando cada skill, `CLAUDE.md`,
     `.agents/skills/` com todas as skills do repositório (nomad-worker só em
     projeto da DKM com job Nomad) e sem drift em relação à fonte, `openspec/config.yaml`
     com o contexto da casa, comandos `/opsx`, garantias instaladas
     (pre-commit ligado no git, CI `sdd-checks`, hooks do Claude);
   - *novo-projeto-sdd* — `valida_rastreabilidade.py` em cada DOC/ (monorepo
     incluído), `openspec validate --strict` e `--archived`, todo cenário T-n
     vigente citado no nome de algum teste do código;
   - *dependências* — Dependabot cobrindo todos os manifestos;
   - *ambientes* — `AMBIENTES.md` quando há artefato de deploy, pares de config
     hml/prod sem chave órfã (respeitando as divergências declaradas);
   - *nomad-worker* — imagem sem `:latest`/sem tag, nenhum segredo literal;
   - *sentinel* — cada arquivo de código alterado foi auditado no conteúdo
     atual.
2. **Mostrar** o relatório agrupado por skill, FURO primeiro, com o arquivo e a
   correção de cada item. Não corrija nada ainda.
3. **Blindar** (com o aval do usuário). Rode o `scaffold.py` na DOC/ existente
   com o porte atual — ele **nunca sobrescreve**: cria só o que falta
   (templates ausentes, `openspec/`, `tools/`, garantias), mescla
   `openspec/config.yaml` e `.claude/settings.json`, liga
   `git config core.hooksPath .githooks` e acrescenta `.sdd/` ao `.gitignore`.
   Sem `openspec/` ainda, rode antes `openspec init --tools claude,agents`.
   Projeto com `System/requisitos-ears.md`: siga a migração em
   `references/estrutura-doc.md`.
4. **O que o scaffold não faz sozinho** — faça com o usuário:
   - `AGENTS.md` já existente não é sobrescrito: acrescente a ele a tabela
     "Skills do projeto" e a seção "Garantias" do template
     `assets/templates/_raiz/AGENTS.md`;
   - cenários vigentes sem teste: renomeie os testes existentes para citar o
     T-n (ou crie o teste); comportamento sem spec vira spec vigente;
   - marque o job `checagens` do workflow `sdd-checks` como obrigatório na
     proteção da branch.
5. **Mapear o código** (seção "Mapa do projeto" abaixo) — num projeto que já
   existe, é o mapa que dá à `conhecimento` o que responder e à `sentinel` o
   contexto de cada área.
6. **Re-auditar** com `python3 <DOC>/tools/checa_projeto.py .` até restarem só
   alertas conscientes. Cada furo que ficar para depois vira item no
   `to-do.md` — nunca silêncio. Registre o antes/depois no `log.md`.

## Mapa do projeto — como o código está organizado

O `INDEX.md` mapeia documentos; o mapa mapeia o **código**. Duas camadas, e a
separação é o que impede o mapa de mentir:

1. **Fatos** — `python3 <DOC>/tools/mapeia_projeto.py .` (ou
   `scripts/mapeia_projeto.py <raiz>` da pasta da skill) grava
   `System/mapa/inventario.json` e `inventario.md`: stack, módulos e quem
   importa quem (Mermaid), rotas/telas, endpoints, chamadas HTTP de saída,
   filas, crons, variáveis de ambiente lidas × declaradas (só nomes), dados e
   infra. Gerado; ninguém edita.
2. **Interpretação** — você escreve `System/mapa/visao-geral.md` e uma página
   por **área de negócio** (não por pasta técnica), seguindo o fio rota → tela
   → serviço → endpoint do back no código. Tudo o que afirmar tem de estar no
   inventário ou citar arquivo:linha; dedução é marcada "(deduzido)". Regras
   de negócio encontradas entram como **candidatas a confirmar** em
   `Negocio/especificacao-negocio.md` — nunca viram RN-n sem o dono do negócio.

Formato das páginas, specs de linha de base e o que levar para as outras
skills: `references/mapa-projeto.md`. Vista navegável para humanos:
`python3 <DOC>/tools/mapa_html.py <DOC>/System/mapa` gera `mapa.html` a partir
das páginas (publique como artifact quando o usuário quiser compartilhar). Acrescente ao `INDEX.md` a linha "como
o código está organizado → `System/mapa/visao-geral.md`". O
`checa_projeto.py` roda `mapeia_projeto.py --check` e alerta quando surge
rota, endpoint, chamada, fila ou variável fora do mapa.
