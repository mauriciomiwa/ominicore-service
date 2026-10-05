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

**Ligar o WhatsApp depois** — `sudo openbao/configurar-openbao.sh hml --so-whatsapp`
pergunta o Provedor (`meta`, `ultramsg` ou `zapi`); reinicie api e worker no host
(`sudo systemctl restart ominicore-hml-api.service ominicore-hml-worker.service`)
e confira o log de boot `whatsapp ligado: <provedor>`.

- **UltraMsg** (gateway não oficial, ADR 0004): uma instância e um número
  **dedicados por ambiente** (hml nunca usa o número de prod), autenticados por
  QR no painel do UltraMsg; o script pede `instance_id` e token. Não há webhook
  a cadastrar: a Tentativa fica `enviada` e `/webhooks/whatsapp` responde 404.
  Instância desconectada (QR expirado) **não dá erro** — o UltraMsg enfileira e
  só envia ao reconectar: acompanhe o status da instância no painel. Antes do
  primeiro envio real, confira o formato da resposta com um `curl` de teste
  (`DOC/Padroes/fontes-externas.md`).
- **Z-API** (gateway não oficial, ADR 0004): como no UltraMsg, uma instância e
  um número **dedicados por ambiente**, autenticados por QR no painel da Z-API;
  o script pede `instance_id`, o token da instância e, opcional, o
  **Client-Token** da conta (aba Segurança). Não há webhook a cadastrar: a
  Tentativa fica `enviada` e `/webhooks/whatsapp` responde 404. Riscos e
  cuidados:
  - **Ordem do Client-Token:** grave o `client_token` pelo script, reinicie api
    e worker e **só então** clique em "Ativar Token" no painel. Ativado antes,
    todo envio vira Falha definitiva (`null not allowed`) e cai no próximo
    Canal.
  - **Token na URL:** a Z-API exige o token da instância no caminho da chamada.
    O OminiCore não o escreve em log nem no motivo da Tentativa; se ele vazar,
    gere outro no painel da instância e rode de novo `--so-whatsapp`.
  - **Falhas depois do aceite não aparecem:** a Z-API responde 200 ao
    enfileirar; número sem WhatsApp, shadow ban ou instância desconectada só
    chegariam pelo webhook dela, que não é ouvido. Acompanhe a instância no
    painel. Antes do primeiro envio real, confira as respostas com um `curl`
    de teste (`DOC/Padroes/fontes-externas.md`).
- **Meta** (conta Meta, número verificado e Modelo `boas_vindas` aprovado):
  o script pede token, `phone_number_id`, App Secret e verify token; depois
  cadastre o webhook da Meta (etapa 5b).
- **Trocar de Provedor ou desligar (rollback):** rode de novo `--so-whatsapp`
  com o outro Provedor, ou apague o caminho
  (`kv delete kv/<prefixo>/whatsapp`, Canal desligado), e reinicie api e worker.

**Trocar o WhatsApp pela API, sem reiniciar (RN-11)** — o caminho do dia a dia
depois da primeira configuração. Uma vez por ambiente:

1. `sudo openbao/configurar-openbao.sh hml --sem-segredos`: a policy passa a
   permitir escrita **só** em `kv/<prefixo>/whatsapp` (ADR 0005).
2. Crie um client de operação (nunca um Sistema de origem comum) com
   `sudo keycloak/criar-sistema-origem.sh hml ops-<nome>` e dê a ele o papel de
   Operador: `sudo keycloak/conceder-operador.sh hml ops-<nome>`. Para revogar:
   o mesmo comando com `--revogar` (os tokens atuais expiram em até 5 min).

Depois, com o token do client de operação (`client_credentials`):

```bash
curl -s -H "Authorization: Bearer $TOKEN" https://<api interna>/v1/config/whatsapp
```

```bash
curl -s -X PUT -H "Authorization: Bearer $TOKEN" https://<api interna>/v1/config/whatsapp -d @config.json
```

```bash
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" https://<api interna>/v1/config/whatsapp
```

- O `config.json` leva `provedor` e os campos dele (ver `api/openapi.yaml`);
  monte-o num arquivo fora do repositório e apague depois — o token do
  Provedor nunca vai em argumento de linha de comando.
- A api passa a usar na hora; o worker, na próxima releitura
  (`OMINICORE_CONFIG_RELEITURA`, 30 s). Confira no log
  `whatsapp reconfigurado: <provedor>`. Configuração inválida editada à mão no
  cofre não derruba nada: o log diz "mantida a anterior".
- Cada troca fica em `config_eventos` (quem, quando, qual Provedor, sem
  segredo). Falha ao registrar = resposta 500 "gravada sem registro" e a troca
  fica no log de erro.
- **Meta pela API:** o App Secret e o verify token passam a morar em
  `kv/<prefixo>/whatsapp`. Antes de voltar para um binário anterior à fatia 10,
  regrave a Meta com `--so-whatsapp` (o binário antigo os procura em
  `webhooks`).

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
