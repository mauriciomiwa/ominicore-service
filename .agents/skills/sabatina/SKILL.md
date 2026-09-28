---
name: sabatina
description: >-
  Entrevista em rodadas até não sobrar decisão em aberto: cada pergunta com
  resposta recomendada, fatos buscados pela IA, nada feito antes do "fechado".
  Use SEMPRE antes de abrir change do OpenSpec, no kickoff, antes de mudança
  grande, ou quando pedirem "me sabatine" ou "vamos alinhar".
---

# Sabatina

O erro mais caro é construir a coisa errada bem feita. A sabatina fecha a
distância entre o que o usuário quer e o que a IA entendeu **antes** de
escrever proposta, requisito ou código.

Adaptada de `grilling`/`grill-with-docs` (Matt Pocock, `mattpocock/skills`,
MIT) para o fluxo SDD da casa: DOC/, OpenSpec, glossário e decisões.

## Antes da primeira rodada

1. Leia o que já existe, para não perguntar o que está escrito: `DOC/INDEX.md`
   (skill `conhecimento`), o glossário (`DOC/…/Negocio/glossario.md`), as
   decisões (`DOC/…/System/decisoes/`), as specs vigentes em
   `openspec/specs/` e, se houver, o mapa (`DOC/…/System/mapa/`).
2. Monte a **árvore de decisões**: cada decisão abre as que dependem dela
   (ex.: "quem pode cancelar" só faz sentido depois de "existe cancelamento
   parcial?").

## Rodadas

A **fronteira** são as decisões cujos pré-requisitos já foram respondidos — as
perguntas que dá para fazer agora sem chutar respostas que você ainda não
ouviu. Pergunte a fronteira inteira numa rodada, numerada, cada uma com a sua
recomendação, e espere as respostas:

```
**P1 — <título curto>**: <pergunta; opções quando houver>
→ Recomendo: <resposta recomendada e o porquê em uma linha>

---

**P2 — <título curto>**: …
→ Recomendo: …
```

- Pergunta que depende de outra ainda aberta **nesta** rodada vai para a
  próxima.
- Mais de 5 perguntas na fronteira? Faça as 5 que destravam mais e diga
  quantas faltam — rodada longa demais vira resposta rasa.
- "Pode seguir a recomendação" vale como resposta para todas as que o usuário
  não contestar.
- A cada rodada, recalcule a fronteira: decisões fechadas destravam as que
  dependiam delas.

## Fatos são da IA; decisões são do usuário

Nunca pergunte o que dá para descobrir: stack, versão instalada, como o código
faz hoje, se já existe regra parecida, onde fica algo. Busque você (lendo o
repositório, o mapa ou mandando um subagente explorar) e traga o fato **com a
fonte**. Enquanto a busca roda, pergunte o resto da fronteira; só esperam as
perguntas que dependem daquele fato.

Quando o usuário afirmar como algo funciona e o código disser outra coisa,
mostre a contradição com o arquivo: "você disse que o cancelamento é parcial,
mas `pedidos/cancelar.go:40` cancela o pedido inteiro — qual vale?".

## Linguagem e decisões, durante a conversa

- **Termos:** use o vocabulário do glossário. Se o usuário usar um termo que
  conflita com ele, ou um termo vago ("conta" = cliente ou usuário?), pare e
  proponha o termo canônico. Termo resolvido entra **na hora** no glossário,
  sem detalhe de implementação, neste formato (o glossário completo segue a
  skill `novo-projeto-sdd`):

  ```md
  **Pedido**:
  Solicitação de compra feita por um Cliente, com um ou mais itens.
  _Evitar_: compra, ordem, transação
  ```
- **Cenários concretos:** para fechar fronteiras entre conceitos, invente um
  caso de borda e pergunte o que acontece ("pedido pago, meio enviado, cliente
  cancela: estorna tudo?"). Caso de borda decidido costuma virar cenário T-n.
- **Decisões:** ofereça registrar decisão (ADR) só quando as três forem
  verdade — difícil de reverter, surpreendente para quem ler o código sem
  contexto, e fruto de escolha real entre alternativas. Senão, não registre.
  Formato: `System/decisoes/NNNN-titulo.md` com título e 1 a 3 frases
  (contexto, o que se decidiu e por quê).

## Fim da sabatina

A sabatina acaba quando a fronteira fica vazia: nenhum ramo sem visitar, nada
assumido em silêncio. Então:

1. Mostre o **resumo do entendimento** em poucas linhas: o que muda, regras
   (RN) novas ou alteradas, casos de borda decididos, termos novos no
   glossário, decisões registradas e o que ficou **fora**.
2. **Não aja antes do "fechado" do usuário.** Com ele, siga o fluxo: RN no
   `Negocio/especificacao-negocio.md` → change do OpenSpec (`/opsx:propose`)
   com REQ-n (`Origem: RN-n`) e cenários T-n tirados dos casos de borda.
3. O que ficou em aberto de propósito vai para o `to-do.md` como decisão em
   aberto, não some.

## Proporcionalidade

Mudança pequena e óbvia (texto, bug com causa clara) não precisa de sabatina:
uma pergunta de confirmação basta. A sabatina é para quando há escolha de
comportamento, regra de negócio, dado ou arquitetura em jogo.
