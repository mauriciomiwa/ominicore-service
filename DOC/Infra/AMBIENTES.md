# Ambientes — contrato canônico

Última revisão: `2026-09-29` — decisões do dono (Mauricio Miwa) na sabatina de
ambientes; ver `log.md` de 2026-09-29.

Regra de conteúdo: **nomes e caminhos sim, valores de segredo nunca.**
hml e prod estão **decididos, não criados**: a criação é a trilha de
infraestrutura (`System/backlog-mvp.md`, `to-do.md`). Até existirem, toda ação
no aviva2.0 é tratada como prod.

## 1. Quais ambientes existem

| Ambiente | Para que serve | Quem usa | Dado real? |
|---|---|---|---|
| local | desenvolvimento (`deploy/local/compose.yaml`) | dev | não |
| hml | validação antes de prod | dono + times dos Sistemas de origem em integração | não — só sintético, Provedores em sandbox |
| prod | Sistemas de origem da casa | Sistemas de origem | sim (PII: RN-9) |

## 2. Como cada ambiente se identifica

| Ambiente | Evidência (comando/variável) | Valor esperado |
|---|---|---|
| local | `OMINICORE_KC_ISSUER` (e `OMINICORE_BAO_TOKEN` preenchido) | `http://localhost:8081/realms/ominicore` · `dev-root` |
| hml | `OMINICORE_KC_ISSUER` + `OMINICORE_BAO_PREFIXO` | `https://auth.flipapi.com.br/realms/ominicore-hml` · `ominicore-hml` |
| prod | `OMINICORE_KC_ISSUER` + `OMINICORE_BAO_PREFIXO` | `https://auth.flipapi.com.br/realms/ominicore` · `ominicore` |

No host: `sudo podman inspect <container> --format '{{range .Config.Env}}{{println .}}{{end}}' | grep -E 'KC_ISSUER|BAO_PREFIXO'`.
hml e prod rodam **no mesmo host** (aviva2.0): a distinção é realm + prefixo
do cofre + schema, não a máquina. As duas variáveis têm de concordar; se
discordarem, é defeito — não execute.

Fonte do host/realm: `../flip/infra/aviva2.0/keycloak/keycloak.env`
(`KC_HOSTNAME`) e o padrão `wbem`/`wbem-hml` de `gerar-realms.py`.

## 3. Endereços e recursos

| Recurso | hml | prod |
|---|---|---|
| API de envio (só rede `aviva`) | `http://ominicore-api-hml:8080` | `http://ominicore-api:8080` |
| Webhooks públicos (Traefik + Cloudflare, ADR 0003) | `https://notify-hml.flipapi.com.br/webhooks/*` | `https://notify.flipapi.com.br/webhooks/*` |
| Keycloak (`iss`) | `https://auth.flipapi.com.br/realms/ominicore-hml` | `https://auth.flipapi.com.br/realms/ominicore` |
| Keycloak backchannel (`OMINICORE_KC_URL`) | `http://keycloak:8080` | `http://keycloak:8080` |
| Banco | Aurora `db-zillionvf-prod` / `new_aviva`, schema `ominicore_hml`, role próprio | Aurora `db-zillionvf-prod` / `new_aviva`, schema `ominicore`, role próprio |
| Fila | tabela `tentativas` do schema (ADR 0001) | idem |
| Host | aviva2.0 (EC2 arm64, us-east-1), quadlets `ominicore-api-hml` / `ominicore-worker-hml` | quadlets `ominicore-api` / `ominicore-worker` |
| OpenBao | `kv/ominicore-hml/*`, auth `jwt-ominicore-hml`, policy/role `ominicore-hml` (só leitura) | `kv/ominicore/*`, auth `jwt-ominicore`, policy/role `ominicore` (só leitura) |
| Podman secrets | `ominicore-hml-kc-secret`, `redis-password` (compartilhado) | `ominicore-kc-secret`, `redis-password` (compartilhado) |
| Imagem | binário em `distroless/static` (sem registry) | idem |
| Observabilidade | `slog` JSON no journald | idem |

