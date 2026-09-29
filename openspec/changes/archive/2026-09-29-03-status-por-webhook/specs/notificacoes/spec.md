## ADDED Requirements

### Requirement: REQ-12 Receber Aviso de status do Provedor
Quando chegar um Aviso de status com assinatura válida do Provedor, o sistema
SHALL atualizar a Tentativa identificada pelo id do Provedor, só fazendo o
status avançar (enviada → entregue → lida; falhou é final), SHALL marcar a
Notificação como entregue quando uma Tentativa for entregue, e SHALL responder
com sucesso a todo aviso válido, inclusive repetido ou de Tentativa
desconhecida. Se a assinatura for inválida ou ausente, o sistema SHALL recusar
o aviso sem nenhum efeito.
Origem: RN-5

#### Scenario: T-34 Entrega confirmada pelo Email
- **WHEN** chega do SendGrid, com assinatura válida, o evento `delivered` de uma Tentativa enviada
- **THEN** a Tentativa e a Notificação ficam entregues

#### Scenario: T-35 Entrega confirmada pelo SMS
- **WHEN** chega do Twilio, com assinatura válida, o status `delivered` de uma Tentativa enviada
- **THEN** a Tentativa e a Notificação ficam entregues

#### Scenario: T-36 Leitura confirmada pelo WhatsApp
- **WHEN** chega da Meta, com assinatura válida, o status `read` de uma Tentativa entregue
- **THEN** a Tentativa fica lida e a Notificação segue entregue

#### Scenario: T-37 Assinatura inválida é recusada
- **WHEN** chega um aviso de qualquer Provedor com assinatura inválida ou ausente
- **THEN** o aviso é recusado como não autorizado e nenhuma Tentativa muda

#### Scenario: T-38 Status não regride
- **WHEN** depois de `read`, chega da Meta um `delivered` atrasado da mesma Tentativa
- **THEN** a Tentativa continua lida e o aviso é respondido com sucesso

#### Scenario: T-39 Aviso de Tentativa desconhecida
- **WHEN** chega um aviso válido com id do Provedor que não existe
- **THEN** o aviso é respondido com sucesso e nada muda

#### Scenario: T-40 Verificação do webhook da Meta
- **WHEN** a Meta verifica o webhook com o verify token correto
- **THEN** a resposta devolve o desafio recebido; com token errado, a verificação é recusada

### Requirement: REQ-13 Falha de entrega depois do aceite
Quando um Aviso de status válido informar que uma Tentativa aceita não foi
entregue, o sistema SHALL marcá-la como falhou (Falha definitiva) e SHALL criar
a Tentativa no próximo Canal da Ordem de Canais com contato (REQ-9); se a
Notificação já estiver entregue por outra Tentativa, o sistema SHALL apenas
registrar a falha.
Origem: RN-5, RN-2

#### Scenario: T-41 Bounce no Email passa ao SMS
- **WHEN** a ordem é Email → SMS, o Email foi aceito e chega do SendGrid o evento `bounce`
- **THEN** a Tentativa de Email fica falhou e existe uma Tentativa pendente no Canal SMS

#### Scenario: T-42 Falha tardia não reabre Notificação entregue
- **WHEN** a Notificação já está entregue e chega a falha de outra Tentativa dela
- **THEN** a Tentativa fica falhou, a Notificação segue entregue e nenhuma Tentativa nova é criada

### Requirement: REQ-14 Consultar Notificação
Quando o Sistema de origem autenticado consultar uma Notificação que ele pediu,
o sistema SHALL devolver o status dela e de cada Tentativa (Canal, número,
status, horários e a classificação da falha: contato inválido, recusado pelo
Provedor, Retentativas esgotadas), MUST NOT devolver contato do Destinatário
nem o texto do motivo, e SHALL responder como inexistente a Notificação de
outro Sistema de origem.
Origem: RN-5, RN-9

#### Scenario: T-43 Dono consulta status e Tentativas
- **WHEN** o Sistema de origem `sistema-a` consulta a Notificação que pediu, entregue pelo SMS depois de bounce no Email
- **THEN** a resposta traz a Notificação entregue, a Tentativa de Email falhou com classificação "contato inválido" e a de SMS entregue, sem e-mail, telefone nem texto do motivo

#### Scenario: T-44 Outro Sistema de origem não enxerga
- **WHEN** o Sistema de origem `sistema-b` consulta uma Notificação de `sistema-a`
- **THEN** a resposta é "não encontrada"
