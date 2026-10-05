# Design

## Context

- **Cofre.** `openbao.Carregar` (`internal/adaptadores/saida/openbao/segredos.go`)
  lê `kv/<prefixo>/whatsapp` exigindo `CamposWhatsApp` (`token`,
  `phone_number_id`). O caminho ausente deixa `WhatsAppLigado` falso. Com o
  Canal ligado, `CamposWebhooksMeta` passa a ser exigido em
  `kv/<prefixo>/webhooks`.
- **Remetentes.** `cmd/ominicore/main.go` (`remetentes`) registra o
  `whatsapp.Remetente` (Meta) só com o Canal ligado. `webhooks.Registrar`
  registra `GET`/`POST /webhooks/whatsapp` só com `WhatsAppLigado`.
- **Despacho.** `aplicacao/despachar.go` (`enviar`) preenche o Template com
  `Template.Preencher` e chama `Remetente.Enviar`. O Remetente da Meta já
  devolve Falha definitiva quando a versão não tem `ModeloAprovado`
  (`whatsapp.go:90`).
- **Template.** `dominio.ConteudoCanal` já tem `Corpo`, que é substituído por
  `Preencher` e entra em `Variaveis()`. A coluna `template_conteudos.corpo`
  existe, vazia na versão WhatsApp do `boas-vindas`
  (`migrations/00004_whatsapp.sql`).
- **Pedido.** `aplicacao/enviar.go` (`Pedir`) valida as variáveis e a existência
  da versão de cada Canal da ordem, com `ErrCanalSemVersao`. Os erros de
  validação viram 422 em `adaptadores/entrada/http/servidor.go:157`.

## Goals / Non-Goals

**Goals:**
- Trocar o Provedor do WhatsApp por ambiente só pelo cofre, com o mesmo binário.
- Não mexer em quem já tem a Meta no cofre: o campo `provedor` é opcional.
- Ter um só lugar que diz se a versão WhatsApp serve ao Provedor ativo, usado
  no pedido e no envio.

**Non-Goals:**
- Os dois Provedores ao mesmo tempo, ou fallback entre eles dentro do Canal.
- Trocar o Provedor a quente: como os demais segredos, ele é lido no boot.
- Aviso de status, checagem prévia de número e mídia no UltraMsg.

## Decisions

1. **`provedor` em `kv/<prefixo>/whatsapp`, com `meta` como padrão.**
   - **Como:** `Carregar` lê o caminho sem lista fixa e escolhe os campos
     exigidos pelo `provedor`: `meta` → `token` e `phone_number_id` (mais
     `CamposWebhooksMeta`); `ultramsg` → `instance_id` e `token`. Um valor
     desconhecido é erro que cita o valor. `Segredos` ganha
     `ProvedorWhatsApp dominio.ProvedorWhatsApp`.
   - **Alternativa:** um caminho por Provedor (`kv/<prefixo>/ultramsg`).
     Rejeitada porque obrigaria uma regra para "os dois caminhos existem", e o
     "caminho ausente = desligado" da change 07 deixaria de ser um ponto só.

2. **A forma exigida mora no domínio.**
   - **Como:** `dominio.ProvedorWhatsApp` (`meta`, `ultramsg`) ganha
     `Servir(ConteudoCanal) error`, que devolve `ErrVersaoWhatsAppSemForma` com
     o motivo "versão whatsapp sem texto" ou "versão whatsapp sem modelo
     aprovado".
   - **No pedido:** `SendMessageUseCase` recebe `ProvedorWhatsApp` (injetado
     como `Desligados`) e, em `Pedir`, chama `Servir` quando o WhatsApp está na
     ordem e não está desligado. O erro vira 422 junto de `ErrCanalSemVersao`.
   - **No envio:** cada Remetente verifica a própria forma, como a Meta já faz
     (`Corpo` vazio → Falha definitiva, T-77). O despacho não muda.
   - **Alternativa:** checar só no envio. Rejeitada porque aceitaria um pedido
     que já se sabe que vai falhar no WhatsApp.

