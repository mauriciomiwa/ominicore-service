# Instruções para agentes de IA

A fonte de verdade deste projeto é `DOC/` (porquê, padrões, operação)
e `openspec/` (comportamento). Antes de responder, planejar ou escrever código:

1. Consulte `DOC/INDEX.md` e responda a partir da documentação.
   Web só para fatos do mundo atual (use `DOC/Padroes/fontes-externas.md`
   antes de busca genérica; versão instalada = manifesto, não web).
2. Disciplina SDD com OpenSpec: nada entra no código sem origem
   RN → REQ → teste. O porquê (RN) está em `DOC/Negocio/`; o
   comportamento (REQ + cenários T-n) em `openspec/specs/`. Toda mudança de
   comportamento nasce como change (`/opsx:propose`), é implementada com
   `/opsx:apply` e fecha com `/opsx:archive`. Formato e regras de ID:
   `openspec/config.yaml`. O nome de cada teste cita o T-n do cenário.
3. Siga `DOC/Padroes/goldrules.md` e os demais `Padroes/` (stack,
   convenções, i18n — nenhum texto hardcoded). Não reinvente o padronizado.
4. Registre decisões no `DOC/log.md`; o que não estiver documentado,
   ofereça documentar em vez de assumir.

## Skills do projeto — leia o SKILL.md antes de agir quando o gatilho bater

| Skill | Gatilho | Onde |
|---|---|---|
| conhecimento | qualquer pergunta sobre este projeto | `.agents/skills/conhecimento/` |
| passagem | ao encerrar a sessão ("vamos parar", "continuo amanhã") ou com o contexto enchendo | `.agents/skills/passagem/` |
| sabatina | antes de abrir change, mudança grande ou de arquitetura; "me sabatine" | `.agents/skills/sabatina/` |
| sentinel | escrever, revisar, refatorar ou concluir código; avaliar PR | `.agents/skills/sentinel/` |
| ambientes | deploy, migration, script com efeito, hml/prod, variável ou segredo | `.agents/skills/ambientes/` |
| nomad-worker | job Nomad, `.hcl`, Vault, Consul (quando o projeto tiver) | `.agents/skills/nomad-worker/` |
| OpenSpec | nova funcionalidade ou mudança de comportamento | `/opsx:*` (`openspec-*` em `.agents/skills/`) |

## Garantias — o que barra, independentemente da IA lembrar

- **Antes de commitar código:** audite o diff com a `sentinel` e registre o
  veredito: `python3 DOC/tools/sdd_hooks.py auditoria --veredito "aprovado"`
  (ou "aprovado com ressalvas" / "corrigir antes de concluir"). O pre-commit
  recusa arquivo de código sem auditoria no conteúdo atual.
- **Checagem completa:** `python3 DOC/tools/checa_projeto.py .` — o
  mesmo conjunto roda no pre-commit (`.githooks/`), no CI (`sdd-checks`, sem a
  sentinel) e, no Claude Code, no hook Stop, que recusa encerrar com FURO.
- Comando com efeito em prod (ou sem alvo identificado) pede permissão ao
  usuário. `git commit --no-verify` é negado: corrija o furo em vez de pular.
