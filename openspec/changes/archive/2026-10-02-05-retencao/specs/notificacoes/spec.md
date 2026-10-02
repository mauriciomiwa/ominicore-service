## MODIFIED Requirements

### Requirement: REQ-14 Consultar Notificação
Quando o Sistema de origem autenticado consultar uma Notificação que ele pediu,
o sistema SHALL devolver o status dela, o horário da anonimização quando já
tiver sido anonimizada, e de cada Tentativa (Canal, número, status, horários e
a classificação da falha: contato inválido, recusado pelo Provedor,
Retentativas esgotadas, Opt-out), MUST NOT devolver contato do Destinatário nem
o texto do motivo, e SHALL responder como inexistente a Notificação de outro
Sistema de origem.
Origem: RN-5, RN-9, RN-8

#### Scenario: T-43 Dono consulta status e Tentativas
- **WHEN** o Sistema de origem `sistema-a` consulta a Notificação que pediu, entregue pelo SMS depois de bounce no Email
- **THEN** a resposta traz a Notificação entregue, a Tentativa de Email falhou com classificação "contato inválido" e a de SMS entregue, sem e-mail, telefone nem texto do motivo

#### Scenario: T-44 Outro Sistema de origem não enxerga
- **WHEN** o Sistema de origem `sistema-b` consulta uma Notificação de `sistema-a`
- **THEN** a resposta é "não encontrada"

#### Scenario: T-56 Consulta mostra quando foi anonimizada
- **WHEN** o dono consulta uma Notificação já anonimizada
- **THEN** a resposta traz o status, as Tentativas e o horário da anonimização

## ADDED Requirements

### Requirement: REQ-19 Anonimizar depois do prazo de retenção
Quando uma Notificação finalizada (entregue, falhou ou enviada sem Tentativa
pendente) passar do prazo de retenção (padrão 30 dias, configurável), o sistema
SHALL apagar o contato do Destinatário, os dados das variáveis e o motivo de
cada Tentativa, SHALL manter status, Canais, classificação da falha e horários,
e SHALL registrar quando anonimizou; Notificação com Tentativa pendente MUST NOT
ser anonimizada. Anonimizar de novo não muda nada, e o Opt-out não é afetado.
Origem: RN-9

#### Scenario: T-57 Notificação finalizada além do prazo é anonimizada
- **WHEN** uma Notificação entregue foi criada há 31 dias, com e-mail, telefone, `nome` e motivo de falha numa Tentativa
- **THEN** depois da rodada, e-mail, telefone, dados e motivo estão vazios, e status, Canais, classificação da falha e horários continuam iguais

#### Scenario: T-58 Notificação dentro do prazo não muda
- **WHEN** uma Notificação entregue foi criada há 29 dias
- **THEN** a rodada não altera nada nela

#### Scenario: T-59 Notificação pendente não é anonimizada
- **WHEN** uma Notificação criada há 31 dias ainda tem Tentativa pendente
- **THEN** a rodada não a anonimiza

#### Scenario: T-60 Prazo configurável
- **WHEN** o prazo configurado é de 7 dias e uma Notificação finalizada foi criada há 8 dias
- **THEN** a rodada a anonimiza

#### Scenario: T-61 Rodada repetida é idempotente
- **WHEN** a rodada roda duas vezes seguidas
- **THEN** o horário da anonimização não muda e nada mais é alterado

#### Scenario: T-62 Opt-out sobrevive à anonimização
- **WHEN** o Destinatário fez Opt-out de SMS e a Notificação dele é anonimizada
- **THEN** um pedido novo para o mesmo telefone continua pulando o Canal SMS
