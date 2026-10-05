# Design

## Context

- `openbao.Carregar` (`internal/adaptadores/saida/openbao/segredos.go`) lê
  todos os caminhos com `ler`, e campo ou caminho ausente é erro. `CamposWebhooks`
  inclui `meta_app_secret` e `meta_verify_token`.
- API e worker são processos separados (`cmd/ominicore`). Os dois chamam
  `Carregar` no boot, então leem o mesmo cofre e enxergam o mesmo estado do Canal.
- A disponibilidade de um Canal fica num só lugar,
  `dominio.Destinatario.Disponivel`: tem contato e não está em Opt-out. Quem
  usa é `PrimeiroComContato`, chamado no pedido (`aplicacao/enviar.go`) e no
  fallback (`dominio/router.go`, que também atende a falha tardia do Aviso de
  status). O Opt-out entra como campo em memória (`Destinatario.OptOut`),
  preenchido pelo caso de uso antes da decisão.
- O despacho já trata Canal sem Remetente registrado como Falha definitiva
  (`aplicacao/despachar.go:55`, motivo "canal sem provedor configurado").
- As rotas `/webhooks/whatsapp` (GET e POST) são registradas sempre em
  `webhooks.Registrar`.

## Goals / Non-Goals

**Goals:**
- Subir API e worker sem nenhuma credencial da Meta.
- Ligar o WhatsApp depois só gravando no cofre e reiniciando os processos,
  sem novo binário.
- Reaproveitar o caminho do Opt-out: o Canal desligado é mais um motivo de
  indisponibilidade, e não um fluxo à parte.

**Non-Goals:**
- Ligar e desligar o Canal a quente: o estado é lido só no boot, como os
  demais segredos.
- Tornar Email ou SMS opcionais.
- Mudar a ordem padrão de algum Template.

## Decisions

1. **O estado do Canal vem da presença de `kv/<prefixo>/whatsapp`.**
   `ler` ganha uma variante opcional. Um caminho inexistente
   (`bao.ErrSecretNotFound`) devolve "ausente", e um caminho existente com
   campo vazio continua sendo erro. `Segredos` ganha `WhatsAppLigado bool`. Os
   campos da Meta saem de `CamposWebhooks` e viram `CamposWebhooksMeta`,
   exigidos só quando o Canal está ligado. Se o Canal está desligado e esses
   campos existem, eles são ignorados.
   *Alternativa:* uma flag `OMINICORE_WHATSAPP_ATIVO`. Foi descartada na
   sabatina (P1) porque seria mais um item de configuração que pode discordar
   do cofre.

2. **`Destinatario.Desligados map[Canal]bool`, em memória e preenchido como o
   `OptOut`.** `Disponivel` passa a exigir também `!Desligados[c]`. O conjunto
   vem de `main` (`canaisDesligados(seg)`) e é injetado em
   `SendMessageUseCase`, `DespacharUseCase` e `ReceiveMessageUseCase`, que o
   copiam para o Destinatário antes de decidir. Assim o pedido, o fallback e a
   falha tardia passam a pular o Canal sem que o Router mude.
   *Alternativa:* passar o conjunto como parâmetro ao `NotificationRouter`.
   Funciona, mas espalha um segundo critério de disponibilidade que o pedido
   também precisaria repetir.

3. **Recusa: `dominio.ErrCanalIndisponivel`** ("canal %s indisponível").
   `Pedir` devolve esse erro em vez de `ErrSemContato` quando não sobra Canal
   e algum Canal da ordem está desligado. O erro cita o primeiro Canal
   desligado da ordem, o que dá "canal whatsapp indisponível" no MVP. O
   servidor HTTP mapeia o erro para 422 com o título "canal indisponível".

4. **Tentativa pendente no Canal desligado:** o despacho não registra o
   Remetente do WhatsApp quando o Canal está desligado. O ramo existente
   (Falha definitiva sem chamar o Provedor) cobre o T-68, e o Router segue a
   ordem pulando o Canal.

5. **Webhook:** `Webhooks` ganha `WhatsAppLigado bool`, e `Registrar` só
   registra as duas rotas da Meta quando ele é verdadeiro. O `ServeMux` já
   responde 404 à rota não registrada.

6. **Script do cofre:** `configurar-openbao.sh <amb>` pergunta
   "Configurar WhatsApp agora? [s/N]". Com a resposta padrão "não", não grava
   `kv/<prefixo>/whatsapp` nem os campos da Meta em `webhooks`. A opção
   `configurar-openbao.sh <amb> --so-whatsapp` grava só o token, o
   `phone_number_id` e, com `kv patch`, os dois campos da Meta, sem tocar nos
   outros segredos. Cada segredo continua sendo passado pelo stdin.

Nenhuma decisão vai para `System/decisoes/`: o caminho é reversível e segue o
padrão já usado pelo Opt-out.

## Risks / Trade-offs

- **Alguém apaga `kv/<prefixo>/whatsapp` em prod por engano** → o processo
  sobe sem o Canal, sem erro de boot, e os pedidos por WhatsApp passam a ser
  recusados com 422. Mitigação: o log de boot "whatsapp desligado" e a recusa
  explícita, nunca uma Falha silenciosa.
- **API e worker reiniciados em momentos diferentes** → durante a janela, a
  API aceita WhatsApp e o worker ainda não tem Remetente. A Tentativa vira
  Falha definitiva e o fallback segue para o próximo Canal (T-68). Mitigação:
  o `publicar.sh` reinicia os dois.
- **Ligar o Canal exige reinício** → é aceito: a leitura dos segredos no boot
  já é o padrão do REQ-6.
