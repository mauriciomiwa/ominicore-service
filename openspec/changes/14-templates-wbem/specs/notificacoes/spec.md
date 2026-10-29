## ADDED Requirements

### Requirement: REQ-44 Templates do wbem para o tomador
Quando o Sistema de origem pedir uma Notificação com o Template
`wbem-link-tomador` ou `wbem-otp-tomador`, o sistema SHALL encontrá-lo com a
versão WhatsApp em texto e preenchê-la com a variável `link` ou `codigo`; o
Template SHALL ter só o Canal WhatsApp.
Origem: RN-13

#### Scenario: T-139 Os dois Templates preenchem o texto do WhatsApp
- **GIVEN** o banco com as migrations aplicadas
- **WHEN** o Template `wbem-link-tomador` é preenchido com `link` e o
  `wbem-otp-tomador` com `codigo`
- **THEN** o texto do WhatsApp traz o link e o código, a Ordem de Canais é só
  WhatsApp e o texto serve à Z-API e ao UltraMsg

#### Scenario: T-140 Pedido sem a variável ou por outro Canal é recusado
- **WHEN** o pedido usa `wbem-otp-tomador` sem `codigo`, ou pede o Canal SMS
- **THEN** o pedido é recusado (422) e nada é registrado
