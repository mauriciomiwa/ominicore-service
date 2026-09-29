## MODIFIED Requirements

### Requirement: REQ-14 Consultar Notificação
Quando o Sistema de origem autenticado consultar uma Notificação que ele pediu,
o sistema SHALL devolver o status dela e de cada Tentativa (Canal, número,
status, horários e a classificação da falha: contato inválido, recusado pelo
Provedor, Retentativas esgotadas, Opt-out), MUST NOT devolver contato do
Destinatário nem o texto do motivo, e SHALL responder como inexistente a
Notificação de outro Sistema de origem.
Origem: RN-5, RN-9, RN-8

#### Scenario: T-43 Dono consulta status e Tentativas
- **WHEN** o Sistema de origem `sistema-a` consulta a Notificação que pediu, entregue pelo SMS depois de bounce no Email
- **THEN** a resposta traz a Notificação entregue, a Tentativa de Email falhou com classificação "contato inválido" e a de SMS entregue, sem e-mail, telefone nem texto do motivo

#### Scenario: T-44 Outro Sistema de origem não enxerga
- **WHEN** o Sistema de origem `sistema-b` consulta uma Notificação de `sistema-a`
- **THEN** a resposta é "não encontrada"

## ADDED Requirements

### Requirement: REQ-15 Registrar e remover Opt-out pela API
Quando um Sistema de origem autenticado registrar o Opt-out de um contato em
um Canal, o sistema SHALL passar a considerar esse Canal indisponível para
esse contato em qualquer Sistema de origem, de forma idempotente; quando
remover, o sistema SHALL voltar a considerá-lo disponível e SHALL guardar qual
Sistema de origem removeu e quando.
Origem: RN-8

#### Scenario: T-45 Registrar Opt-out
- **WHEN** o `sistema-a` registra o Opt-out do telefone `+5511999998888` no Canal SMS
- **THEN** a resposta é sucesso e o Canal SMS desse telefone fica indisponível

#### Scenario: T-46 Registro repetido é idempotente
- **WHEN** o mesmo Opt-out é registrado duas vezes
- **THEN** as duas respostas são sucesso e existe um único Opt-out

#### Scenario: T-47 Remoção guarda quem removeu
- **WHEN** o `sistema-b` remove o Opt-out de SMS daquele telefone
- **THEN** o Canal SMS volta a ficar disponível e o histórico guarda `sistema-b` e o horário da remoção

#### Scenario: T-48 Pedido de Opt-out inválido é recusado
- **WHEN** o pedido traz um Canal desconhecido, um e-mail inválido ou um telefone fora do E.164
- **THEN** o pedido é recusado e nada é registrado

### Requirement: REQ-16 Opt-out pelos sinais do Provedor
Quando o Provedor avisar que o Destinatário se descadastrou (evento
`unsubscribe` ou `spamreport` do SendGrid) ou recusar o envio por STOP (código
21610 do Twilio), o sistema SHALL registrar o Opt-out daquele contato naquele
Canal; a recusa por STOP SHALL ser Falha definitiva, seguindo para o próximo
Canal.
Origem: RN-8

#### Scenario: T-49 Descadastro no SendGrid registra Opt-out de Email
- **WHEN** chega do SendGrid, com assinatura válida, o evento `unsubscribe` para `ana@exemplo.com`
- **THEN** o Canal Email de `ana@exemplo.com` fica em Opt-out

#### Scenario: T-50 STOP no Twilio registra Opt-out de SMS e segue a ordem
- **WHEN** a ordem é SMS → Email e o Twilio recusa o envio com o código 21610
- **THEN** o Canal SMS do telefone fica em Opt-out e a próxima Tentativa é no Canal Email

### Requirement: REQ-17 Canal em Opt-out é pulado
O sistema SHALL tratar Canal em Opt-out como indisponível para aquele contato:
no pedido (sem nenhum Canal disponível, SHALL recusá-lo), no fallback e no
envio, em que SHALL conferir o Opt-out antes de chamar o Provedor e, sem Canal
restante, SHALL encerrar a Notificação como falhou com a classificação Opt-out.
Origem: RN-8

#### Scenario: T-51 Opt-out de SMS faz a ordem SMS → Email sair por Email
- **WHEN** o telefone tem Opt-out de SMS e o pedido tem a ordem SMS → Email
- **THEN** a primeira Tentativa é no Canal Email

#### Scenario: T-52 Todos os Canais em Opt-out recusam o pedido
- **WHEN** o Destinatário tem Opt-out em todos os Canais da ordem
- **THEN** o pedido é recusado com o motivo "destinatário sem contato para os canais" e nada é registrado

#### Scenario: T-53 Opt-out entre o pedido e o envio
- **WHEN** a Notificação foi aceita com a ordem SMS → Email e o Opt-out de SMS é registrado antes do envio
- **THEN** o Provedor de SMS não é chamado, a Tentativa de SMS fica falhou com a classificação Opt-out e a próxima é no Canal Email

#### Scenario: T-54 Opt-out vale para todos os Sistemas de origem
- **WHEN** o `sistema-a` registrou o Opt-out de Email e o `sistema-b` pede uma Notificação só por Email para o mesmo endereço
- **THEN** o pedido do `sistema-b` é recusado

### Requirement: REQ-18 Opt-out sem contato em claro
O sistema MUST NOT guardar o contato do Opt-out em claro: SHALL guardar só um
HMAC do contato normalizado (e-mail em minúsculas e sem espaços; telefone em
E.164), com chave do OpenBao, e o registro SHALL permanecer depois da
retenção das Notificações.
Origem: RN-8, RN-9

#### Scenario: T-55 Tabela de Opt-out sem contato legível
- **WHEN** o Opt-out de `Ana@Exemplo.com` é registrado
- **THEN** nenhuma coluna da tabela de Opt-out contém o endereço, e o Opt-out vale para `ana@exemplo.com`
