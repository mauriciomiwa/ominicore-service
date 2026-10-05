# Proposal

## Why

O deploy do core (Email e SMS) em hml e prod está travado pela conta da Meta.
Hoje o processo não inicia sem `kv/<prefixo>/whatsapp` e sem
`meta_app_secret`/`meta_verify_token` (REQ-6, RN-9). Com valores falsos no
cofre, o pedido por WhatsApp seria aceito e chegaria à Meta só para voltar
com 401. A RN-4 passa a admitir o **Canal desligado**: sem as credenciais da
Meta, o sistema sobe sem o WhatsApp e o liga depois, quando elas forem
gravadas no cofre.

## What Changes

- **Boot**: a ausência total de `kv/<prefixo>/whatsapp` deixa o Canal
  WhatsApp desligado e o boot registra em log "whatsapp desligado". O caminho
  existente com campo faltando, ou o WhatsApp ligado sem os segredos da Meta
  em `kv/<prefixo>/webhooks`, continua impedindo o início. Postgres,
  SendGrid, Twilio, os segredos de webhook do SendGrid e do Twilio e a chave
  do Opt-out continuam obrigatórios.
- **Pedido**: o Canal desligado é pulado na Ordem de Canais, como o Canal sem
  contato e o Canal em Opt-out. Se não sobrar Canal disponível e algum Canal
  pedido estiver desligado, a recusa é 422 com o motivo
  "canal whatsapp indisponível".
- **Fallback e envio**: o fallback pula o Canal desligado. Uma Tentativa que
  já estava pendente no Canal quando ele foi desligado vira Falha definitiva
  e a Notificação segue a ordem.
- **Webhook**: com o Canal desligado, `GET` e `POST /webhooks/whatsapp`
  respondem 404.
- **Script do cofre**: o `configurar-openbao.sh` pergunta
  "Configurar WhatsApp agora? [s/N]" e ganha `--so-whatsapp`, que grava só o
  token, o `phone_number_id` e os dois segredos da Meta, mais tarde.
- A ordem padrão do Template `boas-vindas` não muda: ele já não inclui
  WhatsApp.

## Capabilities

### New Capabilities

Nenhuma.

### Modified Capabilities

- `segredos`: o REQ-6 deixa de exigir as credenciais da Meta para iniciar.
  Elas passam a ligar ou desligar o Canal WhatsApp, e o caminho incompleto
  continua sendo erro.
- `notificacoes`: o novo REQ-20 cobre o Canal desligado no pedido, no
  fallback, no envio e no webhook da Meta.

## Impact

- **Código**: `internal/adaptadores/saida/openbao` (leitura opcional),
  `cmd/ominicore` (registro do Remetente e das rotas conforme o Canal),
  `internal/dominio` e `internal/aplicacao` (disponibilidade do Canal),
  `internal/adaptadores/entrada/webhooks` (rotas da Meta) e
  `internal/adaptadores/entrada/http` (motivo da recusa).
- **Contrato**: `api/openapi.yaml` documenta o novo motivo da recusa 422.
- **Deploy**: `deploy/aviva2/openbao/configurar-openbao.sh` e o runbook
  `deploy/aviva2/README.md` (etapas 4 e 5b, que passam a ser opcionais para a
  Meta). O ambiente local (`deploy/local/bootstrap.sh`) segue com o WhatsApp
  ligado, apontando para a Meta falsa.
- **Fora do escopo**: as decisões abertas do `to-do.md` (chave repetida com
  pedido diferente, envio "pelo menos uma vez", cadastro de Templates).
