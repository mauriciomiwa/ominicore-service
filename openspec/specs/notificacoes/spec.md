# notificacoes Specification

## Purpose
Pedido, registro e envio de Notificações a partir de Templates, com Chave de
idempotência para que o reenvio de um pedido nunca faça o Destinatário receber
duas vezes, e com cada Tentativa registrada com Canal e resultado.

## Requirements

### Requirement: REQ-2 Pedir Notificação por Template
Quando um Sistema de origem autenticado pedir uma Notificação com um Template
existente, o Destinatário e todas as variáveis dele, o sistema SHALL
registrá-la como pendente, com a Ordem de Canais do pedido (ou a do Template) e
a primeira Tentativa no primeiro Canal da ordem com contato do Destinatário, e
responder com o identificador; SHALL recusar o pedido com Canal da
ordem sem versão no Template, sem contato do Destinatário para a ordem ou com
telefone fora do E.164.
Origem: RN-1, RN-2

#### Scenario: T-4 Pedido válido é aceito
- **WHEN** o Sistema de origem pede uma Notificação com Template `boas-vindas`, e-mail do Destinatário e a variável `nome`
- **THEN** a resposta traz o identificador e a Notificação fica pendente com uma Tentativa no Canal Email

#### Scenario: T-5 Template inexistente é recusado
- **WHEN** o pedido informa um Template que não existe
- **THEN** o pedido é recusado com o motivo "template inexistente" e nada é registrado

#### Scenario: T-6 Variável faltando é recusada
- **WHEN** o pedido omite uma variável exigida pelo Template
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

#### Scenario: T-15 Ordem do pedido substitui a do Template
- **WHEN** o pedido informa `canais` SMS → Email e o Destinatário tem telefone e e-mail
- **THEN** a Notificação guarda a ordem SMS → Email e a primeira Tentativa é no Canal SMS

#### Scenario: T-16 Canal pedido sem versão no Template é recusado
- **WHEN** o pedido informa um Canal que o Template não tem
- **THEN** o pedido é recusado citando o Canal e nada é registrado

#### Scenario: T-17 Destinatário sem contato para nenhum Canal é recusado
- **WHEN** a ordem é só SMS e o Destinatário tem apenas e-mail
- **THEN** o pedido é recusado com o motivo "destinatário sem contato para os canais" e nada é registrado

#### Scenario: T-18 Telefone fora do E.164 é recusado
- **WHEN** o Destinatário informa o telefone `11 99999-8888`
- **THEN** o pedido é recusado citando o formato E.164 e nada é registrado

### Requirement: REQ-3 Pedido idempotente
Quando um Sistema de origem repetir a mesma Chave de idempotência em até 24 h,
o sistema SHALL devolver a Notificação já criada sem registrar nova Notificação
nem nova Tentativa; pedido sem Chave de idempotência SHALL ser recusado.
Origem: RN-7

#### Scenario: T-7 Reenvio com a mesma chave não duplica
- **WHEN** o mesmo Sistema de origem envia duas vezes o pedido com a chave `abc-1` dentro de 24 h
- **THEN** as duas respostas trazem o mesmo identificador e existe uma única Notificação com uma única Tentativa

#### Scenario: T-8 Pedido sem chave é recusado
- **WHEN** o pedido chega sem Chave de idempotência
- **THEN** o pedido é recusado com o motivo "chave de idempotência obrigatória" e nada é registrado

#### Scenario: T-9 A chave vale por Sistema de origem
- **WHEN** dois Sistemas de origem diferentes usam a mesma chave `abc-1`
- **THEN** são registradas duas Notificações distintas

### Requirement: REQ-4 Enviar pelo Canal Email
Quando houver uma Tentativa pendente no Canal Email, o sistema SHALL enviá-la
ao Provedor de Email com o conteúdo do Template preenchido e registrar a
Tentativa como enviada.
Origem: RN-3

#### Scenario: T-10 Tentativa de Email é enviada
- **WHEN** uma Notificação pendente no Canal Email é processada e o Provedor aceita
- **THEN** o Provedor recebe o assunto e o corpo com as variáveis preenchidas e a Tentativa fica enviada

