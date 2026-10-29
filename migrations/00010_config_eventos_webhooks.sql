-- +goose Up
-- Fatia 13 (REQ-41): o recadastro dos webhooks da Z-API também é uma troca
-- registrada (ação "webhooks"). Só alarga o CHECK: aditiva, o binário anterior
-- continua gravando 'gravada' e 'apagada'.
ALTER TABLE config_eventos DROP CONSTRAINT config_eventos_acao_check;
ALTER TABLE config_eventos ADD CONSTRAINT config_eventos_acao_check
    CHECK (acao IN ('gravada', 'apagada', 'webhooks'));

-- +goose Down
-- Voltar exige tirar os registros de recadastro, que o CHECK antigo recusa.
DELETE FROM config_eventos WHERE acao = 'webhooks';
ALTER TABLE config_eventos DROP CONSTRAINT config_eventos_acao_check;
ALTER TABLE config_eventos ADD CONSTRAINT config_eventos_acao_check
    CHECK (acao IN ('gravada', 'apagada'));
