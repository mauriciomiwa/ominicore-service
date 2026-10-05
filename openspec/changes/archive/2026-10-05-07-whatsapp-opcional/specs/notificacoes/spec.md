## ADDED Requirements

### Requirement: REQ-20 Canal desligado é indisponível
Enquanto um Canal estiver desligado, o sistema SHALL tratá-lo como
indisponível: SHALL pulá-lo na Ordem de Canais do pedido e do fallback, SHALL
recusar o pedido sem nenhum Canal disponível com o motivo "canal <canal>
indisponível" quando algum Canal pedido estiver desligado, SHALL encerrar como
Falha definitiva a Tentativa pendente nele sem chamar o Provedor e SHALL
responder como rota inexistente aos webhooks do Provedor desse Canal.
Origem: RN-4

#### Scenario: T-65 Pedido só por WhatsApp com o Canal desligado é recusado
- **WHEN** o WhatsApp está desligado e o pedido informa `canais` só WhatsApp para um Destinatário com telefone
- **THEN** o pedido é recusado com o motivo "canal whatsapp indisponível" e nada é registrado

#### Scenario: T-66 Canal desligado no começo da ordem é pulado
- **WHEN** o WhatsApp está desligado e o pedido informa `canais` WhatsApp → SMS para um Destinatário com telefone
- **THEN** o pedido é aceito com a ordem WhatsApp → SMS e a primeira Tentativa é no Canal SMS

#### Scenario: T-67 Fallback pula o Canal desligado
- **WHEN** o WhatsApp está desligado, a ordem é Email → WhatsApp → SMS e o Email falha de forma definitiva
- **THEN** não é criada Tentativa no Canal WhatsApp e a Tentativa seguinte é no Canal SMS

#### Scenario: T-68 Tentativa pendente no Canal desligado segue a ordem
- **WHEN** uma Tentativa de WhatsApp ficou pendente, o Canal é desligado antes do envio e a ordem é WhatsApp → SMS
- **THEN** o Provedor do WhatsApp não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

#### Scenario: T-69 Webhook da Meta com o Canal desligado
- **WHEN** o WhatsApp está desligado e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda
