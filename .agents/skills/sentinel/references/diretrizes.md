# Sentinel — Diretrizes detalhadas

Detalhamento dos itens do checklist do `SKILL.md`. Consulte quando um item
precisar de critério mais fino.

## Antes de escrever (entrada)

- **Requisito claro:** saiba qual REQ o código atende — em
  `openspec/specs/` (vigente) ou no delta da change ativa
  (`openspec/changes/<change>/specs/`). Se não houver, a mudança começa por uma
  change (`/opsx:propose`), não pelo código. Ligue código ↔ T-n ↔ REQ ↔ RN
  (rastreabilidade). Projeto sem OpenSpec: o REQ fica onde o projeto o
  documenta (legado: `System/requisitos-ears.md`).
- **Mudança de comportamento sem change:** código que altera comportamento
  observável sem REQ novo/MODIFIED numa change ativa é spec desatualizada no
  dia do merge — ALERTA (BLOQUEANTE se mexe em contrato público). Refactor sem
  mudança de comportamento não precisa de delta (`skip_specs: true`).
- **Leia as convenções do projeto:** `goldrules.md`, `Padroes/` e
  `stack-tecnologico.md`. Não reinvente o que já está padronizado.
- **Conhecimento externo antes da web:** consulte `fontes-externas.md`; use a
  busca mais econômica (manifesto p/ versões, deep-link oficial p/ API).
- **Proporcionalidade:** decida o nível certo de esforço. Não aplique IaC,
  microsserviços, tracing distribuído ou Zero Trust num projeto pequeno só por
  serem "boas práticas" — canhão em mosquito também é defeito.

## Enquanto escreve

- **Clean Code + SOLID:** responsabilidade única, nomes claros, funções curtas,
  código para humanos lerem.
- **Limites de contexto (DDD):** respeite os módulos/domínios; não vaze regra de
  um contexto para outro.
- **API-First:** se há contrato de API, ele vem antes da implementação dos lados.
- **Sem texto hardcoded:** textos de UI/erros via string-files (i18n).

## Testes (shift-left, TDD, pirâmide)

- Escreva o teste **antes** quando fizer sentido (TDD). Cada cenário
  `#### Scenario: T-n` do REQ vira um teste, e o nome do teste cita o T-n.
- **Pirâmide:** muitos unitários rápidos, alguns de integração, poucos E2E
  críticos. Teste **comportamento**, não cobertura vazia.
- Atualize o `plano-de-testes.md` (status do T-n) e marque a tarefa no
  `tasks.md` da change.
- **Testabilidade:** injete relógio, IO e aleatoriedade (em vez de `now()`/rede
  direto) e evite estado global escondido — é o que torna o teste possível.

## Resiliência e segurança (proporcionais)

- **Design for failure** onde o contexto pede (rede/IO/dependências externas):
  timeouts, retries, circuit breaker.
- **Zero trust** conforme a exposição do sistema. Nunca commitar segredos.
- **Adicionar dependência:** ao incluir uma lib nova, justifique a necessidade,
  prefira lib mantida/popular e não fixe versão com CVE conhecido. O *scan
  contínuo* é do Dependabot, não da sentinela — política em
  `DOC/Infra/seguranca-dependencias.md`. O que a sentinela confere é a
  **cobertura**: manifesto novo (outro diretório ou ecossistema) precisa estar
  no `.github/dependabot.yml` — rode `DOC/tools/dependabot_config.py <raiz>
  --check` quando existir; furo = BLOQUEANTE (dependência que o Dependabot não
  vê não recebe alerta de CVE).
- **PR do Dependabot** (autor `dependabot[bot]`): patch/minor elegíveis entram
  sozinhos pelo workflow de auto-merge — revisar é só para o que ele recusou.
  **Major:** leia changelog/release notes atrás de *breaking changes*, rode os
  testes, e ajuste o código **no mesmo PR** (não mergeie verde-por-acaso com
  teste que não cobre o uso). CI vermelho nunca entra. `ignore:` novo no
  `dependabot.yml` ou alerta dispensado sem linha no `log.md` (o quê, por quê,
  quem, até quando) = BLOQUEANTE.
- **Não afrouxe o fail-closed:** mudança no `dependabot-automerge.yml` que
  remova a exigência de checks obrigatórios, amplie para major ou troque o
  filtro de autor é BLOQUEANTE sem decisão registrada.

## Correção, contrato e observabilidade

- **Idempotência de escrita:** operação que pode ser repetida (retry/reenvio do
  outbox) é segura para repetir — chave de idempotência ou upsert por ID.
- **Migrations seguras:** aditivas e idempotentes, sem perda de dado.
- **Compatibilidade de contrato:** não quebre API/retorno público sem versionar
  ou depreciar.
- **Observabilidade:** log estruturado com contexto nos pontos-chave (proporcional).

## Entrega

- **Small batch / trunk-based:** lotes pequenos, integração frequente.
- **CI verde** antes de concluir (build + testes). Deploy manual é exceção.
- **IA propõe, o dev valida:** a IA pode gerar teste/código/doc, mas a revisão e
  a responsabilidade são humanas.
