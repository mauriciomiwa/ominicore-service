## 1. Template e pedido
- [x] 1.1 Migration 00004: `modelo_aprovado`, `idioma`, `parametros` em `template_conteudos`; versão WhatsApp do `boas-vindas` (`boas_vindas`, `pt_BR`, `[nome]`); `sqlc generate`
- [x] 1.2 Domínio: `CanalWhatsApp`, contato = telefone, `ConteudoCanal` com Modelo aprovado e parâmetros, `Variaveis()` incluindo parâmetros
- [x] 1.3 Teste T-32 (pedido com WhatsApp na ordem é aceito)
- [x] 1.4 Teste T-33 (parâmetro do Modelo aprovado é variável exigida)

## 2. Canal WhatsApp
- [x] 2.1 Conferir na doc oficial da Meta a versão da Graph API e os códigos de erro; registrar em `DOC/Padroes/fontes-externas.md`
- [x] 2.2 Adaptador `saida/whatsapp` + Meta falsa (`internal/testes/metafake`); leitura de `kv/<prefixo>/whatsapp`; worker registra o Canal
- [x] 2.3 Teste T-28 (Tentativa de WhatsApp é enviada pelo Modelo aprovado)
- [x] 2.4 Teste T-29 (número sem WhatsApp passa ao próximo Canal)
- [x] 2.5 Teste T-30 (limite de taxa é Falha temporária)
- [x] 2.6 Teste T-31 (Modelo não aprovado é Falha definitiva)

## 3. Fechamento
- [x] 3.1 Atualizar contrato (descrição de `canais`), `deploy/local/` (segredo `whatsapp`), `deploy/aviva2/openbao/configurar-openbao.sh`, `DOC/Infra/AMBIENTES.md`, mapa e `DOC/System/plano-de-testes.md` (T-28 a T-33)
- [x] 3.2 Rodar `openspec validate --all --strict` e `DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 3.3 Auditar o diff com a skill sentinel
