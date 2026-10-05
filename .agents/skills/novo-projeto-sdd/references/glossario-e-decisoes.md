# Glossário e decisões

Duas memórias curtas que poupam conversa em toda sessão: **qual palavra usar**
(glossário) e **por que algo caro foi feito assim** (decisões). Adaptado de
`domain-modeling` (Matt Pocock, `mattpocock/skills`, MIT).

## Glossário — `Negocio/glossario.md`

A linguagem do domínio compartilhada por negócio, IA e código. Com ela, a
conversa encurta ("a cascata de materialização" em vez de um parágrafo), nomes
de variáveis e arquivos ficam consistentes e a IA gasta menos tokens pensando.

### Formato

```md
**Pedido**:
Solicitação de compra feita por um Cliente, com um ou mais itens.
_Evitar_: compra, ordem, transação
```

- **Opine.** Havendo várias palavras para o mesmo conceito, escolha a melhor e
  liste as outras em `_Evitar_`.
- **Definição curta.** Uma ou duas frases: o que o conceito **é**, não o que o
  sistema faz com ele.
- **Só termos deste domínio.** Conceito genérico de programação (timeout,
  cache, fila) não entra, mesmo que o projeto use muito.
- **Sem implementação.** Nada de tabela, endpoint ou classe: glossário não é
  spec nem rascunho. A regra (RN) mora na especificação; o comportamento, no
  OpenSpec.
- **Grupos** (`## Vendas`, `## Cobrança`) quando surgirem naturalmente.
- **Monorepo** (`DOC/<projeto>/`): um glossário por projeto; termo com sentido
  diferente em dois projetos fica explícito nos dois.

### Durante a conversa (qualquer skill)

- Termo que conflita com o glossário → aponte na hora: "o glossário define
  'cancelamento' como X, mas você parece querer Y — qual vale?".
- Termo vago ou sobrecarregado → proponha o canônico: "'conta' é o Cliente ou o
  Usuário? São coisas diferentes."
- Termo decidido → escreva no glossário **naquele momento**, não no fim.
- Código que usa outra palavra para um termo do glossário → sinal para a
  `sentinel` (nome fora da linguagem do domínio).

## Decisões — `System/decisoes/NNNN-titulo.md`

### Quando registrar

As três precisam ser verdade:

1. **Difícil de reverter** — mudar de ideia depois custa caro.
2. **Surpreendente sem contexto** — quem ler o código vai se perguntar "por que
   fizeram assim?".
3. **Escolha real** — havia alternativas de verdade e uma foi escolhida por
   motivos específicos.

Faltou uma, não registre: fácil de reverter você reverte; óbvio ninguém
pergunta; sem alternativa não há o que explicar.

Costumam passar: forma da arquitetura (monorepo, event sourcing), como dois
contextos conversam (eventos × chamada síncrona), tecnologia que prende (banco,
fila, provedor de identidade, alvo de deploy), limites de escopo e os "nãos"
explícitos, desvios deliberados do caminho óbvio ("SQL à mão em vez de ORM
porque…"), restrições que o código não mostra (compliance, contrato de
latência) e alternativas rejeitadas por motivo não óbvio.

### Formato

```md
# <título curto da decisão>

<1 a 3 frases: o contexto, o que decidimos e por quê.>
```

Um parágrafo basta; o valor está em registrar **que** se decidiu e **por quê**.
Seções opcionais só quando agregam: `Status` (proposta · aceita · substituída
pela NNNN), `Alternativas consideradas`, `Consequências`. Numeração: maior
número existente + 1. A pasta nasce com a primeira decisão.

### Relação com o resto

- Decisão que muda comportamento vira também change no OpenSpec; a decisão
  explica o porquê, a change descreve o quê.
- `design.md` de uma change guarda o raciocínio técnico daquela mudança; se a
  decisão valer além dela (passa no critério), promova para `System/decisoes/`.
- Registro cronológico do que aconteceu continua no `log.md`.
