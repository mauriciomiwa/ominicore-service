# segredos Specification

## Purpose
Guarda das credenciais dos Provedores e dos dados pessoais fora de código,
configuração versionada, banco e log: credenciais vêm só do OpenBao, e o
sistema falha fechado quando não consegue obtê-las.

## Requirements

### Requirement: REQ-6 Credenciais só no OpenBao
O sistema SHALL obter as credenciais dos Provedores exclusivamente do OpenBao
e SHALL recusar-se a iniciar quando não conseguir lê-las, salvo as do Canal
WhatsApp: sem elas, SHALL iniciar com o Canal desligado. Credenciais do
WhatsApp incompletas SHALL impedir o início. O sistema MUST NOT escrever
credencial, contato do Destinatário ou conteúdo da Notificação em log.
Origem: RN-9, RN-4

#### Scenario: T-13 Sem OpenBao o processo não inicia
- **WHEN** o processo inicia e o OpenBao está indisponível ou nega a leitura
- **THEN** o processo encerra com erro que cita o OpenBao, sem aceitar pedidos nem enviar Tentativas

#### Scenario: T-14 Log sem credencial nem dado pessoal
- **WHEN** um pedido é aceito, enviado e falha no Provedor
- **THEN** nenhuma linha de log contém a credencial do Provedor, o e-mail do Destinatário ou o conteúdo preenchido

#### Scenario: T-63 Sem credenciais da Meta o processo inicia com o WhatsApp desligado
- **WHEN** o processo inicia, o OpenBao tem as credenciais de Email, SMS e webhooks do SendGrid e do Twilio e não tem as credenciais do WhatsApp
- **THEN** o processo inicia, aceita pedidos por Email e SMS e registra em log que o WhatsApp está desligado

#### Scenario: T-64 Credenciais da Meta incompletas impedem o início
- **WHEN** o processo inicia e as credenciais do WhatsApp existem sem o identificador do número, ou sem o App Secret ou o verify token da Meta
- **THEN** o processo encerra com erro que cita o campo faltante, sem aceitar pedidos nem enviar Tentativas
