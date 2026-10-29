# Tasks

## 1. Ajuste de redação
- [x] 1.1 Deltas de `segredos` (REQ-6) e `notificacoes` (REQ-10, REQ-11) reescritos abaixo de 500 caracteres (contando a linha Origem), com as mesmas obrigações
- [x] 1.2 Conferir que nenhum cenário foi criado, removido ou reescrito — comparação dos T-n do delta com as specs vigentes

## 2. Fechamento
- [x] 2.1 Rodar `openspec validate ajuste-texto-whatsapp --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 2.2 Arquivar e rodar `openspec validate --all --strict` sem spec reprovada
