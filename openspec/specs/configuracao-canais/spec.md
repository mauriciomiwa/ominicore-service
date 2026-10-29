# configuracao-canais Specification

## Purpose
Permite ao Operador trocar, pela API interna, o Provedor e as credenciais de um
Canal, ou desligá-lo, com efeito sem reiniciar o OminiCore, sem nunca expor os
segredos e registrando quem fez cada troca.

## Requirements

### Requirement: REQ-22 Configurar o Canal WhatsApp pela API
Quando o Operador enviar a configuração do Canal WhatsApp com o Provedor (Meta,
UltraMsg ou Z-API) e todos os campos exigidos por ele, o sistema SHALL
gravá-la no OpenBao no lugar da anterior e ligar o Canal com esse Provedor; com
Provedor desconhecido ou campo exigido faltando, SHALL recusar citando o
Provedor ou o campo, sem gravar nada.
Origem: RN-11, RN-4

#### Scenario: T-92 Operador grava a Z-API
- **WHEN** o Operador envia a configuração com o Provedor Z-API, a instância, o token e o token de segurança da conta
- **THEN** a resposta confirma a gravação, o OpenBao guarda o Provedor Z-API com esses campos e a consulta mostra a Z-API como Provedor ativo

#### Scenario: T-93 Campo exigido faltando é recusado
- **WHEN** o Operador envia a configuração com o Provedor UltraMsg sem o token
- **THEN** a resposta é 422 citando o campo token e a configuração anterior continua no OpenBao

#### Scenario: T-94 Provedor desconhecido é recusado
- **WHEN** o Operador envia a configuração com um Provedor que não é Meta, UltraMsg nem Z-API
- **THEN** a resposta é 422 citando o Provedor e a configuração anterior continua no OpenBao

#### Scenario: T-95 Meta exige e grava os segredos do webhook
- **WHEN** o Operador envia a configuração com o Provedor Meta, com o token, o identificador do número, o App Secret e o verify token
- **THEN** os quatro ficam gravados no caminho do WhatsApp no OpenBao, e sem o App Secret ou o verify token a resposta seria 422 citando o campo

### Requirement: REQ-23 Consultar a configuração sem segredos
Quando o Operador consultar a configuração do Canal WhatsApp, o sistema SHALL
devolver se o Canal está ligado, o Provedor ativo e os campos que não são
segredo, e, para cada segredo do Provedor, só se está configurado; o sistema
MUST NOT devolver o valor nem parte de nenhum segredo.
Origem: RN-11, RN-9

#### Scenario: T-96 Consulta mostra o Provedor sem os segredos
- **WHEN** a Z-API está configurada com instância `inst1`, token e token de segurança da conta, e o Operador consulta a configuração
- **THEN** a resposta traz o Canal ligado, o Provedor Z-API, a instância `inst1` e o token e o token de segurança apenas como configurados, sem nenhum caractere dos seus valores

#### Scenario: T-97 Consulta com o Canal desligado
- **WHEN** não há configuração do WhatsApp no OpenBao e o Operador consulta a configuração
- **THEN** a resposta traz o Canal desligado, sem Provedor

### Requirement: REQ-24 Desligar o Canal WhatsApp pela API
Quando o Operador desligar o Canal WhatsApp, o sistema SHALL apagar a
configuração dele no OpenBao, e o Canal SHALL passar a ser tratado como
desligado (REQ-20).
Origem: RN-11, RN-4

#### Scenario: T-98 Operador desliga o WhatsApp
- **WHEN** o Operador desliga o Canal WhatsApp e, depois da releitura, um Sistema de origem pede uma Notificação só pelo WhatsApp
- **THEN** a configuração não existe mais no OpenBao, a consulta mostra o Canal desligado e o pedido é recusado com 422 "canal whatsapp indisponível"

### Requirement: REQ-25 Só o Operador configura
Se uma requisição de consulta ou alteração da configuração vier sem token
válido, o sistema SHALL recusá-la como não autenticada; se o token for válido
sem o papel de configuração, SHALL recusá-la por falta de permissão; nos dois
casos, nada muda.
Origem: RN-11, RN-1

#### Scenario: T-99 Sistema de origem sem o papel é recusado
- **WHEN** um Sistema de origem com token válido, sem o papel `ominicore-config`, consulta, grava ou desliga a configuração do WhatsApp
- **THEN** a resposta é 403, a configuração no OpenBao não muda e nada é registrado em `config_eventos`

#### Scenario: T-100 Sem token a configuração é recusada
- **WHEN** a consulta ou a alteração da configuração chega sem token ou com token inválido
- **THEN** a resposta é 401 e nada muda

