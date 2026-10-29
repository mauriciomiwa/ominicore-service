## MODIFIED Requirements

### Requirement: REQ-2 Pedir Notificação por Template
Quando um Sistema de origem autenticado pedir uma Notificação com um Template
existente, o Destinatário e todas as variáveis dele, o sistema SHALL
registrá-la como pendente, com a Ordem de Canais do pedido (ou a do Template) e
a primeira Tentativa no primeiro Canal da ordem com contato do Destinatário, e
responder com o identificador; SHALL recusar o pedido com Canal da
ordem sem versão no Template, sem contato do Destinatário para a ordem ou com
telefone fora do E.164.
Origem: RN-1, RN-2

#### Scenario: T-4 Pedido válido é aceito
- **WHEN** o Sistema de origem pede uma Notificação com Template `boas-vindas`, e-mail do Destinatário e a variável `nome`
- **THEN** a resposta traz o identificador e a Notificação fica pendente com uma Tentativa no Canal Email

#### Scenario: T-5 Template inexistente é recusado
- **WHEN** o pedido informa um Template que não existe
- **THEN** o pedido é recusado com o motivo "template inexistente" e nada é registrado

#### Scenario: T-6 Variável faltando é recusada
- **WHEN** o pedido omite uma variável exigida pelo Template
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

#### Scenario: T-15 Ordem do pedido substitui a do Template
- **WHEN** o pedido informa `canais` SMS → Email e o Destinatário tem telefone e e-mail
- **THEN** a Notificação guarda a ordem SMS → Email e a primeira Tentativa é no Canal SMS

#### Scenario: T-16 Canal pedido sem versão no Template é recusado
- **WHEN** o pedido informa um Canal que o Template não tem
- **THEN** o pedido é recusado citando o Canal e nada é registrado

#### Scenario: T-17 Destinatário sem contato para nenhum Canal é recusado
- **WHEN** a ordem é só SMS e o Destinatário tem apenas e-mail
- **THEN** o pedido é recusado com o motivo "destinatário sem contato para os canais" e nada é registrado

#### Scenario: T-18 Telefone fora do E.164 é recusado
- **WHEN** o Destinatário informa o telefone `11 99999-8888`
- **THEN** o pedido é recusado citando o formato E.164 e nada é registrado

### Requirement: REQ-14 Consultar Notificação
Quando o Sistema de origem autenticado consultar uma Notificação que pediu, o
sistema SHALL devolver o status, o horário da anonimização, se houver, e de cada
Tentativa o Canal, o número, o status, os horários e a classificação da falha
(contato inválido, recusado pelo Provedor, Retentativas esgotadas, Opt-out);
MUST NOT devolver contato do Destinatário nem o texto do motivo; e SHALL
responder como inexistente a Notificação de outro Sistema de origem.
Origem: RN-5, RN-9, RN-8

#### Scenario: T-43 Dono consulta status e Tentativas
- **WHEN** o Sistema de origem `sistema-a` consulta a Notificação que pediu, entregue pelo SMS depois de bounce no Email
- **THEN** a resposta traz a Notificação entregue, a Tentativa de Email falhou com classificação "contato inválido" e a de SMS entregue, sem e-mail, telefone nem texto do motivo

#### Scenario: T-44 Outro Sistema de origem não enxerga
- **WHEN** o Sistema de origem `sistema-b` consulta uma Notificação de `sistema-a`
- **THEN** a resposta é "não encontrada"

#### Scenario: T-56 Consulta mostra quando foi anonimizada
- **WHEN** o dono consulta uma Notificação já anonimizada
- **THEN** a resposta traz o status, as Tentativas e o horário da anonimização
