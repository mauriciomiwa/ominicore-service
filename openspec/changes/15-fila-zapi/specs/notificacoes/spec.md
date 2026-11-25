## MODIFIED Requirements

### Requirement: REQ-14 Consultar Notificação
Quando o Sistema de origem autenticado consultar uma Notificação que pediu, o
sistema SHALL devolver o status, o horário da anonimização, se houver, e de cada
Tentativa o Canal, o número, o status, os horários e a classificação da falha
(contato inválido, recusado pelo Provedor, Retentativas esgotadas, Opt-out,
removida pelo Operador); MUST NOT devolver contato do Destinatário nem o texto
do motivo; e SHALL responder como inexistente a Notificação de outro Sistema
de origem.
Origem: RN-5, RN-9, RN-8, RN-12

#### Scenario: T-43 Dono consulta status e Tentativas
- **WHEN** o Sistema de origem `sistema-a` consulta a Notificação que pediu, entregue pelo SMS depois de bounce no Email
- **THEN** a resposta traz a Notificação entregue, a Tentativa de Email falhou com classificação "contato inválido" e a de SMS entregue, sem e-mail, telefone nem texto do motivo

#### Scenario: T-44 Outro Sistema de origem não enxerga
- **WHEN** o Sistema de origem `sistema-b` consulta uma Notificação de `sistema-a`
- **THEN** a resposta é "não encontrada"

#### Scenario: T-56 Consulta mostra quando foi anonimizada
- **WHEN** o dono consulta uma Notificação já anonimizada
- **THEN** a resposta traz o status, as Tentativas e o horário da anonimização

#### Scenario: T-156 Consulta mostra a Tentativa retirada da fila
- **WHEN** o Sistema de origem `wbem` consulta uma Notificação só por WhatsApp cuja Tentativa o Operador retirou da fila da Z-API
- **THEN** a resposta traz a Notificação falhou e a Tentativa de WhatsApp falhou com a classificação `removida_pelo_operador`, sem o telefone nem o texto do motivo
