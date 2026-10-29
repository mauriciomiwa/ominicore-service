# Design

## Context

O ponto de partida é o código da change 08, ainda não arquivada:

- **Cofre.** `openbao.CamposWhatsApp` (`internal/adaptadores/saida/openbao/segredos.go`)
  mapeia cada Provedor aos seus campos exigidos. `Segredos.WhatsApp` guarda os
  campos lidos, e `CamposWebhooksMeta` só é exigido com `meta`.
- **Domínio.** `dominio.ProvedorWhatsApp` (`internal/dominio/provedor_whatsapp.go`)
  tem `meta` e `ultramsg`. `Servir` decide a forma assim: UltraMsg exige
  `Corpo`, e qualquer outro Provedor exige `ModeloAprovado`.
- **Remetentes.** `remetentes` (`cmd/ominicore/main.go`) escolhe o Remetente
  por `ProvedorWhatsApp`. `metaLigada` só registra as rotas da Meta com `meta`,
  então qualquer outro Provedor já responde 404 em `/webhooks/whatsapp`.
- **UltraMsg.** `internal/adaptadores/saida/ultramsg` checa a forma
  (`Servir`), aplica o limite de 4096 runas e classifica a resposta com
  `provedor.Classificar` e `provedor.Inacessivel`, que não repassa o erro do
  cliente HTTP.
- **Z-API, pela doc** (`developer.z-api.io/message/send-text.md` e
  `security/client-token.md`, lidas em 2026-10-05):
  - `POST https://api.z-api.io/instances/{instanceId}/token/{token}/send-text`
    com JSON `{"phone","message"}`, telefone só com dígitos;
  - 200 devolve `{"zaapId","messageId","id"}`;
  - o header opcional `Client-Token`, quando ativado na conta, passa a ser
    obrigatório, e a falta dele dá `{"error":"null not allowed"}`;
  - a doc não traz exemplo de resposta para token errado.

## Goals / Non-Goals

**Goals:**
- Pôr a Z-API como terceiro Provedor só pelo cofre, com o mesmo binário e sem
  mudar o comportamento de Meta e UltraMsg.
- Ter um só lugar para as regras dos Provedores de texto (forma e limite de
  4096), usado pelo UltraMsg e pela Z-API.
- Nunca deixar o token, que vai na URL, chegar ao motivo da Tentativa ou ao
  log.

**Non-Goals:**
- Webhooks da Z-API (*Ao enviar*, *Status*, *Desconectado*).
- `delayMessage`/`delayTyping`, edição de mensagem, mídia e grupos.
- Configurar o webhook da instância pela API da Z-API.

## Decisions

1. **`provedor=zapi` no mesmo caminho do cofre.**
   - **Como:** `CamposWhatsApp[zapi] = {"instance_id", "token"}`.
     `client_token` fica fora da lista de exigidos e é lido se existir (já
     entra em `Segredos.WhatsApp`).
   - **Alternativa:** exigir `client_token`. Rejeitada na sabatina (P4): a
     conta pode não ter o recurso ativado.

2. **"Provedor de texto" no domínio.**
   - **Como:** `ProvedorWhatsApp` ganha `WhatsAppZAPI = "zapi"` e o método
     `EnviaTexto() bool`, verdadeiro para UltraMsg e Z-API. `Servir` e
     `VersaoWhatsAppSemForma.Error` passam a usá-lo em vez de comparar com
     `WhatsAppUltraMsg`. O limite de 4096 vira `dominio.LimiteTextoWhatsApp`,
     com uma checagem compartilhada, e o UltraMsg passa a usá-la.
   - **Efeito:** trocar UltraMsg por Z-API não invalida Tentativas pendentes
     (mesma forma), e o pedido continua recusado com "versão whatsapp sem
     texto" (T-90).
   - **Alternativa:** um `case` a mais em cada ponto. Rejeitada porque
     espalharia a regra "texto até 4096" por dois adaptadores e pelo domínio.

