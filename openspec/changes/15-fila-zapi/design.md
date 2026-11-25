# Design

## Context

- **Z-API em produção** desde 2026-10-08. O envio guarda o `messageId`
  devolvido por `send-text` como `IDProvedor` da Tentativa
  (`internal/adaptadores/saida/zapi/zapi.go`). Os Avisos de status chegam pela
  rota da change 13 e são aplicados por `ReceiveMessageUseCase.Receber`
  (`internal/aplicacao/receber.go`), que trava a Tentativa, decide com
  `Router.DecidirAviso` (falha depois do aceite → próximo Canal) e grava.
- **Credenciais da Z-API**: ficam no OpenBao e são lidas por
  `ConfigurarUseCase.Consultar` (`internal/aplicacao/configurar.go`). O
  recadastro dos webhooks (REQ-41) já mostra o padrão "só com a Z-API ativa"
  com `ErrRecadastroSoZAPI` → 409.
- **Registro de trocas**: vai para `config_eventos` (`migrations/00009`, com o
  CHECK alargado em `00010`), sem coluna para o id do item.
- **Classificação da falha**: é contrato público, em `api/openapi.yaml:454`
  (enum). Não há CHECK no banco.
- **Rotas `/v1/config/*`**: o Traefik já as roteia, e a Cloudflare as fecha ao
  IP do escritório (regra "Block specific paths", ADR 0006).
- **API da Z-API**, pela documentação (que não traz os métodos nem os caminhos
  completos):
  - `POST /queue`, com `pageSize` até 30 e `pagingState`, devolve
    `messages[]`, `pagingState` e `hasMore`. Cada item traz `ZaapId`,
    `MessageId`, `Phone`, `Message` e `Created`;
  - `GET /queue` está depreciado;
  - `GET /status` devolve `connected`, `smartphoneConnected` e `error`;
  - `DELETE /queue/{zaapId}` responde 200.

## Goals / Non-Goals

**Goals:**
- O Operador vê a conexão e a Fila do Provedor e retira um item, sem nunca
  receber o texto do item nem o telefone completo.
- A retirada confirmada vira falha depois do aceite e reaproveita o mesmo
  caminho do Aviso de status.

**Non-Goals:**
- esvaziar a fila inteira;
- reconectar a instância ou mostrar o QR code;
- polling;
- fila do UltraMsg e da Meta;
- guardar o `zaapId` na Tentativa.

## Decisions

1. **Porta nova `FilaProvedorPort`**, implementada pelo pacote `zapi` com as
   mesmas `Credenciais`, o timeout de 10 s e a máscara de token do
   `Remetente`:
   - `Conexao(ctx, cfg) (dominio.ConexaoWhatsApp, error)`;
   - `Listar(ctx, cfg, cursor) (itens, proximoCursor, haMais, error)`;
   - `Retirar(ctx, cfg, zaapID) error`.

   Os erros da Z-API (status fora de 2xx, rede) viram `ErrProvedorSemResposta`,
   com o status. O motivo nunca leva a URL, que contém o token (como em
   T-84). *Alternativa*: chamar a Z-API direto do handler HTTP. Foi descartada
   porque mistura borda com domínio e não dá para testar com o `zapifake`.

2. **Conexão em três estados no domínio**:
   - `connected` → conectado;
   - `!connected` → desconectado;
   - `connected && !smartphoneConnected` → celular sem internet.

   O texto de `error` da Z-API vai junto, como detalhe, sem tradução; o
   Console traduz só o estado.

