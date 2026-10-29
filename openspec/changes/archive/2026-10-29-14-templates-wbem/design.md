# Design

## Context

Template só nasce por migration (`00002`, `00004`, `00008`). A versão WhatsApp
de texto (`corpo`) é a que a Z-API e o UltraMsg usam (RN-4); o Modelo aprovado
é só da Meta. As variáveis saem do texto (`{{nome}}`) e são todas
obrigatórias (`dominio.Template.Variaveis`). Motivação: proposal.md.

## Goals / Non-Goals

**Goals:** os dois Templates disponíveis nos dois schemas (hml e prod), sem
código novo.

**Non-Goals:** cadastro por API, SMS, e-mail e Modelo da Meta.

## Decisions

1. **Migration aditiva, um `INSERT` por tabela**, como a `00002`. O `down`
   apaga os conteúdos e os Templates. Alternativa descartada: cadastro por
   API, que segue em aberto no `to-do.md` e é mudança maior.
2. **Ordem de Canais só `whatsapp`**: o aviva-back pede `canais:
   ["whatsapp"]`; sem versão SMS, um pedido por SMS é recusado (422), o que
   evita mandar o link por um Canal que o dono não decidiu.
3. **Prefixo `wbem-` no id**: identifica o Sistema de origem dono do Template,
   já que os Templates são globais.

## Risks / Trade-offs

- [Meta como Provedor] → o pedido é recusado com 422 "sem modelo aprovado";
  aceito: prod usa a Z-API.
- [Texto mudar] → nova migration (`UPDATE`), como a `00008`.
