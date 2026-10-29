## MODIFIED Requirements

### Requirement: REQ-10 Enviar pelo Canal WhatsApp
Tentativa pendente de WhatsApp SHALL ir ao Provedor ativo, para o telefone do
Destinatário. Meta: o Modelo aprovado (idioma e parâmetros em ordem); sem
WhatsApp ou Modelo não aprovado é Falha definitiva, e limite de taxa ou erro
genérico, temporária. UltraMsg e Z-API: o texto preenchido; limite de taxa ou
indisponibilidade é temporária, e outra recusa, definitiva. Texto com mais de
4096 caracteres ou versão sem a forma do Provedor SHALL ser Falha definitiva,
sem chamada.
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

#### Scenario: T-85 Tentativa de WhatsApp é enviada como texto pela Z-API
- **WHEN** com a Z-API ativa e sem token de segurança da conta, uma Notificação `boas-vindas` com `nome` = Ana e telefone +5511999990000 é processada no Canal WhatsApp e o Provedor aceita devolvendo um `messageId`
- **THEN** o Provedor recebe o texto "Olá Ana, sua conta está pronta." para o telefone 5511999990000, sem o cabeçalho de token de segurança, e a Tentativa fica enviada com esse `messageId`

#### Scenario: T-86 Token de segurança da conta vai em toda chamada à Z-API
- **WHEN** com a Z-API ativa e o token de segurança da conta configurado, uma Tentativa de WhatsApp é enviada
- **THEN** a chamada ao Provedor leva o token de segurança da conta no cabeçalho `Client-Token`

#### Scenario: T-87 Recusa da Z-API passa ao próximo Canal
- **WHEN** com a Z-API ativa, a ordem é WhatsApp → SMS e o Provedor recusa o token, a instância ou o token de segurança da conta
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa, com o motivo do Provedor, e a próxima Tentativa é no Canal SMS

#### Scenario: T-88 Z-API indisponível é Falha temporária
- **WHEN** com a Z-API ativa, o Provedor responde com limite de taxa, erro de servidor ou não responde
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-89 Texto longo demais não chega à Z-API
- **WHEN** com a Z-API ativa, o texto da versão WhatsApp preenchido passa de 4096 caracteres e a ordem é WhatsApp → SMS
- **THEN** o Provedor não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

### Requirement: REQ-11 Versão WhatsApp do Template
A versão WhatsApp do Template SHALL ser um Modelo aprovado (nome, idioma e
parâmetros ordenados, cada um uma variável), um texto com variáveis, ou os
dois, e o pedido com WhatsApp na Ordem de Canais SHALL usar o telefone do
Destinatário. Com o Canal ligado e na ordem, versão sem a forma do Provedor
ativo SHALL ser recusada com "versão whatsapp sem texto" (UltraMsg ou Z-API)
ou "versão whatsapp sem modelo aprovado" (Meta), sem registrar nada.
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

#### Scenario: T-90 Versão sem texto é recusada com a Z-API ativa
- **WHEN** com a Z-API ativa, o pedido informa `canais` WhatsApp → SMS para um Template cuja versão WhatsApp só tem Modelo aprovado
- **THEN** o pedido é recusado com o motivo "versão whatsapp sem texto" e nada é registrado
