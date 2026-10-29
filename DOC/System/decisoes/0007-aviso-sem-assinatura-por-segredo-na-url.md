# Aviso de status sem assinatura, autenticado por segredo na URL

Status: aceita (2026-10-27). Revisa a RN-5 nos Provedores que não assinam.

A Z-API manda os Avisos de status ("Ao enviar" e "Status") sem assinatura nem
cabeçalho de autenticação, e a RN-5 descartava todo aviso sem assinatura
válida. Sem esses avisos, o OminiCore não percebe número sem WhatsApp ou
shadow ban depois do aceite e não passa ao próximo Canal. Decidimos aceitar o
aviso da Z-API em `POST /webhooks/zapi/<segredo>`:

- o segredo tem 32 bytes aleatórios e é comparado em tempo constante;
- o `instanceId` do corpo precisa ser o da instância configurada;
- qualquer falha recebe 404, sem efeito.

O segredo mora só no OpenBao (`kv/<p>/whatsapp`, `webhook_segredo`). Ele é
trocado a cada gravação da Z-API, quando o próprio OminiCore recadastra as URLs
na Z-API pela API dela.

Alternativas consideradas: ficar sem webhook (perde fallback e entrega); exigir
um header com o segredo (a Z-API não deixa configurar headers no webhook); um
proxy com Cloudflare Access na frente (o webhook da Z-API não autentica).

Consequências: quem descobrir a URL (log de terceiros, print, logs da
Cloudflare) consegue forjar avisos até a próxima gravação. O pior caso é marcar
uma Tentativa como entregue, lida ou falhou: isso afeta o status e o fallback,
mas não vaza dado pessoal nem envia mensagem a outro número. O Traefik não
registra `/webhooks/*` (`accessLogs: false`).
