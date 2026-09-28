---
name: conhecimento
description: >-
  Base de conhecimento do projeto: responde pela DOC/ (INDEX.md, estado.md) e
  pelo openspec/ antes de memória ou web, lendo só o necessário. Use SEMPRE em
  pergunta sobre ESTE projeto (arquitetura, negócio, requisitos, padrões, onde
  fica X) ou ao planejar e implementar.
---

# Base de conhecimento do projeto

A fonte de verdade é o repositório: a pasta `DOC/` (porquê, padrões,
operação) e, quando existir, o `openspec/` (comportamento). Esta skill garante consultar o conhecimento
interno **antes** da web ou da memória — e é a **redação canônica** dessa regra:
onde outros documentos (AGENTS.md, CLAUDE.md, skills) a resumirem, esta
prevalece em caso de divergência.

## Passo 0 — Localize a `DOC/`

Convenção da casa (o scaffold da `novo-projeto-sdd` a segue): app único →
`DOC/` na raiz do repositório; monorepo/múltiplos componentes →
`DOC/<projeto>/`. Portanto:

1. Existe `DOC/INDEX.md`? Use-o.
2. Senão, use `DOC/*/INDEX.md` (mais de um = monorepo; escolha pelo componente
   da pergunta).
3. Nenhum dos dois → o projeto não tem base de conhecimento estruturada:
   avise, sugira criá-la com a skill `novo-projeto-sdd` e responda pelo
   código-fonte, deixando claro que não há doc de respaldo.

Comportamento do sistema (requisitos REQ-n e cenários de teste T-n) não mora
na `DOC/`: mora no OpenSpec, em `openspec/` na raiz do repositório — o
`INDEX.md` aponta para lá.

| A pergunta é… | Fonte |
|---|---|
| por que existe / qual regra de negócio (RN) | `DOC/…/Negocio/` |
| o que significa um termo do domínio | `DOC/…/Negocio/glossario.md` |
| por que uma escolha de arquitetura foi feita | `DOC/…/System/decisoes/` (e `design.md` das changes arquivadas) |
| onde fica X no código, qual tela chama qual endpoint, que integrações existem | `DOC/…/System/mapa/visao-geral.md` e a página da área; fatos em `System/mapa/inventario.md` |
| o que o sistema faz hoje (REQ, cenários) | `openspec/specs/<capacidade>/spec.md` (ou `openspec show <capacidade> --type spec`) |
| o que está mudando agora | `openspec/changes/<change>/` (ou `openspec list`) |
| por que uma mudança entregue foi feita | `openspec/changes/archive/<data>-<change>/proposal.md` |
| arquitetura, padrões, infra, ambientes, decisões | `DOC/` via `INDEX.md` |

Spec vigente e código divergindo é informação: diga qual dos dois está
desatualizado (e ofereça abrir a change) em vez de escolher um em silêncio.
Projeto sem `openspec/`: os requisitos estão onde o `INDEX.md` apontar
(legado: `System/requisitos-ears.md`).

## Leia só o que a tarefa pede

Cada arquivo lido custa contexto na conversa inteira. Por isso:

- **Comece pelo `estado.md`** (ao lado do `INDEX.md`): onde o trabalho parou e
  o próximo passo. O hook de início já o entrega; sem hook, leia-o primeiro.
- **`INDEX.md` é roteador, não leitura:** use-o para achar o único doc da
  pergunta e vá direto nele.
- **`log.md` nunca inteiro:** a entrada do topo, ou busque o termo
  (`grep -n "termo" DOC/log.md`) e leia só aquela entrada. Entradas antigas
  estão em `log/arquivo-AAAA.md` — só quando a pergunta for histórica.
- **Mapa:** a página da área em questão; o `inventario.md` só para um fato
  pontual (busque, não leia tudo).
- **Specs:** só a capacidade envolvida (`openspec show <capacidade>`), não
  todas.
- **Código:** o mapa e o glossário dizem onde olhar; abra o arquivo, não a
  pasta inteira.

## Regra

1. **Pergunta sobre o projeto** → abra o `INDEX.md`, ache o doc certo e responda
   a partir dele, citando o arquivo usado.
2. **Pergunta sobre OUTRO software/projeto** → use a **memória compartilhada**:
   chame `kb_search(q, match_count)` (MCP do Supabase, PostgREST ou
   `tools/memoria/query.py`). Responda a partir dos trechos e cite a `source_url`.
   Ver `System/memoria-compartilhada.md`. **Se nenhum desses canais estiver
   disponível na sessão** (MCP desconectado, script ausente), diga isso ao
   usuário e siga para o passo 3 — não invente resultado de memória.
3. **Não achou em lugar nenhum?** Diga que não está documentado e ofereça
   registrar (no `DOC/` do projeto; a ingestão atualiza o índice) antes da web.
4. **Fato do mundo atual** → use as fontes oficiais de `Padroes/fontes-externas.md`
   antes de uma busca genérica. Versão instalada = manifesto, não web.
5. **Fale a língua do projeto.** Havendo `Negocio/glossario.md`, responda com
   os termos dele (nunca os de "Evitar") — é o que deixa a resposta curta e
   sem ambiguidade. Se a pergunta usar um termo que conflita com o glossário,
   aponte antes de responder ("no glossário, 'conta' é o Usuário; você quis
   dizer Cliente?").

## Manutenção

Ao criar ou mover um doc relevante, atualize `DOC/INDEX.md` para a regra
"repositório primeiro" continuar eficaz. Requisito técnico novo não vira doc
na `DOC/`: vira change no OpenSpec. Termo do domínio decidido numa conversa
entra no glossário na hora; decisão difícil de reverter, surpreendente e com
alternativa real vira `System/decisoes/NNNN-titulo.md` (formatos e regras na
skill `novo-projeto-sdd`, seção de glossário e decisões).