#### Scenario: T-11 Primeira Tentativa em até 5 segundos
- **WHEN** um pedido válido é aceito com o sistema ocioso
- **THEN** a primeira Tentativa chega ao Provedor em até 5 segundos após a resposta do pedido

### Requirement: REQ-5 Registrar falha do Provedor
Se o Provedor recusar a Tentativa, o sistema SHALL registrar a Tentativa como
falhou com o motivo informado e, não havendo outro Canal na Ordem de Canais,
SHALL marcar a Notificação como falhou.
Origem: RN-2, RN-3

#### Scenario: T-12 Recusa do Provedor encerra a Notificação sem outro Canal
- **WHEN** o Provedor de Email recusa a Tentativa por endereço inválido e a Ordem de Canais só tem Email
- **THEN** a Tentativa fica falhou com o motivo do Provedor e a Notificação fica falhou

### Requirement: REQ-7 Enviar pelo Canal SMS
Quando houver uma Tentativa pendente no Canal SMS, o sistema SHALL enviá-la ao
Provedor de SMS com o corpo da versão SMS do Template preenchido e registrar o
resultado, tratando recusa por número inválido como Falha definitiva.
Origem: RN-3

#### Scenario: T-19 Tentativa de SMS é enviada
- **WHEN** uma Notificação pendente no Canal SMS é processada e o Provedor aceita
- **THEN** o Provedor recebe o telefone do Destinatário e o corpo com as variáveis preenchidas, e a Tentativa fica enviada

#### Scenario: T-20 Número inválido é Falha definitiva
- **WHEN** o Provedor de SMS recusa o número do Destinatário como inválido
- **THEN** a Tentativa fica falhou com o motivo do Provedor, sem Retentativa

### Requirement: REQ-8 Retentar Falha temporária
Se uma Tentativa terminar em Falha temporária, o sistema SHALL agendar uma
Retentativa no mesmo Canal após 5 s, 30 s e 2 min (±20%), até 3 Retentativas;
esgotadas, SHALL tratar o Canal como falho.
Origem: RN-2

#### Scenario: T-21 Falha temporária agenda Retentativa no mesmo Canal
- **WHEN** o Provedor responde indisponível na primeira Tentativa por Email
- **THEN** existe uma segunda Tentativa pendente no Canal Email, disponível entre 4 s e 6 s depois, e a Notificação segue pendente

#### Scenario: T-22 Espera cresce a cada Retentativa
- **WHEN** as Tentativas 1, 2 e 3 de um Canal terminam em Falha temporária
- **THEN** as Retentativas ficam disponíveis após 5 s, 30 s e 2 min, cada uma dentro de ±20%

#### Scenario: T-23 Retentativas esgotadas encerram o Canal
- **WHEN** as 4 Tentativas de um Canal terminam em Falha temporária
- **THEN** não há quinta Tentativa nesse Canal e a Notificação segue a Ordem de Canais

### Requirement: REQ-9 Fallback pela Ordem de Canais
Quando um Canal falhar (Falha definitiva ou Retentativas esgotadas), o sistema
SHALL criar a Tentativa no próximo Canal da Ordem de Canais para o qual o
Destinatário tem contato; sem Canal restante, SHALL marcar a Notificação como
falhou; com sucesso em qualquer Canal, SHALL marcá-la como enviada e parar.
Origem: RN-2

#### Scenario: T-24 Falha definitiva no Email passa ao SMS
- **WHEN** a ordem é Email → SMS e o Provedor de Email recusa o endereço
- **THEN** a Tentativa seguinte é no Canal SMS e, aceita, a Notificação fica enviada

#### Scenario: T-25 Canal sem contato é pulado
- **WHEN** a ordem é Email → SMS, o Email falha de forma definitiva e o Destinatário não tem telefone
- **THEN** não é criada Tentativa no Canal SMS e a Notificação fica falhou

