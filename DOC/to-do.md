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
