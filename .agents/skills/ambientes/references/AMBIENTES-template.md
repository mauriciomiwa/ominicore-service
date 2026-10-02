# Ambientes — contrato canônico

> Template. Copie para `DOC/Infra/AMBIENTES.md` do projeto e preencha
> **perguntando** cada campo. Campo que você não sabe fica `A DEFINIR` — nunca
> preenchido com o que "costuma ser". Um `A DEFINIR` honesto é informação; um
> chute é armadilha.
>
> Regra de conteúdo: **nomes e caminhos sim, valores de segredo nunca.**
> `secret/data/<proj>/prod` pode estar aqui. O que tem dentro dele, não.

Última revisão: `<AAAA-MM-DD>` — por `<nome>`

---

## 1. Quais ambientes existem

| Ambiente | Para que serve | Quem usa | Dado real? |
|---|---|---|---|
| dev | desenvolvimento local | dev | não |
| hml | validação antes de prod | dev, QA, negócio | mascarado |
| prod | clientes | clientes | sim |

Se o projeto tem mais (`staging`, `sandbox`, `demo`), liste todos. Ambiente que
existe e não está aqui é a origem do próximo susto.

## 2. Como cada ambiente se identifica

A coluna que importa é a **evidência**: o comando ou variável que prova, agora,
onde você está. É o que vai no "Provado por" do bloco de alvo.

| Ambiente | Evidência (comando/variável) | Valor esperado |
|---|---|---|
| dev | `<ex.: echo $APP_ENV>` | `development` |
| hml | `<ex.: nomad status <job> \| grep Namespace>` | `<A DEFINIR>` |
| prod | `<A DEFINIR>` | `<A DEFINIR>` |

Se dois ambientes não dão para distinguir por um comando objetivo, **esse é o
defeito nº 1 a corrigir** — antes de qualquer outra coisa nesta lista. Enquanto
não der, toda ação com efeito fora da máquina local é tratada como prod.

## 3. Endereços e recursos

Resumo para provar o alvo. O detalhe (motor e versão do banco, nome do banco,
backup, path da credencial no cofre, onde roda) fica em `Infra/infraestrutura.md`.

| Recurso | hml | prod |
|---|---|---|
| URL da aplicação | | |
| Banco (host/instância) | | |
| Fila / broker | | |
| Namespace / cluster | | |
| Path base do Vault | | |
| Registry / repositório de imagem | | |
| Observabilidade (log, métrica) | | |

## 4. Configuração

- **Fonte da verdade das chaves:** `<ex.: .env.example na raiz>`
- **Onde ficam os valores:** `<ex.: Vault, path por ambiente>`
- **Quem pode ler/escrever valor de prod:** `<A DEFINIR>`

### Divergências deliberadas

Toda diferença de comportamento entre hml e prod entra aqui, com motivo. O que
não está escrito aqui é defeito, não decisão.

| Chave | hml | prod | Por quê |
|---|---|---|---|
| `<ex.: PAYMENT_MODE>` | `sandbox` | `live` | gateway não tem teste em live |
| | | | |

## 5. Dados de homologação

- **Origem da carga:** `<ex.: dump de prod mascarado>` ou `<gerador sintético>`
- **Script de mascaramento:** `<caminho versionado>`
- **Campos mascarados:** `<nome, e-mail, CPF, telefone, endereço, cartão…>`
- **Frequência / última carga:** `<A DEFINIR>`
- **Restaurar hml em prod:** proibido, sem exceção.

## 6. Promoção e aprovação

| | hml | prod |
|---|---|---|
| Quem aprova | | |
| Como se dispara | | |
| Janela permitida | | `<ex.: dias úteis, 9h–17h; fora disso só incidente>` |
| Comando de rollback | | |
| Tempo estimado de rollback | | |
| Quem observa depois | | |

## 7. Fora do alvo em prod

Ações que ninguém executa em produção sem um incidente aberto e uma segunda
pessoa acompanhando. Liste as do seu contexto:

- restaurar backup por cima da base
- `TRUNCATE` / `DROP` / `DELETE` sem `WHERE`
- rodar seed
- desligar observabilidade
- `<A DEFINIR>`
