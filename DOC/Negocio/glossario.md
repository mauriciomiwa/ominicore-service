# Glossário — ominicore

A linguagem do domínio: um termo por conceito, com o significado *neste*
negócio. Código, RN, specs e conversa usam estes termos; os de "Evitar" não.
Atualize na hora em que um termo for decidido (skill `sabatina`), não depois.
Regras: `references/glossario-e-decisoes.md` da skill `novo-projeto-sdd`.

## Envio

**Notificação**:
Pedido de um Sistema de origem para avisar um Destinatário a partir de um
Template; termina entregue ou falhou, depois de uma ou mais Tentativas.
_Evitar_: mensagem, envio, disparo

**Tentativa**:
Um envio da Notificação por um Canal a um provedor, com seu resultado.
_Evitar_: envio, retry, disparo

**Canal**:
Meio pelo qual a Notificação chega ao Destinatário: Email, SMS (e WhatsApp,
depois).
_Evitar_: provedor, meio, tipo

**Canal desligado**:
Canal cujo Provedor não tem credenciais configuradas no ambiente; o OminiCore
o trata como indisponível no pedido, no fallback e nos Avisos de status.
_Evitar_: canal inativo, canal off, canal sem provedor

**Provedor**:
Empresa externa que executa o envio de um Canal (SendGrid para Email, Twilio
para SMS).
_Evitar_: canal, gateway

**Ordem de Canais**:
Sequência de Canais que a Notificação percorre no fallback; vem do pedido ou,
na falta, do Template.
_Evitar_: prioridade, cascata

**Retentativa**:
Nova Tentativa no mesmo Canal depois de uma Falha temporária, após uma espera.
_Evitar_: retry, reenvio, nova tentativa

**Falha temporária**:
Falha de uma Tentativa que pode dar certo se repetida no mesmo Canal (provedor
fora do ar, limite de taxa).
_Evitar_: erro transitório, soft fail

**Falha definitiva**:
Falha de uma Tentativa que não muda se repetida (contato inválido, recusa do
provedor); leva ao próximo Canal.
_Evitar_: erro permanente, hard fail

**Aviso de status**:
Mensagem que o Provedor manda ao OminiCore dizendo o que aconteceu com uma
Tentativa já aceita (entregue, lida, falhou).
_Evitar_: callback, evento, webhook (é o meio, não o conceito)

**Template**:
Conteúdo cadastrado no OminiCore, com variáveis, em uma versão por Canal, e com
a Ordem de Canais padrão.
_Evitar_: modelo, layout

**Modelo aprovado**:
Template de mensagem aprovado pela Meta para o WhatsApp, referenciado pela
versão WhatsApp de um Template; o texto mora na Meta, o OminiCore envia só os
parâmetros.
_Evitar_: template da Meta, HSM, message template

## Partes

**Sistema de origem**:
Serviço interno da casa que pede Notificações ao OminiCore.
_Evitar_: cliente, consumidor, tenant

**Destinatário**:
Pessoa que recebe a Notificação, identificada pelo contato em cada Canal.
_Evitar_: cliente, usuário, contato

**Opt-out**:
Registro de que um Destinatário não quer receber por um Canal.
_Evitar_: descadastro, unsubscribe, bloqueio

**Chave de idempotência**:
Identificador que o Sistema de origem dá ao pedido para que o reenvio do mesmo
pedido não gere nova Notificação.
_Evitar_: request id, id de correlação
