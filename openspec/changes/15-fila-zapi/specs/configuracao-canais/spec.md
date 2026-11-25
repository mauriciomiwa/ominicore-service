## ADDED Requirements

### Requirement: REQ-45 Consultar a conexão e a Fila do Provedor
Quando o Operador consultar a Fila com a Z-API ativa, o sistema SHALL devolver
a conexão da instância (conectado, desconectado ou celular sem internet) e até
30 itens por página, com cursor, cada um com id, telefone mascarado, entrada e,
se for de Tentativa, a Notificação, o Template e o Sistema de origem; MUST NOT
devolver o texto nem o telefone completo; sem a Z-API ativa, SHALL recusar sem
chamá-la.
Origem: RN-12, RN-11

#### Scenario: T-141 Fila com item do OminiCore e item de fora
- **WHEN** a Z-API está ativa e desconectada, e a fila dela tem um item com o `messageId` de uma Tentativa do Template `wbem-otp-tomador` do Sistema de origem `wbem` e outro item sem Tentativa correspondente
- **THEN** a resposta traz a conexão "desconectado" e os dois itens com o telefone mascarado (só o DDI, o DDD e os 4 últimos dígitos) e a entrada; o primeiro com a Notificação, o Template e o Sistema de origem, o segundo como fora do OminiCore; e nenhum campo traz o texto do item

#### Scenario: T-142 Próxima página pelo cursor
- **WHEN** a fila da Z-API tem 35 itens e o Operador consulta e depois consulta de novo com o cursor devolvido
- **THEN** a primeira resposta traz 30 itens, o cursor e a indicação de que há mais; a segunda traz os 5 restantes e a indicação de que não há mais

#### Scenario: T-143 Sem a Z-API ativa a consulta é recusada
- **WHEN** o UltraMsg está ativo e o Operador consulta a Fila
- **THEN** o pedido é recusado citando que só vale para a Z-API e a Z-API não é chamada

#### Scenario: T-144 Z-API sem resposta ou recusando a credencial
- **WHEN** a Z-API está ativa e responde 401 à consulta da conexão, ou não responde
- **THEN** a resposta é um erro que diz que a Z-API não respondeu como esperado, com o status dela quando houver, e não traz o token nem a URL da Z-API

#### Scenario: T-145 Fila só para o Operador
- **WHEN** um Sistema de origem sem o papel `ominicore-config` consulta a Fila ou pede uma retirada, ou o pedido chega sem token
- **THEN** a resposta é 403 sem o papel e 401 sem token, a Z-API não é chamada e nada é registrado

### Requirement: REQ-46 Retirar um item da Fila do Provedor
Quando o Operador pedir a retirada de um item da Fila, o sistema SHALL pedir a
remoção à Z-API e, só se ela confirmar, deixar a Tentativa do item falhou com a
classificação removida pelo Operador, seguindo o fallback da RN-5, e registrar
o Operador e o id do item; se o item não estiver mais na fila, SHALL responder
que já saiu, sem mudar nada.
Origem: RN-12, RN-5

#### Scenario: T-146 Retirar item de Tentativa com próximo Canal
- **WHEN** a Notificação tem a Ordem de Canais WhatsApp → SMS, a Tentativa de WhatsApp está enviada e na fila da Z-API, e o Operador `ana` retira o item, que a Z-API confirma
- **THEN** a Tentativa de WhatsApp fica falhou com a classificação removida pelo Operador, uma Tentativa de SMS é criada e o registro traz a retirada com o Operador `ana`, o id do item e o horário

#### Scenario: T-147 Item que já saiu da fila
- **WHEN** o Operador pede a retirada de um item que não está mais na fila da Z-API
- **THEN** a resposta diz que o item já saiu da fila, a Tentativa segue como estava e nada é registrado

#### Scenario: T-148 Item fora do OminiCore
- **WHEN** o Operador retira um item sem Tentativa correspondente e a Z-API confirma
- **THEN** o item sai da fila, nenhuma Tentativa muda e o registro traz a retirada com o Operador e o id do item

#### Scenario: T-149 Aviso de entrega depois da retirada é ignorado
- **WHEN** uma Tentativa foi retirada da fila e depois chega da Z-API um Aviso de status de entrega com o mesmo `messageId`
- **THEN** o aviso é aceito e ignorado, e a Tentativa segue falhou com a classificação removida pelo Operador
