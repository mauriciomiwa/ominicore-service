## MODIFIED Requirements

### Requirement: REQ-10 Enviar pelo Canal WhatsApp
Quando houver uma Tentativa pendente no Canal WhatsApp, o sistema SHALL
enviá-la ao Provedor do WhatsApp ativo no ambiente, para o telefone do
Destinatário. Com a Meta, SHALL enviar o Modelo aprovado da versão WhatsApp do
Template, no idioma dela, com os valores das variáveis na ordem dos parâmetros;
recusa por número sem WhatsApp ou por Modelo não aprovado SHALL ser Falha
definitiva, e limite de taxa ou erro genérico SHALL ser Falha temporária. Com
o UltraMsg, SHALL enviar o texto da versão WhatsApp preenchido com as
variáveis; limite de taxa ou indisponibilidade do Provedor SHALL ser Falha
temporária, e qualquer outra recusa SHALL ser Falha definitiva. Texto
preenchido com mais de 4096 caracteres, ou versão WhatsApp sem a forma do
Provedor ativo, SHALL ser Falha definitiva sem chamar o Provedor.
Origem: RN-4

#### Scenario: T-28 Tentativa de WhatsApp é enviada pelo Modelo aprovado
- **WHEN** com a Meta ativa, uma Notificação `boas-vindas` com `nome` = Ana é processada no Canal WhatsApp e o Provedor aceita
- **THEN** o Provedor recebe o Modelo `boas_vindas`, idioma `pt_BR`, parâmetro 1 = Ana e o telefone do Destinatário, e a Tentativa fica enviada

#### Scenario: T-29 Número sem WhatsApp passa ao próximo Canal
- **WHEN** com a Meta ativa, a ordem é WhatsApp → SMS e o Provedor de WhatsApp recusa com o código 131026
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa e a próxima Tentativa é no Canal SMS

#### Scenario: T-30 Limite de taxa é Falha temporária
- **WHEN** com a Meta ativa, o Provedor de WhatsApp recusa com o código 130429
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-31 Modelo não aprovado é Falha definitiva
- **WHEN** com a Meta ativa, o Provedor de WhatsApp recusa com o código 132001
- **THEN** a Tentativa fica falhou com o motivo do Provedor, sem Retentativa

#### Scenario: T-73 Tentativa de WhatsApp é enviada como texto pelo UltraMsg
- **WHEN** com o UltraMsg ativo, uma Notificação `boas-vindas` com `nome` = Ana é processada no Canal WhatsApp e o Provedor aceita devolvendo um id
- **THEN** o Provedor recebe o texto "Olá Ana, sua conta está pronta." para o telefone do Destinatário, e a Tentativa fica enviada com esse id

#### Scenario: T-74 Recusa do UltraMsg passa ao próximo Canal
- **WHEN** com o UltraMsg ativo, a ordem é WhatsApp → SMS e o Provedor recusa o token ou a instância
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa, com o motivo do Provedor, e a próxima Tentativa é no Canal SMS

#### Scenario: T-75 UltraMsg indisponível é Falha temporária
- **WHEN** com o UltraMsg ativo, o Provedor responde com limite de taxa, erro de servidor ou não responde
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-76 Texto longo demais não chega ao UltraMsg
- **WHEN** com o UltraMsg ativo, o texto da versão WhatsApp preenchido passa de 4096 caracteres e a ordem é WhatsApp → SMS
- **THEN** o Provedor não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

#### Scenario: T-77 Tentativa pendente sem a forma do Provedor ativo segue a ordem
- **WHEN** uma Tentativa de WhatsApp ficou pendente, o Provedor ativo passou a ser um cuja forma a versão WhatsApp do Template não tem, e a ordem é WhatsApp → SMS
- **THEN** o Provedor não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

### Requirement: REQ-11 Versão WhatsApp do Template
Onde o Template tiver versão WhatsApp, o sistema SHALL tratá-la como a
referência a um Modelo aprovado (nome, idioma e parâmetros ordenados, cada um
uma variável do Template), como um texto com variáveis, ou como as duas
formas, e SHALL aceitar pedidos com o Canal WhatsApp na Ordem de Canais usando
o telefone do Destinatário como contato. Quando o Canal WhatsApp estiver
ligado e na Ordem de Canais e a versão não tiver a forma do Provedor ativo, o
sistema SHALL recusar o pedido com o motivo "versão whatsapp sem texto"
(UltraMsg) ou "versão whatsapp sem modelo aprovado" (Meta), sem registrar
nada.
Origem: RN-4

#### Scenario: T-32 Pedido com WhatsApp na ordem é aceito
- **WHEN** o pedido informa `canais` WhatsApp → SMS e o Destinatário tem telefone
- **THEN** a Notificação é registrada com a primeira Tentativa no Canal WhatsApp

#### Scenario: T-33 Parâmetro do Modelo aprovado é variável exigida
- **WHEN** o pedido com Canal WhatsApp omite a variável `nome`, usada como parâmetro 1 do Modelo aprovado
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

#### Scenario: T-78 Versão sem texto é recusada com o UltraMsg ativo
- **WHEN** com o UltraMsg ativo, o pedido informa `canais` WhatsApp → SMS para um Template cuja versão WhatsApp só tem Modelo aprovado
- **THEN** o pedido é recusado com o motivo "versão whatsapp sem texto" e nada é registrado

#### Scenario: T-79 Versão sem Modelo aprovado é recusada com a Meta ativa
- **WHEN** com a Meta ativa, o pedido informa `canais` WhatsApp → SMS para um Template cuja versão WhatsApp só tem texto
- **THEN** o pedido é recusado com o motivo "versão whatsapp sem modelo aprovado" e nada é registrado

#### Scenario: T-80 Variável do texto do WhatsApp é exigida
- **WHEN** com o UltraMsg ativo, o pedido com Canal WhatsApp omite a variável `nome`, usada no texto da versão WhatsApp
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

## ADDED Requirements

### Requirement: REQ-21 Provedor do WhatsApp sem Aviso de status
Enquanto o Provedor do WhatsApp ativo for o UltraMsg, o sistema SHALL manter
como enviada a Tentativa aceita por ele, sem esperar Aviso de status, e SHALL
responder como rota inexistente à verificação e aos Avisos de status em
`/webhooks/whatsapp`.
Origem: RN-4, RN-5

#### Scenario: T-81 Webhook da Meta com o UltraMsg ativo
- **WHEN** o UltraMsg é o Provedor do WhatsApp e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda
