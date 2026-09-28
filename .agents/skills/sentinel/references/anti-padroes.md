# Sentinel — Anti-padrões e armadilhas

Detalhamento do item 9 do checklist do `SKILL.md`.

## Banco de dados e desempenho

- **Transação longa:** nunca faça processamento lento (API externa, arquivo
  grande) dentro de uma transação aberta — segura *locks* e enfileira o banco.
  Transações o mais curtas e rápidas possível.
- **N+1:** não consulte em loop (1 query da lista + N por item). Use JOIN/eager
  loading e busque em 1–2 consultas.
- **Índices:** indexe as colunas usadas em busca/filtro (evita *full table
  scan*); não indexe tudo (deixa INSERT/UPDATE lentos). Índice na medida.

## Resiliência

- **Chamada sem timeout:** toda requisição a rede/banco/IO tem tempo limite. Sem
  isso, um destino travado esgota os recursos do servidor.
- **Engolir exceção:** nada de `catch` vazio ou só `print`. Trate ou propague, e
  logue com contexto — mascarar falha torna o bug impossível de achar.
- **Sem circuit breaker:** ao depender de serviço externo, falhe rápido quando
  ele cai (não martele com milhares de requisições); deixe-o recuperar.

## Over-engineering

- **YAGNI:** só construa o que o negócio exige agora; não crie abstração para um
  futuro incerto. Código parado custa caro para manter.
- **Código morto/comentado:** apague — o Git guarda o histórico.
- **God class:** classe/arquivo gigante que faz tudo (validar, salvar, enviar
  e-mail) → quebre por responsabilidade.

## Acoplamento e clareza

- **Lei de Demeter:** fale só com vizinhos diretos; `a.getB().getC().faz()`
  acopla o código à estrutura interna de outro objeto — ela muda, tudo quebra.
  Peça ao vizinho que faça, em vez de atravessá-lo. Há regra de lint para isso
  na maioria das linguagens.
- **Menor surpresa:** método cujo nome promete X mas também faz Y (salva *e*
  envia e-mail; getter que muta estado) — quem chama não tem como saber.
  Efeito colateral vai no nome, ou vai para outro método.
- **DRY com bom senso:** duplicação de *lógica de negócio* é bug em câmera
  lenta — corrige num lugar, esquece no outro. Mas não abstraia duplicação
  acidental de código que evolui separado: abstração errada custa mais que
  duplicação. Detector de duplicação no CI pega o grosso.

## Segurança básica

- **Sem hardcode de segredo:** senhas, chaves de API e URLs de banco em
  variáveis de ambiente/secret manager, nunca no código-fonte.
- **Não confie na entrada do usuário:** valide e sanitize dados do cliente; use
  *queries* parametrizadas (anti SQL injection). Nada vai cru do front ao banco.
- **Dados sensíveis:** proteja PII e segredos em trânsito e repouso quando
  aplicável (não basta evitar logar/hardcodar).

## Outras armadilhas comuns

- **Concorrência:** estado compartilhado sem proteção (condição de corrida),
  variáveis globais mutáveis.
- **Vazamento de recursos:** conexões/arquivos/handles sem fechar — use
  `defer`/try-with-resources/`finally`.
- **Listagem sem paginação/limite:** não retorne a tabela inteira; pagine.
- **Erro ignorado:** não descarte silenciosamente retornos de erro.
- **Log com dado sensível:** nunca logar segredo, token ou PII.
- **Números mágicos:** extraia constantes com nome que explique o valor.
