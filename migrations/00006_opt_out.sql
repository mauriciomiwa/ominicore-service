-- +goose Up
-- Fatia 4 (RN-8): Opt-out por contato + Canal, para toda a casa. O contato só
-- existe como HMAC-SHA256(chave, canal|contato normalizado) — nenhuma coluna
-- guarda e-mail ou telefone (REQ-18), e o registro não expira (fica fora da
-- retenção das Notificações).
CREATE TABLE opt_outs (
    canal         text NOT NULL,
    contato_hmac  bytea NOT NULL,
    origem        text NOT NULL, -- api:<sistema de origem> | sendgrid | twilio
    registrado_em timestamptz NOT NULL DEFAULT now(),
    -- hmac primeiro: cada pedido procura por contato_hmac = ANY(...) (REQ-17)
    PRIMARY KEY (contato_hmac, canal)
);

-- Histórico de registro e remoção (quem e quando, REQ-15).
CREATE TABLE opt_out_eventos (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    canal        text NOT NULL,
    contato_hmac bytea NOT NULL,
    acao         text NOT NULL CHECK (acao IN ('registrado', 'removido')),
    por          text NOT NULL,
    em           timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE opt_out_eventos;
DROP TABLE opt_outs;
