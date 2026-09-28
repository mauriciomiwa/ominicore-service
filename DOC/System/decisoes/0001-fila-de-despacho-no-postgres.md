# Fila de despacho no Postgres, sem broker

Status: aceita (2026-09-28)

O diagrama previa RabbitMQ/SQS, mas no MVP (~100 mil Notificações/dia ≈ 1,2/s)
a fila de despacho é a própria tabela de Tentativas, consumida com
`FOR UPDATE SKIP LOCKED`. Gravar a Notificação e enfileirar a 1ª Tentativa na
mesma transação elimina a perda entre "gravou" e "publicou" sem outbox, e não
acrescenta um componente para operar. Broker externo entra depois, como
adaptador de entrada, se o volume ou a necessidade de os Sistemas de origem
publicarem em fila exigirem.
