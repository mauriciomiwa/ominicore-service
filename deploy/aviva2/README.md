# Deploy no aviva2.0 — runbook

Host `i-044194deaab7ca807` (us-east-1, t4g/arm64), quadlets Podman na rede
`aviva`, ao lado de Traefik, Keycloak, OpenBao e Redis (config deles em
`../flip/infra/aviva2.0/`). **O host é produção** (aviva-back, Keycloak da
casa) e o Aurora `new_aviva` é produção do zillion: toda etapa segue a skill
`ambientes` — bloco de alvo antes, um ambiente por vez, **hml primeiro**.

Contrato de cada ambiente: `DOC/Infra/AMBIENTES.md`; variáveis (sem segredo)
em `ambientes/hml.sh` e `ambientes/prod.sh`.

## Ordem (uma vez por ambiente)

**Antes de tudo, uma vez por host — client de deploy no Keycloak.** O admin
humano do realm `master` tem OTP, e o `kcadm` (login direto) não passa por
OTP. Os scripts de `keycloak/` entram pelo client de serviço `ominicore-deploy`
(`keycloak/lib-admin.sh`):
1. Console admin (`https://auth.flipapi.com.br/admin`), realm **master** ›
   Clients › Create client: Client ID `ominicore-deploy`; Client authentication
   **ligado**; só **Service accounts roles** marcado (Standard flow e Direct
   access desligados).
2. Aba *Service accounts roles* › Assign role › filtro *realm roles* ›
   **`create-realm`** (cria realms e administra só os que criou; não enxerga
   `wbem`/`lara`).
3. Aba *Credentials*: copie o Client secret e grave no host (pede sem eco):
   `ssh -t -i ~/.ssh/aviva2.0.pem ubuntu@44.209.101.158 'read -r -s -p "client secret: " s; echo; printf %s "$s" | sudo podman secret create ominicore-kc-deploy -'`

| # | Etapa | Onde | Quem | Credencial |
|---|---|---|---|---|
| 1 | DNS `notify-hml.flipapi.com.br` → aviva2.0, proxy da Cloudflare ligado | Cloudflare | dono | conta Cloudflare |
| 2 | Schema e role: `aurora/criar-schema.sql` (bloco do ambiente) | DBeaver, como master | dono | senha master do Aurora |
| 3 | Realm + client `ominicore-app` + podman secret: `sudo keycloak/criar-realm.sh hml` | host | dono | nenhuma: os scripts de `keycloak/` entram como `bootstrap-admin` com o podman secret `kc-admin-password` (`lib-admin.sh`) |
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

## Console do Operador (fatia 11, RN-12)

O Console (`https://ominicore.flipapi.com.br`) é um só para os dois Ambientes:
o Operador escolhe homologação ou produção e entra no realm daquele Ambiente.
Desde a fatia 11, o `publicar.sh` instala também a rota pública de
`/v1/config/*` (ADR 0006). Por isso, **o WAF vem antes do primeiro
`publicar.sh` desta versão**, para não haver janela aberta.

| # | Etapa | Onde | Quem | Credencial |
|---|---|---|---|---|
| C1 | IP List `escritorio` com o IP fixo do escritório (Manage Account › Configurations › Lists) | Cloudflare | dono | conta Cloudflare |
| C2 | Regra de WAF (Security › WAF › Custom rules) na zona `flipapi.com.br`, ação **Block**: `(http.host eq "ominicore.flipapi.com.br" or (http.host in {"notify.flipapi.com.br" "notify-hml.flipapi.com.br"} and starts_with(http.request.uri.path, "/v1/"))) and not ip.src in $escritorio` | Cloudflare | dono | conta Cloudflare |
| C3 | IP do escritório no `keycloak-allowlist` de `../flip/infra/aviva2.0/traefik/dynamic/keycloak.yml` (repositório `flip`); sem ele, o login pelo Console é recusado | repositório `flip` + host | dono | chave SSH |
| C4 | Client público: `sudo keycloak/criar-console.sh hml` (PKCE, OTP obrigatório pelo fluxo `navegador-otp`; idempotente) | host | dono | senha admin do Keycloak |
| C5 | Uma conta por pessoa: `sudo keycloak/criar-operador.sh hml <usuario> <email>`. A senha temporária sai uma vez; no primeiro login a pessoa cadastra o OTP, troca a senha e informa nome e sobrenome. Para tirar o acesso, desabilite a pessoa no Keycloak | host | dono | senha admin do Keycloak |
| C6 | `./deploy/aviva2/publicar.sh hml` (rota `/v1/config/*` e `OMINICORE_CONSOLE_ORIGEM`) | máquina local → host | dono ou Claude, com bloco de alvo | chave SSH |
| C7 | GitHub: environment `console` com os secrets `CLOUDFLARE_API_TOKEN` (permissão Workers Scripts: Edit e Workers Routes: Edit na zona) e `CLOUDFLARE_ACCOUNT_ID`; depois Actions › console › Run workflow com "publicar" marcado (a `main` precisa estar verde); depois do primeiro deploy, a variável de repositório `CONSOLE_PUBLICADO=true` liga a verificação diária | GitHub | dono | conta Cloudflare |
| C8 | O próprio workflow roda `verificar-exposicao.sh hml` e `prod` de fora do escritório (T-113 a T-115). Do escritório: entrar, consultar e gravar em homologação | GitHub + escritório | dono | — |

