# Design

## Context

Hoje o isolamento é só por Sistema de origem: o `azp` do token
(`internal/adaptadores/saida/keycloak/validador.go`) vira `sistema_origem` em
`notificacoes` e `idempotencia`, e `ConsultarNotificacao` filtra por ele.
`templates` tem só `id` como chave e as credenciais vivem num prefixo único do
OpenBao. O app conecta ao Postgres com um único role, dono das tabelas (é ele
quem roda o `goose` no `publicar.sh`), e o schema vem do `search_path`.

A sabatina de 2026-11-14 fechou: Organização = empresa ou marca do grupo; uma
coluna com a Organização em cada tabela e row-level security (RLS) como segunda
barreira; Organização derivada do client, nunca de header; cadastro por
migration; os dados atuais vão para `wbem`; Opt-out e retenção globais. Esta
fatia (17) não muda Templates, credenciais nem Console.

## Goals / Non-Goals

**Goals:**
- Organização como conceito no domínio e no banco, com `wbem` cadastrada.
- Client sem Organização recusado com 403 nas rotas do Sistema de origem.
- Toda Notificação nova gravada com a Organização; leitura da API barrada no
  banco por Organização.
- Migration aditiva e reversível; nenhuma mudança visível para o `wbem`.

**Non-Goals:**
- Templates escolhidos por Organização e credenciais por Organização (fatia 18).
  A coluna `templates.organizacao` nasce agora, preenchida com `wbem`, mas ainda
  não filtra o pedido.
- Operadores, `config_eventos` e Console por Organização (fatia 19).
- API para cadastrar Organização (decisão P8: por migration).
- RLS em `tentativas`, `idempotencia` e `opt_outs`: a Tentativa só é lida pela
  Notificação, a idempotência já é por Sistema de origem e o Opt-out é global.

## Decisions

### D1 — Tabelas `organizacoes` e `sistemas_origem`
`organizacoes (id text PK, nome text)` e `sistemas_origem (client text PK,
organizacao text FK)`. A chave é o texto `id` (`wbem`), que já é como o domínio
fala do Sistema de origem e evita um join para mostrar. Alternativa descartada:
pôr a Organização como atributo do client no Keycloak (claim no token). Ela
espalharia a fonte de verdade entre dois sistemas e dois realms, e um erro de
mapper no Keycloak trocaria a Organização sem nenhum rastro no OminiCore.

### D2 — Resolver a Organização na entrada HTTP, com cache curto
O middleware que já extrai o `azp` passa a resolver `client → organização` por
uma porta nova `OrganizacoesPort.DoSistema(ctx, client)`. Client sem linha →
403 `problem+json` "Sistema de origem sem Organização", antes de qualquer caso
de uso. A resposta fica em cache em memória por 60 s, para não custar uma
consulta por pedido, já que o cadastro só muda por migration. O client do
Console continua barrado antes, pelo `soSistemaDeOrigem` (REQ-36).

A Organização entra no contexto da requisição ao lado do Sistema de origem. Os
casos de uso recebem a Organização como argumento explícito (como já recebem
o `sistemaOrigem`), e não a leem do `context`.

### D3 — RLS em `notificacoes` com a Organização fixada na transação
```sql
ALTER TABLE notificacoes ENABLE ROW LEVEL SECURITY;
ALTER TABLE notificacoes FORCE ROW LEVEL SECURITY; -- o app é o dono
CREATE POLICY por_organizacao ON notificacoes
  USING (NULLIF(current_setting('ominicore.organizacao', true), '') IS NULL
         OR organizacao = current_setting('ominicore.organizacao', true));
```
Os caminhos da API (`RegistrarNotificacao`, `ConsultarNotificacao`, busca
por chave de idempotência) rodam numa transação que começa com
`SELECT set_config('ominicore.organizacao', $1, true)`. O valor vale só até o
fim da transação, então nada vaza entre requisições que usam a mesma conexão
do pool. O worker, a retenção e os Avisos de status não fixam a Organização e
continuam vendo tudo (REQ-52).

**Trade-off aceito:** a barreira só vale onde a Organização é fixada. Um caminho
da API que esquecesse de fixá-la veria tudo, mas ainda filtraria pelo Sistema de
origem, como hoje. A alternativa, um segundo role do banco sem `BYPASSRLS` só
para a API, dobra credenciais no OpenBao e no Aurora e foi deixada para quando
houver uma segunda Organização em produção. O T-163 prova a barreira, e um teste
de arquitetura garante que toda leitura de `notificacoes` feita pela API passa
pelo helper `naOrganizacao(tx, org)`.

### D4 — Migration `00013_organizacoes.sql`
Up:
1. cria `organizacoes` e `sistemas_origem`; insere `('wbem', 'wbem')` e o
   vínculo `('wbem' → 'wbem')`;
2. `ALTER TABLE notificacoes ADD COLUMN organizacao text NOT NULL DEFAULT
   'wbem' REFERENCES organizacoes(id)`; o mesmo em `templates`. O default
   fica nesta versão: entre o `migrate` e o restart do deploy, o código anterior
   ainda grava sem Organização, e na `wbem` ele acerta. O código novo sempre
   informa a Organização (o caso de uso recusa a vazia), e a fatia 18 remove o
   default;
3. índice `notificacoes (organizacao, criada_em)` e a política do D3.

O `ADD COLUMN ... DEFAULT` constante não reescreve a tabela no Postgres 11+, e
a tabela em prod é pequena (retenção de 30 dias). Down: remove a política, as
colunas e as duas tabelas. As Notificações e os Templates continuam intactos.

A change 15, ainda sem código, renumera a sua migration para `00014`.

### D5 — Clients de teste
Os testes que usam `sistema-a`/`sistema-b` cadastram as suas Organizações pelo
helper de integração (`org-a`, `org-b`). Com o fake da porta, cadastram no fake.
Nenhuma migration cria Organização de teste.

## Risks / Trade-offs

- **Client real fora do cadastro em hml ou prod** → passa a receber 403. Mitigação:
  antes de publicar, listar os `sistema_origem` distintos dos últimos 30 dias em
  cada ambiente, só com leitura (skill `ambientes`). Se aparecer algum além de
  `wbem`, a migration o cadastra antes do deploy.
- **RLS com `FORCE` pode esconder linha do próprio worker** se alguém fixar a
  Organização numa transação do worker. Mitigação: o T-165 roda o worker com
  duas Organizações.
- **Cache de 60 s** atrasa um vínculo novo por até 1 min. Aceitável, porque o
  vínculo só muda por deploy.

## Migration Plan

1. hml: `publicar.sh hml` (o goose aplica a `00013`), seguido de pedido e
   consulta pelo client do aviva-back de hml.
2. prod: consulta de leitura dos `sistema_origem` distintos, depois
   `HOST=aviva publicar.sh prod` pelo dono, seguido de um pedido real do `wbem`.
3. Rollback: `goose down` para a `00012`, depois a imagem anterior
   (`.anterior`). Não há perda de dados.

## Open Questions

Nenhuma: as decisões estão fechadas pela sabatina (P1 a P10).
