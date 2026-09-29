# To-Do — ominicore

Lista viva de pendências, lembretes e decisões em aberto. As fatias planejadas
ficam no `System/backlog-mvp.md` (cada uma é uma change do OpenSpec); furos de rastreabilidade entram aqui.

## Decisões em aberto
- [ ] Cadastro de Templates: seed por migration (fatia 1) basta no MVP ou precisa de API de administração?
- [ ] Consulta de status pelo Sistema de origem (`GET /v1/notificacoes/{id}`) ou aviso de volta (callback)? Decidir antes da Fatia 3.
- [ ] Mesma Chave de idempotência com pedido diferente dentro de 24 h: devolver a Notificação existente (como hoje) ou recusar como conflito?
- [ ] Prazo e orçamento do MVP.

## Fatia 1 — pendências de verificação
- [ ] VM `podman-machine-default` segue quebrada (sem identidade SSH); hoje o local roda no Docker. Recriar a VM só se o Podman for o runtime local desejado — decisão do dono.
- [ ] Arquivar a change: `/opsx:archive 01-enviar-email` (30/30 tarefas).
- [ ] hml/prod: client `ominicore-app` com audience mapper `openbao` e role `jwt` com `bound_audiences=openbao` (espelhar `deploy/local/`).
- [ ] Fatia 5 (retenção): anonimizar também `tentativas.motivo` — a mensagem do Provedor pode citar o contato do Destinatário.
- [ ] Decidir se envio duplicado na janela de queda (Provedor aceitou, commit falhou) é aceitável: hoje o despacho é "pelo menos uma vez".
- [ ] Caminho do módulo Go (`ominicore` → `github.com/<org>/ominicore`) quando o repositório existir.

## Repositório (GitHub)
- [ ] Criar o repositório no GitHub (org/nome a definir) e fazer o primeiro push.
- [ ] Ligar Dependabot alerts, security updates, grouped updates e allow auto-merge.
- [ ] Marcar `sdd-checks` (job `checagens`) e o CI de Go como obrigatórios na branch `main`.
- [ ] Preencher SLA e donos em `Infra/seguranca-dependencias.md`.

## Infraestrutura (trilha — skill `ambientes`)
- [ ] DNS + Cloudflare para `notify.flipapi.com.br` e `notify-hml.flipapi.com.br` (só `/webhooks/*`).
- [ ] Aurora: schema/role `ominicore_hml` além de `ominicore`, sem acesso cruzado.
- [ ] Realms `ominicore` e `ominicore-hml` no Keycloak; client `ominicore-app` (login no Bao) e um client por Sistema de origem.
- [ ] OpenBao: auth `jwt` com o Keycloak como emissor, role presa ao `azp` `ominicore-app`, policies `ominicore` (`kv/ominicore/*`) e `ominicore-hml` (`kv/ominicore-hml/*`), só leitura; gravar credenciais SendGrid/Twilio.
- [ ] Aurora `new_aviva`: schema e role `ominicore`.
- [ ] Quadlets `ominicore-api` e `ominicore-worker` na rede `aviva`; rota pública só `/webhooks/*` no Traefik; `deploy/aviva2/publicar.sh`.
- [ ] Contas SendGrid e Twilio (sandbox para hml).
