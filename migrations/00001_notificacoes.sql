-- +goose Up
-- Schema `ominicore` (search_path da conexão). Nomes seguem o glossário.

CREATE TABLE templates (
    id           text PRIMARY KEY,
    ordem_canais text[] NOT NULL CHECK (cardinality(ordem_canais) > 0)
);

CREATE TABLE template_conteudos (
    template_id text NOT NULL REFERENCES templates (id),
    canal       text NOT NULL,
    assunto     text NOT NULL DEFAULT '',
    corpo       text NOT NULL,
    PRIMARY KEY (template_id, canal)
);

CREATE TABLE notificacoes (
    id                 uuid PRIMARY KEY,
    sistema_origem     text NOT NULL,
    template_id        text NOT NULL REFERENCES templates (id),
    destinatario_email text NOT NULL DEFAULT '',
    dados              jsonb NOT NULL,
    ordem_canais       text[] NOT NULL,
    status             text NOT NULL,
    criada_em          timestamptz NOT NULL DEFAULT now(),
    atualizada_em      timestamptz NOT NULL DEFAULT now()
);

-- Chave de idempotência por Sistema de origem, válida por 24 h (REQ-3). A linha
-- vencida é reaproveitada pela próxima Notificação com a mesma chave.
CREATE TABLE idempotencia (
    sistema_origem text NOT NULL,
    chave          text NOT NULL,
    notificacao_id uuid NOT NULL REFERENCES notificacoes (id) DEFERRABLE INITIALLY DEFERRED,
    expira_em      timestamptz NOT NULL,
    PRIMARY KEY (sistema_origem, chave)
);

-- Tentativas são a fila de despacho (ADR 0001): o worker reserva com
-- FOR UPDATE SKIP LOCKED as pendentes com disponivel_em vencido.
CREATE TABLE tentativas (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    notificacao_id uuid NOT NULL REFERENCES notificacoes (id),
    canal          text NOT NULL,
    numero         int NOT NULL,
    status         text NOT NULL,
    disponivel_em  timestamptz NOT NULL DEFAULT now(),
    motivo         text NOT NULL DEFAULT '',
    id_provedor    text NOT NULL DEFAULT '',
    criada_em      timestamptz NOT NULL DEFAULT now(),
    atualizada_em  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tentativas_fila ON tentativas (disponivel_em) WHERE status = 'pendente';

-- +goose Down
DROP TABLE tentativas;
DROP TABLE idempotencia;
DROP TABLE notificacoes;
DROP TABLE template_conteudos;
DROP TABLE templates;