O Aurora `new_aviva` já é produção de outro produto (zillion): o role de hml
não enxerga o schema `ominicore`, e nenhum dos dois enxerga os schemas alheios.

## 4. Configuração

- **Fonte da verdade das chaves:** `deploy/local/README.md` (tabela "Variáveis").
- **Onde ficam os valores:** OpenBao (`kv/<prefixo>/postgres`, `kv/<prefixo>/sendgrid`,
  `kv/<prefixo>/twilio` — API Key + Messaging Service, `kv/<prefixo>/whatsapp` —
  token de System User + `phone_number_id`, `kv/<prefixo>/webhooks` — chave pública
  do SendGrid, Auth Token do Twilio, App Secret e verify token da Meta);
  client secret do Keycloak (`ominicore-app`) como podman secret no host (ADR 0002).
- **Quem pode ler/escrever valor de prod:** escrever — só o dono (Mauricio
  Miwa), com token de admin do OpenBao; ler — só o app, pela role `jwt` presa
  ao `azp` `ominicore-app` do realm `ominicore`.

### Divergências deliberadas

| Chave | local | hml | prod | Por quê |
|---|---|---|---|---|
| `OMINICORE_BAO_TOKEN` | `dev-root` | ausente | ausente | OpenBao local em modo dev; hml/prod entram por `jwt` (ADR 0002) |
| `OMINICORE_KC_ISSUER` | realm local | `…/ominicore-hml` | `…/ominicore` | identifica o ambiente |
| `OMINICORE_BAO_PREFIXO` | `ominicore` | `ominicore-hml` | `ominicore` | isola hml de prod no mesmo cofre |
| `OMINICORE_KC_URL` | vazio (usa o emissor) | `http://keycloak:8080` | `http://keycloak:8080` | host público tem allowlist por IP |
| `OMINICORE_DB_SCHEMA` | `ominicore` | `ominicore_hml` | `ominicore` | isola hml de prod no mesmo Aurora |
| `OMINICORE_BAO_AUTH_MOUNT` | `jwt` | `jwt-ominicore-hml` | `jwt-ominicore` | um mount `jwt` aceita um só emissor (realm) |
| `OMINICORE_BAO_ROLE` | `ominicore` | `ominicore-hml` | `ominicore` | policy por ambiente |
| `OMINICORE_REDIS_PASSWORD` | vazio | podman secret `redis-password` | idem | Redis do aviva2.0 com `requirepass` |
| `OMINICORE_WEBHOOK_URL` | vazio | `https://notify-hml.flipapi.com.br` | `https://notify.flipapi.com.br` | origem que o Twilio assina e usa na `StatusCallback` |
| credenciais SendGrid/Twilio/WhatsApp | inválidas/teste | sandbox (WhatsApp: número de teste da Meta) | produção | hml nunca envia a pessoa real |

## 5. Dados de homologação

- **Origem da carga:** só sintética — Templates do seed (migrations),
  Destinatários com endereços de teste, Provedores em sandbox.
- **Extração de prod para hml:** não existe; se um dia existir, mascarando na
  extração (`.agents/skills/ambientes/references/config-e-dados.md`).
- **Restaurar hml em prod:** proibido, sem exceção.

## 6. Promoção e aprovação

| | hml | prod |
|---|---|---|
| Artefato | binário arm64 com o SHA do commit (`deploy/aviva2/publicar.sh hml`) | **o mesmo commit** que rodou em hml (`publicar.sh prod` recusa árvore suja); o binário não embute segredo |
| Quem aprova | dono | dono, a cada promoção, depois de ver o bloco de alvo |
| Migration | `ominicore migrate` com o prefixo de hml | idem com o de prod, antes de trocar o binário; só aditiva |
| Rollback | binário anterior em `/opt/ominicore/ominicore.anterior` + restart | idem (migration aditiva não precisa voltar) |

Checklist: `.agents/skills/ambientes/references/checklist-promocao.md`.
Runbook e scripts: `deploy/aviva2/README.md`.
