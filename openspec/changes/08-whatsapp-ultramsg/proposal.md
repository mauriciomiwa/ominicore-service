# Proposal

## Why

O Canal WhatsApp (RN-4) só tem a Cloud API da Meta, e a conta Meta Business com
o Modelo `boas_vindas` aprovado ainda não existe. Hoje ele está pronto no código
e desligado nos ambientes (change 07). O cliente quer o WhatsApp já, pelo
UltraMsg (`POST https://api.ultramsg.com/{instance_id}/messages/chat`, campos
`token`, `to` e `body` em texto livre de até 4096 caracteres). O UltraMsg é um
gateway **não oficial** sobre o WhatsApp Web: um número comum, autenticado por
QR. O cliente aceitou o risco de bloqueio do número (sabatina de 2026-10-05).

Isso muda a RN-4. Hoje ela diz "só por Modelo aprovado (sem texto livre)" e
passaria a dizer:

> Canal WhatsApp por um de dois Provedores, escolhido por ambiente: a Cloud API
> da Meta, só por Modelo aprovado, ou o UltraMsg, por texto da versão WhatsApp
> do Template preenchido com as variáveis. O envio vai sempre ao telefone do
> Destinatário. Na Meta, número sem WhatsApp ou Modelo não aprovado é Falha
> definitiva; no UltraMsg, credencial ou instância recusada é Falha definitiva.
> Sem as credenciais do Provedor no ambiente, o Canal fica desligado (como hoje).

A RN-5 não muda. O Aviso de status exige assinatura válida, e o UltraMsg não
documenta assinatura. Por isso a Tentativa aceita por ele fica `enviada`.

## What Changes

- **Provedor do WhatsApp por ambiente.** `kv/<prefixo>/whatsapp` ganha o campo
  `provedor` (`meta` ou `ultramsg`). Se o campo faltar, vale `meta`, e quem já
  tem a Meta no cofre não muda nada. Com `ultramsg`, os campos exigidos são
  `instance_id` e `token`, e os segredos de webhook da Meta deixam de ser
  exigidos. Um valor desconhecido ou um campo faltando impede o início.
  O caminho ausente continua deixando o Canal desligado.
- **Envio pelo UltraMsg.** O sistema manda o texto da versão WhatsApp do
  Template, preenchido com as variáveis, ao telefone do Destinatário.
  Resultado: 2xx com id → Tentativa enviada com o id do Provedor. 429 ou 5xx →
  Falha temporária. Token ou instância recusados, ou qualquer outro erro do
  Provedor → Falha definitiva. Texto preenchido com mais de 4096 caracteres →
  Falha definitiva, sem chamar o Provedor.
- **Versão WhatsApp do Template.** Ela pode ter o Modelo aprovado (Meta), o
  texto (UltraMsg) ou os dois. O pedido com o Canal WhatsApp na ordem é recusado
  se a versão não tiver a forma do Provedor ativo, com o motivo "versão whatsapp
  sem texto" ou "versão whatsapp sem modelo aprovado". A Tentativa que já estava
  pendente quando o Provedor foi trocado vira Falha definitiva e segue a ordem.
- **Sem Aviso de status do UltraMsg.** Com `ultramsg`, `/webhooks/whatsapp`
  responde 404, como com o Canal desligado.
- **Seed.** A versão WhatsApp do `boas-vindas` ganha o texto do SMS: "Olá
  {{nome}}, sua conta está pronta."
- **Script do cofre.** O `configurar-openbao.sh` pergunta qual Provedor do
  WhatsApp configurar, e `--so-whatsapp` passa a aceitar o UltraMsg.

## Capabilities

### New Capabilities

Nenhuma.

### Modified Capabilities

- `segredos`: o REQ-6 passa a escolher o Provedor do WhatsApp pelo cofre, com
  os campos exigidos de cada Provedor.
- `notificacoes`: o REQ-10 (envio) e o REQ-11 (versão WhatsApp) ganham o
  UltraMsg. O novo REQ-21 cobre o Provedor do WhatsApp sem Aviso de status.

## Impact

- **Negócio e DOC:** RN-4 em `DOC/Negocio/especificacao-negocio.md`. No
  glossário, "Provedor" passa a citar a Meta e o UltraMsg. Nova decisão em
  `DOC/System/decisoes/0004-whatsapp-por-gateway-nao-oficial.md` com o risco
  aceito pelo cliente. `DOC/Padroes/fontes-externas.md` ganha o UltraMsg e
  `DOC/System/backlog-mvp.md` a Fatia 8.
- **Código:**
  - novo `internal/adaptadores/saida/ultramsg`;
  - `internal/adaptadores/saida/openbao` (campo `provedor` e campos de cada um);
  - `cmd/ominicore` (escolha do Remetente e registro das rotas da Meta);
  - `internal/dominio` (versão WhatsApp com texto);
  - `internal/aplicacao` (recusa pela forma do Provedor ativo);
  - `internal/adaptadores/entrada/http` (motivos da recusa).
- **Contrato:** `api/openapi.yaml` documenta os novos motivos da recusa 422.
- **Banco:** a migration nova só preenche `corpo` na versão WhatsApp do
  `boas-vindas`; o schema não muda.
- **Deploy:** `deploy/aviva2/openbao/configurar-openbao.sh` e o runbook
  `deploy/aviva2/README.md` mudam. Hml e prod usam instâncias e números
  UltraMsg distintos. No ambiente local, o `deploy/local/bootstrap.sh` passa a
  aceitar `WHATSAPP_PROVEDOR=ultramsg` com instância e token por variável;
  sem ela, segue gravando a Meta com valores inválidos, como hoje.
- **Suposição a validar em hml:** a página do endpoint não mostra exemplos de
  resposta. A classificação será tolerante: id na resposta → aceito; campo
  `error` → Falha definitiva, salvo 429 e 5xx. O formato real será conferido
  com a instância de trial antes do deploy.
- **Fora do escopo:**
  - Aviso de status do UltraMsg;
  - checagem prévia de número sem WhatsApp (`contacts/check`);
  - Opt-out de WhatsApp por sinal do Provedor (continua só pela API);
  - respostas do Destinatário (RN-6).
