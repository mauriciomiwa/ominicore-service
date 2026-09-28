# Ambientes — contrato canônico

Última revisão: `2026-09-28` — por Claude (kickoff/fatia 1), **a revisar pelo dono**.

Regra de conteúdo: **nomes e caminhos sim, valores de segredo nunca.**
Hoje só existe o ambiente local; hml e prod estão decididos na sabatina
(`log.md`, 2026-09-28) mas não criados — tudo deles é `A DEFINIR` até a trilha
de infraestrutura (`System/backlog-mvp.md`).

## 1. Quais ambientes existem

| Ambiente | Para que serve | Quem usa | Dado real? |
|---|---|---|---|
| local | desenvolvimento (`deploy/local/compose.yaml`) | dev | não |
| hml | validação antes de prod — **não criado** | A DEFINIR | não (Provedores em sandbox) |
| prod | Sistemas de origem da casa — **não criado** | Sistemas de origem | sim (PII: RN-9) |

## 2. Como cada ambiente se identifica

| Ambiente | Evidência (comando/variável) | Valor esperado |
|---|---|---|
| local | `echo $OMINICORE_BAO_TOKEN` preenchido e `OMINICORE_KC_ISSUER` em `localhost` | `dev-root` · `http://localhost:8081/realms/ominicore` |
| hml | `OMINICORE_KC_ISSUER` | `…/realms/ominicore-hml` (A DEFINIR host) |
| prod | `OMINICORE_KC_ISSUER` | `…/realms/ominicore` (A DEFINIR host) |

hml e prod vão rodar **no mesmo host** (aviva2.0): a distinção é o realm e o
path do OpenBao, não a máquina. Até existirem, toda ação no aviva2.0 é tratada
como prod.

## 3. Endereços e recursos

| Recurso | hml | prod |
|---|---|---|
| URL da aplicação | A DEFINIR (API só na rede `aviva`; webhooks públicos — ADR 0003) | A DEFINIR |
| Banco | Aurora `db-zillionvf-prod` / `new_aviva`, schema A DEFINIR | Aurora `db-zillionvf-prod` / `new_aviva`, schema `ominicore` |
| Fila | tabela `tentativas` (ADR 0001) | idem |
| Host | aviva2.0 (EC2 arm64, us-east-1), quadlets | idem |
| Path base do OpenBao | A DEFINIR | `kv/ominicore/*` |
| Imagem | binário em `distroless/static` (sem registry) | idem |
| Observabilidade | `slog` JSON no journald | idem |

## 4. Configuração

- **Fonte da verdade das chaves:** `deploy/local/README.md` (tabela "Variáveis").
- **Onde ficam os valores:** OpenBao (`kv/ominicore/postgres`, `kv/ominicore/sendgrid`);
  client secret do Keycloak como podman secret no host (ADR 0002).
- **Quem pode ler/escrever valor de prod:** A DEFINIR.

### Divergências deliberadas

| Chave | local | hml/prod | Por quê |
|---|---|---|---|
| `OMINICORE_BAO_TOKEN` | `dev-root` | ausente (login `jwt`) | OpenBao local em modo dev; ADR 0002 |

## 5. Dados de homologação

- **Origem da carga:** A DEFINIR. **Restaurar hml em prod:** proibido.

## 6. Promoção e aprovação

A DEFINIR com a trilha de infraestrutura (`to-do.md`).
