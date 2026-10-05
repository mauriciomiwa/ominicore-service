# decisoes/ — registro de decisões (ADR)

Uma decisão por arquivo, numerada: `0001-titulo-curto.md`, `0002-…` (pegue o
maior número existente e some 1).

**Só registre quando as três forem verdade:** difícil de reverter; surpreendente
para quem ler o código sem contexto ("por que fizeram assim?"); e fruto de uma
escolha real entre alternativas. Faltou uma, não registre.

Formato mínimo — um parágrafo basta:

```md
# <título curto da decisão>

<1 a 3 frases: o contexto, o que decidimos e por quê.>
```

Opcional, só quando agrega: `Status` (proposta · aceita · substituída pela
0007), `Alternativas consideradas` (quando a rejeição não é óbvia) e
`Consequências` (efeitos que não se veem no código).