**Webhooks da Z-API (fatia 13, ADR 0007).** Não se cadastram à mão: ao gravar a
Z-API pelo Console, o OminiCore gera um segredo novo e cadastra na Z-API as URLs
de "Ao enviar" e de "Status" (`https://notify.flipapi.com.br/webhooks/zapi/<segredo>`).
O Console mostra o estado; se der "não cadastrados", use "Cadastrar de novo".
Conferir no painel da Z-API (instância › Webhooks): os dois campos preenchidos
com a mesma URL; os demais (Ao receber, conexão, presença) ficam vazios.
Rollback: voltar o binário (`publicar.sh prod --voltar`) — a rota some (404) e
os avisos que chegarem são ignorados; limpar os campos no painel, se quiser.

**Acesso pela origem fixa (túnel na Hostinger, `72.60.159.29`).** O túnel leva
só o TCP dos IPs da Cloudflare do `flipapi`; o HTTP/3 (QUIC, UDP) sai direto e
aparece com o IP real — o Keycloak responde "Forbidden" e o WAF bloqueia. Desligue o
HTTP/3 (QUIC) no navegador: Edge/Chrome em `edge://flags/#enable-quic` (ou
`chrome://flags/#enable-quic`) → Disabled; Safari em *Desenvolvedor › Sinalizadores
de recurso*. Só trocar de navegador não basta: depois do primeiro acesso por
HTTP/2, a Cloudflare anuncia o HTTP/3 e o navegador passa a usá-lo. Conferir: `https://auth.flipapi.com.br/cdn-cgi/trace`
tem de mostrar `ip=72.60.159.29`.

Repita C4 a C6 em prod (skill `ambientes`). **Trocar o IP do escritório**: só
a IP List `escritorio` (C1) e o `keycloak-allowlist` (C3); até lá, o Console
fica inacessível, o que é uma falha segura.

**Rollback do Console:** `bunx wrangler rollback` em `web/console` (versão
anterior) ou `bunx wrangler delete` (tira do ar). Para fechar a configuração
de fora na hora, apague o router `ominicore<sufixo>-config` de
`/etc/traefik/dynamic/ominicore<sufixo>.yml` no host; o Traefik relê sozinho.

## Conferir que está no ar (hml)

```bash
ssh -i ~/.ssh/aviva2.0.pem ubuntu@44.209.101.158 'sudo podman ps --filter name=ominicore-hml --format "{{.Names}} {{.Status}}"'
```

`publicar.sh` só termina `ok` quando os logs têm `api no ar` e `worker no ar`
— isto é, o login `jwt` no OpenBao, a leitura dos segredos e a conexão com o
Aurora deram certo.

**Religar depois de uma queda (change 12).** O quadlet tenta a cada 10 s e não
desiste (`RestartSec=10s`, `StartLimitIntervalSec=0`): se o Keycloak, o OpenBao
ou o Aurora caírem, api e worker voltam sozinhos quando eles voltarem.
Conferir depois do `publicar.sh`:

```bash
ssh -i ~/.ssh/aviva2.0.pem ubuntu@44.209.101.158 'systemctl show ominicore-api ominicore-worker -p RestartUSec -p StartLimitIntervalUSec'
```

Esperado: `RestartUSec=10s` e `StartLimitIntervalUSec=0` nos dois.

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
| `traefik/ominicore.yml.tmpl` | `/etc/traefik/dynamic/ominicore<sufixo>.yml` (`/webhooks/*` e `/v1/config/*`, ADR 0006) |
| binário | `/opt/ominicore<sufixo>/ominicore` (+ `.anterior`) |