#### Scenario: T-26 Todos os Canais falham
- **WHEN** a ordem é Email → SMS e os dois Provedores recusam de forma definitiva
- **THEN** a Notificação fica falhou com uma Tentativa falhou por Canal

#### Scenario: T-27 Sucesso encerra o fallback
- **WHEN** a ordem é Email → SMS e o Email é aceito
- **THEN** a Notificação fica enviada e nenhuma Tentativa é criada no Canal SMS

### Requirement: REQ-10 Enviar pelo Canal WhatsApp
Tentativa pendente de WhatsApp SHALL ir ao Provedor ativo, para o telefone do
Destinatário. Meta: o Modelo aprovado (idioma e parâmetros em ordem); sem
WhatsApp ou Modelo não aprovado é Falha definitiva, e limite de taxa ou erro
genérico, temporária. UltraMsg e Z-API: o texto preenchido; limite de taxa ou
indisponibilidade é temporária, e outra recusa, definitiva. Texto com mais de
4096 caracteres ou versão sem a forma do Provedor SHALL ser Falha definitiva,
sem chamada.
Origem: RN-4

#### Scenario: T-28 Tentativa de WhatsApp é enviada pelo Modelo aprovado
- **WHEN** com a Meta ativa, uma Notificação `boas-vindas` com `nome` = Ana é processada no Canal WhatsApp e o Provedor aceita
- **THEN** o Provedor recebe o Modelo `boas_vindas`, idioma `pt_BR`, parâmetro 1 = Ana e o telefone do Destinatário, e a Tentativa fica enviada

#### Scenario: T-29 Número sem WhatsApp passa ao próximo Canal
- **WHEN** com a Meta ativa, a ordem é WhatsApp → SMS e o Provedor de WhatsApp recusa com o código 131026
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa e a próxima Tentativa é no Canal SMS

#### Scenario: T-30 Limite de taxa é Falha temporária
- **WHEN** com a Meta ativa, o Provedor de WhatsApp recusa com o código 130429
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-31 Modelo não aprovado é Falha definitiva
- **WHEN** com a Meta ativa, o Provedor de WhatsApp recusa com o código 132001
- **THEN** a Tentativa fica falhou com o motivo do Provedor, sem Retentativa

#### Scenario: T-73 Tentativa de WhatsApp é enviada como texto pelo UltraMsg
- **WHEN** com o UltraMsg ativo, uma Notificação `boas-vindas` com `nome` = Ana é processada no Canal WhatsApp e o Provedor aceita devolvendo um id
- **THEN** o Provedor recebe o texto "Olá Ana, sua conta está pronta." para o telefone do Destinatário, e a Tentativa fica enviada com esse id

#### Scenario: T-74 Recusa do UltraMsg passa ao próximo Canal
- **WHEN** com o UltraMsg ativo, a ordem é WhatsApp → SMS e o Provedor recusa o token ou a instância
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa, com o motivo do Provedor, e a próxima Tentativa é no Canal SMS

#### Scenario: T-75 UltraMsg indisponível é Falha temporária
- **WHEN** com o UltraMsg ativo, o Provedor responde com limite de taxa, erro de servidor ou não responde
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-76 Texto longo demais não chega ao UltraMsg
- **WHEN** com o UltraMsg ativo, o texto da versão WhatsApp preenchido passa de 4096 caracteres e a ordem é WhatsApp → SMS
- **THEN** o Provedor não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

#### Scenario: T-77 Tentativa pendente sem a forma do Provedor ativo segue a ordem
- **WHEN** uma Tentativa de WhatsApp ficou pendente, o Provedor ativo passou a ser um cuja forma a versão WhatsApp do Template não tem, e a ordem é WhatsApp → SMS
- **THEN** o Provedor não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

#### Scenario: T-85 Tentativa de WhatsApp é enviada como texto pela Z-API
- **WHEN** com a Z-API ativa e sem token de segurança da conta, uma Notificação `boas-vindas` com `nome` = Ana e telefone +5511999990000 é processada no Canal WhatsApp e o Provedor aceita devolvendo um `messageId`
- **THEN** o Provedor recebe o texto "Olá Ana, sua conta está pronta." para o telefone 5511999990000, sem o cabeçalho de token de segurança, e a Tentativa fica enviada com esse `messageId`

