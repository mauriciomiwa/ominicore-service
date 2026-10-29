# To-Do — ominicore

Lista viva de pendências, lembretes e decisões em aberto. As fatias planejadas
ficam no `System/backlog-mvp.md` (cada uma é uma change do OpenSpec); furos de rastreabilidade entram aqui.

## Decisões em aberto
- [ ] Cadastro de Templates: seed por migration (fatia 1) basta no MVP ou precisa de API de administração?
- [ ] Mesma Chave de idempotência com pedido diferente dentro de 24 h: devolver a Notificação existente (como hoje) ou recusar como conflito?
- [ ] Prazo e orçamento do MVP.

## Fatia 1 — pendências de verificação
- [ ] VM `podman-machine-default` segue quebrada (sem identidade SSH); hoje o local roda no Docker. Recriar a VM só se o Podman for o runtime local desejado — decisão do dono.
- [ ] hml/prod: client `ominicore-app` com audience mapper `openbao` e role `jwt` com `bound_audiences=openbao` (espelhar `deploy/local/`).
- [ ] Fatia 5 (retenção): anonimizar também `tentativas.motivo` — a mensagem do Provedor pode citar o contato do Destinatário.
- [ ] Decidir se envio duplicado na janela de queda (Provedor aceitou, commit falhou) é aceitável: hoje o despacho é "pelo menos uma vez".
- [ ] Caminho do módulo Go (`ominicore` → `github.com/<org>/ominicore`) quando o repositório existir.

## Prod — operação (2026-10-27)
- [x] (change `12-defeitos-prod`: quadlet com `RestartSec=10s` e `StartLimitIntervalSec=0`; Console com "Tentar de novo"; falta publicar) api e worker ficaram parados de 16/10 a 27/10: subiram com o Keycloak fora, o systemd estourou o limite de reinícios (quadlet sem `RestartSec`) e desistiu. Corrigir no `deploy/aviva2/quadlet/ominicore.container.tmpl` (intervalo + limite) e fazer o Console mostrar erro quando a API não responde — change 12.
- [ ] Sistema de origem `wbem` criado no realm `ominicore` (prod); credenciais em `kv/aviva/wbem-ominicore` (`client_id`, `client_secret`, `token_url`, `api_url`). Falta a integração no aviva-back.
- [x] (change `12-defeitos-prod`: `LerWhatsApp` trata a versão apagada como desligado; falta publicar e o T-98 contra o OpenBao real) Defeito: "Desligar o WhatsApp" (DELETE `/v1/config/whatsapp`) faz soft delete no KV v2 e `LerWhatsApp` (`internal/adaptadores/saida/openbao/segredos.go`) lê a versão apagada como "incompleta: sem o campo token" → GET 503 no Console e worker preso na configuração anterior. Em 27/10 restaurado com `kv undelete` (versão 2). Tratar versão apagada como desligado + teste — change 12. Até lá, não usar o botão.
- [ ] Apagar o client `teste-zapi` (prod) se não for mais usado.
- [ ] `publicar.sh` à prova de execução duplicada: em 2026-10-27 dois `publicar.sh prod` sobrepostos deixaram `/opt/ominicore/ominicore.anterior` cortado no meio (segfault) — o `--voltar` derrubaria a API. Trava no host (`flock`) e cópia atômica do `.anterior` (`cp` para temporário + `mv`). Até lá: um deploy por vez. O `.anterior` foi recriado pelo dono a partir do binário atual (ac762e0, sha256 igual) em 2026-10-27.

## WAF e origem fixa (2026-10-28)
- [ ] Cloudflare: lista `escritorio` = `72.60.159.29` (túnel na Hostinger); regra de WAF ativa, ação Block, expressão da ADR 0006 — hoje não bloqueia (dono).
- [ ] Rotear `ominicore.flipapi.com.br` e `notify.flipapi.com.br` pelo túnel da Hostinger, como o `auth.flipapi.com.br` (dono).
- [ ] Repositório `flip`: `traefik/dynamic/keycloak.yml` com o `72.60.159.29`, hoje só no host.
- [ ] Túnel da Hostinger levar UDP (HTTP/3): hoje o Safari sai direto e leva "Forbidden"; o Edge (HTTP/2) passa (2026-10-29).
- [ ] Reteste: celular no 4G → 403 no Console e em `/v1/config/*`; depois fechar a 7.3 da change 11 (`verificar-exposicao.sh` no CI).

