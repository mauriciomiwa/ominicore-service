# notificacoes Specification

## Purpose
Pedido, registro e envio de Notificações a partir de Templates, com Chave de
idempotência para que o reenvio de um pedido nunca faça o Destinatário receber
duas vezes, e com cada Tentativa registrada com Canal e resultado.

## Requirements

### Requirement: REQ-2 Pedir Notificação por Template
Quando um Sistema de origem autenticado pedir uma Notificação informando um
Template existente, o Destinatário e todas as variáveis do Template, o sistema
SHALL registrar a Notificação como pendente, com a primeira Tentativa no
primeiro Canal da Ordem de Canais, e responder com o identificador da
Notificação.
Origem: RN-1

#### Scenario: T-4 Pedido válido é aceito
- **WHEN** o Sistema de origem pede uma Notificação com Template `boas-vindas`, e-mail do Destinatário e a variável `nome`
- **THEN** a resposta traz o identificador e a Notificação fica pendente com uma Tentativa no Canal Email

#### Scenario: T-5 Template inexistente é recusado
- **WHEN** o pedido informa um Template que não existe
- **THEN** o pedido é recusado com o motivo "template inexistente" e nada é registrado

#### Scenario: T-6 Variável faltando é recusada
- **WHEN** o pedido omite uma variável exigida pelo Template
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

### Requirement: REQ-3 Pedido idempotente
Quando um Sistema de origem repetir a mesma Chave de idempotência em até 24 h,
o sistema SHALL devolver a Notificação já criada sem registrar nova Notificação
nem nova Tentativa; pedido sem Chave de idempotência SHALL ser recusado.
Origem: RN-7

#### Scenario: T-7 Reenvio com a mesma chave não duplica
- **WHEN** o mesmo Sistema de origem envia duas vezes o pedido com a chave `abc-1` dentro de 24 h
- **THEN** as duas respostas trazem o mesmo identificador e existe uma única Notificação com uma única Tentativa

#### Scenario: T-8 Pedido sem chave é recusado
- **WHEN** o pedido chega sem Chave de idempotência
- **THEN** o pedido é recusado com o motivo "chave de idempotência obrigatória" e nada é registrado

#### Scenario: T-9 A chave vale por Sistema de origem
- **WHEN** dois Sistemas de origem diferentes usam a mesma chave `abc-1`
- **THEN** são registradas duas Notificações distintas

### Requirement: REQ-4 Enviar pelo Canal Email
Quando houver uma Tentativa pendente no Canal Email, o sistema SHALL enviá-la
ao Provedor de Email com o conteúdo do Template preenchido e registrar a
Tentativa como enviada.
Origem: RN-3

#### Scenario: T-10 Tentativa de Email é enviada
- **WHEN** uma Notificação pendente no Canal Email é processada e o Provedor aceita
- **THEN** o Provedor recebe o assunto e o corpo com as variáveis preenchidas e a Tentativa fica enviada

#### Scenario: T-11 Primeira Tentativa em até 5 segundos
- **WHEN** um pedido válido é aceito com o sistema ocioso
- **THEN** a primeira Tentativa chega ao Provedor em até 5 segundos após a resposta do pedido

### Requirement: REQ-5 Registrar falha do Provedor
Se o Provedor recusar a Tentativa, o sistema SHALL registrar a Tentativa como
falhou com o motivo informado e, não havendo outro Canal na Ordem de Canais,
SHALL marcar a Notificação como falhou.
Origem: RN-2, RN-3

#### Scenario: T-12 Recusa do Provedor encerra a Notificação sem outro Canal
- **WHEN** o Provedor de Email recusa a Tentativa por endereço inválido e a Ordem de Canais só tem Email
- **THEN** a Tentativa fica falhou com o motivo do Provedor e a Notificação fica falhou
