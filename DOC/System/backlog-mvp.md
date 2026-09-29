# Backlog do MVP — ominicore

Fatias verticais ordenadas por dependência: cada uma entrega valor fim-a-fim
(API → banco → Provedor) e tem "pronto" = seus testes passando. "Fazer todo o
banco" ou "fazer todos os adaptadores" não é fatia — é camada.

**Cada fatia é uma change do OpenSpec** (`openspec/changes/<change>/`): o
proposal diz o porquê (RN), os deltas em `specs/` trazem os REQ e cenários
(T-n), o `tasks.md` é o passo a passo. Este arquivo é só a ordem e o mapa —
o detalhe mora na change. Status: change ativa = em curso; movida para
`openspec/changes/archive/` = entregue.

### Fatia 1 — Enviar e-mail por Template
Change: `01-enviar-email`
Valor: um Sistema de origem autenticado pede uma Notificação por Template e ela
chega por e-mail, sem duplicar em reenvio; esqueleto hexagonal, contrato,
`podman compose` e segredos no OpenBao de pé.
Cobre: RN-1 (REQ-1, REQ-2), RN-7 (REQ-3), RN-3 (REQ-4, REQ-5), RN-9 (REQ-6)
Depende de: —
Pronto quando: T-1 a T-14 passando e change arquivada.
**Entregue em 2026-09-29** (`openspec/changes/archive/2026-09-29-01-enviar-email/`).

### Fatia 2 — SMS e fallback entre Canais
Change: `02-sms-e-fallback`
Valor: Notificação sai por SMS; Falha temporária é retentada até 3 vezes e
Falha definitiva passa ao próximo Canal da Ordem de Canais.
Cobre: RN-3 (REQ-7), RN-2 (REQ-8, REQ-9), RN-1/RN-2 (REQ-2 modificado)
Depende de: Fatia 1
Pronto quando: testes dos REQ-7 a REQ-9 passando e change arquivada.
**Entregue em 2026-09-29** (`openspec/changes/archive/2026-09-29-02-sms-e-fallback/`).

### Fatia 3 — Status de entrega por webhook
Change: `03-status-por-webhook`
Valor: o Aviso de status de SendGrid, Twilio e Meta atualiza a Tentativa (entregue,
lida); falha de entrega depois do aceite passa ao próximo Canal; o Sistema de
origem consulta a Notificação; aviso sem assinatura válida é recusado.
Cobre: RN-5 (REQ-12, REQ-13, REQ-14)
Depende de: Fatias 2 e 6
Pronto quando: testes dos REQ-12 a REQ-14 passando e change arquivada.
**Entregue em 2026-09-29** (`openspec/changes/archive/2026-09-29-03-status-por-webhook/`).

### Fatia 4 — Opt-out por Canal
Change: `04-opt-out`
Valor: Destinatário com Opt-out de um Canal não recebe por ele; a Notificação
segue pelos demais Canais da Ordem.
Cobre: RN-8 (REQ-15)
Depende de: Fatia 2
Pronto quando: testes do REQ-16 passando e change arquivada.

### Fatia 5 — Retenção de 30 dias
Change: `05-retencao`
Valor: conteúdo e contato do Destinatário ficam ilegíveis após o prazo
(padrão 30 dias, configurável).
Cobre: RN-9 (REQ-16)
Depende de: Fatia 1
Pronto quando: testes do REQ-15 passando e change arquivada.

### Fatia 6 — WhatsApp por Modelo aprovado
Change: `06-whatsapp`
Valor: Notificação sai pelo WhatsApp (Cloud API da Meta) quando o pedido inclui
o Canal; número sem WhatsApp cai no próximo Canal da ordem.
Cobre: RN-4 (REQ-10, REQ-11)
Depende de: Fatia 2
Pronto quando: T-28 a T-33 passando e change arquivada.
**Entregue em 2026-09-29** (`openspec/changes/archive/2026-09-29-06-whatsapp/`). Incluída no MVP em
2026-09-29 (antes era "depois"); status de entrega do WhatsApp entra com a Fatia 3.

## Trilha de infraestrutura (não é fatia)

Deploy no aviva2.0 (quadlets `ominicore-api` e `ominicore-worker`, rota
`/webhooks/*` no Traefik, realm `ominicore`/`ominicore-hml`, role e policy no
OpenBao, schema e role no Aurora, `publicar.sh`). Segue a skill `ambientes`;
itens no `to-do.md`. Precisa estar pronta antes da Fatia 3 (webhooks exigem
rota pública).

## Depois do MVP
RN-6 respostas do Destinatário · RN-10 janela de horário e
agendamento · consumo de pedidos por fila externa.
