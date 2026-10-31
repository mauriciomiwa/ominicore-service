## ADDED Requirements

### Requirement: REQ-49 Modelos da régua de comunicação do wbem
Quando o Sistema de origem pedir uma Notificação com um dos 16 Templates da régua do `wbem` (RN-14), o sistema SHALL encontrá-lo só com a versão WhatsApp em texto e preenchê-la com as variáveis do modelo; faltando uma delas, SHALL recusar o pedido (422) sem registrar nada. Os Templates da RN-13 SHALL continuar como estão.
Origem: RN-14

#### Scenario: T-157 Os 16 modelos preenchem o texto do WhatsApp
- **GIVEN** o banco com as migrations aplicadas
- **WHEN** cada modelo da régua é preenchido com as suas variáveis
- **THEN** o texto usa todas elas, começa pelo título em negrito, termina no aviso de segurança, a Ordem de Canais é só WhatsApp e o texto serve à Z-API e ao UltraMsg

#### Scenario: T-158 Sem variável é recusado e a RN-13 não muda
- **WHEN** o pedido usa `wbem-pix-nao-confirmado` sem `reservada_ate`
- **THEN** o pedido é recusado (422) e nada é registrado
- **WHEN** o `wbem-link-tomador` é preenchido com `link`
- **THEN** o texto é o mesmo da RN-13