3. **Novo adaptador `internal/adaptadores/saida/ultramsg`.**
   - **Envio:** `POST {URL}/{instance_id}/messages/chat`, corpo
     `application/x-www-form-urlencoded` com `token`, `to` (telefone E.164) e
     `body`, e timeout de 10 s, como os demais. `URLPadrao` é
     `https://api.ultramsg.com`, e `OMINICORE_ULTRAMSG_URL` aponta os testes
     para um UltraMsg falso (`httptest`).
   - **Limite de tamanho:** mais de 4096 runas (`utf8.RuneCountInString`) é
     Falha definitiva antes de qualquer chamada (T-76).
   - **Classificação, tolerante ao formato ainda não confirmado:**
     - erro de rede → `provedor.Inacessivel`;
     - 429 ou 5xx → Falha temporária;
     - JSON com `error` não vazio, ou `sent` falso → Falha definitiva, com
       motivo "ultramsg <status>: <error>";
     - 2xx sem `error` → Aceito, guardando o `id` como texto (aceita número ou
       string, `json.RawMessage`);
     - 2xx sem id e sem erro → Aceito com id vazio. Tratar como falha levaria o
       fallback a reenviar uma mensagem que pode ter saído.
   - **Segurança:** o token vai no corpo, nunca na URL, e não aparece no motivo.
     A `instance_id` vai no caminho, mas `Inacessivel` não repassa o erro do
     cliente HTTP.
   - **Alternativa:** JSON no corpo. A doc e o exemplo dela usam parâmetros de
     formulário, então fica o formulário.

4. **As rotas da Meta dependem do Provedor, não só do Canal.** O campo
   `webhooks.WhatsAppLigado` vira `MetaLigada`, e o `main` passa
   `seg.WhatsAppLigado && seg.ProvedorWhatsApp == meta`. Com o UltraMsg, as
   rotas não existem e respondem 404 (T-81).

5. **Seed por migration nova** (`migrations/00008_whatsapp_texto.sql`): `UPDATE` do `corpo` da versão WhatsApp do `boas-vindas` para o
   texto do SMS. O `Down` volta o campo a `''`. O schema não muda.

6. **ADR `0004-whatsapp-por-gateway-nao-oficial`.** Passa nos três critérios: é
   difícil de reverter (número e base de clientes), é surpreendente (gateway
   não oficial num sistema que exige assinatura) e houve alternativa real
   (esperar a Meta). Registra o risco aceito pelo cliente, o número dedicado
   por ambiente e o uso só transacional.

## Risks / Trade-offs

- **Bloqueio do número pelo WhatsApp** (gateway não oficial) → número
  dedicado por ambiente, só mensagens transacionais a clientes, e a Meta
  continua pronta para a troca pelo cofre. O risco foi aceito pelo cliente
  (ADR 0004).
- **Instância desconectada (QR expirado):** o UltraMsg aceita, enfileira e só
  envia ao reconectar → a Tentativa fica `enviada` sem ter saído. Mitigação
  operacional no runbook: monitorar o status da instância no painel do
  UltraMsg. O endpoint de status fica para uma fatia futura.
- **Número sem WhatsApp não é detectado:** o UltraMsg aceita e a mensagem
  nunca chega, sem fallback → risco conhecido; `contacts/check` fica fora do
  escopo.
- **Formato da resposta não confirmado** → classificação tolerante
  (Decisão 3) e conferência com a instância de trial antes do deploy em hml
  (tarefa 5.2). Se o formato divergir, só o `classificar` do adaptador muda.
- **Sem Aviso de status** → Tentativa e Notificação ficam `enviada` (nunca
  `entregue`). O Sistema de origem precisa saber disso; vai no runbook e no
  `api/openapi.yaml`.

## Migration Plan

1. Fazer o deploy do binário com o cofre intacto. Quem não tem `provedor` segue
   na Meta (ou desligado), sem mudança.
2. Em hml: `configurar-openbao.sh hml --so-whatsapp`, escolhendo `ultramsg` e
   informando a instância e o token de hml; reiniciar api e worker; conferir no
   log "whatsapp ligado: ultramsg" e enviar um `boas-vindas` real.
3. Em prod, o mesmo procedimento com a instância e o número de prod, pela skill
   `ambientes`.
4. **Rollback:** `kv delete kv/<prefixo>/whatsapp` (Canal desligado) ou voltar
   `provedor=meta` e reiniciar. A migration é reversível pelo `Down`.

## Open Questions

- O formato exato das respostas de sucesso e de erro do UltraMsg. Será
  confirmado com a instância de trial; não muda specs nem tarefas, só o teste
  de classificação.
