---
name: sentinel
description: >-
  Audita o código EM FOCO (arquivo, função, diff) contra goldrules, Padroes/,
  Clean Code/SOLID, testes, resiliência e segurança, com rastreabilidade
  RN→REQ→teste. Use SEMPRE ao escrever, revisar, refatorar ou concluir código,
  implementar fatia ou avaliar PR — só o trecho em foco, não o projeto todo.
---

# Sentinel — o código em foco

Audita o que o desenvolvedor está mexendo agora, não o projeto inteiro.
Agnóstica de linguagem: o *como* vem das convenções do projeto.

**Fonte é a `DOC/`** (comece pelo `INDEX.md`; `DOC/` ou `DOC/NOME/`): cite a
regra e o arquivo exato em cada achado (`goldrules.md`, `Padroes/…`,
`openspec/specs/…`, `openspec/changes/<change>/…`). Sem `DOC/`/goldrules:
audite pelos princípios gerais e avise que não há padrão documentado.

## Antes de ler: o que o script já prova

Unidade de auditoria = o **diff da tarefa** (`git diff`), mais o contexto que
ele exigir (a função inteira, o chamador, o teste). Não releia o que não mudou.

Com as garantias SDD, rode **uma vez** `python3 <DOC>/tools/checa_projeto.py .`
e use a saída como achado, citando a linha: ela já prova T-n sem teste,
`openspec validate`, Dependabot (9), chave sem par hml/prod (11), mapa
desatualizado (12) e imagem/segredo em job Nomad. Não refaça isso lendo
arquivos. O script não julga design, nomes, resiliência nem anti-padrões
(2, 3, 6, 9, 10): esses continuam na leitura do diff, sem atalho.

## Formato

```
[BLOQUEANTE|ALERTA|SUGESTÃO] arquivo:linha — problema
Regra: <regra violada + fonte>
Correção: <o que fazer, curto>
```

BLOQUEANTE = requisito, segurança, perda de dado, contrato quebrado. ALERTA =
padrão do projeto ou anti-padrão com risco real. SUGESTÃO = legibilidade; não
trava. Sem emojis. Termine com o veredito: *aprovado*, *aprovado com
ressalvas* ou *corrigir antes de concluir* — e por quê.

**Registre o veredito** (projetos com as garantias SDD):
`python3 <DOC>/tools/sdd_hooks.py auditoria --veredito "<veredito>"`. Amarra o
veredito ao conteúdo atual de cada arquivo alterado; pre-commit e hook Stop
cobram. Mudou depois? Audite de novo. Registrar sem auditar anula a garantia.
Código gerado ou de terceiros vai no `.sentinelignore` — nunca código do
projeto.

## Checklist

1. **Rastreável:** REQ em `openspec/specs/` ou no delta de uma change ativa, e
   o teste do trecho cita o T-n? Comportamento novo sem delta = spec
   desatualizada (BLOQUEANTE; refactor sem mudança de comportamento não precisa
   de delta). Mexeu em spec, backlog ou teste: rode
   `<DOC>/tools/valida_rastreabilidade.py`.
2. Convenções do projeto (`Padroes/`, stack), sem reinventar.
3. Clean Code/SOLID/KISS, limites de contexto. **Nomes do glossário**
   (`Negocio/glossario.md`), nunca os de "Evitar" — ALERTA; BLOQUEANTE se mudar
   o sentido. Conceito novo sem termo: sugira.
4. Sem texto hardcoded (i18n).
5. Teste no nível certo; tarefa marcada no `tasks.md`; `plano-de-testes.md`
   atualizado (quando existir); dependências injetadas (relógio, IO, random).
6. Resiliência e segurança proporcionais: timeout, retry/circuit breaker em
   rede/IO, entrada validada, nenhum segredo no código.
7. Build e testes do trecho rodados localmente; o que não der, aponte como
   pendência.
8. Lote pequeno, com revisão humana.
9. Sem anti-padrões (transação longa, N+1, `catch` vazio, código morto, God
   class, DRY, Demeter, efeito colateral surpresa, recurso aberto, sem
   paginação, log de PII/segredo). Dependência nova justificada, sem CVE e
   **coberta pelo `.github/dependabot.yml`** (senão BLOQUEANTE; confira com
   `<DOC>/tools/dependabot_config.py --check`). Detalhe:
   `references/anti-padroes.md`.
10. Idempotente onde há retry; contrato não quebrado sem versionar; migration
    aditiva ou com plano de volta escrito e testado em hml, e compatível com a
    versão anterior do código (mesmo critério do checklist de promoção da
    `ambientes`); log estruturado nos pontos-chave.
11. **Fronteira de ambiente:** host, URL, namespace ou path de Vault fora do
    código; nada de `if (env === 'prod')` espalhado; nenhum default apontando
    para prod; fixture sem dado real. Chave nova sem par no outro ambiente:
    BLOQUEANTE pela regra de configuração da `ambientes` (Passo 2) — o
    `checa_projeto` já a acusa. *Executar* deploy ou migration é da `ambientes`.
12. **Mapa:** criou ou removeu rota, tela, endpoint, chamada externa, fila, cron
    ou variável? Rode `<DOC>/tools/mapeia_projeto.py .` e atualize a página da
    área (sem `System/mapa/`, não se aplica).

Critério fino de qualquer item, mudança de comportamento sem change e PR do
Dependabot: `references/diretrizes.md`.

## Proporcionalidade

Regra-mãe: o princípio certo, no grau certo, ligado a um requisito e a um
teste. Sempre: 2, 3, 9 e segredos/entrada do 6. Script interno ou protótipo
declarado: pode relaxar 1, 4, 5 e 10 (diga). Rede/IO/banco: 6, 9 e 10 pesam.
Contrato público ou migration: 10 é bloqueante. Config, deploy ou infra: 11 é
bloqueante. Não reporte SUGESTÃO em cascata num trecho trivial.
