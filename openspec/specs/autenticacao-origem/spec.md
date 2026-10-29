# autenticacao-origem Specification

## Purpose
Identificação dos Sistemas de origem que chamam a API de envio, por token do
Keycloak da casa, para que só serviços internos autorizados peçam Notificações
e cada Notificação saiba quem a pediu.

## Requirements

### Requirement: REQ-1 Autenticar o Sistema de origem
Se uma requisição à API de envio não trouxer um token válido emitido pelo realm
do OminiCore para o público `ominicore`, o sistema SHALL recusá-la como não
autenticada, sem registrar nada.
Origem: RN-1

#### Scenario: T-1 Pedido sem token é recusado
- **WHEN** um pedido de Notificação chega sem token
- **THEN** a resposta é "não autenticado" e nenhuma Notificação é registrada

#### Scenario: T-2 Token de outro emissor ou público é recusado
- **WHEN** um pedido chega com token expirado, de outro realm ou sem o público `ominicore`
- **THEN** a resposta é "não autenticado" e nenhuma Notificação é registrada

#### Scenario: T-3 Token válido identifica o Sistema de origem
- **WHEN** um pedido válido chega com token do client `sistema-a`
- **THEN** a Notificação registrada tem `sistema-a` como Sistema de origem

### Requirement: REQ-36 Conta do Console não pede Notificações
Se uma requisição para pedir ou consultar Notificação, ou para registrar ou
remover Opt-out, vier com token válido emitido para o client do Console, o
sistema SHALL recusá-la por falta de permissão, sem registrar nada.
Origem: RN-12, RN-1

#### Scenario: T-119 Token do Console recusado no envio e no Opt-out
- **WHEN** um token válido do client `ominicore-console`, com o papel `ominicore-config`, pede uma Notificação, consulta uma Notificação ou registra um Opt-out
- **THEN** a resposta é 403, nenhuma Notificação nem Opt-out é criado e nada é despachado