#### Scenario: T-86 Token de segurança da conta vai em toda chamada à Z-API
- **WHEN** com a Z-API ativa e o token de segurança da conta configurado, uma Tentativa de WhatsApp é enviada
- **THEN** a chamada ao Provedor leva o token de segurança da conta no cabeçalho `Client-Token`

#### Scenario: T-87 Recusa da Z-API passa ao próximo Canal
- **WHEN** com a Z-API ativa, a ordem é WhatsApp → SMS e o Provedor recusa o token, a instância ou o token de segurança da conta
- **THEN** a Tentativa de WhatsApp fica falhou sem Retentativa, com o motivo do Provedor, e a próxima Tentativa é no Canal SMS

#### Scenario: T-88 Z-API indisponível é Falha temporária
- **WHEN** com a Z-API ativa, o Provedor responde com limite de taxa, erro de servidor ou não responde
- **THEN** existe uma Retentativa pendente no Canal WhatsApp

#### Scenario: T-89 Texto longo demais não chega à Z-API
- **WHEN** com a Z-API ativa, o texto da versão WhatsApp preenchido passa de 4096 caracteres e a ordem é WhatsApp → SMS
- **THEN** o Provedor não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

### Requirement: REQ-11 Versão WhatsApp do Template
A versão WhatsApp do Template SHALL ser um Modelo aprovado (nome, idioma e
parâmetros ordenados, cada um uma variável), um texto com variáveis, ou os
dois, e o pedido com WhatsApp na Ordem de Canais SHALL usar o telefone do
Destinatário. Com o Canal ligado e na ordem, versão sem a forma do Provedor
ativo SHALL ser recusada com "versão whatsapp sem texto" (UltraMsg ou Z-API)
ou "versão whatsapp sem modelo aprovado" (Meta), sem registrar nada.
Origem: RN-4

#### Scenario: T-32 Pedido com WhatsApp na ordem é aceito
- **WHEN** o pedido informa `canais` WhatsApp → SMS e o Destinatário tem telefone
- **THEN** a Notificação é registrada com a primeira Tentativa no Canal WhatsApp

#### Scenario: T-33 Parâmetro do Modelo aprovado é variável exigida
- **WHEN** o pedido com Canal WhatsApp omite a variável `nome`, usada como parâmetro 1 do Modelo aprovado
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

#### Scenario: T-78 Versão sem texto é recusada com o UltraMsg ativo
- **WHEN** com o UltraMsg ativo, o pedido informa `canais` WhatsApp → SMS para um Template cuja versão WhatsApp só tem Modelo aprovado
- **THEN** o pedido é recusado com o motivo "versão whatsapp sem texto" e nada é registrado

#### Scenario: T-79 Versão sem Modelo aprovado é recusada com a Meta ativa
- **WHEN** com a Meta ativa, o pedido informa `canais` WhatsApp → SMS para um Template cuja versão WhatsApp só tem texto
- **THEN** o pedido é recusado com o motivo "versão whatsapp sem modelo aprovado" e nada é registrado

#### Scenario: T-80 Variável do texto do WhatsApp é exigida
- **WHEN** com o UltraMsg ativo, o pedido com Canal WhatsApp omite a variável `nome`, usada no texto da versão WhatsApp
- **THEN** o pedido é recusado citando a variável faltante e nada é registrado

#### Scenario: T-90 Versão sem texto é recusada com a Z-API ativa
- **WHEN** com a Z-API ativa, o pedido informa `canais` WhatsApp → SMS para um Template cuja versão WhatsApp só tem Modelo aprovado
- **THEN** o pedido é recusado com o motivo "versão whatsapp sem texto" e nada é registrado