3. **`FilaUseCase` em `internal/aplicacao`**, que depende de
   `ConfigurarUseCase.Consultar`, da porta, do repositório e do
   `ReceiveMessageUseCase`:
   - **Consultar**:
     - sem a Z-API ativa → `ErrFilaSoZAPI` (409);
     - chama `Conexao` e `Listar`, nessa ordem; qualquer falha → 502;
     - busca de uma vez as Tentativas pelos `MessageId` da página, com a
       consulta nova `TentativasPorIDProvedor(ids)`, que devolve Notificação,
       Template e Sistema de origem;
     - devolve itens com o telefone **mascarado no servidor**
       (`+55 11 *****-6095`), para o número completo nunca chegar ao
       navegador.
   - **Retirar(zaapId, por)**:
     - percorre a fila pelo cursor até achar o `zaapId`, no máximo 20 páginas
       (600 itens);
     - não achou → `ErrJaSaiuDaFila` (404), sem chamar o `DELETE`;
     - achou → chama `Retirar`; 200 → registra `fila_retirada` com `por` e o
       `zaapId` e, se o `MessageId` for de uma Tentativa, aplica
       `Receber(AvisoStatus{IDProvedor: MessageId, Status: falhou, Categoria:
       FalhaRemovidaPeloOperador, Motivo: "removida da fila pelo Operador"})`.

   Assim a falha usa a mesma trava, a mesma regra de avanço (`Avanca`) e o
   mesmo fallback do Aviso (REQ-13), e o T-149 sai de graça, porque falhou é
   final. *Alternativa*: o Console mandar o `MessageId` junto com o `zaapId`.
   Foi descartada porque o servidor marcaria falhou numa Tentativa indicada
   pelo cliente. A busca na fila também resolve a corrida da P8 sem depender
   de como a Z-API responde a um `zaapId` que não existe.

4. **Categoria nova `removida_pelo_operador`** (`dominio.FalhaRemovidaPeloOperador`),
   acrescentada ao enum do OpenAPI. A mudança é aditiva. *Alternativa*: usar
   `recusado_pelo_provedor`. Foi descartada porque mentiria ao Sistema de
   origem sobre quem recusou.

5. **Registro**: a migration `00014_config_eventos_fila.sql` acrescenta a
   coluna `item text` (nula) e alarga o CHECK com `fila_retirada`, como a
   `00010`. O `Down` apaga essas linhas antes de estreitar o CHECK.
   `EventoConfig` ganha `Item`. A falha no registro depois de a Z-API
   confirmar segue o padrão `ErrTrocaSemRegistro`: a retirada valeu, e o log
   diz que o registro falhou.

6. **Rotas** (com `operador(...)` e preflight CORS como as de config):
   - `GET /v1/config/whatsapp/fila?cursor=…` →
     `200 {conexao: {estado, detalhe}, itens: [{id, telefone, entrou_em,
     notificacao_id?, template_id?, sistema_origem?}], cursor, ha_mais}`;
   - `DELETE /v1/config/whatsapp/fila/{id}` → 204, 404 (já saiu), 409 (não é
     Z-API) ou 502 (`{status}` da Z-API).

   Item fora do OminiCore vem sem `notificacao_id`.

7. **Console**: uma seção nova em `whatsapp.vue`, num componente
   `FilaZapi.vue` para não inchar a página (379 linhas), só quando `zapiAtiva`.
   - `criarApi` ganha `consultarFila(cursor?)` e `retirarDaFila(id)`.
   - "Retirar" usa o `DialogoConfirmar` que já cita produção (REQ-31).
   - O tempo na fila é calculado no navegador a partir de `entrou_em`.
   - Todos os textos ficam em `locales/pt-BR.json`.

8. **`zapifake`** ganha `POST /queue` (com `pagingState`), `GET /status` e
   `DELETE /queue/{id}`, com fila e conexão configuráveis pelo teste.

## Risks / Trade-offs

- **[Risco] Rotas da Z-API diferentes da documentação.**
  *Mitigação*: antes do deploy, um script só de leitura (`GET /status`,
  `POST /queue`) contra a instância real confere o método, o caminho e se o
  `MessageId` da fila é o `messageId` do envio. O `DELETE` só é provado em
  prod, com a permissão do dono, quando houver um item real na fila.
- **[Trade-off] A busca pela fila custa até 20 chamadas numa fila grande.**
  É aceitável: a fila normal é pequena, e passar disso responde 404 como "já
  saiu", o que é o pior caso seguro, porque não marca nada.
- **[Risco] Sistema de origem com enum fechado** quebrar com o valor novo.
  *Mitigação*: confirmar com o aviva-back, antes do deploy, se ele lê a
  classificação da falha e aceita um valor novo.
- **[Trade-off] O texto do item nunca sai do servidor.** O Operador não vê o
  conteúdo; identifica o item pelo Template e pela Notificação.
