## Purpose
Define a que Organização (empresa ou marca do grupo) pertencem cada Sistema de
origem e cada Notificação, e garante que nenhuma Organização enxerga os dados
de outra.

## ADDED Requirements

### Requirement: REQ-50 Sistema de origem pertence a uma Organização
Se uma requisição para pedir ou consultar Notificação, ou para registrar ou
remover Opt-out, vier com token válido de um client que não pertence a nenhuma
Organização, o sistema SHALL recusá-la por falta de permissão, sem registrar
nem despachar nada.
Origem: RN-15, RN-1

#### Scenario: T-159 Client sem Organização é recusado
- **WHEN** um token válido do client `sistema-x`, sem Organização, pede uma Notificação, consulta uma Notificação, registra um Opt-out ou remove um Opt-out
- **THEN** a resposta é 403 "Sistema de origem sem Organização", nenhuma Notificação nem Opt-out é criado ou removido e nada é despachado

#### Scenario: T-160 Client com Organização segue como antes
- **WHEN** o client `wbem`, da Organização `wbem`, pede uma Notificação válida
- **THEN** o pedido é aceito com o mesmo contrato de hoje e a Notificação é registrada na Organização `wbem`

### Requirement: REQ-51 Notificação fica com a Organização de quem pediu
Quando um Sistema de origem tiver um pedido de Notificação aceito, o sistema
SHALL registrar a Notificação na Organização desse Sistema de origem; e, ao
passar a ter Organizações, SHALL colocar na Organização `wbem` toda Notificação
e todo Template que já existiam, sem mudar status, Tentativas nem conteúdo.
Origem: RN-15

#### Scenario: T-161 Dados existentes vão para a Organização wbem
- **WHEN** a base tem Notificações, Tentativas e Templates gravados antes das Organizações e passa a ter Organizações
- **THEN** todas as Notificações e Templates ficam na Organização `wbem`, com status, Tentativas e conteúdo iguais aos de antes, e o client `wbem` pertence à Organização `wbem`

#### Scenario: T-162 Voltar atrás não perde dados
- **WHEN** a passagem para Organizações é desfeita
- **THEN** Notificações, Tentativas e Templates continuam iguais aos de antes da passagem

### Requirement: REQ-52 Uma Organização não enxerga a outra
Quando a API ler Notificações para um Sistema de origem, o sistema SHALL
devolver só Notificações da Organização dele, mesmo que a busca não filtre
pelo Sistema de origem; o despacho, a retenção e os Avisos de status SHALL
continuar tratando Notificações de todas as Organizações.
Origem: RN-15, RN-5

#### Scenario: T-163 Barreira do banco entre Organizações
- **WHEN** existem Notificações das Organizações `wbem` e `org-b` e uma leitura da API, fixada na Organização `org-b`, busca uma Notificação da `wbem` sem filtrar pelo Sistema de origem
- **THEN** a leitura não encontra a Notificação da `wbem`

#### Scenario: T-164 Consulta de outra Organização é inexistente
- **WHEN** o client `sistema-b`, da Organização `org-b`, consulta uma Notificação pedida pelo client `wbem`
- **THEN** a resposta é "não encontrada", como para Notificação de outro Sistema de origem

#### Scenario: T-165 Despacho e Avisos seguem para todas as Organizações
- **WHEN** há Notificações pendentes das Organizações `wbem` e `org-b` e chega um Aviso de entrega para uma Tentativa da `org-b`
- **THEN** o worker despacha as Notificações das duas Organizações e o Aviso marca a Tentativa da `org-b` como entregue
