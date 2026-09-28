# To-Do — ominicore

Lista viva de pendências, lembretes e decisões em aberto. As fatias planejadas
ficam no `System/backlog-mvp.md` (cada uma é uma change do OpenSpec); furos de rastreabilidade entram aqui.

## Decisões em aberto
- [ ] Cadastro de Templates: seed por migration (fatia 1) basta no MVP ou precisa de API de administração?
- [ ] Consulta de status pelo Sistema de origem (`GET /v1/notificacoes/{id}`) ou aviso de volta (callback)? Decidir antes da Fatia 3.
- [ ] Mesma Chave de idempotência com pedido diferente dentro de 24 h: devolver a Notificação existente (como hoje) ou recusar como conflito?
- [ ] Prazo e orçamento do MVP.

## Fatia 1 — pendências de verificação
- [ ] **Runtime de containers:** a VM `podman-machine-default` não sobe (falta a identidade SSH `~/.local/share/containers/podman/machine/machine`). Consertar recriando a VM (`podman machine rm` + `podman machine init`), decisão do dono.
- [ ] Rodar `make itest` e ver T-4, T-7, T-9, T-10, T-11, T-12 passarem contra Postgres real (hoje escritos e pulando); só então marcar 1.4, 3.5, 3.6, 3.8, 4.1 e 4.4 no `tasks.md`.
- [ ] Subir `make compose-up` uma vez e seguir o `deploy/local/README.md` de ponta a ponta (tarefa 1.3); exercita também o login `jwt` do OpenBao contra o Keycloak local (tarefa 5.1 — hoje só o caminho por token e a falha fechada estão testados).
- [ ] Fatia 5 (retenção): anonimizar também `tentativas.motivo` — a mensagem do Provedor pode citar o contato do Destinatário.
- [ ] Decidir se envio duplicado na janela de queda (Provedor aceitou, commit falhou) é aceitável: hoje o despacho é "pelo menos uma vez".
- [ ] Caminho do módulo Go (`ominicore` → `github.com/<org>/ominicore`) quando o repositório existir.

## Repositório (GitHub)
- [ ] Criar o repositório no GitHub (org/nome a definir) e fazer o primeiro push.
- [ ] Ligar Dependabot alerts, security updates, grouped updates e allow auto-merge.
- [ ] Marcar `sdd-checks` (job `checagens`) e o CI de Go como obrigatórios na branch `main`.
- [ ] Preencher SLA e donos em `Infra/seguranca-dependencias.md`.

## Infraestrutura (trilha — skill `ambientes`)
- [ ] Realms `ominicore` e `ominicore-hml` no Keycloak; client `ominicore-app` (login no Bao) e um client por Sistema de origem.
- [ ] OpenBao: auth `jwt` com o Keycloak como emissor, role presa ao `azp` `ominicore-app`, policy leitura em `kv/ominicore/*`; gravar credenciais SendGrid/Twilio.
- [ ] Aurora `new_aviva`: schema e role `ominicore`.
- [ ] Quadlets `ominicore-api` e `ominicore-worker` na rede `aviva`; rota pública só `/webhooks/*` no Traefik; `deploy/aviva2/publicar.sh`.
- [ ] Completar os `A DEFINIR` de `Infra/AMBIENTES.md` (hml/prod) antes do primeiro deploy.
- [ ] Contas SendGrid e Twilio (sandbox para hml).
