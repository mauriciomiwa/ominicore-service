-- +goose Up
-- Organização (RN-15, REQ-50 a REQ-52): empresa ou marca do grupo. Cadastro e
-- vínculo client → Organização só por migration (sabatina de 2026-11-14). Tudo
-- o que existia antes pertence à `wbem`.
CREATE TABLE organizacoes (
    id   text PRIMARY KEY,
    nome text NOT NULL
);

CREATE TABLE sistemas_origem (
    client      text PRIMARY KEY, -- azp do token do Keycloak
    organizacao text NOT NULL REFERENCES organizacoes (id)
);

INSERT INTO organizacoes (id, nome) VALUES ('wbem', 'wbem');
INSERT INTO sistemas_origem (client, organizacao) VALUES ('wbem', 'wbem');

-- DEFAULT constante não reescreve a tabela. Ele fica nesta versão para o código
-- anterior, que ainda roda entre o migrate e o restart do deploy, continuar
-- gravando (na wbem, a única Organização em prod); a fatia 18 o remove.
ALTER TABLE notificacoes ADD COLUMN organizacao text NOT NULL DEFAULT 'wbem' REFERENCES organizacoes (id);
ALTER TABLE templates ADD COLUMN organizacao text NOT NULL DEFAULT 'wbem' REFERENCES organizacoes (id);

CREATE INDEX notificacoes_organizacao ON notificacoes (organizacao, criada_em);

-- Barreira do banco (REQ-52, design D3): a API fixa a Organização na transação
-- (set_config local); sem ela (worker, retenção, Avisos de status) vale tudo.
-- FORCE porque o role do app é o dono das tabelas.
ALTER TABLE notificacoes ENABLE ROW LEVEL SECURITY;
ALTER TABLE notificacoes FORCE ROW LEVEL SECURITY;
CREATE POLICY por_organizacao ON notificacoes
    USING (NULLIF(current_setting('ominicore.organizacao', true), '') IS NULL
           OR organizacao = current_setting('ominicore.organizacao', true));

-- +goose Down
DROP POLICY por_organizacao ON notificacoes;
ALTER TABLE notificacoes NO FORCE ROW LEVEL SECURITY;
ALTER TABLE notificacoes DISABLE ROW LEVEL SECURITY;
DROP INDEX notificacoes_organizacao;
ALTER TABLE templates DROP COLUMN organizacao;
ALTER TABLE notificacoes DROP COLUMN organizacao;
DROP TABLE sistemas_origem;
DROP TABLE organizacoes;
