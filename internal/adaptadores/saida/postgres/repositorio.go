// Package postgres implementa MessageRepositoryPort sobre o schema `ominicore`.
package postgres

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ominicore/internal/adaptadores/saida/postgres/db"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// Repositorio guarda Notificações, Tentativas e Templates.
type Repositorio struct {
	Pool *pgxpool.Pool
	// ChaveOptOut é a chave do HMAC dos contatos em Opt-out (kv/<prefixo>/optout).
	// Permanente: trocá-la invalida todos os Opt-outs (REQ-18).
	ChaveOptOut []byte
}

var _ portas.MessageRepositoryPort = Repositorio{}

// BuscarTemplate devolve o Template com uma versão por Canal.
func (r Repositorio) BuscarTemplate(ctx context.Context, id string) (dominio.Template, error) {
	linhas, err := db.New(r.Pool).BuscarTemplate(ctx, id)
	if err != nil {
		return dominio.Template{}, fmt.Errorf("buscar template: %w", err)
	}
	if len(linhas) == 0 {
		return dominio.Template{}, fmt.Errorf("%w: %s", dominio.ErrTemplateInexistente, id)
	}
	tpl := dominio.Template{ID: id, OrdemCanais: canais(linhas[0].OrdemCanais), Conteudos: map[dominio.Canal]dominio.ConteudoCanal{}}
	for _, l := range linhas {
		tpl.Conteudos[dominio.Canal(l.Canal)] = dominio.ConteudoCanal{
			Assunto: l.Assunto, Corpo: l.Corpo,
			ModeloAprovado: l.ModeloAprovado, Idioma: l.Idioma, Parametros: l.Parametros,
		}
	}
	return tpl, nil
}

// RegistrarNotificacao grava Notificação e primeira Tentativa numa transação.
// Chave já reivindicada nas últimas 24 h devolve a Notificação existente (REQ-3).
func (r Repositorio) RegistrarNotificacao(ctx context.Context, n portas.NovaNotificacao) (portas.Registro, error) {
	dados, err := json.Marshal(n.Dados)
	if err != nil {
		return portas.Registro{}, fmt.Errorf("serializar dados: %w", err)
	}
	var reg portas.Registro
	err = pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		id, err := q.ReivindicarChave(ctx, db.ReivindicarChaveParams{SistemaOrigem: n.SistemaOrigem, Chave: n.ChaveIdempotencia})
		if errors.Is(err, pgx.ErrNoRows) {
			existente, err := q.BuscarPorChave(ctx, db.BuscarPorChaveParams{SistemaOrigem: n.SistemaOrigem, Chave: n.ChaveIdempotencia})
			reg = portas.Registro{ID: existente.ID, Status: dominio.StatusNotificacao(existente.Status), Reutilizada: true}
			return err
		}
		if err != nil {
			return err
		}
		if err := q.InserirNotificacao(ctx, db.InserirNotificacaoParams{
			ID: id, SistemaOrigem: n.SistemaOrigem, TemplateID: n.TemplateID,
			DestinatarioEmail: n.Destinatario.Email, DestinatarioTelefone: n.Destinatario.Telefone,
			Dados: dados, OrdemCanais: textos(n.OrdemCanais),
		}); err != nil {
			return err
		}
		reg = portas.Registro{ID: id, Status: dominio.NotificacaoPendente}
		return q.InserirTentativa(ctx, db.InserirTentativaParams{NotificacaoID: id, Canal: string(n.PrimeiroCanal), Numero: 1})
	})
	if err != nil {
		return portas.Registro{}, fmt.Errorf("registrar notificação: %w", err)
	}
	return reg, nil
}