### Requirement: REQ-12 Receber Aviso de status do Provedor
Quando chegar um Aviso de status com assinatura válida do Provedor, o sistema
SHALL atualizar a Tentativa identificada pelo id do Provedor, só fazendo o
status avançar (enviada → entregue → lida; falhou é final), SHALL marcar a
Notificação como entregue quando uma Tentativa for entregue, e SHALL responder
com sucesso a todo aviso válido, inclusive repetido ou de Tentativa
desconhecida. Se a assinatura for inválida ou ausente, o sistema SHALL recusar
o aviso sem nenhum efeito.
Origem: RN-5

#### Scenario: T-34 Entrega confirmada pelo Email
- **WHEN** chega do SendGrid, com assinatura válida, o evento `delivered` de uma Tentativa enviada
- **THEN** a Tentativa e a Notificação ficam entregues

#### Scenario: T-35 Entrega confirmada pelo SMS
- **WHEN** chega do Twilio, com assinatura válida, o status `delivered` de uma Tentativa enviada
- **THEN** a Tentativa e a Notificação ficam entregues

#### Scenario: T-36 Leitura confirmada pelo WhatsApp
- **WHEN** chega da Meta, com assinatura válida, o status `read` de uma Tentativa entregue
- **THEN** a Tentativa fica lida e a Notificação segue entregue

#### Scenario: T-37 Assinatura inválida é recusada
- **WHEN** chega um aviso de qualquer Provedor com assinatura inválida ou ausente
- **THEN** o aviso é recusado como não autorizado e nenhuma Tentativa muda

#### Scenario: T-38 Status não regride
- **WHEN** depois de `read`, chega da Meta um `delivered` atrasado da mesma Tentativa
- **THEN** a Tentativa continua lida e o aviso é respondido com sucesso

#### Scenario: T-39 Aviso de Tentativa desconhecida
- **WHEN** chega um aviso válido com id do Provedor que não existe
- **THEN** o aviso é respondido com sucesso e nada muda

#### Scenario: T-40 Verificação do webhook da Meta
- **WHEN** a Meta verifica o webhook com o verify token correto
- **THEN** a resposta devolve o desafio recebido; com token errado, a verificação é recusada

### Requirement: REQ-13 Falha de entrega depois do aceite
Quando um Aviso de status válido informar que uma Tentativa aceita não foi
entregue, o sistema SHALL marcá-la como falhou (Falha definitiva) e SHALL criar
a Tentativa no próximo Canal da Ordem de Canais com contato (REQ-9); se a
Notificação já estiver entregue por outra Tentativa, o sistema SHALL apenas
registrar a falha.
Origem: RN-5, RN-2

#### Scenario: T-41 Bounce no Email passa ao SMS
- **WHEN** a ordem é Email → SMS, o Email foi aceito e chega do SendGrid o evento `bounce`
- **THEN** a Tentativa de Email fica falhou e existe uma Tentativa pendente no Canal SMS

#### Scenario: T-42 Falha tardia não reabre Notificação entregue
- **WHEN** a Notificação já está entregue e chega a falha de outra Tentativa dela
- **THEN** a Tentativa fica falhou, a Notificação segue entregue e nenhuma Tentativa nova é criada

### Requirement: REQ-14 Consultar Notificação
Quando o Sistema de origem autenticado consultar uma Notificação que pediu, o
sistema SHALL devolver o status, o horário da anonimização, se houver, e de cada
Tentativa o Canal, o número, o status, os horários e a classificação da falha
(contato inválido, recusado pelo Provedor, Retentativas esgotadas, Opt-out);
MUST NOT devolver contato do Destinatário nem o texto do motivo; e SHALL
responder como inexistente a Notificação de outro Sistema de origem.
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

### Requirement: REQ-15 Registrar e remover Opt-out pela API
Quando um Sistema de origem autenticado registrar o Opt-out de um contato em
um Canal, o sistema SHALL passar a considerar esse Canal indisponível para
esse contato em qualquer Sistema de origem, de forma idempotente; quando
remover, o sistema SHALL voltar a considerá-lo disponível e SHALL guardar qual
Sistema de origem removeu e quando.
Origem: RN-8

