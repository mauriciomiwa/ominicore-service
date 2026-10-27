## ADDED Requirements

### Requirement: REQ-42 Estado dos webhooks da Z-API no Console
Enquanto a Z-API for o Provedor ativo, o Console SHALL mostrar se os webhooks
estão cadastrados na Z-API e oferecer "Cadastrar de novo", que pede o
recadastro e mostra o resultado; o Console MUST NOT mostrar o segredo nem a
URL do webhook.
Origem: RN-12, RN-5

#### Scenario: T-136 Webhooks não cadastrados e recadastro
- **WHEN** a consulta devolve a Z-API com os webhooks não cadastrados e o Operador clica em "Cadastrar de novo", e o recadastro dá certo
- **THEN** o Console mostra primeiro o aviso de não cadastrados e depois os webhooks cadastrados, sem exibir segredo nem URL
