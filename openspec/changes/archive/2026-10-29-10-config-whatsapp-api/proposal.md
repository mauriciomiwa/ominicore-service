# Proposal

## Why

O Provedor e as credenciais do Canal WhatsApp (RN-4) mudam hoje só pelo
`configurar-openbao.sh --so-whatsapp`, rodado pelo dono no host, seguido do
reinício de api e worker, porque o cofre só é lido no boot (REQ-6). Com três
Provedores possíveis (Meta, UltraMsg e Z-API), trocar de um para outro, ou
renovar um token, virou operação frequente. A nova **RN-11** (sabatina de
2026-10-05) diz que o **Operador** faz isso pela API, sem reiniciar. As
credenciais continuam só no OpenBao (RN-9), a API nunca as devolve, e cada
troca registra quem fez e quando.

## What Changes

- **API interna de configuração** do Canal WhatsApp:
  - `GET /v1/config/whatsapp` → Provedor ativo, campos não secretos em claro e,
    para cada segredo, só se está configurado;
  - `PUT /v1/config/whatsapp` → grava a configuração inteira de `meta`,
    `ultramsg` ou `zapi`. Campo obrigatório faltando ou Provedor desconhecido
    → 422 citando o campo, sem gravar nada. A credencial não é testada no
    Provedor;
  - `DELETE /v1/config/whatsapp` → desliga o Canal.
- **Só o Operador.** As três rotas exigem o papel `ominicore-config` do
  Keycloak no token. Sem o papel → 403, inclusive para Sistema de origem; sem
  token válido → 401 (REQ-1).
- **Grava no OpenBao, só em `kv/<prefixo>/whatsapp`.** A policy do OminiCore
  ganha escrita nesse caminho e em nenhum outro (ADR 0005, que revisa a 0002).
  Com a Meta, o App Secret e o verify token do webhook passam a ser gravados
  ali. O boot os lê de lá e, na falta, de `kv/<prefixo>/webhooks`, como hoje.
- **Vale sem reiniciar.** Api e worker releem `kv/<prefixo>/whatsapp` a cada
  `OMINICORE_CONFIG_RELEITURA` (padrão 30 s) e passam a usar o Provedor novo:
  Remetente, forma exigida no pedido, Canal ligado ou desligado e rotas do
  webhook da Meta. Se a releitura encontrar uma configuração inválida (editada
  à mão), mantém a anterior e registra o erro no log.
- **Registro das trocas.** Cada `PUT` e `DELETE` grava em `config_eventos` o
  Canal, a ação, o Provedor, o Operador (`azp`) e o horário, sem segredo, e
  registra a troca no log.
- **Infra.** `configurar-openbao.sh` dá a policy de escrita. Um script novo
  cria o papel `ominicore-config` no realm e o concede a um client.

## Capabilities

### New Capabilities

- `configuracao-canais`: configuração, pela API, do Provedor e das credenciais
  de um Canal pelo Operador (hoje só o WhatsApp), com efeito sem reinício e
  registro das trocas.

### Modified Capabilities

- `segredos`: novo requisito para os segredos do webhook da Meta no caminho do
  WhatsApp, com o local antigo como alternativa. O REQ-6 não muda de texto.

## Impact

- **Dependência:** as changes `08-whatsapp-ultramsg` e `09-whatsapp-zapi`
  devem estar arquivadas antes da 10, porque o comportamento por Provedor vem
  delas.
- **Negócio e DOC:** RN-11, o termo Operador no glossário e a Fatia 10 no
  backlog (já feitos). Ficam para a implementação: a ADR 0005, o `log.md` e o
  `to-do.md` (status da instância e QR, configuração de Email e SMS por API,
  teste da credencial no Provedor).
- **Código:**
  - `internal/adaptadores/entrada/http` (rotas e checagem do papel);
  - `internal/adaptadores/saida/keycloak` (papéis do token);
  - `internal/adaptadores/saida/openbao` (gravar, apagar e reler o caminho do
    WhatsApp);
  - `internal/aplicacao` (caso de uso de configuração);
  - `cmd/ominicore` (Provedor trocável a quente em api e worker);
  - `internal/adaptadores/entrada/webhooks` (rotas da Meta decididas a cada
    requisição);
  - `internal/adaptadores/saida/postgres` (`config_eventos`).
- **Contrato:** `api/openapi.yaml` ganha `/v1/config/whatsapp`, com o
  esquema de leitura sem segredos e as respostas 401, 403 e 422. A mudança é
  aditiva.
- **Banco:** migration `00009_config_eventos.sql` (tabela nova, `Down` a
  remove).
- **Deploy:**
  - `deploy/aviva2/openbao/configurar-openbao.sh` (policy);
  - `deploy/aviva2/keycloak/` (papel e concessão);
  - `deploy/local/` (papel no realm local);
  - runbook `deploy/aviva2/README.md` e `DOC/Infra/AMBIENTES.md`.
- **Fora do escopo:**
  - status da instância e QR;
  - testar a credencial no Provedor;
  - configuração de Email e SMS por API;
  - webhooks da Z-API;
  - histórico de valores (o registro guarda só quem, quando e qual Provedor).