// ReservarTentativa abre uma transação e trava a próxima Tentativa pendente
// (FOR UPDATE SKIP LOCKED, ADR 0001). A trava dura até Concluir;
// se o processo cair no meio, a Tentativa volta para a fila.
func (r Repositorio) ReservarTentativa(ctx context.Context) (portas.TentativaReservada, bool, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return portas.TentativaReservada{}, false, fmt.Errorf("abrir transação: %w", err)
	}
	linha, err := db.New(tx).ReservarTentativa(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return portas.TentativaReservada{}, false, nil
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		return portas.TentativaReservada{}, false, fmt.Errorf("reservar tentativa: %w", err)
	}
	n, err := notificacaoDe(linha)
	if err != nil {
		_ = tx.Rollback(ctx)
		return portas.TentativaReservada{}, false, err
	}
	if n.Destinatario.OptOut, err = r.optOuts(ctx, db.New(tx), n.Destinatario); err != nil {
		_ = tx.Rollback(ctx)
		return portas.TentativaReservada{}, false, err
	}
	t := dominio.Tentativa{ID: linha.ID, NotificacaoID: linha.NotificacaoID, Canal: dominio.Canal(linha.Canal), Numero: int(linha.Numero), Status: dominio.TentativaPendente}
	return portas.TentativaReservada{
		Tentativa:   t,
		Notificacao: n,
		Concluir: func(ctx context.Context, d dominio.Decisao, res dominio.ResultadoEnvio) error {
			return concluir(ctx, tx, t, d, res)
		},
	}, true, nil
}