#### Scenario: T-45 Registrar Opt-out
- **WHEN** o `sistema-a` registra o Opt-out do telefone `+5511999998888` no Canal SMS
- **THEN** a resposta é sucesso e o Canal SMS desse telefone fica indisponível

#### Scenario: T-46 Registro repetido é idempotente
- **WHEN** o mesmo Opt-out é registrado duas vezes
- **THEN** as duas respostas são sucesso e existe um único Opt-out

#### Scenario: T-47 Remoção guarda quem removeu
- **WHEN** o `sistema-b` remove o Opt-out de SMS daquele telefone
- **THEN** o Canal SMS volta a ficar disponível e o histórico guarda `sistema-b` e o horário da remoção

#### Scenario: T-48 Pedido de Opt-out inválido é recusado
- **WHEN** o pedido traz um Canal desconhecido, um e-mail inválido ou um telefone fora do E.164
- **THEN** o pedido é recusado e nada é registrado

### Requirement: REQ-16 Opt-out pelos sinais do Provedor
Quando o Provedor avisar que o Destinatário se descadastrou (evento
`unsubscribe` ou `spamreport` do SendGrid) ou recusar o envio por STOP (código
21610 do Twilio), o sistema SHALL registrar o Opt-out daquele contato naquele
Canal; a recusa por STOP SHALL ser Falha definitiva, seguindo para o próximo
Canal.
Origem: RN-8

#### Scenario: T-49 Descadastro no SendGrid registra Opt-out de Email
- **WHEN** chega do SendGrid, com assinatura válida, o evento `unsubscribe` para `ana@exemplo.com`
- **THEN** o Canal Email de `ana@exemplo.com` fica em Opt-out

#### Scenario: T-50 STOP no Twilio registra Opt-out de SMS e segue a ordem
- **WHEN** a ordem é SMS → Email e o Twilio recusa o envio com o código 21610
- **THEN** o Canal SMS do telefone fica em Opt-out e a próxima Tentativa é no Canal Email

### Requirement: REQ-17 Canal em Opt-out é pulado
O sistema SHALL tratar Canal em Opt-out como indisponível para aquele contato:
no pedido (sem nenhum Canal disponível, SHALL recusá-lo), no fallback e no
envio, em que SHALL conferir o Opt-out antes de chamar o Provedor e, sem Canal
restante, SHALL encerrar a Notificação como falhou com a classificação Opt-out.
Origem: RN-8

#### Scenario: T-51 Opt-out de SMS faz a ordem SMS → Email sair por Email
- **WHEN** o telefone tem Opt-out de SMS e o pedido tem a ordem SMS → Email
- **THEN** a primeira Tentativa é no Canal Email

#### Scenario: T-52 Todos os Canais em Opt-out recusam o pedido
- **WHEN** o Destinatário tem Opt-out em todos os Canais da ordem
- **THEN** o pedido é recusado com o motivo "destinatário sem contato para os canais" e nada é registrado

#### Scenario: T-53 Opt-out entre o pedido e o envio
- **WHEN** a Notificação foi aceita com a ordem SMS → Email e o Opt-out de SMS é registrado antes do envio
- **THEN** o Provedor de SMS não é chamado, a Tentativa de SMS fica falhou com a classificação Opt-out e a próxima é no Canal Email

#### Scenario: T-54 Opt-out vale para todos os Sistemas de origem
- **WHEN** o `sistema-a` registrou o Opt-out de Email e o `sistema-b` pede uma Notificação só por Email para o mesmo endereço
- **THEN** o pedido do `sistema-b` é recusado

### Requirement: REQ-18 Opt-out sem contato em claro
O sistema MUST NOT guardar o contato do Opt-out em claro: SHALL guardar só um
HMAC do contato normalizado (e-mail em minúsculas e sem espaços; telefone em
E.164), com chave do OpenBao, e o registro SHALL permanecer depois da
retenção das Notificações.
Origem: RN-8, RN-9

#### Scenario: T-55 Tabela de Opt-out sem contato legível
- **WHEN** o Opt-out de `Ana@Exemplo.com` é registrado
- **THEN** nenhuma coluna da tabela de Opt-out contém o endereço, e o Opt-out vale para `ana@exemplo.com`

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

