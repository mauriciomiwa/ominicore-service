package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"ominicore/migrations"
)

// Schema é o schema do OminiCore; a conexão usa search_path=ominicore.
const Schema = "ominicore"

// Conectar abre o pool com search_path no schema do OminiCore.
func Conectar(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("dsn do postgres inválido: %w", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = Schema
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
func Migrar(ctx context.Context, pool *pgxpool.Pool) error {
	var existe bool
	if err := pool.QueryRow(ctx, `SELECT to_regnamespace($1) IS NOT NULL`, Schema).Scan(&existe); err != nil {
		return fmt.Errorf("verificar schema: %w", err)
	}
	if !existe {
		if _, err := pool.Exec(ctx, `CREATE SCHEMA `+Schema); err != nil {
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
