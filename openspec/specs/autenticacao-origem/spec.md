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