## Fatia 14 — Templates do wbem (em andamento noutra sessão, 2026-10-29)
- [x] (resolvido em 29/10: commit `1a123da` e change `14-templates-wbem` com RN-13, REQ-44, T-139 e T-140) `migrations/00011_templates_wbem.sql` (templates `wbem-link-tomador` e `wbem-otp-tomador`, só WhatsApp) está fora do Git na branch `fatia/14-templates-wbem` e **sem change no OpenSpec**: abrir a change (RN/REQ/T-n — mudança de comportamento: Templates novos) e auditar com a sentinel quando a sessão que a escreve fechar a tarefa. Auditar no meio da escrita não vale.
- [x] (resolvido em 29/10: RN-13 criada e diff auditado pela sentinel no commit `9ccf62a`) Furos da fatia 14 apontados pelo hook em 2026-10-29 (trabalho da outra sessão, ainda em escrita): REQ-44 da change `14-templates-wbem` cita a RN-13, que não existe em `especificacao-negocio.md`; `internal/integracao/integracao_test.go` alterado sem auditoria da sentinel. Fechar na sessão da fatia 14 (criar a RN-13 ou corrigir a referência; auditar o diff).
- [ ] Rastreabilidade: o REQ-21 aparece como MODIFIED na change `09-whatsapp-zapi`, mas ainda não existe na spec vigente (nasce na `08-whatsapp-ultramsg`, não arquivada). Some ao arquivar na ordem 08 → 09 → 10 → 11 → 12 → 13; antes disso, a 08 precisa da 5.2 (UltraMsg) e a 11, da 1.2 e da 7.3.

## WAF do Console (achado e corrigido em 2026-10-29)
- [x] A regra C2 e a lista `escritorio` nunca tinham sido criadas (zona no plano Free, 5 regras em uso). Corrigido pelo dono: IP do escritório na regra "Block specific paths". CI (run 40960531950, de 135.119.237.66): T-113 a T-115 com 403 da Cloudflare. Variável `CONSOLE_PUBLICADO=true` criada: a verificação diária das 06:17 passa a rodar.

## Console — deploy (2026-10-29)
- [ ] O login OAuth do `wrangler` expira em dias; trocar por um API Token da Cloudflare só com Workers (secret `CLOUDFLARE_API_TOKEN` do environment `console`) e publicar pelo `console.yml` no GitHub Actions.

## Furos de validação
- [x] `openspec validate --all --strict` reprova `spec/notificacoes`: requirements[0] e [11] têm texto > 500 caracteres (warning vira erro no strict). Encurtar movendo exemplos e bordas para cenários — numa change (ex.: junto do archive 08→11), não editando a spec vigente à mão. Visto em 2026-10-08. Causa (2026-10-27): o OpenSpec 1.15.0 passou o aviso de texto longo de INFO para WARNING, e o `--strict` reprova warning — a spec não mudou. Resolvido em 2026-10-27 pela change `ajuste-texto-notificacoes` (arquivada).

## Fatia 13 — Avisos de status da Z-API — fora do escopo
- [ ] Avisos de status do UltraMsg (formato e autenticação a levantar; a resposta do envio ainda nem foi conferida — 5.2 da 08).

## Fatia 11 — Console do Operador (change `11-console-operador`) — fora do escopo, para decidir depois
- [ ] Criar o hml (DNS `notify-hml`, realm/cofre/schema `-hml`) e devolver Homologação ao `web/console/src/console/ambientes-publicados.json` (prod foi publicado primeiro, decisão de 2026-10-06).
- [ ] Configuração por cliente da casa (W Bem, Lara), cada um com o próprio WhatsApp: change de backend própria, antes de o Console ganhar o seletor de cliente.
- [ ] Email, SMS, status da instância e QR no Console (entram quando a API tiver as rotas).
- [ ] Worker intermediário (BFF) com Cloudflare Tunnel, se a rota pública de `/v1/config/*` deixar de ser aceitável.
- [ ] Outros idiomas no Console (hoje só pt-BR).
- [ ] Tema do Keycloak com a cara do Console; hoje o login, o OTP e o perfil (nome e sobrenome pedidos no primeiro login pelo Keycloak 26) usam o tema padrão.
- [ ] Deploy (dono): IP do escritório na IP List da Cloudflare e no `ipAllowList` do `auth.flipapi.com.br` (repositório `flip`), regra de WAF, secrets `CLOUDFLARE_API_TOKEN`/`CLOUDFLARE_ACCOUNT_ID` e environment `console` no GitHub.
- [ ] Materio free fixado na 2.3.0 (01/2025), com ESLint 8 legado: migrar para o ESLint 9 (flat config) quando o template ou o Dependabot pedirem.

