package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ominicore/internal/adaptadores/saida/postgres/db"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// naOrganizacao roda fn numa transação com a Organização fixada (set_config
// local): a política por_organizacao de `notificacoes` só deixa ver e gravar
// as dela (REQ-52, design D3). Toda leitura da API passa por aqui; o worker,
// a retenção e os Avisos de status não, e veem todas as Organizações.
func naOrganizacao(ctx context.Context, pool *pgxpool.Pool, org dominio.Organizacao, fn func(*db.Queries) error) error {
	if org == "" {
		return dominio.ErrSemOrganizacao
	}
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		if err := q.FixarOrganizacao(ctx, string(org)); err != nil {
			return fmt.Errorf("fixar organização: %w", err)
		}
		return fn(q)
	})
}

// ValidadeCacheOrganizacao: o vínculo client → Organização só muda por
// migration, então um minuto de atraso é aceitável (design D2).
const ValidadeCacheOrganizacao = time.Minute

// Organizacoes resolve a Organização do client do token, com cache em memória.
type Organizacoes struct {
	Pool *pgxpool.Pool
	// Agora é o relógio do cache; nil = time.Now.
	Agora func() time.Time

	mu    sync.Mutex
	cache map[string]entradaOrganizacao
}

type entradaOrganizacao struct {
	org    dominio.Organizacao
	ok     bool
	expira time.Time
}

var _ portas.OrganizacoesPort = (*Organizacoes)(nil)

// DoSistema devolve a Organização do client; ok=false quando não há vínculo.
// Também guarda o "sem Organização", para não consultar o banco a cada recusa.
func (o *Organizacoes) DoSistema(ctx context.Context, client string) (dominio.Organizacao, bool, error) {
	agora := time.Now
	if o.Agora != nil {
		agora = o.Agora
	}
	o.mu.Lock()
	e, achou := o.cache[client]
	o.mu.Unlock()
	if achou && agora().Before(e.expira) {
		return e.org, e.ok, nil
	}
	org, err := db.New(o.Pool).OrganizacaoDoSistema(ctx, client)
	ok := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", false, fmt.Errorf("organização do sistema de origem: %w", err)
	}
	e = entradaOrganizacao{org: dominio.Organizacao(org), ok: ok, expira: agora().Add(ValidadeCacheOrganizacao)}
	o.mu.Lock()
	if o.cache == nil {
		o.cache = map[string]entradaOrganizacao{}
	}
	o.cache[client] = e
	o.mu.Unlock()
	return e.org, e.ok, nil
}
