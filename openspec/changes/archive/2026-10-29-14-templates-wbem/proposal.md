# Proposal

## Why

O aviva-back (Sistema de origem `wbem`) passa a mandar ao tomador, pelo
WhatsApp, o link de acesso à proposta e o código de acesso (RN-99 do
aviva-back). Ele pede a Notificação por Template (RN-1), mas o OminiCore só
tem o `boas-vindas`, e Template não se cadastra por API (decisão em aberto,
`DOC/to-do.md`): só por migration.

O dono decidiu (rodada de 29/10/2026, na sessão do aviva-back) dois Templates
de texto curto, só com versão WhatsApp em texto (Z-API). A change propõe a
**RN-13**:

> RN-13 — O OminiCore oferece ao Sistema de origem `wbem` dois Templates para
> o tomador, só pelo Canal WhatsApp: o do link de acesso (`wbem-link-tomador`,
> variável `link`) e o do código de acesso (`wbem-otp-tomador`, variável
> `codigo`).

## What Changes

- Migration `00011_templates_wbem.sql` com os dois Templates, Ordem de Canais
  só `whatsapp` e versão WhatsApp em texto:
  - `wbem-link-tomador`: "WBem: acesse sua proposta de crédito: {{link}} (vale
    24 h, uso único)."
  - `wbem-otp-tomador`: "WBem: seu código de acesso é {{codigo}}. Não
    compartilhe."
- Sem mudança de código nem de contrato: o pedido segue o `POST
  /v1/notificacoes` de hoje.

Fora de escopo: cadastro de Template por API, versões SMS e e-mail, e Modelo
aprovado da Meta para esses Templates (com a Meta como Provedor, o pedido é
recusado com 422 "versão whatsapp sem modelo aprovado", como já acontece).

## Capabilities

### New Capabilities
- nenhuma

### Modified Capabilities
- `notificacoes`: REQ-44 (Templates do `wbem` disponíveis).

## Impact

- Banco: migration aditiva `00011` (só `INSERT`; o `down` apaga os dois
  Templates). Publicar antes do aviva-back ligar o canal.
- Consumidor: aviva-back, change `whatsapp-pelo-ominicore`.
