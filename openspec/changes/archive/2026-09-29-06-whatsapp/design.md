## Context
O Canal SMS (fatia 2) já dá contato por telefone, Retentativa, fallback e a
regra comum de classificação (`saida/provedor`). A versão de um Template por
Canal é `template_conteudos(assunto, corpo)`, com variáveis `{{nome}}` no texto.

## Decisions
- **Modelo aprovado em colunas novas de `template_conteudos`** (migration
  00004, aditiva): `modelo_aprovado`, `idioma`, `parametros text[]`. Para o
  WhatsApp, `corpo` fica vazio: o texto mora na Meta.
- **Domínio:** `ConteudoCanal` ganha `ModeloAprovado`, `Idioma` e
  `Parametros`; `Template.Variaveis()` passa a incluir os parâmetros (assim
  T-33 cai na validação que já existe); `Preencher` devolve os valores dos
  parâmetros na ordem. `Destinatario.Contato(whatsapp)` = telefone.
- **Adaptador `saida/whatsapp`:** `POST {URL}/{versão}/{phone_number_id}/messages`
  com `messaging_product: whatsapp`, `type: template`, componente `body` com os
  parâmetros `text`; `Authorization: Bearer <token de System User>`. Versão da
  Graph API em constante, conferida na doc oficial na implementação.
- **Classificação:** lê `error.code` da resposta; 131026 e 132001 forçam
  Falha definitiva, 130429 e 131000 forçam Falha temporária; sem código
  conhecido, vale `provedor.Classificar` pelo status HTTP.

## Risks / Trade-offs
- "Enviada" no WhatsApp = aceita pela Meta; entrega/leitura só com os webhooks
  da Fatia 3.
- Validação real depende da conta Meta Business, do número verificado e do
  Modelo `boas_vindas` aprovado (pendência no `to-do.md`); até lá, só Meta falsa.
- A Meta desativa versões antigas da Graph API: a troca vira tarefa datada.
