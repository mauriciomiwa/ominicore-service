# Exemplos de formato — requisitos no OpenSpec e fatias verticais

Modelos autocontidos para os passos 5 e 6 do fluxo. Se houver um projeto-molde
acessível, prefira o formato dele; senão, use estes. A convenção é a mesma que
o `openspec/config.yaml` injeta em todo artefato gerado pelo `/opsx:propose`.

## Fatias verticais (`DOC/System/backlog-mvp.md`)

Cada fatia entrega valor fim-a-fim (UI → API → banco), tem "pronto" = seus
testes passando e é **uma change do OpenSpec**. Ordene por dependência.

```
### Fatia 1 — Registrar venda simples
Change: `01-registrar-venda`
Valor: operador registra uma venda e ela aparece na listagem.
Cobre: RN-1 (REQ-1, REQ-2)
Depende de: —
Pronto quando: T-1 e T-2 passando e change arquivada.

### Fatia 2 — Operação offline
Change: `02-operacao-offline`
Valor: vender sem rede e sincronizar depois, sem duplicar.
Cobre: RN-3 (REQ-3)
Depende de: Fatia 1
Pronto quando: T-3 passando e change arquivada.
```

Anti-exemplo: "fatia = camada" ("fazer todo o banco", "fazer todas as telas")
não é fatia vertical — não entrega valor testável fim-a-fim.

## Change de uma fatia (`openspec/changes/01-registrar-venda/`)

### `proposal.md`

```markdown
## Why
RN-1: o operador precisa registrar vendas no balcão e vê-las na listagem —
hoje isso é feito em planilha e se perde.

## What Changes
- Registro de venda com UUID gerado no cliente e timestamp de negócio.
- Reenvio da mesma venda não duplica.

Fora do escopo: cancelamento (RN-4, fatia 3), operação offline (fatia 2).

## Impact
- Nova capacidade `vendas`; endpoint POST /vendas; tabela vendas.
```

### `specs/vendas/spec.md` (delta)

Frase normativa em EARS com a palavra-chave SHALL/MUST em inglês (o
`openspec validate` depende dela); títulos estruturais e GIVEN/WHEN/THEN também
em inglês; o resto em português.

```markdown
## Purpose
Registro e consulta das vendas feitas no balcão, com gravação idempotente
para que reenvios e reconexões nunca dupliquem uma venda.

## ADDED Requirements

### Requirement: REQ-1 Registrar venda
Quando o operador registrar uma venda, o sistema SHALL gravá-la com UUID
gerado no cliente e timestamp de negócio.
Origem: RN-1

#### Scenario: T-1 Venda gravada aparece na listagem
- **WHEN** o operador registra uma venda
- **THEN** a consulta retorna a venda com UUID e timestamp corretos

### Requirement: REQ-2 Reenvio idempotente
Se a mesma venda (mesmo UUID) for reenviada, o sistema SHALL gravá-la uma
única vez.
Origem: RN-1

#### Scenario: T-2 Reenvio não duplica
- **WHEN** o mesmo payload é enviado duas vezes
- **THEN** existe um único registro no banco
```

Padrões EARS para a frase normativa: *Quando <evento>…* (event-driven) ·
*Se <condição indesejada>, …* (unwanted behavior) · *Enquanto <estado>…*
(state-driven) · *Onde <feature presente>…* (opcional) · sem gatilho = ubíquo
("O sistema SHALL…").

Regras que o `valida_rastreabilidade.py` confere:

- `REQ-n` no título do requisito; ID global, sequencial, nunca reaproveitado.
- `Origem: RN-n` (uma ou mais) apontando RN que existe na especificação.
- Ao menos um `#### Scenario: T-n`; cada T-n pertence a um único REQ.
- Mudar um REQ vigente = `## MODIFIED Requirements` repetindo o **mesmo
  título** (ID incluído) com o texto novo completo. Remover =
  `## REMOVED Requirements` com o motivo. REQ novo com ID de um vigente é furo.

### `tasks.md`

```markdown
## 1. Registrar venda
- [ ] 1.1 Endpoint POST /vendas com UUID do cliente
- [ ] 1.2 Teste T-1 (venda aparece na listagem)
- [ ] 1.3 Upsert idempotente por UUID
- [ ] 1.4 Teste T-2 (reenvio não duplica)

## 2. Fechamento
- [ ] 2.1 Rodar openspec validate e DOC/tools/valida_rastreabilidade.py sem furo
- [ ] 2.2 Auditoria final do diff completo com a skill sentinel (cada tarefa já foi auditada no apply)
```

O nome do teste no código cita o T-n (`test_T1_venda_aparece_na_listagem`,
`it("T-1 …")`) — é isso que liga cenário, teste e plano sem tabela manual.

Teste que só dá para fazer à mão (abrir o app, clicar, olhar) vira roteiro em
`tests/manual/*.md`: um título por T-n (`## T-8 Link externo abre no navegador`),
os passos e o resultado esperado. O roteiro versionado conta como teste que cita
o T-n; prefira sempre automatizar quando der.
