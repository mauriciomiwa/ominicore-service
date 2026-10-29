## Purpose

Dá ao Operador uma página web, aberta só a partir do escritório, para escolher o
Ambiente, entrar com a própria conta e configurar o Canal pela API, sem nunca
ver segredos e sem abrir o resto da API à internet.

## ADDED Requirements

### Requirement: REQ-29 Entrar no Ambiente escolhido
Quando o Operador escolher um Ambiente no Console, o Console SHALL levá-lo ao
login do Keycloak daquele Ambiente e só chamar a API daquele Ambiente com o
token obtido ali; o Console MUST NOT guardar segredo de client nem gravar o
token em armazenamento persistente do navegador.
Origem: RN-12

#### Scenario: T-106 Escolher homologação leva ao realm de homologação
- **WHEN** o Operador abre o Console sem sessão e escolhe homologação
- **THEN** é levado ao login do realm `ominicore-hml` com o client `ominicore-console` e desafio PKCE, sem segredo de client, e depois do login as chamadas vão só para `notify-hml.flipapi.com.br` com o token desse realm, e nenhum token (de acesso, de atualização ou de identidade) fica gravado em `localStorage`, `sessionStorage` ou cookie

#### Scenario: T-107 Trocar de Ambiente exige novo login
- **WHEN** o Operador está em homologação com sessão e troca para produção
- **THEN** o token de homologação é descartado, o Console leva ao login do realm `ominicore` e nenhuma chamada à API de produção sai antes desse login

### Requirement: REQ-30 Console sem o papel não configura
Se o Operador entrar com uma conta sem o papel de configuração, o Console SHALL
mostrar que falta permissão e não mostrar nenhuma configuração.
Origem: RN-12, RN-11

#### Scenario: T-108 Conta sem o papel vê falta de permissão
- **WHEN** uma conta sem o papel `ominicore-config` entra no Console e a API responde 403 à consulta
- **THEN** o Console mostra a mensagem de falta de permissão e nenhum formulário de configuração

### Requirement: REQ-31 Produção sempre destacada
Enquanto o Ambiente escolhido for produção, o Console SHALL mostrar em toda tela
o nome do Ambiente com destaque de alerta, e SHALL pedir confirmação citando
produção antes de gravar ou desligar.
Origem: RN-12

#### Scenario: T-109 Produção mostra o destaque e pede confirmação
- **WHEN** o Operador está em produção e grava uma configuração
- **THEN** toda tela mostra "Produção" com o destaque de alerta, e a gravação só é enviada depois de ele confirmar uma pergunta que cita produção; em homologação o destaque não aparece

### Requirement: REQ-32 Configurar o WhatsApp pelo Console sem ver segredos
Quando o Operador consultar, gravar ou desligar o WhatsApp pelo Console, o
Console SHALL usar a API de configuração do Ambiente e mostrar o resultado; os
segredos MUST NOT aparecer preenchidos em nenhuma tela, só se estão
configurados, e a recusa da API SHALL ser mostrada citando o campo.
Origem: RN-12, RN-11

#### Scenario: T-110 Consulta mostra o Provedor e os segredos só como configurados
- **WHEN** a API devolve a Z-API ativa, com a instância `inst1` e os segredos configurados, e o Operador abre a tela do WhatsApp
- **THEN** o Console mostra o Canal ligado, o Provedor Z-API, a instância `inst1` e cada segredo só como "configurado", com os campos de segredo vazios

#### Scenario: T-111 Gravar mostra a confirmação ou o campo recusado
- **WHEN** o Operador preenche o UltraMsg sem o token e grava, e depois preenche o token e grava de novo
- **THEN** na primeira vez o Console mostra a recusa citando o campo token, sem limpar o formulário, e na segunda mostra a confirmação e a consulta atualizada, com os segredos de novo vazios

#### Scenario: T-112 Desligar pede confirmação
- **WHEN** o Operador pede para desligar o WhatsApp e cancela, e depois pede de novo e confirma
- **THEN** ao cancelar nenhuma chamada é feita, e ao confirmar o Console chama o desligamento e mostra o Canal desligado

### Requirement: REQ-33 Acesso de fora só do escritório e só à configuração
Se uma requisição da internet vier de fora do IP do escritório, o sistema SHALL
recusá-la antes de chegar ao Console ou à API de configuração; a API SHALL
continuar sem rota pública para pedido de Notificação e Opt-out, venha de onde
vier.
Origem: RN-12

#### Scenario: T-113 Console recusado de fora do escritório
- **WHEN** uma requisição de um IP fora do escritório pede `https://ominicore.flipapi.com.br`
- **THEN** a resposta é 403 da Cloudflare e nenhum arquivo do Console é entregue

#### Scenario: T-114 Configuração recusada de fora do escritório
- **WHEN** uma requisição de um IP fora do escritório chama `/v1/config/whatsapp` em `notify.flipapi.com.br` ou `notify-hml.flipapi.com.br`
- **THEN** a resposta é 403 da Cloudflare e a requisição não chega à API

#### Scenario: T-115 Envio e Opt-out continuam sem rota pública
- **WHEN** uma requisição da internet chama `/v1/notificacoes` ou `/v1/opt-outs` em `notify.flipapi.com.br` ou `notify-hml.flipapi.com.br`
- **THEN** a resposta é 403 ou 404 e a requisição não chega à API

### Requirement: REQ-34 CORS só para o Console
Quando uma requisição à configuração vier com a origem do Console, a API SHALL
autorizá-la por CORS, inclusive no preflight; com outra origem, SHALL NOT
devolver cabeçalho que a autorize, e as rotas de envio e de Opt-out nunca
devolvem cabeçalho de CORS.
Origem: RN-12

#### Scenario: T-116 Preflight da origem do Console é aceito
- **WHEN** chega um `OPTIONS /v1/config/whatsapp` com origem `https://ominicore.flipapi.com.br`, pedindo `PUT` e o cabeçalho `Authorization`
- **THEN** a resposta é 204, sem exigir token, autorizando essa origem, os métodos `GET`, `PUT` e `DELETE` e o cabeçalho `Authorization`

#### Scenario: T-117 Outra origem não é autorizada
- **WHEN** chega um `OPTIONS` ou um `GET /v1/config/whatsapp` com origem `https://exemplo.com`, ou qualquer requisição a `/v1/notificacoes` com a origem do Console
- **THEN** nenhuma resposta traz `Access-Control-Allow-Origin`

### Requirement: REQ-35 Operador entra com segundo fator
Quando um Operador pessoa for criado num Ambiente, o sistema SHALL exigir que
ele configure o segundo fator (OTP) e troque a senha no primeiro login, e SHALL
pedir o OTP em todo login no Console.
Origem: RN-12

#### Scenario: T-118 Operador novo configura o OTP no primeiro login
- **WHEN** o script de criação de Operador cria a pessoa `ana` no realm local com o papel `ominicore-config`
- **THEN** `ana` tem o papel, as ações obrigatórias de configurar o OTP e de trocar a senha, e o fluxo de login do client `ominicore-console` exige o OTP
