# Instruções para agentes de IA

Fonte de verdade: `{{DOC_REL}}/` (porquê, padrões, operação) e `openspec/`
(comportamento). Antes de responder, planejar ou escrever código:

1. Responda pela documentação, a partir de `{{DOC_REL}}/INDEX.md`. Web só para
   fato do mundo atual (antes, `{{DOC_REL}}/Padroes/fontes-externas.md`, quando
   existir; versão instalada = manifesto, não web).
2. SDD: nada entra no código sem origem RN → REQ → teste. RN em
   `{{DOC_REL}}/Negocio/`; REQ e cenários T-n em `openspec/specs/` (formato em
   `openspec/config.yaml`); o nome de cada teste cita o T-n.
3. Ciclo de mudança de comportamento: `sabatina` → `/opsx:propose` →
   `/opsx:apply` (a `sentinel` audita o diff de cada tarefa) → `openspec validate`
   + `valida_rastreabilidade` → deploy pela `ambientes` → `/opsx:archive` → `passagem`.
4. Siga `{{DOC_REL}}/Padroes/` (goldrules, stack, i18n: nenhum texto
   hardcoded). Decisão vai para `{{DOC_REL}}/log.md`; o que não estiver
   documentado, ofereça documentar em vez de assumir.

## Skills do projeto — leia o SKILL.md quando o gatilho bater

| Skill | Gatilho | Onde |
|---|---|---|
| conhecimento | pergunta sobre este projeto | `.agents/skills/conhecimento/` |
| passagem | fim de sessão ("vamos parar") ou contexto enchendo | `.agents/skills/passagem/` |
| sabatina | antes de change ou mudança grande; "me sabatine" | `.agents/skills/sabatina/` |
| sentinel | escrever, revisar ou concluir código; avaliar PR | `.agents/skills/sentinel/` |
| ambientes | deploy, migration, hml/prod, variável ou segredo | `.agents/skills/ambientes/` |
| novo-projeto-sdd | kickoff, DOC/, OpenSpec, auditar garantias, mapear o código | `.agents/skills/novo-projeto-sdd/` |
| nomad-worker | job Nomad, `.hcl`, Vault, Consul (projeto da DKM) | `.agents/skills/nomad-worker/` |
| OpenSpec | mudança de comportamento | `/opsx:*` (`openspec-*` em `.agents/skills/`) |

## Garantias — o que barra, independentemente da IA lembrar

- Antes de commitar código: `sentinel` no diff e o veredito registrado com
  `python3 {{DOC_REL}}/tools/sdd_hooks.py auditoria --veredito "aprovado"`; o
  pre-commit recusa código sem auditoria.
- `python3 {{DOC_REL}}/tools/checa_projeto.py .` roda no pre-commit, no CI e no
  hook Stop, que não deixa encerrar com FURO.
- Comando com efeito em prod (ou sem alvo) pede permissão; `--no-verify` é negado.
