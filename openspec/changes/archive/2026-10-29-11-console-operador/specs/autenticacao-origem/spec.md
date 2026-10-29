## ADDED Requirements

### Requirement: REQ-36 Conta do Console não pede Notificações
Se uma requisição para pedir ou consultar Notificação, ou para registrar ou
remover Opt-out, vier com token válido emitido para o client do Console, o
sistema SHALL recusá-la por falta de permissão, sem registrar nada.
Origem: RN-12, RN-1

#### Scenario: T-119 Token do Console recusado no envio e no Opt-out
- **WHEN** um token válido do client `ominicore-console`, com o papel `ominicore-config`, pede uma Notificação, consulta uma Notificação ou registra um Opt-out
- **THEN** a resposta é 403, nenhuma Notificação nem Opt-out é criado e nada é despachado
