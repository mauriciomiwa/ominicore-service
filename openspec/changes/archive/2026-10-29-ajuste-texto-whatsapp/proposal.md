# Proposal

## Why

O archive das changes 08, 09 e 10 (2026-10-29) deixou três requisitos acima de
500 caracteres, e o `openspec validate --all --strict` (CI, pre-commit e hook
Stop) passou a reprovar as specs `segredos` (REQ-6) e `notificacoes` (REQ-10 e
REQ-11). A mudança é só de redação, como a `ajuste-texto-notificacoes`: nenhuma
regra muda (a origem continua RN-4 e RN-9).

## What Changes

- **REQ-6**, **REQ-10** e **REQ-11** reescritos de forma mais enxuta, com as
  mesmas obrigações.
- Todos os cenários copiados sem alteração: os testes que citam T-n não mudam.
  Nenhum REQ novo.

## Capabilities

### New Capabilities

### Modified Capabilities

- `segredos`: redação do REQ-6 abaixo de 500 caracteres, sem mudar comportamento.
- `notificacoes`: redação do REQ-10 e do REQ-11 abaixo de 500 caracteres, sem
  mudar comportamento.

## Impact

- Só `openspec/specs/segredos/spec.md` e `openspec/specs/notificacoes/spec.md`,
  pelo archive. Sem código, contrato, banco ou deploy.
