## MODIFIED Requirements

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