### Requirement: REQ-26 Configuração vale sem reiniciar
Enquanto o OminiCore estiver rodando, o sistema SHALL reler a configuração do
Canal WhatsApp no OpenBao periodicamente e passar a usar o Provedor, as
credenciais e o estado ligado ou desligado lidos, no pedido, no envio e nos
webhooks, sem reiniciar; se a configuração lida for inválida, SHALL manter a
anterior e registrar o erro em log, sem a credencial.
Origem: RN-11, RN-4

#### Scenario: T-101 Troca do UltraMsg pela Z-API sem reinício
- **WHEN** o UltraMsg está ativo, o Operador grava a Z-API e, depois da releitura, uma Notificação com WhatsApp é processada
- **THEN** a Tentativa sai pela Z-API, o UltraMsg não é chamado e nenhum processo foi reiniciado

#### Scenario: T-102 Configuração inválida na releitura mantém a anterior
- **WHEN** a Z-API está ativa e a configuração no OpenBao é trocada por fora da API por uma sem o token
- **THEN** depois da releitura a Z-API continua ativa com as credenciais anteriores e o log registra o erro citando o campo faltante, sem nenhuma credencial

#### Scenario: T-103 Rotas da Meta acompanham o Provedor ativo
- **WHEN** a Z-API está ativa e o Operador grava a Meta, com os segredos do webhook
- **THEN** depois da releitura a verificação em `/webhooks/whatsapp` com o verify token gravado é aceita, e antes dela respondia 404

### Requirement: REQ-27 Registro das trocas de configuração
Quando a configuração de um Canal for gravada ou apagada pela API, o sistema
SHALL registrar o Canal, a ação, o Provedor, o Operador e o horário, e MUST
NOT registrar nenhum segredo. Quando a troca vier do Console, o Operador
registrado SHALL ser o nome de usuário da pessoa no Keycloak; vinda de outro
client com o papel, SHALL ser o client.
Origem: RN-11, RN-12

#### Scenario: T-104 Troca fica registrada sem segredo
- **WHEN** o Operador `ops-a` grava a Z-API e depois desliga o Canal WhatsApp
- **THEN** existem dois registros, gravado com Z-API e apagado, ambos com o Operador `ops-a` e o horário, e nenhum registro nem linha de log contém o token ou o token de segurança

#### Scenario: T-120 Troca pelo Console registra a pessoa
- **WHEN** a pessoa `ana`, entrando pelo Console, grava a Z-API
- **THEN** o registro da troca traz o Operador `ana`, e não o client `ominicore-console`

### Requirement: REQ-40 Cadastrar os webhooks ao gravar a Z-API
Quando o Operador gravar a Z-API, o sistema SHALL gerar um segredo novo, que
invalida o anterior, guardá-lo no OpenBao com a configuração e cadastrar na
Z-API as URLs de "Ao enviar" e de status com esse segredo; se o cadastro falhar,
SHALL manter a gravação e informar a falha na resposta e na consulta, que
mostram o estado do cadastro e nunca o segredo.
Origem: RN-5, RN-11

#### Scenario: T-131 Gravar a Z-API cadastra os webhooks com segredo novo
- **WHEN** o Operador grava a Z-API e a Z-API aceita o cadastro das duas URLs
- **THEN** as duas URLs cadastradas levam o mesmo segredo novo, o segredo anterior passa a receber 404 e a consulta mostra os webhooks cadastrados

#### Scenario: T-132 Falha no cadastro não desfaz a gravação
- **WHEN** o Operador grava a Z-API e a chamada de cadastro na Z-API falha
- **THEN** a Z-API fica gravada e ativa, a resposta e a consulta mostram os webhooks como não cadastrados, e o log registra a falha sem o segredo nem o token

#### Scenario: T-133 Consulta não mostra o segredo do webhook
- **WHEN** a Z-API está gravada com webhooks cadastrados e o Operador consulta a configuração
- **THEN** a resposta mostra o segredo do webhook apenas como configurado e o estado do cadastro, sem nenhum caractere do segredo

### Requirement: REQ-41 Recadastrar os webhooks da Z-API
Quando o Operador pedir o recadastro dos webhooks, o sistema SHALL cadastrar de
novo na Z-API as duas URLs com o segredo atual, sem trocá-lo, e atualizar o
estado do cadastro; sem a Z-API ativa, SHALL recusar o pedido sem efeito.
Origem: RN-5, RN-11

#### Scenario: T-134 Recadastro com o segredo atual
- **WHEN** os webhooks estão como não cadastrados e o Operador pede o recadastro, e a Z-API aceita
- **THEN** as URLs cadastradas levam o segredo atual, que continua valendo, e a consulta mostra os webhooks cadastrados

#### Scenario: T-135 Recadastro sem a Z-API ativa é recusado
- **WHEN** o UltraMsg está ativo e o Operador pede o recadastro dos webhooks
- **THEN** o pedido é recusado citando que só vale para a Z-API e nada é chamado
