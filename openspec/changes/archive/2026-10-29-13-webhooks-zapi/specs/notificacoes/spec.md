## ADDED Requirements

### Requirement: REQ-37 Aviso da Z-API autenticado pela URL
Quando chegar um aviso da Z-API com o segredo da URL e a instância configurados
para a Z-API ativa, o sistema SHALL tratá-lo como Aviso de status válido
(REQ-12, REQ-13) e responder `200 {"value": true}`, inclusive se nada mudar; com
segredo errado, outra instância ou a Z-API inativa, SHALL responder como rota
inexistente, sem efeito.
Origem: RN-5, RN-4

#### Scenario: T-121 Aviso com o segredo e a instância certos é aceito
- **WHEN** a Z-API está ativa e chega em `/webhooks/zapi/<segredo atual>` o status `RECEIVED` de uma Tentativa enviada, com o `instanceId` configurado
- **THEN** a resposta é 200 `{"value": true}` e a Tentativa fica entregue

#### Scenario: T-122 Segredo errado é recusado como rota inexistente
- **WHEN** chega um aviso com um segredo diferente do atual, inclusive o segredo anterior a uma regravação
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-123 Aviso de outra instância é recusado
- **WHEN** chega um aviso com o segredo atual e um `instanceId` diferente do configurado
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-124 Sem a Z-API ativa a rota não existe
- **WHEN** o UltraMsg ou a Meta é o Provedor ativo, ou o WhatsApp está desligado, e chega um aviso em `/webhooks/zapi/<qualquer segredo>`
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-125 Aviso autêntico sem efeito responde sucesso
- **WHEN** chega um aviso autêntico de uma mensagem que o OminiCore não conhece, ou um `RECEIVED` repetido de Tentativa já lida
- **THEN** a resposta é 200 `{"value": true}` e nada muda

### Requirement: REQ-38 Status da mensagem pela Z-API
Quando chegar da Z-API um aviso válido de status, o sistema SHALL aplicar a
cada id da lista `ids` o status entregue para `RECEIVED` e lida para `READ` ou
`PLAYED`, só fazendo avançar (REQ-12), e SHALL ignorar `SENT` e `READ_BY_ME`.
Origem: RN-5, RN-4

#### Scenario: T-126 Leitura confirmada pela Z-API
- **WHEN** chega da Z-API o status `READ` com o id de uma Tentativa entregue
- **THEN** a Tentativa fica lida e a Notificação segue entregue

#### Scenario: T-127 Status SENT não muda nada
- **WHEN** chega da Z-API o status `SENT` com o id de uma Tentativa enviada
- **THEN** a Tentativa continua enviada e a resposta é 200 `{"value": true}`

### Requirement: REQ-39 Erro de envio da Z-API depois do aceite
Quando chegar da Z-API um aviso válido "Ao enviar" com `error`, o sistema SHALL
marcar a Tentativa como falhou, como contato inválido para "Phone number does
not exist" ou número inválido e como recusado pelo Provedor nos demais, e
seguir o REQ-13; sem `error`, SHALL manter a Tentativa como está.
Origem: RN-5, RN-4, RN-2

#### Scenario: T-128 Número sem WhatsApp passa ao SMS
- **WHEN** a ordem é WhatsApp → SMS, a Z-API aceitou a Tentativa e chega "Ao enviar" com `error` "Phone number does not exist"
- **THEN** a Tentativa de WhatsApp fica falhou com classificação contato inválido e existe uma Tentativa pendente no Canal SMS

#### Scenario: T-129 Shadow ban é recusa do Provedor
- **WHEN** chega "Ao enviar" com `error` "The message was not sent due to a likely shadow ban" de uma Tentativa enviada
- **THEN** a Tentativa fica falhou com classificação recusado pelo Provedor

#### Scenario: T-130 Ao enviar sem erro não muda nada
- **WHEN** chega "Ao enviar" sem `error` de uma Tentativa enviada
- **THEN** a Tentativa continua enviada e a resposta é 200 `{"value": true}`

## MODIFIED Requirements

### Requirement: REQ-21 Provedor do WhatsApp sem Aviso de status
Enquanto o Provedor do WhatsApp ativo for o UltraMsg, o sistema SHALL manter
como enviada a Tentativa aceita por ele, sem esperar Aviso de status; com o
UltraMsg ou a Z-API ativos, SHALL responder como rota inexistente à
verificação e aos Avisos de status em `/webhooks/whatsapp` (os avisos da Z-API
chegam pelo REQ-37).
Origem: RN-4, RN-5

#### Scenario: T-81 Webhook da Meta com o UltraMsg ativo
- **WHEN** o UltraMsg é o Provedor do WhatsApp e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-91 Webhook da Meta com a Z-API ativa
- **WHEN** a Z-API é o Provedor do WhatsApp e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda
