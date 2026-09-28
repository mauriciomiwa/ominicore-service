# Checklist de promoção hml → prod

Cada item é **verificável**: ou você mostra a evidência, ou o item falhou. "Acho
que sim" conta como falhou. Se qualquer BLOQUEANTE falhar, o veredito é **não
promove**, e a resposta diz qual.

**Itens que só um humano pode atestar** (aprovação, backup já restaurado
alguma vez, quem está observando, quem foi avisado): evidência válida é a
afirmação explícita do usuário nesta conversa — registre como
`atestado por <usuário>`. Perguntar uma vez e anotar a resposta é o
procedimento; tratá-los como automaticamente falhos faria o checklist reprovar
toda promoção e, no terceiro dia, ser ignorado por inteiro.

---

## Antes — artefato

- [ ] **BLOQUEANTE** O artefato que vai para prod é **byte a byte** o que rodou
      em hml, identificado por tag imutável (SHA), não por `latest` nem por
      branch. Rebuild na promoção invalida o teste: o que você validou deixou
      de ser o que você subiu.
- [ ] A tag está registrada onde a equipe consegue achar depois (release, nota
      de deploy, ticket). Deploy que ninguém sabe reconstituir não tem rollback.
- [ ] O que mudou desde a versão em prod está listado — commits ou changelog.

## Antes — configuração

- [ ] **BLOQUEANTE** `diff_ambientes.py` rodado; chave nova existe nos dois
      lados, com valor e dono definidos.
- [ ] **BLOQUEANTE** Cada divergência que sobrou está na tabela de divergências
      deliberadas do contrato, com motivo. Divergência não escrita é defeito,
      não decisão — e trava a promoção até virar uma linha no contrato.
- [ ] Segredo novo criado no path de prod **antes** do deploy, não durante.
- [ ] **BLOQUEANTE** Nenhum valor de produção foi para o repositório, log,
      print, ticket ou resposta de chat no processo. (Verificável no que passou
      por esta conversa e pelo diff; o resto é atestado pelo usuário.)

## Antes — banco

- [ ] **BLOQUEANTE** Migration é aditiva, ou tem plano de volta escrito e
      testado em hml.
- [ ] **BLOQUEANTE** A migration é compatível com a **versão anterior** do
      código durante a janela de deploy. Coluna removida ou renomeada quebra as
      instâncias que ainda não trocaram — separe em duas entregas.
- [ ] Tempo estimado da migration medido em hml com volume comparável ao de
      prod. Migration que trava tabela grande é incidente, não deploy.
- [ ] Backup recente e **restaurável** — restore já testado alguma vez, não só
      o backup existindo (item humano: atestado pelo usuário).

## Antes — reversão

- [ ] **BLOQUEANTE** Comando de rollback escrito, com a tag anterior explícita.
- [ ] Tempo estimado de rollback conhecido.
- [ ] **O que não volta** está dito: dado escrito no meio, e-mail enviado,
      cobrança feita, webhook disparado, mensagem publicada na fila. Rollback
      desfaz código; não desfaz efeito no mundo.

## Antes — gente

- [ ] **BLOQUEANTE** Quem o contrato define como aprovador aprovou, nesta
      versão (item humano: atestado pelo usuário).
- [ ] Alguém está observando depois do deploy, e sabe **o que** olhar: qual
      métrica, qual log, por quanto tempo (item humano).
- [ ] Janela respeitada. Sexta às 18h não é janela; é o começo de um fim de
      semana ruim.
- [ ] Quem é afetado foi avisado, se houver indisponibilidade (item humano).

## Durante

- [ ] Um ambiente por vez, um comando por vez.
- [ ] Bloco de alvo declarado e confirmado antes de cada comando com efeito.
- [ ] Saída de cada passo lida antes do passo seguinte. Encadear comandos no
      escuro é como um erro pequeno vira um grande.

## Depois

- [ ] Health check e as métricas-chave observadas pelo tempo combinado — não
      "subiu sem erro no console".
- [ ] Log de erro comparado com o padrão de antes do deploy.
- [ ] O caminho crítico exercitado de verdade, uma vez, em prod.
- [ ] Se algo destoou: **rollback primeiro, diagnóstico depois** — desde que o
      rollback seja seguro no estado atual do banco. Investigar com prod
      quebrado é escolher a indisponibilidade mais longa.

      **Exceção que importa:** se a migration já rodou e não é reversível,
      voltar o código contra o schema novo piora a situação. Aí o caminho é
      *fix-forward*, pelo plano de volta escrito antes de subir. É por isso que
      aquele item é bloqueante: sem ele, no momento do incidente não há
      decisão a tomar, só improviso.
- [ ] Contrato atualizado se algo mudou (chave nova, endereço, aprovador).

---

## Veredito

Encerre com uma linha:

- **pode promover** — todos os bloqueantes verdes.
- **pode promover com ressalvas** — só não-bloqueantes falharam; liste-os.
- **não promove** — diga qual bloqueante falhou e o que resolve.
