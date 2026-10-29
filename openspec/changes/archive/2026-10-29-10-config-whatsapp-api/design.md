# Design

## Context

- **Estado do Canal congelado no boot.** `cmd/ominicore` lê o cofre uma vez
  (`openbao.Carregar`) e espalha cópias:
  - `canaisDesligados(seg)` vai para `SendMessageUseCase`, `DespacharUseCase`
    e `ReceiveMessageUseCase` (campo `Desligados`);
  - `seg.ProvedorWhatsApp` vai para `SendMessageUseCase`;
  - `remetentes(cfg, seg)` vai para o worker;
  - `metaLigada(seg)` decide, no `webhooks.Registrar`, se as rotas da Meta
    **existem** no mux;
  - os segredos da Meta entram em `webhooks.Credenciais`.

  Nada disso muda depois do boot.
- **Cofre.** O login é JWT do Keycloak + role com policy **só de leitura**
  (ADR 0002), com `token_ttl` de 1 h. `lerWhatsApp` já trata o caminho
  ausente (`ErrSecretNotFound`) como Canal desligado. Os campos exigidos por
  Provedor estão em `openbao.CamposWhatsApp` e `CamposOpcionaisWhatsApp`. Os
  segredos da Meta estão em `kv/<prefixo>/webhooks`.
- **Autenticação.** `keycloak.Validador.SistemaOrigem` devolve só o `azp`;
  papéis do token não são lidos. O realm é um por ambiente do OminiCore
  (`criar-realm.sh`), e os clients usam `client_credentials`.
- **Eventos.** `opt_out_eventos` (migration 00006) é o modelo de registro de
  "quem fez o quê" sem dado sensível.

## Goals / Non-Goals

**Goals:**
- Um único ponto que diz qual é a configuração atual do WhatsApp, consultado
  por pedido, envio, Aviso de status e webhooks, e trocado a quente.
- A escrita no cofre restrita a um caminho, sem ampliar o que a app lê.
- Validar a configuração com as **mesmas** regras do boot.

**Non-Goals:**
- Notificar o outro processo na hora da troca (pub/sub). A releitura
  periódica basta.
- Versionar ou desfazer configurações pela API. O KV v2 guarda versões, mas
  não as expomos.
- Configurar Email e SMS.

## Decisions

1. **Regras de configuração no domínio.**
   - **Como:** `dominio.ConfigWhatsApp{Ligado, Provedor, Campos}`, com:
     - `Validar()`, que usa o mapa de campos exigidos e opcionais por
       Provedor, movido de `openbao` para o `dominio`;
     - `SegredosDe(Provedor)`: `token`, `client_token`, `meta_app_secret`,
       `meta_verify_token`. Os demais campos (`instance_id`,
       `phone_number_id`) são públicos.
   - **Quem usa:** o boot, a releitura e o `PUT` chamam a mesma validação.
   - **Alternativa:** validar no handler HTTP. Rejeitada porque duplicaria a
     regra do boot.

2. **Configuração atual trocável: `configwa.Atual`.**
   - **O que guarda:** um `atomic.Pointer` com o retrato
     `{Config, Remetente}`.
   - **O que expõe:**
     - `Atual()` devolve o retrato;
     - `Aplicar(ConfigWhatsApp)` monta o Remetente do Provedor e troca o
       ponteiro.
   - **Consumidores:**
     - os casos de uso trocam o mapa `Desligados` e o `ProvedorWhatsApp`
       fixos por uma porta `portas.EstadoCanais` (`Desligados()`,
       `ProvedorWhatsApp()`), lida a cada pedido;
     - o worker recebe um Remetente de WhatsApp que delega ao
       `Atual().Remetente`;
     - o webhook registra sempre as rotas da Meta, e cada requisição consulta
       `Atual()`: sem Meta ativa responde 404 (REQ-21, T-103), com Meta usa os
       segredos do retrato.
   - **Alternativa:** reiniciar o processo ao detectar mudança. Rejeitada na
     sabatina (P4).

3. **Releitura periódica.**
   - **Como:** um laço em api e worker relê só `kv/<prefixo>/whatsapp` a cada
     `OMINICORE_CONFIG_RELEITURA` (padrão `30s`, `time.ParseDuration`).
     - Configuração válida e diferente da atual → `Aplicar` e log
       "whatsapp reconfigurado: <provedor>".
     - Inválida → mantém a atual e registra o erro citando o campo (T-102).
     - Cofre inacessível → mantém a atual e registra o erro.
   - **Sessão no cofre:** o token do login JWT vale 1 h. O cliente do cofre
     guarda a sessão e, em 403 ou token expirado, refaz o login (novo JWT do
     Keycloak) uma vez antes de dar erro.
   - **Na api:** depois de um `PUT` ou `DELETE` bem-sucedido, a api aplica a
     configuração na hora; o worker a pega na próxima releitura.

4. **Escrita no cofre, policy mínima (ADR 0005, revisa a 0002).**
   - **Policy:**
     - leitura em `kv/data/<p>/*`, como hoje;
     - `create`, `update` e `delete` só em `kv/data/<p>/whatsapp`.
   - **Operações:**
     - `PUT` = `KVv2.Put` com o mapa inteiro, que substitui a versão;
     - `DELETE` = `KVv2.Delete`, um soft delete da versão atual, depois do
       qual o `Get` devolve `ErrSecretNotFound` e o Canal fica desligado.
       As versões antigas continuam no cofre como histórico de quem tem acesso
       de admin.
   - **Alternativa:** `kv/metadata` com `destroy`. Rejeitada porque apagaria o
     histórico e exigiria mais permissão.

