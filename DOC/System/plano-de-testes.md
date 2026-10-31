# Plano de testes — ominicore

Um teste por cenário marcado com T-n nas specs do OpenSpec
(`#### Scenario: T-n …` em `openspec/specs/` ou nos deltas de
`openspec/changes/`). A spec diz *o que* verificar e a qual REQ o T-n
pertence; aqui fica só *como* e *em que pé está* — sem repetir o REQ, para não
haver dois lugares a manter. Tipo: **Exemplo** (caso concreto) ou
**Propriedade** (invariante). Status: pendente / escrito (não rodado) /
passando.

T-1 a T-62 vigentes desde 2026-10-02 e T-63 a T-69 desde 2026-10-05 (`openspec/specs/{autenticacao-origem,notificacoes,segredos}`); T-70 a T-81 na change ativa `08-whatsapp-ultramsg`; T-82 a T-91 na change ativa `09-whatsapp-zapi`; T-92 a T-105 na change ativa `10-config-whatsapp-api`; T-106 a T-120 na change ativa `11-console-operador`; T-121 a T-136 na change ativa `13-webhooks-zapi`; T-137 e T-138 na change ativa `12-defeitos-prod`.

Console do Operador (`web/console`): Vitest + `@vue/test-utils` + happy-dom,
`bun run test`; o nome do teste começa pelo T-n (`it('T-106 …')`). T-113 a
T-115 são verificações de borda (`curl` contra os hosts reais) que só valem de
fora do escritório: rodam no GitHub Actions depois do deploy e uma vez por dia.

Nome do teste em Go cita o T-n: `func TestT7_ReenvioComAMesmaChaveNaoDuplica(t *testing.T)`.
Unitários com fakes (`internal/testes/`: repositório em memória, realm JWT,
SendGrid via `httptest`); integração em `internal/integracao/` com
`testcontainers-go` (Postgres real) — pula sozinho sem runtime de containers;
rode com `make itest` (Docker Desktop, Colima ou Podman: usa o socket do contexto ativo). Última execução: 2026-10-02, 30/30 contra Postgres 17 (Colima). Provedores reais nunca entram em teste.

