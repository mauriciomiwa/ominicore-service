-- Schema e role do OminiCore no database new_aviva (Aurora db-zillionvf-prod),
-- no mesmo molde do keycloak/aviva (flip/infra/aviva2.0/keycloak/criar-banco-keycloak.sql).
-- Quem roda: o dono, como master, no DBeaver (auto-commit LIGADO).
-- ATENÇÃO: new_aviva é produção do zillion — rode um ambiente por vez e confira.
-- A senha nunca fica aqui: gere, rode, e grave o DSN só no OpenBao
-- (deploy/aviva2/openbao/configurar-openbao.sh).

-- ===== hml =====
-- 1) qualquer database (role é global):
create role ominicore_hml login password '<GERAR-E-COLAR-AQUI>';
-- 2) conectado ao new_aviva:
grant connect on database new_aviva to ominicore_hml;
create schema ominicore_hml;
grant usage, create on schema ominicore_hml to ominicore_hml;
alter role ominicore_hml in database new_aviva set search_path = ominicore_hml;

-- ===== prod ===== (só quando for promover; mesmo roteiro)
-- create role ominicore login password '<GERAR-E-COLAR-AQUI>';
-- grant connect on database new_aviva to ominicore;
-- create schema ominicore;
-- grant usage, create on schema ominicore to ominicore;
-- alter role ominicore in database new_aviva set search_path = ominicore;

-- DSN para o OpenBao (kv/<prefixo>/postgres, campo dsn):
-- postgres://ominicore_hml:<senha>@db-zillionvf-prod.cluster-cphzw9bdavhn.sa-east-1.rds.amazonaws.com:5432/new_aviva?sslmode=verify-full&sslrootcert=/etc/ominicore/rds-ca.pem
