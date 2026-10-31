-- +goose Up
-- Régua de comunicação do aviva-back (Sistema de origem wbem, RN-14): os 16
-- modelos da apresentação comercial v9 (31/10/2026), só a versão WhatsApp em
-- texto (Z-API). O botão da v9 vira uma linha com o link; o PDF (Mensagens 7,
-- 8 e 11) vai como link. Os dois Templates da RN-13 não mudam.
INSERT INTO templates (id, ordem_canais) VALUES
    ('wbem-financiamento-na-clinica', ARRAY['whatsapp']),
    ('wbem-open-finance-confirmado', ARRAY['whatsapp']),
    ('wbem-open-finance-nao-confirmado', ARRAY['whatsapp']),
    ('wbem-pix-autorizado', ARRAY['whatsapp']),
    ('wbem-pix-nao-confirmado', ARRAY['whatsapp']),
    ('wbem-proposta-vence-hoje', ARRAY['whatsapp']),
    ('wbem-debito-cancelado', ARRAY['whatsapp']),
    ('wbem-boleto-disponivel', ARRAY['whatsapp']),
    ('wbem-parcela-nao-debitada', ARRAY['whatsapp']),
    ('wbem-contratacao-liberada', ARRAY['whatsapp']),
    ('wbem-contratacao-adiada', ARRAY['whatsapp']),
    ('wbem-contrato-assinado', ARRAY['whatsapp']),
    ('wbem-lembrete-open-finance', ARRAY['whatsapp']),
    ('wbem-lembrete-pix', ARRAY['whatsapp']),
    ('wbem-lembrete-biometria', ARRAY['whatsapp']),
    ('wbem-lembrete-assinatura', ARRAY['whatsapp']);
