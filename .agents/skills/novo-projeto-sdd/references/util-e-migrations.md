# util.md e migrations.md — as colas fixas do projeto

Os dois ficam ao lado do `INDEX.md` (`DOC/` ou `DOC/<projeto>/`) e aparecem fixos
no topo do leitor do Sentinel, junto com a cola de Git. O `verifica-projeto.py`
aponta como ALERTA quando faltam: o `util.md` sempre, o `migrations.md` só quando
o projeto tem migrations (pasta com "migrat" no caminho e arquivo de código ou
`.sql`, fora da DOC e das dependências).

## Antes de escrever

1. **Levante os comandos de verdade**, não invente: README, `Taskfile.yml`,
   `Makefile`, `package.json` (scripts), `go.mod`/`pyproject.toml`, `.github/workflows/`,
   `tools/` e `scripts/`. Para migrations, também a ferramenta (golang-migrate,
   Prisma, Supabase CLI, Alembic, Django, Rails, Flyway, Liquibase) e a pasta.
2. **Confira o que der** rodando o comando de ajuda ou de listagem (`--help`,
   `task --list`, `npm run`). Comando que não deu para confirmar leva
   "(não conferido)".
3. **Nunca** coloque segredo, token, senha ou URL com credencial: use a variável
   (`$DATABASE_URL`) e diga de onde ela vem.

## util.md

```md
# Util — comandos mais usados do <projeto>

> Fixo no Sentinel, ao lado do Git. Um bloco por tarefa; rode da raiz do repositório.

## Rodar em desenvolvimento
## Testes
## Compilar / instalar
## Checagens (lint, SDD)
## Versão e release
## Deploy            ← só o comando e: "hml/prod passam pela skill ambientes"
```

- Um bloco `sh` por tarefa, com comentário curto no fim da linha quando ajudar.
- Seção sem comando no projeto sai (não deixe cabeçalho vazio).
- Tem migrations? Uma linha no fim apontando para `migrations.md`; não
  duplique os comandos aqui.
- Até ~80 linhas: é cola, não manual. Explicação longa vai para a DOC.

## migrations.md

```md
# Migrations — <projeto>

> Fixo no Sentinel. Dev pode rodar daqui; hml e prod só pela skill ambientes.

- Ferramenta: <nome e versão do manifesto>
- Pasta: `<caminho>` · nome do arquivo: `<padrão, ex.: 20260930_descricao.sql>`
- Banco de dev: `<como subir>` · variável: `$DATABASE_URL` (vem de <arquivo/Vault>)
- Bancos de hml e prod (host, instância, credencial no cofre, quem pode escrever):
  no `Infra/infraestrutura.md`, não aqui.

## Criar uma migration
## Aplicar em dev
## Ver o status (quais já rodaram)
## Desfazer a última em dev
## hml e prod
Não rode daqui. Use a skill `ambientes`: prova o alvo antes de executar e segue o
checklist de promoção (migration aditiva, compatível com a versão anterior).
```

- Não repita aqui o que é do ambiente (hosts, instâncias, cofre): o
  `Infra/infraestrutura.md` aponta para este arquivo e vice-versa.
- Diga se a migration é **reversível** na ferramenta (tem `down`?) e qual o
  cuidado de dados (coluna nova com default, backfill em lote).
- Comando que aponta para um ambiente leva o ambiente escrito no próprio bloco
  (`# dev`), nunca um default que caia em prod.

## Depois de escrever

Rode `python3 <DOC>/tools/checa_projeto.py .` (ou Verificar no Sentinel): o
ALERTA some e o fixo deixa de aparecer como pendente.
