-- +goose Up
-- Fatia 8 (RN-4): versão WhatsApp do boas-vindas ganha o texto enviado pelo
-- UltraMsg (o mesmo do SMS); o Modelo aprovado continua para a Meta.
UPDATE template_conteudos SET corpo = 'Olá {{nome}}, sua conta está pronta.'
WHERE template_id = 'boas-vindas' AND canal = 'whatsapp';

-- +goose Down
UPDATE template_conteudos SET corpo = ''
WHERE template_id = 'boas-vindas' AND canal = 'whatsapp';
