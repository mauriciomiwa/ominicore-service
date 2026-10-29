# Configuração pública só do escritório, para o Console do Operador

Status: aceita (2026-10-05). Revisa a 0003 no ponto "API de envio só na rede interna".

Pela RN-12, o Operador configura o WhatsApp por uma página web, o Console do
Operador (`ominicore.flipapi.com.br`). A página é estática na Cloudflare (Worker
com Static Assets) e chama a API direto do navegador. Por isso, **só o prefixo
`/v1/config/`** ganha rota pública no Traefik, em `notify[-hml].flipapi.com.br`.
O acesso a ele fica protegido em três camadas:

- uma regra de WAF da Cloudflare só deixa passar o IP fixo do escritório, no
  Console e nesse prefixo (a 443 do host já só aceita a Cloudflare);
- a API exige o JWT do realm do Ambiente com o papel `ominicore-config`;
- o CORS só autoriza a origem do Console.

`/v1/notificacoes` e `/v1/opt-outs` continuam sem rota pública, e o token do
client do Console é recusado nelas.

Alternativas consideradas: um Worker intermediário falando com a API por
Cloudflare Tunnel (a API não ganharia rota pública, mas haveria mais peças no
host; recusado na sabatina); o Cloudflare Access na frente (atrito de CORS e
cookie entre domínios); `ipAllowList` no Traefik (põe o IP no repositório e não
cobre o Console, que não passa pelo Traefik).

Consequências: um endpoint que escreve no cofre fica alcançável pela internet,
contido pelo WAF. Se a regra for desligada, a única barreira passa a ser o JWT;
a verificação diária do CI (`deploy/aviva2/verificar-exposicao.sh`) existe
para pegar isso. A troca do IP do escritório se faz na IP List da Cloudflare,
e o mesmo IP precisa estar no `ipAllowList` do `auth.flipapi.com.br`
(`../flip/infra/aviva2.0/traefik/dynamic/keycloak.yml`).

Atualização (2026-10-28): o escritório deixou de ter IP fixo. A origem fixa
passa a ser o túnel exclusivo do dono na Hostinger (`72.60.159.29`), que entra
na lista `escritorio` e no `ipAllowList` do Keycloak; o Console e a API são
roteados por ele. A decisão não muda: só a origem.

Atualização (2026-10-29): a zona está no plano Free, com 5 regras personalizadas,
todas em uso. Em vez de uma lista e de uma regra própria, o IP fica escrito
direto na regra "Block specific paths", com um `or` (`deploy/aviva2/README.md`,
C1 e C2). Até essa data, a regra não existia, e a verificação pelo CI foi o que
mostrou isso.
