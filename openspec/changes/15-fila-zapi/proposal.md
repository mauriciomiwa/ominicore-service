# Proposal

## Why

Com a instância da Z-API desconectada, as Tentativas aceitas ficam presas na
fila da própria Z-API: o OminiCore as marca como enviadas e o Operador não tem
como ver o que está parado, nem por quê, nem tirar da fila uma mensagem que
não deve mais sair (um código de acesso vencido, por exemplo). A sabatina de
2026-10-30 estendeu a **RN-12** (o que o Operador faz pelo Console): ver a
conexão da instância e a Fila do Provedor e retirar um item dela. Uma Tentativa
retirada é falha depois do aceite e segue a **RN-5** (fallback para o próximo
Canal).

## What Changes

- **RN-12 estendida** em `DOC/Negocio/especificacao-negocio.md`: "o Operador vê
  a conexão do WhatsApp e a Fila do Provedor e retira um item dela"; termo
  **Fila do Provedor** no glossário.
- **API de configuração**, só para o Operador e só com a Z-API ativa:
  - `GET /v1/config/whatsapp/fila` devolve a conexão da instância (conectado,
    desconectado ou celular sem internet) e até 30 itens por página, com cursor
    para a próxima. Cada item traz o id do item, o telefone mascarado, quando
    entrou e, se for de uma Tentativa do OminiCore, a Notificação, o Template e
    o Sistema de origem. **Nunca** o texto nem o telefone completo.
  - `DELETE /v1/config/whatsapp/fila/{id}` pede a retirada à Z-API. Só se ela
    confirmar, a Tentativa do item fica falhou, com a classificação nova
    **removida pelo Operador**, e segue o fallback da RN-5. A retirada fica
    registrada com o Operador e o id do item. Item que já saiu da fila: nada
    muda e a resposta diz isso. Item fora do OminiCore: é retirado e
    registrado, sem Tentativa a mudar.
  - Z-API inacessível ou recusando a credencial: erro com o status da Z-API.
- **Consulta da Notificação (REQ-14 modificado):** a classificação da falha
  ganha "removida pelo Operador". É aditiva no contrato do Sistema de origem.
- **Console:** na tela do WhatsApp, com a Z-API ativa, aparecem a conexão e a
  Fila, com "Atualizar", "Carregar mais" e "Retirar". "Retirar" pede
  confirmação citando produção (REQ-31). A tela não consulta sozinha
  (sem polling) e não reconecta a instância (sem QR code).

## Capabilities

### New Capabilities

### Modified Capabilities

- `configuracao-canais`: consulta da conexão e da Fila da Z-API e retirada de
  um item.
- `console-operador`: a conexão e a Fila na tela do WhatsApp e o "Retirar".
- `notificacoes`: REQ-14 com a classificação "removida pelo Operador".

## Impact

- **Código:**
  - `internal/adaptadores/saida/zapi`: status, fila e retirada;
  - `internal/aplicacao`: caso de uso da Fila, que reaproveita a aplicação do
    Aviso de status para a falha;
  - `internal/adaptadores/saida/postgres`: Tentativas por id do Provedor e
    registro da retirada;
  - `internal/adaptadores/entrada/http`: duas rotas e CORS;
  - `internal/testes/zapifake`: fila e status falsos;
  - `web/console`: seção da Fila, i18n.
- **Banco:** migration aditiva em `config_eventos` (ação `fila_retirada` e o id
  do item).
- **Contrato:** `api/openapi.yaml` ganha as duas rotas e o valor
  `removida_pelo_operador` na classificação da falha.
- **Borda:** nada novo. As rotas ficam sob `/v1/config/*`, já roteado pelo
  Traefik e fechado ao escritório pela regra da Cloudflare (ADR 0006).
- **A conferir com a instância real** (a documentação da Z-API não traz os
  métodos nem os caminhos completos):
  - `POST /queue` (cursor), `GET /status` e `DELETE /queue/{zaapId}`;
  - se o `MessageId` do item da fila é o `messageId` do envio;
  - o que a Z-API responde ao retirar um item que já saiu.
- **Fora do escopo:**
  - esvaziar a fila inteira;
  - reconectar a instância ou mostrar o QR code;
  - consultar sozinho de tempos em tempos;
  - fila do UltraMsg e da Meta.
