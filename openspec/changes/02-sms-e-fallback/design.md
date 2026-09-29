## Context
Fatia 1 entregou o despacho com uma Tentativa por Notificação e um
`NotificationRouter` que encerra em qualquer falha. A tabela `tentativas` já
tem `numero` e `disponivel_em`, e a fila reserva só as pendentes com
`disponivel_em <= now()` (ADR 0001).

## Decisions
- **Retentativa = nova linha em `tentativas`**, no mesmo Canal, `numero + 1` e
  `disponivel_em = now() + espera`; gravada na mesma transação que conclui a
  Tentativa atual. Sem schema novo, sem job à parte: a fila já respeita
  `disponivel_em`.
- **Router puro no domínio:** recebe resultado, Canal, número da Tentativa,
  Ordem de Canais e contatos; devolve a decisão (Retentativa em X após d,
  próxima Tentativa no Canal Y, ou fim enviada/falhou). A espera com jitter
  recebe a fonte de aleatoriedade injetada, para o teste T-22 ser determinístico.
- **Twilio:** `POST /2010-04-01/Accounts/{AccountSid}/Messages.json`
  (form-encoded) com `To`, `MessagingServiceSid` e `Body`, autenticado por API
  Key (Basic `api_key_sid:api_key_secret`). Classificação igual ao SendGrid
  (2xx aceito; 429/5xx/rede temporária; demais 4xx definitiva).
- **Contatos:** coluna `destinatario_telefone` em `notificacoes` (migration
  00003, aditiva); validação E.164 (`^\+[1-9]\d{7,14}$`) no adaptador HTTP.

## Risks / Trade-offs
- Mantém "pelo menos uma vez" da fatia 1: queda entre o aceite do Provedor e o
  commit pode repetir o envio (já no `to-do.md`).
- Jitter só local a cada Notificação; rajada de Retentativas simultâneas depois
  de uma queda longa do Provedor fica para o rate limit por Provedor (sem REQ).