| Teste | Tipo | Onde (arquivo do teste) | Status |
|---|---|---|---|
| T-1 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-2 | Exemplo | `internal/adaptadores/saida/keycloak/validador_test.go` | passando |
| T-3 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-4 | Exemplo | `internal/aplicacao/aplicacao_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-5 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-6 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-7 | Exemplo | `internal/integracao/integracao_test.go` | passando |
| T-8 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-9 | Exemplo | `internal/integracao/integracao_test.go` | passando |
| T-10 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-11 | Exemplo | `internal/integracao/integracao_test.go` | passando |
| T-12 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-13 | Exemplo | `cmd/ominicore/main_test.go` | passando |
| T-14 | Propriedade | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-15 | Exemplo | `internal/aplicacao/aplicacao_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-16 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-17 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-18 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-19 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-20 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` · `internal/adaptadores/saida/twilio/twilio_test.go` | passando |
| T-21 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-22 | Propriedade | `internal/dominio/dominio_test.go` | passando |
| T-23 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres, relógio adiantado) | passando (domínio e Postgres) |
| T-24 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-25 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-26 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-27 | Exemplo | `internal/dominio/dominio_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (domínio e Postgres) |
| T-28 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-29 | Exemplo | `internal/adaptadores/saida/whatsapp/whatsapp_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-30 | Exemplo | `internal/adaptadores/saida/whatsapp/whatsapp_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-31 | Exemplo | `internal/adaptadores/saida/whatsapp/whatsapp_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-32 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-33 | Exemplo | `internal/aplicacao/aplicacao_test.go` | passando |
| T-34 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-35 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-36 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-37 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-38 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-39 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-40 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-41 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` · `internal/integracao/integracao_test.go` (Postgres; + 8 avisos simultâneos = 1 fallback) | passando (fake e Postgres) |
| T-42 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-43 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-44 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-45 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-46 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-47 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-48 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-49 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-50 | Exemplo | `internal/adaptadores/entrada/worker/worker_test.go` | passando |
| T-51 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-52 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-53 | Exemplo | `internal/aplicacao/aplicacao_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-54 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-55 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-56 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-57 | Exemplo | `internal/aplicacao/aplicacao_test.go` (fake) · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres) |
| T-58 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-59 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-60 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-61 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-62 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-63 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` · `cmd/ominicore/main_test.go` | passando |
| T-64 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` | passando |
| T-65 | Exemplo | `internal/aplicacao/desligado_test.go` · `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-66 | Exemplo | `internal/aplicacao/desligado_test.go` | passando |
| T-67 | Exemplo | `internal/aplicacao/desligado_test.go` | passando |
| T-68 | Exemplo | `internal/aplicacao/desligado_test.go` | passando |
| T-69 | Exemplo | `internal/adaptadores/entrada/webhooks/webhooks_test.go` | passando |
| T-70 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` · `cmd/ominicore/main_test.go` | passando (change 08 ativa) |
| T-71 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` | passando (change 08 ativa) |
| T-72 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` | passando (change 08 ativa) |
| T-73 | Exemplo | `internal/adaptadores/saida/ultramsg/ultramsg_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (fake e Postgres; formato real da resposta a conferir em hml) |
| T-74 | Exemplo | `internal/adaptadores/saida/ultramsg/ultramsg_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 08 ativa) |
| T-75 | Exemplo | `internal/adaptadores/saida/ultramsg/ultramsg_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 08 ativa) |
| T-76 | Exemplo | `internal/adaptadores/saida/ultramsg/ultramsg_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 08 ativa) |
| T-77 | Exemplo | `internal/adaptadores/saida/ultramsg/ultramsg_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 08 ativa) |
| T-78 | Exemplo | `internal/aplicacao/provedor_whatsapp_test.go` · `internal/adaptadores/entrada/http/servidor_test.go` | passando (change 08 ativa) |
| T-79 | Exemplo | `internal/aplicacao/provedor_whatsapp_test.go` | passando (change 08 ativa) |
| T-80 | Exemplo | `internal/aplicacao/provedor_whatsapp_test.go` | passando (change 08 ativa) |
| T-81 | Exemplo | `cmd/ominicore/main_test.go` | passando (change 08 ativa) |
| T-82 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` · `cmd/ominicore/main_test.go` | passando (change 09 ativa) |
| T-83 | Exemplo | `internal/adaptadores/saida/openbao/segredos_test.go` | passando (change 09 ativa) |
| T-84 | Exemplo | `internal/adaptadores/saida/zapi/zapi_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 09 ativa) |
| T-85 | Exemplo | `internal/adaptadores/saida/zapi/zapi_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (fake; respostas reais a conferir em hml) |
| T-86 | Exemplo | `internal/adaptadores/saida/zapi/zapi_test.go` | passando (change 09 ativa) |
| T-87 | Exemplo | `internal/adaptadores/saida/zapi/zapi_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (fake; corpo da recusa real a conferir em hml) |
| T-88 | Exemplo | `internal/adaptadores/saida/zapi/zapi_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 09 ativa) |
| T-89 | Exemplo | `internal/adaptadores/saida/zapi/zapi_test.go` · `internal/adaptadores/entrada/worker/worker_test.go` | passando (change 09 ativa) |
| T-90 | Exemplo | `internal/aplicacao/provedor_whatsapp_test.go` · `internal/adaptadores/entrada/http/servidor_test.go` | passando (change 09 ativa) |
| T-91 | Exemplo | `cmd/ominicore/main_test.go` | passando (change 09 ativa) |
| T-92 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-93 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-94 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-95 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-96 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-97 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-98 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` (memória) · `internal/adaptadores/saida/openbao/segredos_test.go` (KV v2 falso com soft delete) · `internal/integracao/openbao_test.go` (OpenBao 2.6.3 real) | passando (memória, KV falso e OpenBao 2.6.3 real; `make itest` 2026-10-27, Docker Desktop) |
| T-99 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-100 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando (change 10 ativa) |
| T-101 | Exemplo | `cmd/ominicore/config_test.go` | passando (change 10 ativa) |
| T-102 | Exemplo | `cmd/ominicore/config_test.go` | passando (change 10 ativa) |
| T-103 | Exemplo | `cmd/ominicore/config_test.go` | passando (change 10 ativa) |
| T-104 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` · `internal/integracao/integracao_test.go` (Postgres) | passando (change 10 ativa) |
| T-105 | Exemplo | `internal/adaptadores/saida/openbao/cofre_test.go` | passando (change 10 ativa) |
| T-106 | Exemplo | `web/console/src/console/sessao.spec.ts` (Vitest, `oidc-client-ts` real) | passando |
| T-107 | Exemplo | `web/console/src/console/sessao.spec.ts` (Vitest) | passando |
| T-108 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-109 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-110 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-111 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-112 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-113 | Exemplo | `deploy/aviva2/verificar-exposicao.sh` (CI `console.yml`, de fora do escritório) | escrito (roda depois do deploy) |
| T-114 | Exemplo | `deploy/aviva2/verificar-exposicao.sh` (CI `console.yml`, de fora do escritório) | escrito (roda depois do deploy) |
| T-115 | Exemplo | `deploy/aviva2/verificar-exposicao.sh` (CI `console.yml`, de fora do escritório) | escrito (roda depois do deploy) |
| T-116 | Exemplo | `internal/adaptadores/entrada/http/cors_test.go` | passando |
| T-117 | Exemplo | `internal/adaptadores/entrada/http/cors_test.go` | passando |
| T-118 | Exemplo | `internal/integracao/keycloak_test.go` (testcontainers, Keycloak 26.3) | passando |
| T-119 | Exemplo | `internal/adaptadores/entrada/http/servidor_test.go` | passando |
| T-120 | Exemplo | `internal/adaptadores/entrada/http/config_test.go` | passando |
| T-121 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-122 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-123 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-124 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-125 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-126 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-127 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-128 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` · `internal/integracao/zapi_test.go` (Postgres) | passando (memória e Postgres; `make itest` 2026-10-29) |
| T-129 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-130 | Exemplo | `internal/adaptadores/entrada/webhooks/zapi_test.go` | passando |
| T-131 | Exemplo | `internal/adaptadores/entrada/http/webhooks_zapi_test.go` | passando |
| T-132 | Exemplo | `internal/adaptadores/entrada/http/webhooks_zapi_test.go` | passando |
| T-133 | Exemplo | `internal/adaptadores/entrada/http/webhooks_zapi_test.go` | passando |
| T-134 | Exemplo | `internal/adaptadores/entrada/http/webhooks_zapi_test.go` | passando |
| T-135 | Exemplo | `internal/adaptadores/entrada/http/webhooks_zapi_test.go` | passando |
| T-136 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-137 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-138 | Exemplo | `web/console/src/pages/whatsapp.spec.ts` (Vitest) | passando |
| T-139 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-140 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-157 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
| T-158 | Exemplo | `internal/integracao/integracao_test.go` (Postgres) | passando |
