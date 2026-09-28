-- +goose Up
-- Template da fatia 1. Cadastro por API é decisão em aberto (DOC/to-do.md).
INSERT INTO templates (id, ordem_canais) VALUES ('boas-vindas', ARRAY['email']);
INSERT INTO template_conteudos (template_id, canal, assunto, corpo) VALUES
    ('boas-vindas', 'email', 'Bem-vindo, {{nome}}',
     'Olá {{nome}}, sua conta está pronta.');

-- +goose Down
DELETE FROM template_conteudos WHERE template_id = 'boas-vindas';
DELETE FROM templates WHERE id = 'boas-vindas';
