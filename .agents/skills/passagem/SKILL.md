---
name: passagem
description: >-
  Passagem de bastão: reescreve DOC/estado.md (até 40 linhas) com change em
  curso, próximo passo, decisões, bloqueios e arquivos que importam, para o
  próximo chat começar dali. Use SEMPRE ao encerrar a sessão ("vamos parar",
  "continuo amanhã"), com o contexto enchendo ou quando o hook pedir.
---

# Passagem de bastão

O chat novo não lembra de nada. Sem um retrato do momento, ele relê o
`log.md` inteiro ou explora o código de novo: gasta tokens e ainda pode
retomar do ponto errado. O `estado.md` é esse retrato, e é o **único** arquivo
de contexto que o início da sessão entrega à IA.

Inspirada na `handoff` (Matt Pocock, `mattpocock/skills`, MIT), adaptada ao
fluxo SDD da casa.

## Onde

`estado.md` fica ao lado do `INDEX.md`: `DOC/estado.md` (app único) ou
`DOC/<projeto>/estado.md` (monorepo, um por projeto). Não existe? Crie.

## Formato (no máximo 40 linhas)

```md
# Estado — <projeto>

> Atualizado em <AAAA-MM-DD HH:MM> por <quem estava na sessão>

## Em curso
- Change: `<nome>` — tarefa <n.m> de `tasks.md` (<o que é>)
- Próximo passo: <uma ação concreta, com arquivo ou comando>

## Decisões recentes
- <decisão curta> (detalhe: `log.md` <data> ou `System/decisoes/NNNN-…`)

## Bloqueios e pendências
- <o que impede seguir e de quem depende>

## Arquivos que importam agora
- `<caminho>` — <por quê>
```

Regras:

- **Retrato, não histórico.** Reescreva o arquivo inteiro a cada passagem. O
  que já não vale sai; a história mora no `log.md` e nas changes arquivadas.
- **Aponte, não copie.** Decisão, RN, requisito e erro já estão em algum
  arquivo: cite o caminho. Nada de colar trecho de spec, diff ou saída de
  comando.
- **Próximo passo acionável.** "Rodar `go test ./internal/sdd`" serve;
  "continuar o trabalho" não.
- **Sem segredo e sem dado pessoal**, como em todo o repositório.
- Seções sem conteúdo ficam com "—". Passou de 40 linhas? Corte: o excesso é
  histórico.

## Ao fazer a passagem

1. Confira o que mudou na sessão: `git status`, a change ativa
   (`openspec list`) e o `tasks.md` dela.
2. Marque no `tasks.md` o que foi concluído; registre no `log.md` a decisão
   que tiver sido tomada (entrada curta, mais recente no topo).
3. Reescreva o `estado.md` no formato acima.
4. Diga ao usuário, em uma linha, qual é o próximo passo registrado.

## Ao começar uma sessão

O hook de início já entrega o `estado.md`. Sem hook (outra IA, outra
ferramenta), leia-o antes de qualquer outro arquivo e confirme com o usuário
se o próximo passo ainda vale. Depois, leia só o que a tarefa pedir, pelo
`INDEX.md`.
