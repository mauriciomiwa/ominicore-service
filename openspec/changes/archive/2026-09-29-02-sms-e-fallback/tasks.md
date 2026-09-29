## 1. Pedido com Ordem de Canais e telefone
- [x] 1.1 Contrato `api/openapi.yaml`: `canais` opcional e `destinatario.telefone` (E.164); lint do contrato
- [x] 1.2 Migration 00003: `destinatario_telefone`, versão SMS do `boas-vindas`, ordem padrão Email → SMS; `sqlc generate`
- [x] 1.3 Caso de uso e handler: ordem do pedido, Canal sem versão, nenhum contato, E.164
- [x] 1.4 Teste T-15 (ordem do pedido substitui a do Template)
- [x] 1.5 Teste T-16 (Canal pedido sem versão no Template é recusado)
- [x] 1.6 Teste T-17 (Destinatário sem contato para nenhum Canal é recusado)
- [x] 1.7 Teste T-18 (telefone fora do E.164 é recusado)
- [x] 1.8 Rodar de novo T-4, T-5 e T-6 (REQ-2 modificado)

## 2. Canal SMS
- [x] 2.1 Adaptador Twilio (API Key + Messaging Service) e leitura de `kv/<prefixo>/twilio` no OpenBao; worker registra o Canal SMS
- [x] 2.2 Teste T-19 (Tentativa de SMS é enviada)
- [x] 2.3 Teste T-20 (número inválido é Falha definitiva)

## 3. Retentativa
- [x] 3.1 Router no domínio com espera 5 s / 30 s / 2 min ±20% (aleatoriedade injetada); repositório grava a Retentativa na transação da conclusão
- [x] 3.2 Teste T-21 (Falha temporária agenda Retentativa no mesmo Canal)
- [x] 3.3 Teste T-22 (espera cresce a cada Retentativa)
- [x] 3.4 Teste T-23 (Retentativas esgotadas encerram o Canal)

## 4. Fallback
- [x] 4.1 Router escolhe o próximo Canal com contato; repositório cria a Tentativa ou encerra a Notificação
- [x] 4.2 Teste T-24 (Falha definitiva no Email passa ao SMS)
- [x] 4.3 Teste T-25 (Canal sem contato é pulado)
- [x] 4.4 Teste T-26 (todos os Canais falham)
- [x] 4.5 Teste T-27 (sucesso encerra o fallback)
- [x] 4.6 Rodar de novo T-12 e T-10 (fatia 1) sobre o router novo

## 5. Fechamento
- [x] 5.1 Atualizar `deploy/local/` (segredo `twilio` no bootstrap), `DOC/Infra/AMBIENTES.md`, mapa e `DOC/System/plano-de-testes.md` (T-15 a T-27)
- [x] 5.2 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 5.3 Auditar o diff com a skill sentinel
