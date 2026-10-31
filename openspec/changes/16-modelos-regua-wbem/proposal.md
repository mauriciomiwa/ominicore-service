# Proposal

## Why

O dono declarou mandatória a apresentação comercial e operacional v9 do
aviva-back (31/10/2026). Ela traz a régua de comunicação com o paciente: o link
inicial do financiamento e as Mensagens 1 a 15 (Open Finance, Pix Automático,
reserva e limite do fundo, Unico e contrato, pós-venda). O aviva-back pede cada
mensagem por Template (RN-1), e o OminiCore só tem os dois da RN-13.

Decisões do dono (31/10/2026): só os modelos agora, sem os disparos; texto com
o link escrito no lugar do botão da v9; o PDF das Mensagens 7, 8 e 11 vai como
link. A change propõe a **RN-14**.

## What Changes

- Migration `00012_modelos_regua_wbem.sql` com 16 Templates, Ordem de Canais só
  `whatsapp` e versão WhatsApp em texto: título em negrito, o corpo da v9, a
  linha do botão com `{{link}}` e o rodapé "A W Bem nunca pede senha, código ou
  depósito." em itálico.
- O link inicial é um Template novo (`wbem-financiamento-na-clinica`); o
  `wbem-link-tomador` e o `wbem-otp-tomador` não mudam (o aviva-back os pede só
  com `link` e `codigo`).
- Sem mudança de código nem de contrato.

Fora de escopo: botão de link e anexo no WhatsApp, versões SMS e e-mail,
Modelo aprovado da Meta, disparos no aviva-back.

## Capabilities

### Modified Capabilities
- `notificacoes`: REQ-49 (modelos da régua do wbem).

## Impact

- Banco: migration aditiva `00012` (só `INSERT`; o `down` apaga os 16).
- A `15-fila-zapi` passa a usar a `00013`.
- Consumidor: aviva-back, quando cada evento existir.
