## Why
RN-2: hoje qualquer falha encerra a Notificação — uma queda momentânea do
SendGrid vira Notificação perdida, e um e-mail inválido não tenta outro Canal.
A regra é retentar a Falha temporária no mesmo Canal (até 3 Retentativas, 5 s /
30 s / 2 min) e passar ao próximo Canal da Ordem de Canais na Falha
definitiva ou quando as Retentativas acabam, pulando Canal sem contato. RN-3:
o segundo Canal do MVP é o SMS (Twilio). RN-1: o Sistema de origem pode
informar a Ordem de Canais no pedido.

É a Fatia 2 do `DOC/System/backlog-mvp.md`; depende da Fatia 1 (arquivada).

## What Changes
- Pedido aceita `canais` opcional (Ordem de Canais); sem ele vale a do
  Template. Canal pedido sem versão no Template é recusado.
- Destinatário ganha `telefone` (E.164). Canal sem contato é pulado; pedido
  sem contato para nenhum Canal da ordem é recusado.
- Canal SMS pelo Twilio (API Key + Messaging Service, credenciais em
  `kv/<prefixo>/twilio`).
- Retentativa de Falha temporária no mesmo Canal: 3 Retentativas (4 Tentativas),
  esperas de 5 s, 30 s e 2 min com jitter de ±20%.
- Fallback para o próximo Canal da Ordem de Canais; sucesso em qualquer Canal
  marca a Notificação `enviada`; sem Canal restante, `falhou`.
- Template `boas-vindas` ganha versão SMS; ordem padrão Email → SMS (migration).

Fora do escopo: status por webhook (Fatia 3), Opt-out (Fatia 4), retenção
(Fatia 5), WhatsApp (RN-4, depois).

## Impact
- Capacidade `notificacoes`: REQ-2 modificado; REQ-7, REQ-8 e REQ-9 novos.
- Contrato `api/openapi.yaml`: campos `canais` e `destinatario.telefone`
  (aditivos, sem quebra).
- Migration 00003 (aditiva): versão SMS do `boas-vindas` e coluna do telefone.
- Novo adaptador `saida/twilio`; `openbao` passa a ler `<prefixo>/twilio`.
