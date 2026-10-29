# Proposal

## Why

O OpenSpec 1.15.0 passou o aviso de texto de requisito acima de 500 caracteres
de INFO para WARNING, e o `openspec validate --all --strict` (CI, pre-commit e
hook Stop) passou a reprovar a spec `notificacoes` por causa do REQ-2 e do
REQ-14. A mudança é só de redação: nenhuma regra muda (origem continua RN-1,
RN-2, RN-5, RN-8 e RN-9).

## What Changes

- **REQ-2** e **REQ-14** reescritos de forma mais enxuta, com as mesmas
  obrigações (aceite, recusas, o que a consulta devolve e o que não devolve,
  só o dono enxerga).
- Todos os cenários (T-4, T-5, T-6, T-15 a T-18, T-43, T-44, T-56) copiados sem
  alteração: os testes que citam T-n não mudam. Nenhum REQ novo.

## Capabilities

### New Capabilities

### Modified Capabilities

- `notificacoes`: redação do REQ-2 e do REQ-14 abaixo de 500 caracteres, sem
  mudar comportamento.

## Impact

- Só `openspec/specs/notificacoes/spec.md`, pelo archive. Sem código,
  contrato, banco ou deploy.
- Nenhuma change ativa altera o REQ-2 nem o REQ-14 (a 08 e a 09 mexem nos
  REQ-10, 11 e 21), então os archives delas não desfazem este ajuste.
