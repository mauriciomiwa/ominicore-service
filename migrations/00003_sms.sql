-- +goose Up
-- Fatia 2 (RN-2, RN-3): contato SMS do Destinatário e versão SMS do boas-vindas.
ALTER TABLE notificacoes ADD COLUMN destinatario_telefone text NOT NULL DEFAULT '';

INSERT INTO template_conteudos (template_id, canal, corpo) VALUES
    ('boas-vindas', 'sms', 'Olá {{nome}}, sua conta está pronta.');
UPDATE templates SET ordem_canais = ARRAY['email', 'sms'] WHERE id = 'boas-vindas';

-- +goose Down
UPDATE templates SET ordem_canais = ARRAY['email'] WHERE id = 'boas-vindas';
DELETE FROM template_conteudos WHERE template_id = 'boas-vindas' AND canal = 'sms';
ALTER TABLE notificacoes DROP COLUMN destinatario_telefone;
