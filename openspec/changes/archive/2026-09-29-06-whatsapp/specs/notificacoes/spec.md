## ADDED Requirements

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
