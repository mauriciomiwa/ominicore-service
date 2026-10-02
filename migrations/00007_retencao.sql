-- +goose Up
-- Fatia 5 (RN-9): Notificação finalizada é anonimizada depois do prazo de
-- retenção; o índice parcial serve à busca das ainda não anonimizadas.
ALTER TABLE notificacoes ADD COLUMN anonimizada_em timestamptz;
CREATE INDEX notificacoes_a_anonimizar ON notificacoes (criada_em) WHERE anonimizada_em IS NULL;

-- +goose Down
DROP INDEX notificacoes_a_anonimizar;
ALTER TABLE notificacoes DROP COLUMN anonimizada_em;