### Requirement: REQ-20 Canal desligado é indisponível
Enquanto um Canal estiver desligado, o sistema SHALL tratá-lo como
indisponível: SHALL pulá-lo na Ordem de Canais do pedido e do fallback, SHALL
recusar o pedido sem nenhum Canal disponível com o motivo "canal <canal>
indisponível" quando algum Canal pedido estiver desligado, SHALL encerrar como
Falha definitiva a Tentativa pendente nele sem chamar o Provedor e SHALL
responder como rota inexistente aos webhooks do Provedor desse Canal.
Origem: RN-4

#### Scenario: T-65 Pedido só por WhatsApp com o Canal desligado é recusado
- **WHEN** o WhatsApp está desligado e o pedido informa `canais` só WhatsApp para um Destinatário com telefone
- **THEN** o pedido é recusado com o motivo "canal whatsapp indisponível" e nada é registrado

#### Scenario: T-66 Canal desligado no começo da ordem é pulado
- **WHEN** o WhatsApp está desligado e o pedido informa `canais` WhatsApp → SMS para um Destinatário com telefone
- **THEN** o pedido é aceito com a ordem WhatsApp → SMS e a primeira Tentativa é no Canal SMS

#### Scenario: T-67 Fallback pula o Canal desligado
- **WHEN** o WhatsApp está desligado, a ordem é Email → WhatsApp → SMS e o Email falha de forma definitiva
- **THEN** não é criada Tentativa no Canal WhatsApp e a Tentativa seguinte é no Canal SMS

#### Scenario: T-68 Tentativa pendente no Canal desligado segue a ordem
- **WHEN** uma Tentativa de WhatsApp ficou pendente, o Canal é desligado antes do envio e a ordem é WhatsApp → SMS
- **THEN** o Provedor do WhatsApp não é chamado, a Tentativa de WhatsApp fica falhou como Falha definitiva e a próxima é no Canal SMS

#### Scenario: T-69 Webhook da Meta com o Canal desligado
- **WHEN** o WhatsApp está desligado e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda

### Requirement: REQ-21 Provedor do WhatsApp sem Aviso de status
Enquanto o Provedor do WhatsApp ativo for o UltraMsg, o sistema SHALL manter
como enviada a Tentativa aceita por ele, sem esperar Aviso de status; com o
UltraMsg ou a Z-API ativos, SHALL responder como rota inexistente à
verificação e aos Avisos de status em `/webhooks/whatsapp` (os avisos da Z-API
chegam pelo REQ-37).
Origem: RN-4, RN-5

#### Scenario: T-81 Webhook da Meta com o UltraMsg ativo
- **WHEN** o UltraMsg é o Provedor do WhatsApp e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-91 Webhook da Meta com a Z-API ativa
- **WHEN** a Z-API é o Provedor do WhatsApp e chega a verificação ou um Aviso de status em `/webhooks/whatsapp`
- **THEN** a resposta é 404 e nenhuma Tentativa muda

### Requirement: REQ-37 Aviso da Z-API autenticado pela URL
Quando chegar um aviso da Z-API com o segredo da URL e a instância configurados
para a Z-API ativa, o sistema SHALL tratá-lo como Aviso de status válido
(REQ-12, REQ-13) e responder `200 {"value": true}`, inclusive se nada mudar; com
segredo errado, outra instância ou a Z-API inativa, SHALL responder como rota
inexistente, sem efeito.
Origem: RN-5, RN-4

#### Scenario: T-121 Aviso com o segredo e a instância certos é aceito
- **WHEN** a Z-API está ativa e chega em `/webhooks/zapi/<segredo atual>` o status `RECEIVED` de uma Tentativa enviada, com o `instanceId` configurado
- **THEN** a resposta é 200 `{"value": true}` e a Tentativa fica entregue

