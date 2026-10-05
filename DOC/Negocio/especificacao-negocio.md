# Especificação de negócio — ominicore

## Visão e problema
O OminiCore é o **serviço interno de notificações multicanal** da casa. Hoje
cada sistema que precisa avisar alguém (e-mail, SMS) integra sozinho com o
provedor, repete credenciais, retry e controle de descadastro. O OminiCore
centraliza isso: o Sistema de origem pede uma Notificação a partir de um
Template, e o OminiCore escolhe o Canal, tenta, faz fallback e registra o
resultado.

## Usuários e stakeholders
- **Quem opera:** os Sistemas de origem internos (serviços da casa), que chamam
  a API de envio. Não há usuário humano final operando o OminiCore no MVP.
- **Quem é impactado:** o Destinatário (pessoa que recebe a mensagem) e os times
  donos dos Sistemas de origem.
- **Quem decide:** dono do projeto (Mauricio Miwa).

## Objetivo e métricas de sucesso
- **Taxa de entrega final > 98%** (contando o fallback entre Canais), medida
  pelo status de entrega devolvido pelos provedores.
- **Tempo até a primeira Tentativa < 5 s** (do aceite do pedido ao envio ao
  provedor), no volume do MVP.
- Volume de referência do MVP: **até ~100 mil Notificações/dia**.

## Situação atual
Greenfield. Não há sistema a substituir nem dado a migrar; os Sistemas de
origem passam a chamar o OminiCore à medida que forem integrados.

## Escopo / MVP
- **Dentro:** envio por Template com dados; Canais Email (SendGrid), SMS
  (Twilio) e WhatsApp (Meta, só Modelo aprovado — incluído em 2026-09-29); ordem de Canais com retry e fallback; status de entrega por webhook;
  idempotência; Opt-out por Destinatário e Canal; retenção de 30 dias;
  credenciais só no OpenBao.
- **Fora (próximas etapas):** repasse das respostas do
  Destinatário ao Sistema de origem; janela de horário e agendamento; consumo de
  pedidos por fila externa (RabbitMQ/SQS).

## Requisitos de negócio

