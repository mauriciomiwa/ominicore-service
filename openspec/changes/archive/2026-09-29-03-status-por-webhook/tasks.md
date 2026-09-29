## 1. Domínio e dados
- [x] 1.1 Conferir na doc oficial os formatos de evento e assinatura de SendGrid, Twilio e Meta; registrar em `DOC/Padroes/fontes-externas.md`
- [x] 1.2 Domínio: status `entregue`/`lida`, máquina de status monotônica, `AvisoStatus`, classificação da falha
- [x] 1.3 Migration 00005 (índice por `id_provedor`, `categoria_falha`); queries de aviso, falha tardia e consulta; `sqlc generate`
- [x] 1.4 Caso de uso `ReceiveMessageUseCase` (aplica o aviso; falha tardia chama o router)

## 2. Webhooks
- [x] 2.1 Segredos `kv/<prefixo>/webhooks` no OpenBao; `OMINICORE_WEBHOOK_URL`; Twilio envia `StatusCallback`
- [x] 2.2 Handlers `/webhooks/sendgrid`, `/webhooks/twilio`, `/webhooks/whatsapp` com validação de assinatura; fakes de assinatura nos testes
- [x] 2.3 Teste T-34 (entrega confirmada pelo Email)
- [x] 2.4 Teste T-35 (entrega confirmada pelo SMS)
- [x] 2.5 Teste T-36 (leitura confirmada pelo WhatsApp)
- [x] 2.6 Teste T-37 (assinatura inválida é recusada)
- [x] 2.7 Teste T-38 (status não regride)
- [x] 2.8 Teste T-39 (aviso de Tentativa desconhecida)
- [x] 2.9 Teste T-40 (verificação do webhook da Meta)

## 3. Falha tardia
- [x] 3.1 Teste T-41 (bounce no Email passa ao SMS)
- [x] 3.2 Teste T-42 (falha tardia não reabre Notificação entregue)

## 4. Consulta
- [x] 4.1 `GET /v1/notificacoes/{id}` no contrato e no handler (filtro por Sistema de origem)
- [x] 4.2 Teste T-43 (dono consulta status e Tentativas)
- [x] 4.3 Teste T-44 (outro Sistema de origem não enxerga)

## 5. Fechamento
- [x] 5.1 Atualizar `deploy/local/` e `deploy/aviva2/` (segredo `webhooks`, `OMINICORE_WEBHOOK_URL`, cadastro nos painéis no runbook), `DOC/Infra/AMBIENTES.md`, mapa e `DOC/System/plano-de-testes.md` (T-34 a T-44)
- [x] 5.2 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 5.3 Auditar o diff com a skill sentinel
