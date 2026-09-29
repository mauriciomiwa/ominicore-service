-- +goose Up
-- Fatia 6 (RN-4): versão WhatsApp de um Template referencia um Modelo aprovado
-- pela Meta (nome + idioma) e as variáveis que viram os parâmetros {{1}}, {{2}}…
ALTER TABLE template_conteudos
    ADD COLUMN modelo_aprovado text NOT NULL DEFAULT '',
    ADD COLUMN idioma          text NOT NULL DEFAULT '',
    ADD COLUMN parametros      text[] NOT NULL DEFAULT '{}';

-- ordem padrão do boas-vindas NÃO muda: WhatsApp só quando o pedido pede.
INSERT INTO template_conteudos (template_id, canal, corpo, modelo_aprovado, idioma, parametros)
VALUES ('boas-vindas', 'whatsapp', '', 'boas_vindas', 'pt_BR', ARRAY['nome']);

-- +goose Down
DELETE FROM template_conteudos WHERE template_id = 'boas-vindas' AND canal = 'whatsapp';
ALTER TABLE template_conteudos DROP COLUMN parametros, DROP COLUMN idioma, DROP COLUMN modelo_aprovado;
