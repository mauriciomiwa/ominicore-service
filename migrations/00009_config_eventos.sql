-- +goose Up
-- Fatia 10 (RN-11): cada troca da configuração de um Canal pela API — quem
-- (o azp do Operador), o quê e quando. Nenhuma coluna guarda segredo (REQ-27).
CREATE TABLE config_eventos (
    id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    canal    text NOT NULL,
    acao     text NOT NULL CHECK (acao IN ('gravada', 'apagada')),
    provedor text, -- nulo quando apagada
    por      text NOT NULL,
    em       timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE config_eventos;
