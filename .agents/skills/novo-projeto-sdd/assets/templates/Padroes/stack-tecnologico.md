# Stack tecnológico — {{PROJETO}}

Fonte de verdade das **versões** = manifesto/lockfile do repositório, não este
doc. Cite ao lado de cada item o arquivo que o originou (inventário do
kickoff) — assim ninguém precisa adivinhar se foi decisão ou chute.

## Stack

| Camada | Tecnologia | Origem |
|---|---|---|
| <backend/cliente/banco> | <a preencher> | <pom.xml, package.json, confirmado pelo usuário…> |

## Verificação automática

O que o lint pega de graça a revisão não gasta atenção. Configure conforme a
stack: função máxima + complexidade ciclomática (Clean Code), detector de
duplicação (DRY), regra de encadeamento (Lei de Demeter), formatador.

| Verificação | Ferramenta (desta stack) | Onde roda |
|---|---|---|
| <lint/duplicação/Demeter> | <a preencher> | <pre-commit, CI> |

## Decisões em aberto
- <o que foi detectado mas não confirmado; versões a fixar>
