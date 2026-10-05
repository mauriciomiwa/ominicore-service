# Escrita no cofre pela API do Operador, só na configuração do WhatsApp

Status: aceita (2026-10-05) — revisa a 0002 no ponto "policy só de leitura"

Trocar o Provedor do WhatsApp (Meta, UltraMsg ou Z-API) ou renovar um token
exigia SSH no host, o `configurar-openbao.sh` e o reinício de api e worker.
Pela RN-11, o Operador (papel de realm `ominicore-config` no Keycloak) faz
isso pela API, e o OminiCore passa a **escrever** no OpenBao — mas a policy dá
`create`/`update`/`delete` só em `kv/<prefixo>/whatsapp`; o resto continua só
leitura. Os segredos do webhook da Meta foram para esse mesmo caminho (com o
de `webhooks` como alternativa de leitura) para não dar escrita sobre os
segredos do SendGrid e do Twilio. A troca vale sem reinício: api e worker
releem o caminho a cada 30 s.

Alternativas consideradas: seguir só com o script (mantém a operação manual e
o reinício); guardar a configuração no Postgres cifrada (viola a RN-9);
dar escrita em `webhooks` também (amplia o estrago de um vazamento).

Consequências: um bug ou invasão da app pode trocar ou apagar a configuração
do WhatsApp — o KV v2 guarda as versões anteriores para recuperação manual e
cada troca fica em `config_eventos`; voltar para um binário anterior à fatia
10 com a Meta configurada pela API exige regravá-la pelo script; o login no
cofre é refeito quando a sessão de 1 h expira.
