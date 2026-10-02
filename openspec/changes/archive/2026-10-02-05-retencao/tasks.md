## 1. Dados e rodada
- [x] 1.1 Migration 00007 (`anonimizada_em`, índice parcial); queries `AnonimizarLote` e `ContarPendentesVencidas`; `sqlc generate`
- [x] 1.2 Repositório e porta `Anonimizar(ctx, antes time.Time, lote int)`; caso de uso de retenção
- [x] 1.3 `worker.Retencao` (rodada no boot e a cada 1 h, relógio injetado); `OMINICORE_RETENCAO_DIAS` no `main`
- [x] 1.4 Teste T-57 (Notificação finalizada além do prazo é anonimizada)
- [x] 1.5 Teste T-58 (Notificação dentro do prazo não muda)
- [x] 1.6 Teste T-59 (Notificação pendente não é anonimizada)
- [x] 1.7 Teste T-60 (prazo configurável)
- [x] 1.8 Teste T-61 (rodada repetida é idempotente)
- [x] 1.9 Teste T-62 (Opt-out sobrevive à anonimização)

## 2. Consulta
- [x] 2.1 `anonimizada_em` na consulta (repositório, handler e contrato)
- [x] 2.2 Teste T-56 (consulta mostra quando foi anonimizada)
- [x] 2.3 Rodar de novo T-43 e T-44 (REQ-14 modificado)

## 3. Fechamento
- [x] 3.1 Atualizar `deploy/` (env `OMINICORE_RETENCAO_DIAS`), `DOC/Infra/AMBIENTES.md`, `DOC/Infra/infraestrutura.md`, mapa e `DOC/System/plano-de-testes.md` (T-56 a T-62)
- [x] 3.2 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 3.3 Auditar o diff com a skill sentinel
