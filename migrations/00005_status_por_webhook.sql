-- +goose Up
-- Fatia 3 (RN-5): Aviso de status acha a Tentativa pelo id do Provedor; a
-- consulta expõe só a classificação da falha (REQ-14), nunca o motivo.
ALTER TABLE tentativas ADD COLUMN categoria_falha text NOT NULL DEFAULT '';
CREATE INDEX tentativas_id_provedor ON tentativas (id_provedor) WHERE id_provedor <> '';

-- +goose Down
DROP INDEX tentativas_id_provedor;
ALTER TABLE tentativas DROP COLUMN categoria_falha;
