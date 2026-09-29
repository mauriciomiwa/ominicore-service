// Package postgres implementa MessageRepositoryPort sobre o schema `ominicore`.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ominicore/internal/adaptadores/saida/postgres/db"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// Repositorio guarda Notificações, Tentativas e Templates.
type Repositorio struct {
	Pool *pgxpool.Pool
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
		tpl.Conteudos[dominio.Canal(l.Canal)] = dominio.ConteudoCanal{Assunto: l.Assunto, Corpo: l.Corpo}
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
	if err := q.ConcluirTentativa(ctx, db.ConcluirTentativaParams{ID: t.ID, Status: string(d.Tentativa), Motivo: d.Motivo, IDProvedor: res.IDProvedor}); err != nil {
		return err
	}
	if p := d.Proxima; p != nil { // Retentativa (REQ-8) ou próximo Canal (REQ-9), na mesma transação
		if err := q.InserirTentativa(ctx, db.InserirTentativaParams{
			NotificacaoID: t.NotificacaoID, Canal: string(p.Canal), Numero: int32(p.Numero), //nolint:gosec // Numero ≤ 4
			EsperaMs: p.Espera.Milliseconds(),
		}); err != nil {
			return err
		}
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
