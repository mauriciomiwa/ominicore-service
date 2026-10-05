# Mapa do projeto — procedimento

O mapa responde "como este código está organizado e o que cada parte faz".
Ele tem duas camadas, e a separação é a regra que impede o mapa de mentir:

| Camada | Arquivo | Quem escreve | Pode editar? |
|---|---|---|---|
| Fatos | `System/mapa/inventario.json` e `inventario.md` | `tools/mapeia_projeto.py` | não — rode o script de novo |
| Interpretação | `System/mapa/visao-geral.md` e `System/mapa/<area>.md` | a IA, revisada por você | sim |

Tudo o que a interpretação afirma sobre rota, endpoint, chamada, fila ou
variável tem de estar no inventário (ou citar o arquivo:linha de onde veio).
O que for dedução — o propósito de uma área, uma regra de negócio — é marcado
como tal.

## Passo 1 — Inventário

```bash
python3 <DOC>/tools/mapeia_projeto.py .            # grava DOC/System/mapa/inventario.{json,md}
```

Monorepo: rode por projeto, com `--saida DOC/<projeto>/System/mapa` e a raiz
do projeto como argumento. Leia o `inventario.md` inteiro antes de escrever.

## Passo 2 — Áreas

Agrupe o código em **áreas de negócio**, não em pastas técnicas. Num front,
uma área costuma ser um grupo de rotas/telas + os serviços de API que elas
chamam + o estado (store) que usam (ex.: FlipCred › Cobrança). Num back, um
grupo de endpoints + filas + tabelas. Pastas técnicas transversais
(components, helpers, plugins) entram na visão geral, não viram área.

Para cada área, siga o fio no código: rota → componente da tela → serviço de
API chamado → endpoint do back. Leia os arquivos que o inventário aponta; não
adivinhe pelo nome.

## Passo 3 — `visao-geral.md`

```markdown
# Mapa — <projeto>

> Interpretação sobre o inventário de <data> (`inventario.md`). Revisado por: <nome>, <data>.

## O que é
<2–4 frases: para quem é o sistema e o que faz>

## Como está organizado
<diagrama Mermaid: áreas, módulos transversais e sistemas externos (back, bancos, parceiros)>

## Áreas
| Área | O que faz | Entradas (rotas/endpoints) | Depende de |
|---|---|---|---|
| [Cobrança](cobranca.md) | … | `flipcred/cobranca`… | flipsaude-back `/checkout-parcelas/*` |

## Integrações externas
| Sistema | Como é chamado | Onde |

## Configuração
<variáveis lidas × declaradas; aponte as lidas e não declaradas — são candidatas a furo de ambiente>

## Pontos de atenção
<duplicação (mesmo endpoint chamado de dois lugares), código morto aparente, acoplamento — cada um com arquivo>
```

## Passo 4 — Uma página por área (`<area>.md`)

```markdown
# <Área>

> Fatos do inventário de <data>; deduções marcadas com (deduzido).

## Para que serve
<o que o usuário consegue fazer aqui>

## Telas / entradas
| Rota | Tela | Arquivo |

## Fluxo principal
<Mermaid sequenceDiagram: usuário → tela → serviço → endpoint do back → resposta>

## Chamadas ao back
| Método | Endpoint | Chamado em | Para quê |

## Estado e dados
<store/módulos de estado, entidades>

## Regras de negócio encontradas (a confirmar)
| Candidata | Evidência (arquivo:linha) |
<ex.: "parcela não pode vencer antes de D+1" — EditChargeModal.vue:42 (deduzido)>

## Specs
<capacidade do OpenSpec que cobre esta área, se houver; senão "sem spec — próxima mudança aqui cria">
```

## Passo 5 — Levar para o resto das skills

- **RN candidatas** vão para `Negocio/especificacao-negocio.md` numa seção
  "Descobertas no código — a confirmar", com a evidência. Só viram RN-n
  depois que o dono do negócio confirmar.
- **Specs de linha de base** (opcional, para áreas críticas): comportamento
  que já existe vai direto para `openspec/specs/<capacidade>/spec.md`, com
  REQ-n, `Origem: RN-n` (só RN confirmada) e cenários T-n que tenham teste.
- **Ambientes:** variável lida e não declarada vira item no `to-do.md` e, se
  houver deploy, linha no `Infra/AMBIENTES.md`.
- **INDEX.md:** linha "como o código está organizado → `System/mapa/visao-geral.md`".
- **Visualização (opcional):** `python3 <DOC>/tools/mapa_html.py <DOC>/System/mapa`
  gera `System/mapa/mapa.html` (navegação por área, filtro, diagramas, fatos
  do inventário) a partir destas páginas — pronto para abrir no navegador ou
  publicar como artifact. Só biblioteca padrão do Python. A fonte continua
  sendo a DOC/: regenere depois de revisar as páginas.

## Manter vivo

- `checa_projeto.py` roda `mapeia_projeto.py --check` e alerta quando surge
  rota, endpoint, chamada, fila ou variável fora do mapa.
- A `sentinel` cobra atualizar o mapa quando o trecho cria uma dessas coisas.
- Atualizar = rodar o script de novo e revisar só as páginas das áreas
  afetadas (o `--check` diz quais itens mudaram).

## Situação do projeto (na vista `mapa.html`)

O `mapa_html.py` abre a vista com a seção **Situação do projeto**, lida na hora
dos arquivos que o fluxo já mantém — ninguém escreve à mão:

- **Onde parou:** o `estado.md` (skill `passagem`).
- **Fatias do backlog:** cada fatia de `System/backlog-mvp.md` como *concluída*
  (change arquivada, ou sem change mas com todos os REQ vigentes), *em curso*
  (change ativa, com tarefas feitas/total do `tasks.md`) ou *a fazer*.
- **Aonde voltar:** as pendências abertas (`- [ ]`) do `to-do.md`.

Para atualizar, gere a vista de novo (no Sentinel, botão "Atualizar mapa").

