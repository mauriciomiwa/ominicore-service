## Why
RN-4 (incluída no MVP em 2026-09-29): os Sistemas de origem querem avisar pelo
WhatsApp, o Canal de maior leitura, sem perder a garantia de entrega — número
sem WhatsApp tem de cair no próximo Canal da Ordem de Canais (RN-2, já
entregue). A Meta só aceita, fora da janela de 24 h, mensagem por Modelo
aprovado; texto livre depende de RN-6 (respostas), que segue "depois".

É a Fatia 6 do `DOC/System/backlog-mvp.md`; depende da Fatia 2 (arquivada).

## What Changes
- Canal `whatsapp` pela Cloud API da Meta (Graph API, versão fixada no
  adaptador), token de System User e `phone_number_id` em
  `kv/<prefixo>/whatsapp`.
- Versão WhatsApp do Template: referencia o Modelo aprovado (nome + idioma) e
  a lista ordenada das variáveis do Template que viram os parâmetros `{{1}}`,
  `{{2}}`… do Modelo.
- Contato do Canal WhatsApp = `destinatario.telefone` (E.164), o mesmo do SMS.
- Classificação dos erros da Meta: 131026 (número sem WhatsApp) e 132001
  (Modelo inexistente ou não aprovado) = Falha definitiva; 130429 (limite) e
  131000 (erro genérico) = Falha temporária; demais pela regra comum.
- `boas-vindas` ganha versão WhatsApp (Modelo `boas_vindas`, `pt_BR`,
  parâmetros `[nome]`); a ordem padrão **não muda** (Email → SMS).

Fora do escopo: texto livre / janela de 24 h (RN-6), status de entrega do
WhatsApp por webhook (Fatia 3), cadastro de Modelo aprovado pela API da Meta.

## Impact
- Capacidade `notificacoes`: REQ-10 e REQ-11 novos.
- Migration 00004 (aditiva): colunas do Modelo aprovado em `template_conteudos`
  e versão WhatsApp do `boas-vindas`.
- Novo adaptador `saida/whatsapp`; OpenBao passa a ler `<prefixo>/whatsapp`.
- Contrato: nenhum campo novo (`canais` já aceita `whatsapp`; a descrição muda).
