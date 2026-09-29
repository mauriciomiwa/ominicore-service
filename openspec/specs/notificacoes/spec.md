# notificacoes Specification

## Purpose
Pedido, registro e envio de Notificações a partir de Templates, com Chave de
idempotência para que o reenvio de um pedido nunca faça o Destinatário receber
duas vezes, e com cada Tentativa registrada com Canal e resultado.

## Requirements

### Requirement: REQ-2 Pedir Notificação por Template
Quando um Sistema de origem autenticado pedir uma Notificação informando um
Template existente, o Destinatário e todas as variáveis do Template, o sistema
SHALL registrar a Notificação como pendente, com a Ordem de Canais do pedido
(ou, sem ela, a do Template), a primeira Tentativa no primeiro Canal dessa
ordem para o qual o Destinatário tem contato, e responder com o identificador
da Notificação. O sistema SHALL recusar o pedido cuja ordem cite Canal sem
versão no Template, cujo Destinatário não tenha contato para nenhum Canal da
ordem ou cujo telefone não esteja no formato E.164.
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

### Requirement: REQ-7 Enviar pelo Canal SMS
Quando houver uma Tentativa pendente no Canal SMS, o sistema SHALL enviá-la ao
Provedor de SMS com o corpo da versão SMS do Template preenchido e registrar o
resultado, tratando recusa por número inválido como Falha definitiva.
Origem: RN-3

#### Scenario: T-19 Tentativa de SMS é enviada
- **WHEN** uma Notificação pendente no Canal SMS é processada e o Provedor aceita
- **THEN** o Provedor recebe o telefone do Destinatário e o corpo com as variáveis preenchidas, e a Tentativa fica enviada

#### Scenario: T-20 Número inválido é Falha definitiva
- **WHEN** o Provedor de SMS recusa o número do Destinatário como inválido
- **THEN** a Tentativa fica falhou com o motivo do Provedor, sem Retentativa

### Requirement: REQ-8 Retentar Falha temporária
Se uma Tentativa terminar em Falha temporária, o sistema SHALL agendar uma
Retentativa no mesmo Canal após 5 s, 30 s e 2 min (±20%), até 3 Retentativas;
esgotadas, SHALL tratar o Canal como falho.
Origem: RN-2

#### Scenario: T-21 Falha temporária agenda Retentativa no mesmo Canal
- **WHEN** o Provedor responde indisponível na primeira Tentativa por Email
- **THEN** existe uma segunda Tentativa pendente no Canal Email, disponível entre 4 s e 6 s depois, e a Notificação segue pendente

#### Scenario: T-22 Espera cresce a cada Retentativa
- **WHEN** as Tentativas 1, 2 e 3 de um Canal terminam em Falha temporária
- **THEN** as Retentativas ficam disponíveis após 5 s, 30 s e 2 min, cada uma dentro de ±20%

#### Scenario: T-23 Retentativas esgotadas encerram o Canal
- **WHEN** as 4 Tentativas de um Canal terminam em Falha temporária
- **THEN** não há quinta Tentativa nesse Canal e a Notificação segue a Ordem de Canais

### Requirement: REQ-9 Fallback pela Ordem de Canais
Quando um Canal falhar (Falha definitiva ou Retentativas esgotadas), o sistema
SHALL criar a Tentativa no próximo Canal da Ordem de Canais para o qual o
Destinatário tem contato; sem Canal restante, SHALL marcar a Notificação como
falhou; com sucesso em qualquer Canal, SHALL marcá-la como enviada e parar.
Origem: RN-2

#### Scenario: T-24 Falha definitiva no Email passa ao SMS
- **WHEN** a ordem é Email → SMS e o Provedor de Email recusa o endereço
- **THEN** a Tentativa seguinte é no Canal SMS e, aceita, a Notificação fica enviada

#### Scenario: T-25 Canal sem contato é pulado
- **WHEN** a ordem é Email → SMS, o Email falha de forma definitiva e o Destinatário não tem telefone
- **THEN** não é criada Tentativa no Canal SMS e a Notificação fica falhou

#### Scenario: T-26 Todos os Canais falham
- **WHEN** a ordem é Email → SMS e os dois Provedores recusam de forma definitiva
- **THEN** a Notificação fica falhou com uma Tentativa falhou por Canal

#### Scenario: T-27 Sucesso encerra o fallback
- **WHEN** a ordem é Email → SMS e o Email é aceito
- **THEN** a Notificação fica enviada e nenhuma Tentativa é criada no Canal SMS

### Requirement: REQ-10 Enviar pelo Canal WhatsApp
Quando houver uma Tentativa pendente no Canal WhatsApp, o sistema SHALL
enviá-la ao Provedor de WhatsApp como o Modelo aprovado da versão WhatsApp do
Template, no idioma dela, com os valores das variáveis na ordem dos parâmetros,
para o telefone do Destinatário; recusa por número sem WhatsApp ou por Modelo
não aprovado SHALL ser Falha definitiva, e limite de taxa ou erro genérico do
Provedor SHALL ser Falha temporária.
Origem: RN-4

#### Scenario: T-28 Tentativa de WhatsApp é enviada pelo Modelo aprovado
- **WHEN** uma Notificação `boas-vindas` com `nome` = Ana é processada no Canal WhatsApp e o Provedor aceita
- **THEN** o Provedor recebe o Modelo `boas_vindas`, idioma `pt_BR`, parâmetro 1 = Ana e o telefone do Destinatário, e a Tentativa fica enviada

#### Scenario: T-29 Número sem WhatsApp passa ao próximo Canal
- **WHEN** a ordem é WhatsApp → SMS e o Provedor de WhatsApp recusa com o código 131026
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa e a próxima Tentativa é no Canal SMS

#### Scenario: T-30 Limite de taxa é Falha temporária
- **WHEN** o Provedor de WhatsApp recusa com o código 130429
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-31 Modelo não aprovado é Falha definitiva
- **WHEN** o Provedor de WhatsApp recusa com o código 132001
- **THEN** a Tentativa fica falhou com o motivo do Provedor, sem Retentativa

### Requirement: REQ-11 Versão WhatsApp do Template
Onde o Template tiver versão WhatsApp, o sistema SHALL tratá-la como a
referência a um Modelo aprovado (nome, idioma e parâmetros ordenados, cada um
uma variável do Template) e SHALL aceitar pedidos com o Canal WhatsApp na
Ordem de Canais usando o telefone do Destinatário como contato.
Origem: RN-4

#### Scenario: T-32 Pedido com WhatsApp na ordem é aceito
- **WHEN** o pedido informa `canais` WhatsApp → SMS e o Destinatário tem telefone
- **THEN** a Notificação é registrada com a primeira Tentativa no Canal WhatsApp

#### Scenario: T-33 Parâmetro do Modelo aprovado é variável exigida
- **WHEN** o pedido com Canal WhatsApp omite a variável `nome`, usada como parâmetro 1 do Modelo aprovado
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado
