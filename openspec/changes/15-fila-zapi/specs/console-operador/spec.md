## ADDED Requirements

### Requirement: REQ-47 Conexão e Fila do Provedor no Console
Enquanto a Z-API for o Provedor ativo, o Console SHALL mostrar a conexão e a
Fila do Provedor ao abrir a tela do WhatsApp e no "Atualizar", com telefone
mascarado, tempo na fila e Template, Sistema de origem e Notificação ou "fora
do OminiCore", e "Carregar mais" enquanto houver itens; MUST NOT mostrar o
texto do item nem consultar sozinho; sem resposta da Z-API, SHALL mostrar o
erro com o status e "Tentar de novo".
Origem: RN-12

#### Scenario: T-150 Conexão e itens da Fila
- **WHEN** a consulta da configuração devolve a Z-API ativa, a Fila devolve a conexão "desconectado", um item de Tentativa e um item fora do OminiCore, e o Operador abre a tela do WhatsApp
- **THEN** o Console mostra a instância desconectada, os dois itens com o telefone mascarado e o tempo na fila, o primeiro com o Template, o Sistema de origem e a Notificação, o segundo como "fora do OminiCore", e nenhuma nova consulta à Fila sai sem o Operador clicar em "Atualizar"

#### Scenario: T-151 Carregar mais
- **WHEN** a Fila devolve 30 itens com a indicação de que há mais e o Operador clica em "Carregar mais"
- **THEN** o Console consulta com o cursor devolvido, acrescenta os novos itens abaixo dos anteriores e esconde "Carregar mais" quando não há mais

#### Scenario: T-152 Z-API sem resposta e nova tentativa
- **WHEN** a consulta da Fila responde que a Z-API não respondeu como esperado, com o status 401, e, ao clicar em "Tentar de novo", responde com a Fila
- **THEN** o Console mostra primeiro o erro citando o status 401 e depois a conexão e os itens

#### Scenario: T-153 Com outro Provedor a Fila não aparece
- **WHEN** a consulta da configuração devolve o UltraMsg ativo
- **THEN** a tela do WhatsApp não mostra a conexão nem a Fila e não consulta a Fila

### Requirement: REQ-48 Retirar um item da Fila pelo Console
Quando o Operador pedir a retirada de um item da Fila, o Console SHALL pedir
confirmação, citando produção em produção, e só depois pedir a retirada à API;
depois SHALL mostrar o resultado e recarregar a Fila, avisando quando o item
já tiver saído.
Origem: RN-12

#### Scenario: T-154 Retirar pede confirmação
- **WHEN** o Operador em produção pede a retirada de um item e cancela, e depois pede de novo e confirma
- **THEN** a pergunta cita produção, ao cancelar nenhuma chamada é feita e, ao confirmar, o Console pede a retirada, mostra a confirmação e recarrega a Fila sem o item

#### Scenario: T-155 Item que já saiu
- **WHEN** o Operador confirma a retirada e a API responde que o item já saiu da fila
- **THEN** o Console avisa que a mensagem já saiu da fila e recarrega a Fila