3. **Novo adaptador `internal/adaptadores/saida/zapi`, nos moldes do UltraMsg.**
   - **Envio:**
     - `POST {URL}/instances/{instance_id}/token/{token}/send-text`, com os dois
       segmentos passados por `url.PathEscape`;
     - corpo JSON `{"phone": <E.164 sem "+">, "message": <texto>}`;
     - `Client-Token` no header só quando configurado;
     - timeout de 10 s;
     - `URLPadrao` é `https://api.z-api.io`, e `OMINICORE_ZAPI_URL` aponta os
       testes para um falso (`internal/testes/zapifake`, `httptest`).
   - **Classificação:**
     - erro de rede → `provedor.Inacessivel("zapi")`, sem o `err`, porque o
       `*url.Error` do Go traz a URL com o token (T-84);
     - 429 ou 5xx → Falha temporária;
     - campo `error` não vazio, ou outro status fora de 2xx → Falha definitiva
       com o motivo "zapi <status>: <error>";
     - 2xx sem `error` → Aceito, com id = `messageId`, ou `id` se aquele faltar;
     - 2xx sem id → Aceito com id vazio. É o mesmo argumento do UltraMsg: não
       reenviar pelo fallback uma mensagem que pode ter saído.
   - **Motivo:** o motivo só usa o status e o texto do `error` do corpo, nunca a
     URL ou os headers da requisição.
   - **Alternativa:** um adaptador genérico de "gateway de texto"
     parametrizado. Rejeitada porque os contratos diferem (formulário × JSON,
     token no corpo × na URL, header extra), e o parâmetro esconderia isso.

4. **Rotas da Meta sem mudança.** `metaLigada` já devolve falso para `zapi`,
   então T-91 só precisa de teste.

5. **ADR: adendo na 0004, sem ADR nova.** O risco do gateway não oficial já
   está decidido ali. O adendo registra que vale também para a Z-API e que
   continua a regra de um número dedicado por ambiente.

## Risks / Trade-offs

- **Falha assíncrona invisível.** A Z-API responde 200 ao enfileirar. Número
  sem WhatsApp, shadow ban ou timeout só chegam pelo webhook *Ao enviar*, que
  não vamos ouvir. A Tentativa fica `enviada` e o fallback não ocorre. A mesma
  classe de risco já existe no UltraMsg. Os webhooks com segredo na URL vão
  para o `to-do.md`, junto da revisão da RN-5.
- **Token na URL.** Ele pode ficar em logs de proxy ou do lado da Z-API. A
  saída é direta para `api.z-api.io` por HTTPS. O token não entra em motivo nem
  em log nosso (T-84), e o runbook ensina a rotacioná-lo no painel.
- **Client-Token ativado no painel antes de estar no cofre** faz todo envio
  virar Falha definitiva. Ordem no runbook: gravar `client_token` no cofre,
  reiniciar e só então ativar no painel.
- **Instância desconectada (QR expirado)**, igual ao UltraMsg. Mitigação
  operacional: monitorar a instância no painel.
- **Formato de erro não documentado** para token errado, além do exemplo
  `null not allowed` → a classificação é tolerante (Decisão 3). A conferência
  com uma instância real entra como tarefa antes do deploy em hml.

## Migration Plan

1. Arquivar a change 08 (os deltas desta partem dos textos dela).
2. Fazer o deploy do binário com o cofre intacto: nada muda para quem está em
   `meta` ou `ultramsg`.
3. Em hml: `configurar-openbao.sh hml --so-whatsapp`, escolhendo `zapi` e
   informando a instância, o token e, se houver, o Client-Token. Depois,
   reiniciar api e worker, conferir no log "whatsapp ligado: zapi" e enviar um
   `boas-vindas` real.
4. Em prod, o mesmo procedimento com a instância de prod, pela skill
   `ambientes`.
5. **Rollback:** voltar `provedor` para o valor anterior, ou `kv delete` do
   caminho (Canal desligado), e reiniciar. Não há migration.

## Open Questions

- O corpo exato da recusa por token ou instância inválidos, e o status HTTP
  que a acompanha. Será conferido com uma instância real. Não muda specs nem
  tarefas, só o teste de classificação do adaptador.
