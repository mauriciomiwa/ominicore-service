# Regras de ouro — princípios de engenharia

Base que a skill `sentinel` audita. Regra-mãe: **proporcionalidade +
rastreabilidade** — o princípio certo, no grau certo, ligado a um requisito e
a um teste. Canhão em mosquito também é defeito.

Cada princípio tem três camadas de garantia: **definido** aqui, **auditado**
pela `sentinel` no código em foco e — quando a stack permite — **mecanizado**
em lint/CI (ver "Verificação automática" em `stack-tecnologico.md`). Princípio
sem camada de verificação é aspiração, não regra.

- **Clean Code / SOLID** — nomes que revelam intenção, funções pequenas com
  uma responsabilidade, dependências injetadas e apontando para abstrações.
  *Verifica:* lint (tamanho, complexidade, parâmetros) + revisão.
- **DRY** — cada conhecimento tem representação única; duplicar lógica de
  negócio é bug em câmera lenta (corrige num lugar, esquece no outro).
  *Verifica:* detector de duplicação no CI + revisão.
- **KISS** — a solução mais simples que passa nos mesmos testes; complexidade
  só entra com justificativa registrada.
  *Verifica:* revisão — "existe caminho mais simples com os testes verdes?".
- **YAGNI** — nada se constrói antes de um requisito exigir.
  *Verifica:* rastreabilidade — código sem REQ de origem é violação por
  definição (`tools/valida_rastreabilidade.py` + sentinel, item 1).
- **Menor Surpresa** — o componente faz o que o nome promete, sem efeito
  colateral escondido.
  *Verifica:* revisão — nome promete X, o código faz X+Y?
- **Lei de Demeter** — fale só com vizinhos diretos; `a.getB().getC().faz()`
  acopla você à estrutura interna alheia.
  *Verifica:* regra de lint (maioria das linguagens) + revisão.
- **API-First** — contrato (OpenAPI/schema) definido e versionado antes da
  implementação; quebra de contrato exige versionar, nunca surpreender.
- **Modularidade / DDD** — limites de contexto explícitos; o domínio não
  conhece infraestrutura; módulos conversam por interfaces.
- **CI/CD** — todo commit passa por build + testes automatizados; entregar é
  decisão de negócio, não evento técnico.
- **Trunk-Based** — lotes pequenos integrados com frequência; branch de vida
  curta; feature flag em vez de branch longa.
- **IaC** — infraestrutura descrita em código versionado; nada criado "no
  clique" que não dê para recriar do zero.
- **Pirâmide de testes** — muitos unitários, alguns de integração, poucos
  E2E; prove o requisito no nível mais barato que o prova.
- **Shift-left** — segurança, qualidade e validação o mais cedo possível:
  lint, scan e revisão antes do merge, não depois do deploy.
- **Observabilidade** — log estruturado nos pontos-chave, métricas e traces
  proporcionais ao risco; nunca PII ou segredo em log.
- **Design for failure** — rede e IO falham: timeout, retry com backoff,
  circuit breaker e idempotência onde há reenvio.
- **Zero trust** — não confie na rede interna; autentique e autorize cada
  chamada; segredos em cofre, nunca em código ou commit.
- **Scan de dependências** — dependência nova é decisão: justificada, sem CVE
  conhecido, atualizada por rotina via Dependabot (`.github/dependabot.yml`).
  *Verifica:* `tools/dependabot_config.py <raiz> --check` sem furo; política
  (SLA, regra de merge, ignore) em `Infra/seguranca-dependencias.md`.
- **IA como assistente** — IA acelera, humano decide: código gerado passa
  pela mesma revisão e pelos mesmos testes de qualquer código.
