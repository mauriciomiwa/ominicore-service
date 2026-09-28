// Package redis implementa o cache do JWKS sobre o Redis da casa.
package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Cache guarda valores com validade.
type Cache struct {
	Cliente *goredis.Client
}

// Get devolve ok=false quando a chave não existe.
func (c Cache) Get(ctx context.Context, chave string) ([]byte, bool, error) {
	v, err := c.Cliente.Get(ctx, chave).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

// Set grava com validade.
func (c Cache) Set(ctx context.Context, chave string, valor []byte, ttl time.Duration) error {
	return c.Cliente.Set(ctx, chave, valor, ttl).Err()
}
