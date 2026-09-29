-- name: BuscarTemplate :many
SELECT t.id, t.ordem_canais, c.canal, c.assunto, c.corpo, c.modelo_aprovado, c.idioma, c.parametros
FROM templates t
JOIN template_conteudos c ON c.template_id = t.id
WHERE t.id = $1;

-- name: ReivindicarChave :one
-- Grava a chave para uma Notificação nova; devolve linha só quando a chave é
-- nova ou estava vencida (REQ-3).
INSERT INTO idempotencia (sistema_origem, chave, notificacao_id, expira_em)
VALUES ($1, $2, gen_random_uuid(), now() + interval '24 hours')
ON CONFLICT (sistema_origem, chave) DO UPDATE
    SET notificacao_id = EXCLUDED.notificacao_id, expira_em = EXCLUDED.expira_em
    WHERE idempotencia.expira_em < now()
RETURNING notificacao_id;

-- name: BuscarPorChave :one
SELECT n.id, n.status
FROM idempotencia i
JOIN notificacoes n ON n.id = i.notificacao_id
WHERE i.sistema_origem = $1 AND i.chave = $2;

-- name: InserirNotificacao :exec
INSERT INTO notificacoes (id, sistema_origem, template_id, destinatario_email, destinatario_telefone, dados, ordem_canais, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'pendente');

-- espera_ms conta a partir do relógio do banco, o mesmo que a fila compara.
-- name: InserirTentativa :exec
INSERT INTO tentativas (notificacao_id, canal, numero, status, disponivel_em)
VALUES (sqlc.arg(notificacao_id), sqlc.arg(canal), sqlc.arg(numero), 'pendente', now() + (sqlc.arg(espera_ms)::bigint * interval '1 millisecond'));

-- name: ReservarTentativa :one
SELECT t.id, t.notificacao_id, t.canal, t.numero,
       n.sistema_origem, n.template_id, n.destinatario_email, n.destinatario_telefone, n.dados, n.ordem_canais
FROM tentativas t
JOIN notificacoes n ON n.id = t.notificacao_id
WHERE t.status = 'pendente' AND t.disponivel_em <= now()
ORDER BY t.disponivel_em
LIMIT 1
FOR UPDATE OF t SKIP LOCKED;

-- name: ConcluirTentativa :exec
UPDATE tentativas
SET status = $2, motivo = $3, id_provedor = $4, atualizada_em = now()
WHERE id = $1;

-- name: AtualizarStatusNotificacao :exec
UPDATE notificacoes SET status = $2, atualizada_em = now() WHERE id = $1;
