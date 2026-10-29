## ADDED Requirements

### Requirement: REQ-43 API do Ambiente sem resposta
Se a API do Ambiente não responder ou responder erro do servidor, o Console
SHALL mostrar que a API daquele Ambiente não respondeu, com o status quando
houver, e um botão "Tentar de novo" que repete a chamada, e MUST NOT ficar
carregando indefinidamente.
Origem: RN-12

#### Scenario: T-137 API fora do ar e nova tentativa
- **WHEN** o Operador abre a tela do WhatsApp em Produção, a chamada à API falha por rede e, ao clicar em "Tentar de novo", a API responde com a configuração
- **THEN** o Console mostra primeiro que a API de Produção não respondeu, sem barra de carregamento, e depois a configuração

#### Scenario: T-138 Erro do servidor mostra o status
- **WHEN** a consulta da configuração responde 503
- **THEN** o Console mostra que a API não respondeu como esperado, citando o status 503, e o botão "Tentar de novo"