func concluir(ctx context.Context, tx pgx.Tx, t dominio.Tentativa, d dominio.Decisao, res dominio.ResultadoEnvio) error {
	defer func() { _ = tx.Rollback(ctx) }()
	q := db.New(tx)
	if err := q.ConcluirTentativa(ctx, db.ConcluirTentativaParams{
		ID: t.ID, Status: string(d.Tentativa), Motivo: d.Motivo, IDProvedor: res.IDProvedor, CategoriaFalha: string(d.Categoria),
	}); err != nil {
		return err
	}
	if err := inserirProxima(ctx, q, t.NotificacaoID, d.Proxima); err != nil { // REQ-8/REQ-9, mesma transação
		return err
	}
	if err := q.AtualizarStatusNotificacao(ctx, db.AtualizarStatusNotificacaoParams{ID: t.NotificacaoID, Status: string(d.Notificacao)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func notificacaoDe(l db.ReservarTentativaRow) (dominio.Notificacao, error) {
	var dados map[string]string
	if err := json.Unmarshal(l.Dados, &dados); err != nil {
		return dominio.Notificacao{}, fmt.Errorf("ler dados da notificação %s: %w", l.NotificacaoID, err)
	}
	return dominio.Notificacao{
		ID: l.NotificacaoID, SistemaOrigem: l.SistemaOrigem, TemplateID: l.TemplateID,
		Destinatario: dominio.Destinatario{Email: l.DestinatarioEmail, Telefone: l.DestinatarioTelefone}, Dados: dados,
		OrdemCanais: canais(l.OrdemCanais), Status: dominio.NotificacaoPendente,
	}, nil
}

func canais(s []string) []dominio.Canal {
	out := make([]dominio.Canal, len(s))
	for i, c := range s {
		out[i] = dominio.Canal(c)
	}
	return out
}

func textos(c []dominio.Canal) []string {
	out := make([]string, len(c))
	for i, v := range c {
		out[i] = string(v)
	}
	return out
}

func inserirProxima(ctx context.Context, q *db.Queries, notificacao string, p *dominio.ProximaTentativa) error {
	if p == nil {
		return nil
	}
	return q.InserirTentativa(ctx, db.InserirTentativaParams{
		NotificacaoID: notificacao, Canal: string(p.Canal), Numero: int32(p.Numero), //nolint:gosec // Numero ≤ 4
		EsperaMs: p.Espera.Milliseconds(),
	})
}

// ReservarPorIDProvedor abre a transação do aviso e trava Tentativa e
// Notificação (REQ-12, REQ-13).
func (r Repositorio) ReservarPorIDProvedor(ctx context.Context, idProvedor string) (portas.AvisoReservado, bool, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return portas.AvisoReservado{}, false, fmt.Errorf("abrir transação: %w", err)
	}
	l, err := db.New(tx).ReservarPorIDProvedor(ctx, idProvedor)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return portas.AvisoReservado{}, false, nil
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		return portas.AvisoReservado{}, false, fmt.Errorf("reservar aviso: %w", err)
	}
	n, err := notificacaoDe(db.ReservarTentativaRow{
		NotificacaoID: l.NotificacaoID, SistemaOrigem: l.SistemaOrigem, TemplateID: l.TemplateID,
		DestinatarioEmail: l.DestinatarioEmail, DestinatarioTelefone: l.DestinatarioTelefone, Dados: l.Dados, OrdemCanais: l.OrdemCanais,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		return portas.AvisoReservado{}, false, err
	}
	n.Status = dominio.StatusNotificacao(l.StatusNotificacao)
	if n.Destinatario.OptOut, err = r.optOuts(ctx, db.New(tx), n.Destinatario); err != nil {
		_ = tx.Rollback(ctx)
		return portas.AvisoReservado{}, false, err
	}
	t := dominio.Tentativa{ID: l.ID, NotificacaoID: l.NotificacaoID, Canal: dominio.Canal(l.Canal), Numero: int(l.Numero), Status: dominio.StatusTentativa(l.StatusTentativa)}
	return portas.AvisoReservado{
		Tentativa: t, Notificacao: n,
		Aplicar: func(ctx context.Context, a dominio.AtualizacaoAviso) error { return aplicarAviso(ctx, tx, t, a) },
		Liberar: func(ctx context.Context) { _ = tx.Rollback(ctx) },
	}, true, nil
}

func aplicarAviso(ctx context.Context, tx pgx.Tx, t dominio.Tentativa, a dominio.AtualizacaoAviso) error {
	defer func() { _ = tx.Rollback(ctx) }()
	q := db.New(tx)
	if err := q.AplicarAviso(ctx, db.AplicarAvisoParams{ID: t.ID, Status: string(a.Tentativa), Motivo: a.Motivo, CategoriaFalha: string(a.Categoria)}); err != nil {
		return err
	}
	if err := inserirProxima(ctx, q, t.NotificacaoID, a.Proxima); err != nil {
		return err
	}
	if a.Notificacao != "" {
		if err := q.AtualizarStatusNotificacao(ctx, db.AtualizarStatusNotificacaoParams{ID: t.NotificacaoID, Status: string(a.Notificacao)}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ConsultarNotificacao filtra pelo Sistema de origem na própria query (REQ-14).
func (r Repositorio) ConsultarNotificacao(ctx context.Context, sistemaOrigem, id string) (portas.Consulta, bool, error) {
	q := db.New(r.Pool)
	n, err := q.ConsultarNotificacao(ctx, db.ConsultarNotificacaoParams{ID: id, SistemaOrigem: sistemaOrigem})
	if errors.Is(err, pgx.ErrNoRows) {
		return portas.Consulta{}, false, nil
	}
	if err != nil {
		return portas.Consulta{}, false, fmt.Errorf("consultar notificação: %w", err)
	}
	linhas, err := q.ListarTentativas(ctx, id)
	if err != nil {
		return portas.Consulta{}, false, fmt.Errorf("listar tentativas: %w", err)
	}
	c := portas.Consulta{
		ID: n.ID, TemplateID: n.TemplateID, Status: dominio.StatusNotificacao(n.Status),
		CriadaEm: n.CriadaEm, AtualizadaEm: n.AtualizadaEm, AnonimizadaEm: n.AnonimizadaEm,
	}
	for _, l := range linhas {
		c.Tentativas = append(c.Tentativas, portas.TentativaConsultada{
			Canal: dominio.Canal(l.Canal), Numero: int(l.Numero), Status: dominio.StatusTentativa(l.Status),
			Categoria: dominio.CategoriaFalha(l.CategoriaFalha), CriadaEm: l.CriadaEm, AtualizadaEm: l.AtualizadaEm,
		})
	}
	return c, true, nil
}

// hmacContato: HMAC-SHA256(chave, canal|contato normalizado) — o que fica no banco (REQ-18).
func (r Repositorio) hmacContato(c dominio.Canal, contato string) ([]byte, error) {
	norm, err := dominio.NormalizarContato(c, contato)
	if err != nil {
		return nil, err
	}
	m := hmac.New(sha256.New, r.ChaveOptOut)
	m.Write([]byte(string(c) + "|" + norm))
	return m.Sum(nil), nil
}

func (r Repositorio) optOuts(ctx context.Context, q *db.Queries, d dominio.Destinatario) (map[dominio.Canal]bool, error) {
	var hmacs [][]byte
	for _, c := range []dominio.Canal{dominio.CanalEmail, dominio.CanalSMS, dominio.CanalWhatsApp} {
		if d.Contato(c) == "" {
			continue
		}
		if h, err := r.hmacContato(c, d.Contato(c)); err == nil { // contato malformado não tem Opt-out
			hmacs = append(hmacs, h)
		}
	}
	if len(hmacs) == 0 {
		return nil, nil
	}
	linhas, err := q.CanaisEmOptOut(ctx, hmacs)
	if err != nil {
		return nil, fmt.Errorf("consultar opt-out: %w", err)
	}
	out := map[dominio.Canal]bool{}
	for _, c := range linhas {
		out[dominio.Canal(c)] = true
	}
	return out, nil
}

// OptOutsDe diz em quais Canais os contatos estão em Opt-out (pedido, REQ-17).
func (r Repositorio) OptOutsDe(ctx context.Context, d dominio.Destinatario) (map[dominio.Canal]bool, error) {
	return r.optOuts(ctx, db.New(r.Pool), d)
}

// RegistrarOptOut grava o Opt-out e, se for novo, o evento de registro.
func (r Repositorio) RegistrarOptOut(ctx context.Context, c dominio.Canal, contato, por string) error {
	h, err := r.hmacContato(c, contato)
	if err != nil {
		return err
	}
	return pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		n, err := q.RegistrarOptOut(ctx, db.RegistrarOptOutParams{Canal: string(c), ContatoHmac: h, Origem: por})
		if err != nil || n == 0 {
			return err // já existia: idempotente, sem evento novo
		}
		return q.InserirEventoOptOut(ctx, db.InserirEventoOptOutParams{Canal: string(c), ContatoHmac: h, Acao: "registrado", Por: por})
	})
}

// RemoverOptOut apaga o Opt-out e, se existia, grava quem removeu (REQ-15).
func (r Repositorio) RemoverOptOut(ctx context.Context, c dominio.Canal, contato, por string) error {
	h, err := r.hmacContato(c, contato)
	if err != nil {
		return err
	}
	return pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		n, err := q.RemoverOptOut(ctx, db.RemoverOptOutParams{Canal: string(c), ContatoHmac: h})
		if err != nil || n == 0 {
			return err
		}
		return q.InserirEventoOptOut(ctx, db.InserirEventoOptOutParams{Canal: string(c), ContatoHmac: h, Acao: "removido", Por: por})
	})
}

// AnonimizarLote apaga contato, dados e motivo de um lote de Notificações
// finalizadas fora do prazo (REQ-19), numa transação.
func (r Repositorio) AnonimizarLote(ctx context.Context, antes time.Time, lote int) (int, error) {
	ids, err := db.New(r.Pool).AnonimizarLote(ctx, db.AnonimizarLoteParams{Antes: antes, Lote: int32(lote)}) //nolint:gosec // lote ≤ 1000
	if err != nil {
		return 0, fmt.Errorf("anonimizar lote: %w", err)
	}
	return len(ids), nil
}

// PendentesVencidas conta Notificações pendentes que já passaram do prazo.
func (r Repositorio) PendentesVencidas(ctx context.Context, antes time.Time) (int, error) {
	n, err := db.New(r.Pool).ContarPendentesVencidas(ctx, antes)
	return int(n), err
}
