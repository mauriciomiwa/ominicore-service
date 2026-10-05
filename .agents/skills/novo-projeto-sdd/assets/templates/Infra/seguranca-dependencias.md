# Segurança de dependências — {{PROJETO}}

Dependência é código de terceiro rodando com as nossas permissões. A casa usa
**Dependabot** (GitHub) para as três frentes: alerta de CVE, PR de correção de
segurança e PR de atualização de rotina. Este doc é a política; os arquivos
que a executam são `.github/dependabot.yml` e
`.github/workflows/dependabot-automerge.yml`.

## Ligado no repositório (Settings > Code security)

| Recurso | Estado | Para quê |
|---|---|---|
| Dependency graph | ligado | base de tudo |
| Dependabot alerts | ligado | avisa CVE em dependência declarada |
| Dependabot security updates | ligado | abre PR de correção sozinho |
| Grouped security updates | ligado | um PR por ecossistema, não um por CVE |
| Allow auto-merge (Settings > General) | ligado | pré-requisito do workflow |
| Status checks obrigatórios na default branch | ligado | sem isso o auto-merge **não** roda (fail-closed) |

## Version updates (`.github/dependabot.yml`)

- **Semanal**, segunda 06:00 (America/São Paulo), um bloco por ecossistema.
- **minor + patch agrupados** num PR por ecossistema; **major em PR próprio**.
- Máximo **5 PRs abertos** por ecossistema — fila maior que isso vira ruído
  ignorado.
- Cooldown padrão do GitHub (3 dias após o release) vale para version updates;
  não vale para security updates, que chegam na hora.
- Gerado por `tools/dependabot_config.py <raiz>`, que lê os manifestos.
  Manifesto novo (outro diretório, outro ecossistema)? Rode
  `tools/dependabot_config.py <raiz> --check` — furo de cobertura sai com rc 1.

## Regra de merge

| PR | Quem aprova |
|---|---|
| patch (qualquer dependência) | auto-merge, CI obrigatório verde |
| minor de dev-dependency | auto-merge, CI obrigatório verde |
| minor com correção de segurança (GHSA) | auto-merge, CI obrigatório verde |
| minor de produção | revisão humana |
| **major** | revisão humana: ler changelog/breaking changes, rodar testes, ajustar código no mesmo PR |
| CI vermelho | nunca entra — conserte ou feche com motivo |

PR agrupado herda o maior tipo do grupo: um minor de produção no meio do lote
leva o lote inteiro para revisão humana. É o preço de ter menos PRs.

> Merge feito pelo `GITHUB_TOKEN` **não dispara** workflows de `push`. Se o
> deploy roda em push na default branch, troque o token do passo de merge por
> um GitHub App token — ou aceite que o deploy sai no próximo merge humano.

## SLA por severidade (alerta aberto → corrigido em produção)

| Severidade | Prazo |
|---|---|
| Crítica | <preencher — sugestão: 2 dias úteis> |
| Alta | <preencher — sugestão: 7 dias> |
| Média | <preencher — sugestão: 30 dias> |
| Baixa | próxima janela de rotina |

Estourou o prazo sem correção disponível? Registre a mitigação (flag, WAF,
versão fixada, uso não exposto) no `log.md` e deixe o alerta aberto.

## Ignorar / dispensar

Ignorar é decisão, não silêncio: todo `ignore:` no `dependabot.yml` e todo
alerta dispensado ("dismiss") tem uma linha no `log.md` com **o quê, por quê,
quem e até quando** revisitar. `ignore` sem prazo vira dívida invisível.

## Dono

| Papel | Quem |
|---|---|
| Triagem semanal dos PRs e alertas | <preencher> |
| Aprovação de major | <preencher> |