INSERT INTO template_conteudos (template_id, canal, corpo) VALUES
    -- link inicial
    ('wbem-financiamento-na-clinica', 'whatsapp',
     E'*Seu financiamento na {{clinica}}*\n\nOlá, {{nome}}! A {{clinica}} registrou o financiamento do seu tratamento de {{valor_tratamento}}.\n\nCondição escolhida na clínica: {{condicao}}, primeira parcela em {{primeira_parcela}}.\n\nPara concluir, confirme pelo seu celular. O link é seu e vale por 24 horas.\n\nConfirmar meu financiamento: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 1
    ('wbem-open-finance-confirmado', 'whatsapp',
     E'*Autorização confirmada*\n\nOlá, {{nome}}! O seu banco confirmou a autorização do Open Finance.\n\nToque no link abaixo para continuar a contratação do seu financiamento na {{clinica}}. Por segurança, vamos pedir um código de confirmação antes de seguir.\n\nContinuar contratação: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 2
    ('wbem-open-finance-nao-confirmado', 'whatsapp',
     E'*Autorização não confirmada*\n\nOlá, {{nome}}! Não foi possível confirmar a autorização do Open Finance no seu banco.\n\nSua proposta de financiamento na {{clinica}} continua reservada até {{reservada_ate}}. Para seguir, é só fazer um novo pedido de autorização — leva poucos minutos.\n\nAutorizar novamente: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 3
    ('wbem-pix-autorizado', 'whatsapp',
     E'*Débito automático autorizado*\n\nOlá, {{nome}}! O seu banco confirmou a autorização do Pix Automático para as parcelas do seu financiamento na {{clinica}}.\n\nFalta pouco: toque no link abaixo para confirmar sua identidade e assinar o contrato. Por segurança, vamos pedir um código de confirmação antes de seguir.\n\nContinuar contratação: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 4
    ('wbem-pix-nao-confirmado', 'whatsapp',
     E'*Autorização não confirmada*\n\nOlá, {{nome}}! Não foi possível confirmar a autorização do Pix Automático no seu banco.\n\nSua proposta de financiamento na {{clinica}} continua reservada até {{reservada_ate}}. Para seguir, é só retomar a autorização de onde você parou.\n\nRetomar autorização: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 5
    ('wbem-proposta-vence-hoje', 'whatsapp',
     E'*Sua proposta vence hoje*\n\nOlá, {{nome}}! Sua proposta de financiamento na {{clinica}} fica reservada só até hoje às {{hora_limite}}. Depois disso, será preciso fazer uma nova solicitação na clínica.\n\nPara seguir, conclua a autorização pendente no seu banco.\n\nConcluir autorização: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 6
    ('wbem-debito-cancelado', 'whatsapp',
     E'*Débito automático cancelado*\n\nOlá, {{nome}}! Identificamos que a autorização do Pix Automático das parcelas do seu financiamento na {{clinica}} foi cancelada no seu banco.\n\nA partir de agora, as parcelas passam a ser pagas por boleto. Cada boleto é registrado 10 dias antes do vencimento, enviado por aqui, pelo WhatsApp, e também fica disponível para download na sua área do cliente.\n\nVer minhas parcelas: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 7
    ('wbem-boleto-disponivel', 'whatsapp',
     E'*Boleto da parcela {{parcela}}*\n\nOlá, {{nome}}! Segue o boleto da parcela {{parcela}} do seu financiamento na {{clinica}}.\n\nValor: {{valor}}\nVencimento: {{vencimento}}\nBoleto em PDF: {{link_pdf}}\n\nVocê também encontra este e os próximos boletos na sua área do cliente.\n\nVer minhas parcelas: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 8
    ('wbem-parcela-nao-debitada', 'whatsapp',
     E'*Parcela {{parcela}} não debitada*\n\nOlá, {{nome}}! Não conseguimos concluir o débito automático da parcela {{parcela}} do seu financiamento na {{clinica}}.\n\nVocê pode pagar pelo boleto: {{link_pdf}}\nValor: {{valor}}. Vencimento: {{vencimento}}.\n\nEnquanto o boleto não for pago, vamos tentar o débito de novo nos próximos dias. As próximas parcelas continuam no débito automático.\n\nVer minhas parcelas: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 9
    ('wbem-contratacao-liberada', 'whatsapp',
     E'*Contratação liberada*\n\nOlá, {{nome}}! Já é possível concluir a contratação do seu financiamento na {{clinica}}, com a mesma taxa aprovada. Os valores são atualizados para a data de hoje.\n\nToque no link abaixo para continuar. Por segurança, vamos pedir um código de confirmação antes de seguir.\n\nConcluir contratação: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 10
    ('wbem-contratacao-adiada', 'whatsapp',
     E'*Sua contratação vai levar mais alguns dias*\n\nOlá, {{nome}}! A conclusão da contratação do seu financiamento na {{clinica}} vai levar mais alguns dias. Sua condição continua garantida até {{garantida_ate}}, e você recebe por aqui o link assim que for possível finalizar.\n\nVer minha proposta: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 11
    ('wbem-contrato-assinado', 'whatsapp',
     E'*Financiamento contratado*\n\nOlá, {{nome}}! O seu financiamento na {{clinica}} foi contratado. A sua Cédula de Crédito Bancário está em {{link_pdf}}, e você acompanha as parcelas na sua área do cliente.\n\nAcessar meu financiamento: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 12
    ('wbem-lembrete-open-finance', 'whatsapp',
     E'*Falta autorizar no seu banco*\n\nOlá, {{nome}}! Para seguir com o financiamento na {{clinica}}, falta autorizar o compartilhamento dos seus dados pelo Open Finance no aplicativo do seu banco.\n\nSe o aplicativo não abriu ou você saiu antes de concluir, toque no link abaixo para retomar a autorização.\n\nRetomar autorização: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 13
    ('wbem-lembrete-pix', 'whatsapp',
     E'*Autorize o débito automático*\n\nOlá, {{nome}}! Enviamos ao seu banco o pedido de débito automático das parcelas do seu financiamento na {{clinica}}.\n\nPara seguir com a contratação, abra o aplicativo do seu banco e autorize o Pix Automático. Ele aparece em nome do W Capital Health FIDC, parceiro da W Bem, nas notificações ou na área Pix do aplicativo.\n\nVer como autorizar: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 14
    ('wbem-lembrete-biometria', 'whatsapp',
     E'*Falta confirmar a sua identidade*\n\nOlá, {{nome}}! Falta só a foto do seu rosto para confirmar a sua identidade e seguir com o financiamento na {{clinica}}.\n\nLeva menos de um minuto. Toque no link abaixo para continuar de onde parou.\n\nContinuar verificação: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._'),
    -- Mensagem 15
    ('wbem-lembrete-assinatura', 'whatsapp',
     E'*Falta assinar o contrato*\n\nOlá, {{nome}}! A sua identidade foi confirmada. Falta assinar o contrato do seu financiamento na {{clinica}} para concluir a contratação.\n\nAssinar contrato: {{link}}\n\n_A W Bem nunca pede senha, código ou depósito._');

-- +goose Down
DELETE FROM template_conteudos WHERE template_id IN ('wbem-financiamento-na-clinica', 'wbem-open-finance-confirmado', 'wbem-open-finance-nao-confirmado', 'wbem-pix-autorizado', 'wbem-pix-nao-confirmado', 'wbem-proposta-vence-hoje', 'wbem-debito-cancelado', 'wbem-boleto-disponivel', 'wbem-parcela-nao-debitada', 'wbem-contratacao-liberada', 'wbem-contratacao-adiada', 'wbem-contrato-assinado', 'wbem-lembrete-open-finance', 'wbem-lembrete-pix', 'wbem-lembrete-biometria', 'wbem-lembrete-assinatura');
DELETE FROM templates WHERE id IN ('wbem-financiamento-na-clinica', 'wbem-open-finance-confirmado', 'wbem-open-finance-nao-confirmado', 'wbem-pix-autorizado', 'wbem-pix-nao-confirmado', 'wbem-proposta-vence-hoje', 'wbem-debito-cancelado', 'wbem-boleto-disponivel', 'wbem-parcela-nao-debitada', 'wbem-contratacao-liberada', 'wbem-contratacao-adiada', 'wbem-contrato-assinado', 'wbem-lembrete-open-finance', 'wbem-lembrete-pix', 'wbem-lembrete-biometria', 'wbem-lembrete-assinatura');
