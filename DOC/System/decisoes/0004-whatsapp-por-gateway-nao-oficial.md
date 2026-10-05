# WhatsApp por gateway não oficial (UltraMsg)

Status: aceita (2026-10-05)

A conta Meta Business não existe e o cliente quer o WhatsApp já. Adotamos o
UltraMsg, um gateway sobre o WhatsApp Web (número comum autenticado por QR),
como segundo Provedor do Canal WhatsApp, escolhido por ambiente no cofre
(`provedor` em `kv/<prefixo>/whatsapp`); a Cloud API da Meta continua pronta
para a troca. O cliente aceitou o risco: o número pode ser bloqueado pelo
WhatsApp e o uso pode conflitar com os termos dele.

Alternativas consideradas: esperar a conta Meta (adia o Canal sem data);
trocar a Meta pelo UltraMsg (perde a saída oficial quando a conta existir).

Consequências: número dedicado por ambiente (hml e prod), só mensagem
transacional a quem já é cliente; sem Aviso de status (a Tentativa fica
`enviada`, nunca `entregue`); número sem WhatsApp não é detectado; instância
desconectada enfileira no UltraMsg sem erro — monitorar no painel.

## Adendo (2026-10-05) — Z-API

A Z-API entra como terceiro Provedor do Canal WhatsApp (`provedor=zapi`,
change `09-whatsapp-zapi`), da mesma família do UltraMsg: gateway sobre o
WhatsApp Web, número autenticado por QR. O risco aceito pelo cliente nesta
decisão vale também para ela, com as mesmas consequências e a mesma regra de
número dedicado por ambiente. Diferenças que pesam: a Z-API exige o token da
instância na URL (o OminiCore não o escreve em log nem no motivo) e tem um
token de segurança da conta (`Client-Token`) opcional. Os webhooks dela não
têm assinatura documentada: seguem sem uso até a RN-5 ser revista.
