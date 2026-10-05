# Deploy no aviva2.0 — runbook

Host `i-044194deaab7ca807` (us-east-1, t4g/arm64), quadlets Podman na rede
`aviva`, ao lado de Traefik, Keycloak, OpenBao e Redis (config deles em
`../flip/infra/aviva2.0/`). **O host é produção** (aviva-back, Keycloak da
casa) e o Aurora `new_aviva` é produção do zillion: toda etapa segue a skill
`ambientes` — bloco de alvo antes, um ambiente por vez, **hml primeiro**.

Contrato de cada ambiente: `DOC/Infra/AMBIENTES.md`; variáveis (sem segredo)
em `ambientes/hml.sh` e `ambientes/prod.sh`.

## Ordem (uma vez por ambiente)

| # | Etapa | Onde | Quem | Credencial |
|---|---|---|---|---|
| 1 | DNS `notify-hml.flipapi.com.br` → aviva2.0, proxy da Cloudflare ligado | Cloudflare | dono | conta Cloudflare |
| 2 | Schema e role: `aurora/criar-schema.sql` (bloco do ambiente) | DBeaver, como master | dono | senha master do Aurora |
| 3 | Realm + client `ominicore-app` + podman secret: `sudo keycloak/criar-realm.sh hml` | host | dono | senha admin do Keycloak |
| 4 | Cofre: `sudo openbao/configurar-openbao.sh hml` (auth `jwt-ominicore-hml`, policy, role, segredos). "Configurar WhatsApp agora? [s/N]": sem conta Meta, responda N — o core sobe com o WhatsApp desligado | host | dono | token admin do OpenBao; chaves SendGrid/Twilio de sandbox (Meta só se ligar o WhatsApp) |
| 5 | `./deploy/aviva2/publicar.sh hml` (build arm64, migrate, quadlets, rota) | máquina local → host | dono ou Claude, com bloco de alvo | chave SSH `~/.ssh/aviva2.0.pem` |
| 5b | Cadastrar os webhooks nos painéis: SendGrid Event Webhook → `https://notify-hml.flipapi.com.br/webhooks/sendgrid` (eventos delivered, bounce, dropped; **Signed Event Webhook** ligado); Meta (só com o WhatsApp ligado) → `https://notify-hml.flipapi.com.br/webhooks/whatsapp` com o verify token do passo 4 (campo `messages`). O Twilio não precisa: cada SMS leva a `StatusCallback` | painéis SendGrid e Meta | dono | contas dos Provedores |
| 6 | Sistemas de origem: `sudo keycloak/criar-sistema-origem.sh hml <client-id>` | host | dono | senha admin do Keycloak |

**Ligar o WhatsApp depois** (conta Meta, número verificado e Modelo `boas_vindas`
aprovado): `sudo openbao/configurar-openbao.sh hml --so-whatsapp`, reiniciar
api e worker no host (`sudo systemctl restart ominicore-hml-api.service ominicore-hml-worker.service`; o log de boot deixa de dizer
`whatsapp desligado`) e cadastrar o webhook da Meta (etapa 5b).

Para copiar os scripts ao host: `scp -i ~/.ssh/aviva2.0.pem -r deploy/aviva2 ubuntu@44.209.101.158:/tmp/`.

Credencial nunca vai para o repositório, para o chat nem para argumento de
linha de comando: os scripts pedem no terminal (`read -s`) e gravam direto no
OpenBao ou em podman secret.

## Conferir que está no ar (hml)

```bash
ssh -i ~/.ssh/aviva2.0.pem ubuntu@44.209.101.158 'sudo podman ps --filter name=ominicore-hml --format "{{.Names}} {{.Status}}"'
```

`publicar.sh` só termina `ok` quando os logs têm `api no ar` e `worker no ar`
— isto é, o login `jwt` no OpenBao, a leitura dos segredos e a conexão com o
Aurora deram certo.

## Chave do Opt-out — não rotacionar

`kv/<prefixo>/optout` guarda a chave do HMAC dos contatos em Opt-out (RN-8,
REQ-18). O OminiCore não guarda o contato em claro, então **trocar a chave faz
todos os Opt-outs pararem de valer** — as pessoas voltariam a receber. O
`configurar-openbao.sh` cria a chave uma vez e nunca a sobrescreve; incluí-la
nos snapshots do OpenBao (`openbao-snapshot.timer`) é o backup dela.

## Rollback

`./deploy/aviva2/publicar.sh hml --voltar` troca pelo binário anterior. As
migrations são só aditivas (`DOC/Infra/AMBIENTES.md` §6), então o binário
anterior roda sobre o schema novo.

## O que cada peça gera no host

| Arquivo | Destino |
|---|---|
| `quadlet/ominicore.container.tmpl` | `/etc/containers/systemd/ominicore<sufixo>-{api,worker}.container` |
| `ominicore.env.tmpl` | `/etc/ominicore<sufixo>/ominicore.env` |
| `traefik/ominicore.yml.tmpl` | `/etc/traefik/dynamic/ominicore<sufixo>.yml` (só `/webhooks/*`) |
| binário | `/opt/ominicore<sufixo>/ominicore` (+ `.anterior`) |