| RN | Descrição | Critério de aceite (verificável) | MVP? |
|---|---|---|---|
| RN-1 | Um Sistema de origem autenticado pede uma Notificação informando o Template, o Destinatário e os dados das variáveis. | Pedido válido é aceito e recebe um identificador; pedido de quem não se autenticou, com Template inexistente ou com variável faltando é recusado com o motivo. | sim |
| RN-2 | A Notificação percorre os Canais na ordem pedida (ou na ordem padrão do Template), pulando Canal sem contato do Destinatário. Falha temporária é retentada no mesmo Canal até 3 vezes (4 Tentativas no total), com espera crescente de 5 s, 30 s e 2 min; falha definitiva ou Retentativas esgotadas passam ao próximo Canal. | Com o 1º Canal falhando de forma definitiva, a Notificação sai pelo 2º; com falha temporária, há até 4 Tentativas no mesmo Canal antes de mudar; Canal sem contato é pulado; sem Canal restante, a Notificação termina `falhou`. | sim |
| RN-3 | Canais disponíveis: Email (SendGrid) e SMS (Twilio). | Uma Notificação com Canal Email chega por e-mail; com Canal SMS chega por SMS. | sim |
| RN-4 | Canal WhatsApp por um de dois Provedores, escolhido por ambiente: a Cloud API da Meta, só por Modelo aprovado, ou o UltraMsg (gateway não oficial, risco aceito pelo cliente — ADR 0004), por texto da versão WhatsApp do Template preenchido com as variáveis (até 4096 caracteres); sempre com o telefone do Destinatário. Na Meta, número sem WhatsApp ou Modelo não aprovado é Falha definitiva; no UltraMsg, credencial ou instância recusada é Falha definitiva. Versão WhatsApp sem a forma do Provedor ativo recusa o pedido. Sem as credenciais do Provedor no ambiente, o Canal fica desligado: o pedido só com Canais desligados é recusado e, na ordem, o Canal desligado é pulado. O UltraMsg não manda Aviso de status: a Tentativa aceita por ele fica enviada. | Notificação com Canal WhatsApp chega pelo WhatsApp com os parâmetros do Modelo aprovado (Meta) ou com o texto preenchido (UltraMsg); número sem WhatsApp na Meta passa ao próximo Canal da ordem; sem credenciais no ambiente, o sistema sobe sem o Canal e o pedido só por WhatsApp é recusado. | sim (fatias 06, 07 e 08) |
| RN-5 | O Aviso de status do Provedor (entregue, lida, falhou) atualiza a Tentativa e só a faz avançar (enviada → entregue → lida; falhou é final); falha de entrega depois do aceite dispara o fallback para o próximo Canal (RN-2), salvo se a Notificação já foi entregue; aviso sem assinatura válida é descartado. O Sistema de origem dono consulta o status da Notificação e das Tentativas, sem contato nem texto do motivo. | Aviso válido de entrega marca a Tentativa e a Notificação entregues; bounce no Email leva ao SMS; aviso com assinatura inválida é recusado e nada muda; outro Sistema de origem não enxerga a Notificação. | sim |
| RN-6 | Respostas do Destinatário (ex.: resposta no WhatsApp) são repassadas ao Sistema de origem. | Resposta do Destinatário chega ao Sistema de origem que enviou. | depois |
| RN-7 | Todo pedido traz uma chave de idempotência; a mesma chave do mesmo Sistema de origem em até 24 h devolve a Notificação já criada e nada é reenviado. | Dois pedidos com a mesma chave em 24 h geram uma única Notificação e uma única sequência de Tentativas; pedido sem chave é recusado. | sim |
| RN-8 | O Destinatário pode fazer Opt-out de um Canal, valendo para toda a casa (por contato + Canal); o OminiCore pula esse Canal no pedido, no fallback e no envio. O Opt-out é registrado pelo Sistema de origem (API) ou pelos sinais do Provedor (descadastro no SendGrid, STOP no Twilio); só a API remove, com registro de quem e quando. O contato do Opt-out é guardado só como HMAC, e o registro não expira. | Com Opt-out de SMS registrado, uma Notificação com ordem SMS → Email sai só por Email; com todos os Canais da ordem em Opt-out, o pedido é recusado; o STOP no Twilio registra o Opt-out de SMS. | sim |
| RN-9 | O conteúdo das Notificações é dado pessoal: depois do prazo de retenção (padrão 30 dias, global e configurável), a Notificação **finalizada** é anonimizada — contato, dados das variáveis e motivo das Tentativas são apagados; status, Canais, classificação da falha e horários ficam (medem a entrega). Notificação ainda pendente não é anonimizada. O Opt-out não expira (só HMAC, RN-8). Credenciais dos provedores ficam só no OpenBao. | Após o prazo, conteúdo e contato do Destinatário não são mais legíveis no histórico, e a consulta diz quando a Notificação foi anonimizada; nenhuma credencial de provedor existe em código, config versionada, banco ou log. | sim |
| RN-10 | Janela de horário permitido e agendamento de envio. | Notificação fora da janela espera o próximo horário permitido. | depois |

## Regras de negócio / políticas
- **Nunca** enviar duas vezes a mesma Notificação por reenvio do pedido (RN-7).
- **Nunca** enviar por um Canal em que o Destinatário fez Opt-out (RN-8).
- **Nunca** aceitar pedido de quem não é um Sistema de origem autenticado (RN-1).
- **Sempre** registrar cada Tentativa com Canal, resultado e motivo da falha.
- **Nunca** gravar conteúdo, contato do Destinatário ou credencial em log.

## Restrições
- **LGPD:** telefone, e-mail e conteúdo são dados pessoais (RN-9).
- **Segredos:** só no OpenBao; autenticação dos Sistemas de origem pelo Keycloak
  da casa.
- Prazo e orçamento: não definidos no kickoff.

## Glossário
Ver `Negocio/glossario.md`.
