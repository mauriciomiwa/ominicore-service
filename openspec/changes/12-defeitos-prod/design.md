# Design

## Context

O motivo está no `proposal.md`, em Why. Pontos do código:

- `Cofre.LerWhatsApp` (`internal/adaptadores/saida/openbao/segredos.go:133`)
  só trata `bao.ErrSecretNotFound` como desligado. Depois do *soft delete* do
  KV v2 (`ApagarWhatsApp`, linha 207), a leitura da versão atual não dá
  "não encontrado": o OpenBao devolve a versão com os dados vazios e a
  `deletion_time` preenchida. O código passa esse mapa vazio adiante e recusa
  como incompleto. O T-98 roda contra o cofre em memória
  (`internal/testes/memoria`), que apaga de verdade, e por isso não pegou.
- O quadlet tem só `Restart=always` em `[Service]`. O padrão do systemd é
  `RestartSec=100ms`, com `StartLimitBurst=5` em `StartLimitIntervalSec=10s`.
- O Console (`web/console/src/console/api.ts:39`) não trata a exceção do
  `fetch`, e a tela (`src/pages/whatsapp.vue:93`) não desliga o carregamento
  quando a promessa rejeita.

## Goals / Non-Goals

**Goals:** cada defeito com um teste que falha antes da correção.

**Non-Goals:** mudar o *soft delete* (a sabatina decidiu mantê-lo) e alertar
quando os serviços caem.

## Decisions

### D1 — Versão apagada é Canal desligado
Em `LerWhatsApp`, a resposta com `Data` vazio, ou com
`VersionMetadata.DeletionTime` preenchida, devolve `ConfigWhatsApp{}`
(desligado, REQ-20), como já acontece com `ErrSecretNotFound`. A releitura e o
boot passam pelo mesmo caminho e por isso também se corrigem. Gravar de novo
cria uma versão nova, e o Canal volta.
- *Alternativa:* destruir o caminho (`metadata delete`). Recusada na sabatina:
  perde o histórico e a recuperação.
- **Teste:** o T-98 ganha uma versão em `internal/integracao/` contra o OpenBao
  2.6.3 em modo dev (testcontainers, a mesma versão do aviva2.0). O teste
  grava, apaga pela API do OminiCore e confere a consulta "desligado" e a
  releitura sem erro. Antes da correção ele dá 503.

### D2 — Quadlet religa sem desistir
`[Unit]` ganha `StartLimitIntervalSec=0` e `[Service]` ganha `RestartSec=10s`.
O resultado é uma tentativa a cada 10 s, para sempre. Com o Keycloak, o OpenBao
ou o Aurora fora, o processo falha no boot e tenta de novo, sem estourar o
limite.
- *Alternativa:* `RestartSec=30s` com limite de 1 h. Recusada: exigiria
  intervenção humana depois de uma queda longa, que foi o problema de 16/10.

### D3 — Console com a API fora
- `criarApi.chamar` captura a exceção do `fetch` e devolve
  `{ tipo: 'falha', status: 0 }`. Os demais `!r.ok` já viram
  `{ tipo: 'falha', status }`.
- A tela mostra o texto `whatsapp.semResposta` (status 0, "A API de
  {ambiente} não respondeu") ou `whatsapp.falha` (com o status), além do botão
  "Tentar de novo", que chama `carregar()` de novo.
- O `carregar` garante o fim do carregamento em `finally`.

## Risks / Trade-offs

- **[Trade-off] Religar a cada 10 s para sempre gera log a cada falha** durante
  uma queda longa. Aceito: um log a cada 10 s é barato e dá o sinal do
  problema.
- **[Risco] A leitura de uma versão apagada mudar no OpenBao** → o teste de
  integração com a versão fixada pega isso na próxima atualização.

## Migration Plan

1. Arquivar a 11 (e, antes dela, a 08, a 09 e a 10).
2. `publicar.sh prod`, que instala o binário e o quadlet novo; o
   `systemctl daemon-reload` já faz parte do script.
3. Conferir com `systemctl show ominicore-api -p RestartUSec -p StartLimitIntervalUSec`.
4. Publicar o Console (`wrangler deploy`).

**Rollback:** `publicar.sh prod --voltar` (binário anterior). O quadlet
anterior volta com o arquivo do commit anterior.