## Fatia 10 — Configuração do WhatsApp pela API (change `10-config-whatsapp-api`) — fora do escopo, para decidir depois
- [ ] Status da instância e QR do UltraMsg e da Z-API pela API (hoje só no painel do Provedor).
- [ ] Testar a credencial no Provedor antes de gravar (hoje só os campos são conferidos).
- [ ] Configuração de Email (SendGrid) e SMS (Twilio) pela API do Operador.

## Fatia 09 — WhatsApp pela Z-API (change `09-whatsapp-zapi`, 26/27; auditada em 2026-10-05: aprovado com ressalvas)
- [x] (decidido em 2026-10-27, change `13-webhooks-zapi`) Webhooks *Ao enviar* (`DeliveryCallback`) e *Status* da Z-API com segredo na URL: detectariam número sem WhatsApp e shadow ban depois do aceite e disparariam o fallback, mas a Z-API não documenta assinatura — exige revisar a RN-5 (aviso sem assinatura válida é descartado). Decisão do negócio.
- [x] Auditar com a sentinel o diff de código da 09 (`cmd/ominicore`, `internal/adaptadores/saida/{openbao,ultramsg,zapi}`, `internal/dominio`, `internal/testes/zapifake`, scripts de `deploy/`) e registrar o veredito quando a sessão que implementa fechar as tarefas — auditar no meio da escrita não vale (o veredito se amarra ao conteúdo atual).

## Fatia 08 — WhatsApp por UltraMsg (change `08-whatsapp-ultramsg`, 31/32; auditada em 2026-10-05: aprovado com ressalvas)
- [x] Auditar com a sentinel todo o diff de código da change (hoje em `internal/adaptadores/saida/openbao/`, `internal/adaptadores/saida/ultramsg/` e `internal/dominio/`, e crescendo) quando a implementação fechar, e registrar o veredito. É trabalho em curso de outra sessão: auditar no meio da escrita não vale (o veredito se amarra ao conteúdo atual).

## Fatia 06 — WhatsApp
Não bloqueia o deploy (Fatia 7): sem estes itens o core sobe com o WhatsApp
desligado; para ligar, `configurar-openbao.sh <amb> --so-whatsapp`.
- [ ] Conta Meta Business + WhatsApp Business Account + número verificado; Modelo aprovado `boas_vindas` (`pt_BR`, 1 parâmetro = nome). Sem isso, só testes com Meta falsa.
- [ ] Token permanente de System User e `phone_number_id` em `kv/<prefixo>/whatsapp` (dono).
- [ ] Revisar a versão da Graph API fixada no adaptador antes da data de desativação da Meta.

## Dados
- [ ] Backup do Aurora `db-zillionvf-prod` tem só 2 dias de retenção (`../flip/aws/runbook-migracao-aurora-us-east-1.md`): os Opt-outs (`opt_outs`) não podem ser perdidos — decidir retenção maior ou dump próprio do schema `ominicore` (dono).
- [ ] Conferir se a migração do Aurora para us-east-1 já aconteceu antes do primeiro deploy (muda host e latência).

## Repositório (GitHub)
- [ ] Criar o repositório no GitHub (org/nome a definir) e fazer o primeiro push.
- [ ] Ligar Dependabot alerts, security updates, grouped updates e allow auto-merge.
- [ ] Marcar `sdd-checks` (job `checagens`) e o CI de Go como obrigatórios na branch `main`.
- [ ] Preencher SLA e donos em `Infra/seguranca-dependencias.md`.

## Infraestrutura (trilha — skill `ambientes`; runbook `deploy/aviva2/README.md`)
Artefatos prontos no repositório (2026-09-29); **nada executado no host ainda**. Ordem, hml primeiro:
- [ ] 1. DNS + Cloudflare `notify-hml.flipapi.com.br` (dono).
- [ ] 2. Aurora: `deploy/aviva2/aurora/criar-schema.sql`, bloco hml (dono, DBeaver).
- [ ] 3. Keycloak: `sudo deploy/aviva2/keycloak/criar-realm.sh hml` (dono, no host).
- [ ] 4. OpenBao: `sudo deploy/aviva2/openbao/configurar-openbao.sh hml` (dono, no host; contas SendGrid/Twilio de sandbox; WhatsApp: responder N até a conta Meta existir).
- [ ] Instalar `shellcheck` local (`brew install shellcheck`) — `configurar-openbao.sh` da Fatia 7 só passou por `bash -n`.
- [ ] 5. `./deploy/aviva2/publicar.sh hml` (bloco de alvo antes).
- [ ] 6. Sistema de origem de teste em hml (`keycloak/criar-sistema-origem.sh hml <id>`) e um pedido de ponta a ponta.
- [ ] Depois, prod: mesmos passos com `prod`, só com a promoção aprovada (`AMBIENTES.md` §6).
