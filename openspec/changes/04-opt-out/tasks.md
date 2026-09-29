## 1. Domínio e dados
- [ ] 1.1 Domínio: `Destinatario.OptOut`, `Disponivel`, `FalhaOptOut`; `PrimeiroComContato` respeita Opt-out
- [ ] 1.2 Migration 00006 (`opt_outs`, `opt_out_eventos`); queries; `sqlc generate`; HMAC com chave de `kv/<prefixo>/optout`
- [ ] 1.3 Repositório: registrar/remover Opt-out; carregar Opt-out no pedido, na reserva da fila e na reserva do aviso

## 2. API de Opt-out
- [ ] 2.1 `POST`/`DELETE /v1/opt-outs` no contrato e no handler; caso de uso de Opt-out
- [ ] 2.2 Teste T-45 (registrar Opt-out)
- [ ] 2.3 Teste T-46 (registro repetido é idempotente)
- [ ] 2.4 Teste T-47 (remoção guarda quem removeu)
- [ ] 2.5 Teste T-48 (pedido de Opt-out inválido é recusado)

## 3. Sinais do Provedor
- [ ] 3.1 SendGrid `unsubscribe`/`spamreport` no webhook; Twilio 21610 = `FalhaOptOut` e registro no despacho
- [ ] 3.2 Teste T-49 (descadastro no SendGrid registra Opt-out de Email)
- [ ] 3.3 Teste T-50 (STOP no Twilio registra Opt-out de SMS e segue a ordem)

## 4. Canal em Opt-out é pulado
- [ ] 4.1 Worker confere o Opt-out antes de chamar o Provedor
- [ ] 4.2 Teste T-51 (Opt-out de SMS faz a ordem SMS → Email sair por Email)
- [ ] 4.3 Teste T-52 (todos os Canais em Opt-out recusam o pedido)
- [ ] 4.4 Teste T-53 (Opt-out entre o pedido e o envio)
- [ ] 4.5 Teste T-54 (Opt-out vale para todos os Sistemas de origem)
- [ ] 4.6 Teste T-55 (tabela de Opt-out sem contato legível)
- [ ] 4.7 Rodar de novo T-43 e T-44 (REQ-14 modificado)

## 5. Fechamento
- [ ] 5.1 Atualizar `deploy/local/` e `deploy/aviva2/` (segredo `optout`, "não rotacionar" no runbook), `DOC/Infra/AMBIENTES.md`, mapa e `DOC/System/plano-de-testes.md` (T-45 a T-55)
- [ ] 5.2 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [ ] 5.3 Auditar o diff com a skill sentinel