#### Scenario: T-122 Segredo errado é recusado como rota inexistente
- **WHEN** chega um aviso com um segredo diferente do atual, inclusive o segredo anterior a uma regravação
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-123 Aviso de outra instância é recusado
- **WHEN** chega um aviso com o segredo atual e um `instanceId` diferente do configurado
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-124 Sem a Z-API ativa a rota não existe
- **WHEN** o UltraMsg ou a Meta é o Provedor ativo, ou o WhatsApp está desligado, e chega um aviso em `/webhooks/zapi/<qualquer segredo>`
- **THEN** a resposta é 404 e nenhuma Tentativa muda

#### Scenario: T-125 Aviso autêntico sem efeito responde sucesso
- **WHEN** chega um aviso autêntico de uma mensagem que o OminiCore não conhece, ou um `RECEIVED` repetido de Tentativa já lida
- **THEN** a resposta é 200 `{"value": true}` e nada muda

### Requirement: REQ-38 Status da mensagem pela Z-API
Quando chegar da Z-API um aviso válido de status, o sistema SHALL aplicar a
cada id da lista `ids` o status entregue para `RECEIVED` e lida para `READ` ou
`PLAYED`, só fazendo avançar (REQ-12), e SHALL ignorar `SENT` e `READ_BY_ME`.
Origem: RN-5, RN-4

#### Scenario: T-126 Leitura confirmada pela Z-API
- **WHEN** chega da Z-API o status `READ` com o id de uma Tentativa entregue
- **THEN** a Tentativa fica lida e a Notificação segue entregue

#### Scenario: T-127 Status SENT não muda nada
- **WHEN** chega da Z-API o status `SENT` com o id de uma Tentativa enviada
- **THEN** a Tentativa continua enviada e a resposta é 200 `{"value": true}`

### Requirement: REQ-39 Erro de envio da Z-API depois do aceite
Quando chegar da Z-API um aviso válido "Ao enviar" com `error`, o sistema SHALL
marcar a Tentativa como falhou, como contato inválido para "Phone number does
not exist" ou número inválido e como recusado pelo Provedor nos demais, e
seguir o REQ-13; sem `error`, SHALL manter a Tentativa como está.
Origem: RN-5, RN-4, RN-2

#### Scenario: T-128 Número sem WhatsApp passa ao SMS
- **WHEN** a ordem é WhatsApp → SMS, a Z-API aceitou a Tentativa e chega "Ao enviar" com `error` "Phone number does not exist"
- **THEN** a Tentativa de WhatsApp fica falhou com classificação contato inválido e existe uma Tentativa pendente no Canal SMS

#### Scenario: T-129 Shadow ban é recusa do Provedor
- **WHEN** chega "Ao enviar" com `error` "The message was not sent due to a likely shadow ban" de uma Tentativa enviada
- **THEN** a Tentativa fica falhou com classificação recusado pelo Provedor

#### Scenario: T-130 Ao enviar sem erro não muda nada
- **WHEN** chega "Ao enviar" sem `error` de uma Tentativa enviada
- **THEN** a Tentativa continua enviada e a resposta é 200 `{"value": true}`

### Requirement: REQ-44 Templates do wbem para o tomador
Quando o Sistema de origem pedir uma Notificação com o Template
`wbem-link-tomador` ou `wbem-otp-tomador`, o sistema SHALL encontrá-lo com a
versão WhatsApp em texto e preenchê-la com a variável `link` ou `codigo`; o
Template SHALL ter só o Canal WhatsApp.
Origem: RN-13

#### Scenario: T-139 Os dois Templates preenchem o texto do WhatsApp
- **GIVEN** o banco com as migrations aplicadas
- **WHEN** o Template `wbem-link-tomador` é preenchido com `link` e o
  `wbem-otp-tomador` com `codigo`
- **THEN** o texto do WhatsApp traz o link e o código, a Ordem de Canais é só
  WhatsApp e o texto serve à Z-API e ao UltraMsg

#### Scenario: T-140 Pedido sem a variável ou por outro Canal é recusado
- **WHEN** o pedido usa `wbem-otp-tomador` sem `codigo`, ou pede o Canal SMS
- **THEN** o pedido é recusado (422) e nada é registrado
