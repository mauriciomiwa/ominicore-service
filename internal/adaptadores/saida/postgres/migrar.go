package postgres

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"ominicore/migrations"
)

// SchemaPadrao é o schema de prod; hml usa ominicore_hml (DOC/Infra/AMBIENTES.md).
const SchemaPadrao = "ominicore"

// reSchema aceita só identificador simples: o nome vai para CREATE SCHEMA.
var reSchema = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

func validarSchema(schema string) error {
	if !reSchema.MatchString(schema) {
		return fmt.Errorf("schema inválido: %q", schema)
	}
	return nil
}

// Conectar abre o pool com search_path no schema do ambiente.
func Conectar(ctx context.Context, dsn, schema string) (*pgxpool.Pool, error) {
	if err := validarSchema(schema); err != nil {
		return nil, err
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("dsn do postgres inválido: %w", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("conectar ao postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectar ao postgres: %w", err)
	}
	return pool, nil
}

// Migrar aplica as migrations embutidas. Cria o schema só se ele não existir
// (em hml/prod ele já vem criado, com role próprio).
func Migrar(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	if err := validarSchema(schema); err != nil {
		return err
	}
	var existe bool
	if err := pool.QueryRow(ctx, `SELECT to_regnamespace($1) IS NOT NULL`, schema).Scan(&existe); err != nil {
		return fmt.Errorf("verificar schema: %w", err)
	}
	if !existe {
		if _, err := pool.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
			return fmt.Errorf("criar schema: %w", err)
		}
	}
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() { _ = sqlDB.Close() }()
	prov, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		return fmt.Errorf("preparar migrations: %w", err)
	}
	if _, err := prov.Up(ctx); err != nil {
		return fmt.Errorf("aplicar migrations: %w", err)
	}
	return nil
}
