## ADDED Requirements

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
