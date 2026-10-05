# Garantias — como as skills rodam sem depender de a IA lembrar


Instrução aumenta a chance de uma skill disparar; só código fora do modelo
garante. O scaffold instala três camadas, em qualquer porte:

| Camada | Arquivo | O que faz |
|---|---|---|
| Sessão (Claude Code) | `.claude/settings.json` → `tools/sdd_hooks.py` | `SessionStart` injeta as skills e o ciclo; `UserPromptSubmit` diz qual skill o pedido aciona; `PreToolUse` (Bash) pede permissão ao usuário para comando com efeito em prod ou sem alvo e nega `--no-verify`; `Stop` roda a checagem e **não deixa encerrar com FURO** (inclusive diff sem auditoria da sentinel) |
| Commit | `.githooks/pre-commit` | `checa_projeto.py --modo precommit` (rápido, sem rede): rastreabilidade, T-n nos testes, ambientes, Nomad **e a sentinel** — commit com arquivo de código não auditado é recusado |
| Merge | `.github/workflows/sdd-checks.yml` | `checa_projeto.py --modo ci` (com `openspec validate`); a sentinel não roda no CI |

**A sentinel roda só no git local, sem API.** Ela fecha o laço registrando o
veredito com `tools/sdd_hooks.py auditoria --veredito "..."`, que grava o
conteúdo (blob do git) de cada arquivo de código alterado em
`.sdd/auditoria.jsonl`. O pre-commit confere os arquivos do commit e o hook
Stop confere o trabalho em aberto: arquivo sem auditoria no conteúdo atual,
ou com "corrigir antes de concluir", barra. Commit parcial funciona, porque a
checagem é por arquivo. O que isso não cobre: quem commita com `--no-verify`
fora do Claude Code (dentro dele o hook nega) e veredito registrado sem
auditoria real — a revisão humana do PR é a última linha. Outras ferramentas
(Codex, Cursor) não têm os hooks do Claude, mas passam pelo mesmo pre-commit.

Código gerado ou de terceiros (modelos do framework, stubs gerados) fica fora
da sentinel listando os padrões em `.sentinelignore` na raiz (uma linha por
padrão; `pasta/` ignora a pasta inteira). Só entra ali o que ninguém escreve à
mão — código do projeto nunca.

Os arquivos que sustentam as garantias — `.githooks/`, `.claude/settings.json`,
`.github/workflows/sdd-checks.yml` e o próprio `.sentinelignore` — passam pela
sentinel em qualquer mudança, qualquer que seja a extensão, e o
`.sentinelignore` não os tira: mexer neles desliga ou afrouxa a checagem.
