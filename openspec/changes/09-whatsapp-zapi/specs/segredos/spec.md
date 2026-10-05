## MODIFIED Requirements

### Requirement: REQ-6 Credenciais só no OpenBao
O sistema SHALL obter as credenciais dos Provedores exclusivamente do OpenBao
e SHALL recusar-se a iniciar quando não conseguir lê-las, salvo as do Canal
WhatsApp: sem elas, SHALL iniciar com o Canal desligado. As credenciais do
WhatsApp SHALL indicar o Provedor do Canal, Meta, UltraMsg ou Z-API (sem
indicação, Meta), e SHALL impedir o início quando indicarem Provedor
desconhecido ou faltar algum campo exigido por esse Provedor. O sistema MUST
NOT escrever credencial, contato do Destinatário ou conteúdo da Notificação em
log.
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
- **WHEN** o processo inicia e as credenciais do WhatsApp, sem indicar o Provedor ou indicando a Meta, existem sem o identificador do número, ou sem o App Secret ou o verify token da Meta
- **THEN** o processo encerra com erro que cita o campo faltante, sem aceitar pedidos nem enviar Tentativas

#### Scenario: T-70 WhatsApp pelo UltraMsg inicia sem os segredos da Meta
- **WHEN** o processo inicia e as credenciais do WhatsApp indicam o UltraMsg, com a instância e o token, e não existem o App Secret nem o verify token da Meta
- **THEN** o processo inicia com o Canal WhatsApp ligado pelo UltraMsg e registra em log o Provedor do WhatsApp, sem a credencial

#### Scenario: T-71 Credenciais do UltraMsg incompletas impedem o início
- **WHEN** o processo inicia e as credenciais do WhatsApp indicam o UltraMsg sem a instância ou sem o token
- **THEN** o processo encerra com erro que cita o campo faltante, sem aceitar pedidos nem enviar Tentativas

#### Scenario: T-72 Provedor do WhatsApp desconhecido impede o início
- **WHEN** o processo inicia e as credenciais do WhatsApp indicam um Provedor que não é Meta, UltraMsg nem Z-API
- **THEN** o processo encerra com erro que cita o Provedor desconhecido, sem aceitar pedidos nem enviar Tentativas

#### Scenario: T-82 WhatsApp pela Z-API inicia sem os segredos da Meta
- **WHEN** o processo inicia e as credenciais do WhatsApp indicam a Z-API, com a instância e o token, sem o token de segurança da conta, e não existem o App Secret nem o verify token da Meta
- **THEN** o processo inicia com o Canal WhatsApp ligado pela Z-API e registra em log o Provedor do WhatsApp, sem a credencial

#### Scenario: T-83 Credenciais da Z-API incompletas impedem o início
- **WHEN** o processo inicia e as credenciais do WhatsApp indicam a Z-API sem a instância ou sem o token
- **THEN** o processo encerra com erro que cita o campo faltante, sem aceitar pedidos nem enviar Tentativas

#### Scenario: T-84 Z-API inacessível não expõe o token
- **WHEN** com a Z-API ativa, a chamada ao Provedor falha por erro de rede ou é recusada
- **THEN** nem o motivo da Tentativa nem as linhas de log contêm o token da instância ou o token de segurança da conta