5. **Segredos da Meta no caminho do WhatsApp (REQ-28).**
   - O `PUT` com `meta` exige e grava `meta_app_secret` e `meta_verify_token`
     em `kv/<p>/whatsapp`.
   - O boot e a releitura leem esses campos de lá e, se faltarem, de
     `kv/<p>/webhooks`. Quem foi configurado pelo script antigo não muda.
   - `CamposWebhooksMeta` deixa de ser exigido em `webhooks` quando os campos
     estão no caminho do WhatsApp.

6. **Papel do Operador: realm role `ominicore-config`.**
   - **Leitura:** o Validador passa a ler também `realm_access.roles` do
     token. Surge `Identidade{AZP, Papeis}`, e `SistemaOrigem` continua
     existindo para as rotas atuais.
   - **Na rota:** o middleware `exigirPapel("ominicore-config")` responde 403
     (`problem+json` "sem permissão") sem chamar o caso de uso.
   - **Por que realm role:** o realm é exclusivo do OminiCore por ambiente, e
     o papel aparece no token de qualquer client sem mapper extra. Um client
     role exigiria um client "ominicore" que hoje não existe, porque o público
     vem de um audience mapper.
   - **Concessão:** `deploy/aviva2/keycloak/conceder-operador.sh <client-id>`
     cria o papel (idempotente) e o atribui ao service account do client.

7. **Contrato HTTP.**
   - **`PUT /v1/config/whatsapp`**:
     - corpo JSON plano `{"provedor": "zapi", "instance_id": "...", "token": "...", "client_token": "..."}`;
     - só os campos conhecidos do Provedor são gravados, e os demais são
       ignorados;
     - resposta 200 com a mesma visão do `GET`; 422 com o motivo.
   - **`GET /v1/config/whatsapp`**:
     - responde `{"ligado": true, "provedor": "zapi", "campos": {"instance_id": "inst1"}, "segredos": {"token": true, "client_token": false}}`;
     - com o Canal desligado: `{"ligado": false}`.
   - **`DELETE /v1/config/whatsapp`**: 204, idempotente.
   - **Erros:** 401 e 403 como em REQ-1 e REQ-25.

8. **Registro das trocas: `config_eventos`.**
   - **Tabela** (migration 00009): `id`, `canal`, `acao` (`gravada` |
     `apagada`), `provedor` (nulo no `apagada`), `por` (azp), `em`.
   - **Ordem:** grava o cofre, depois o evento.
     - Se o evento falhar, a api registra em log, nível erro, a ação, o
       Provedor e o Operador (o log vira o registro) e responde 500 "gravada
       sem registro". Não desfaz o cofre, porque o desfazer também pode falhar
       e deixaria um estado pior.
   - **Alternativa:** evento antes do cofre. Rejeitada porque registraria uma
     troca que não aconteceu.

## Risks / Trade-offs

- **Credencial do Operador vazada** → quem a tiver troca o número do
  WhatsApp de todos.
  - Mitigação:
    - papel dado só a clients de operação, nunca a Sistemas de origem;
    - API interna (ADR 0003);
    - toda troca registrada;
    - revogar = tirar o papel ou desabilitar o client.
- **A app passa a escrever no cofre** → um bug ou uma invasão da app pode
  apagar ou trocar a configuração do WhatsApp. O dano fica limitado a esse
  caminho pela policy, e o KV v2 guarda as versões anteriores para
  recuperação manual.
- **Janela de até 30 s entre api e worker:** o pedido aceito pela api com o
  Provedor novo pode encontrar o worker no antigo. A forma exigida é checada
  de novo no envio (T-77), então a Tentativa falha e segue a ordem em vez de
  sair errada.
- **Releitura a cada 30 s por processo** → carga desprezível no OpenBao. Um
  re-login a cada hora é normal.
- **Cofre fora do ar** → o Canal continua com a última configuração boa.
  Uma troca feita nesse intervalo não chega ao worker até o cofre voltar.

## Migration Plan

1. Arquivar a 08 e a 09.
2. Deploy em hml, com a skill `ambientes`:
   - rodar `configurar-openbao.sh hml --sem-segredos`, que agora inclui a
     policy de escrita;
   - rodar `conceder-operador.sh hml <client-de-operacao>`;
   - fazer o deploy do binário; a migration 00009 é aditiva.
3. Conferir em hml:
   - `GET` sem o papel → 403;
   - `PUT` da Z-API → a próxima Tentativa sai por ela sem reinício;
   - `config_eventos` com o registro.
4. Em prod, o mesmo procedimento.
5. **Rollback:**
   - voltar o binário anterior, que ignora a tabela nova e lê `whatsapp` como
     antes;
   - a policy de escrita pode ficar ou ser retirada pelo script;
   - a migration tem `Down` (`DROP TABLE config_eventos`).
   - A Meta configurada pela API, com os segredos em `whatsapp`, não sobe no
     binário antigo. Antes de voltar para ele, gravar a Meta pelo script
     antigo.
