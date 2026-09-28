# Configuração, segredos e dados de homologação

Detalhe dos passos 2 e 3 da skill `ambientes`.

## configuração e segredo divergentes

A divergência entre ambientes só dói quando promove, e aí dói no pior momento.
Antídoto: **toda chave existe nos dois lados, declarada, mesmo que o valor
mude.**

- Fonte da verdade das **chaves** (não dos valores) é um arquivo versionado —
  `.env.example`, `values.yaml`, o bloco `variable` do HCL, o que o projeto usar.
  O contrato aponta qual é.
- **Valor de produção nunca entra no repositório**, nem em exemplo, nem em
  comentário, nem em log, nem na resposta ao usuário. Path e nome do segredo
  são públicos; conteúdo não.
- **Chave nova nasce nos dois ambientes na mesma entrega**, com valor e dono
  definidos. Introduzir só em hml é criar um bug agendado para a promoção.
- Se a configuração diverge de propósito (sandbox vs. produção do gateway,
  quota menor, feature flag), a divergência vai **documentada no contrato** com
  o motivo. Divergência não escrita é defeito, não decisão.

Para levantar o estado atual (o script fica ao lado deste arquivo, não no
repositório do projeto — use o caminho da skill):

```bash
python3 <caminho-da-skill>/ambientes/scripts/diff_ambientes.py \
    .env.hml .env.prod --rotulos hml prod --valores-iguais
```

Ele compara **chaves declaradas** (`.env`, YAML aninhado, JSON, ou `variable`
de HCL) e reporta: só em hml, só em prod, e presentes nos dois. Compara nomes,
nunca imprime valores. O `--valores-iguais` acrescenta o achado mais útil e
menos óbvio: chave cujo **valor é idêntico nos dois ambientes** — se for
credencial, comprometer hml compromete prod. Rode antes de toda promoção e
sempre que criar variável. Demais opções: `--help`.


## dados de homologação

Homologação existe para que prod não seja o primeiro contato com a realidade.
Dois defeitos opostos, ambos comuns:

- **Dado irreal demais:** hml passa, prod quebra no volume, no encoding, no
  caso de borda que só o cliente real produz. O que hml precisa espelhar é a
  *forma* dos dados de prod — cardinalidade, tamanho, valores sujos, nulos —
  não o conteúdo.
- **Dado real demais:** cópia de produção sem mascarar. Isso é vazamento de
  dado pessoal com outro nome, e vale para dump, print, log e anexo em ticket.
  Nome, e-mail, documento, telefone, endereço e cartão saem **mascarados na
  origem** — o dado sensível não chega a existir em hml.

Regra prática: o mascaramento acontece **durante a extração, antes de o dump
sair de produção** — nunca alterando dado em prod, e nunca depois, em hml, com
o dump cru já parado em disco. E é script versionado, não passo manual: passo
manual é o que alguém esquece na madrugada. O contrato registra qual é o script
e quando foi a última carga.

Sentido inverso — restaurar dado de hml em prod — é **BLOQUEANTE**, sempre.


