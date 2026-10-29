-- +goose Up
-- Templates do aviva-back (sistema de origem wbem, RN-99 do aviva-back): link de
-- acesso e código de acesso do tomador. Só a versão WhatsApp, em texto (Z-API).
INSERT INTO templates (id, ordem_canais) VALUES
    ('wbem-link-tomador', ARRAY['whatsapp']),
    ('wbem-otp-tomador', ARRAY['whatsapp']);
INSERT INTO template_conteudos (template_id, canal, corpo) VALUES
    ('wbem-link-tomador', 'whatsapp',
     'WBem: acesse sua proposta de crédito: {{link}} (vale 24 h, uso único).'),
    ('wbem-otp-tomador', 'whatsapp',
     'WBem: seu código de acesso é {{codigo}}. Não compartilhe.');

-- +goose Down
DELETE FROM template_conteudos WHERE template_id IN ('wbem-link-tomador', 'wbem-otp-tomador');
DELETE FROM templates WHERE id IN ('wbem-link-tomador', 'wbem-otp-tomador');
