## Purpose
Guarda das credenciais dos Provedores e dos dados pessoais fora de código,
configuração versionada, banco e log: credenciais vêm só do OpenBao, e o
sistema falha fechado quando não consegue obtê-las.

## ADDED Requirements

### Requirement: REQ-6 Credenciais só no OpenBao
O sistema SHALL obter as credenciais dos Provedores exclusivamente do OpenBao,
SHALL recusar-se a iniciar quando não conseguir lê-las e MUST NOT escrever
credencial, contato do Destinatário ou conteúdo da Notificação em log.
Origem: RN-9

#### Scenario: T-13 Sem OpenBao o processo não inicia
- **WHEN** o processo inicia e o OpenBao está indisponível ou nega a leitura
- **THEN** o processo encerra com erro que cita o OpenBao, sem aceitar pedidos nem enviar Tentativas

#### Scenario: T-14 Log sem credencial nem dado pessoal
- **WHEN** um pedido é aceito, enviado e falha no Provedor
- **THEN** nenhuma linha de log contém a credencial do Provedor, o e-mail do Destinatário ou o conteúdo preenchido
