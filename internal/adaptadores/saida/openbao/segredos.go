// Package openbao lê as credenciais do OminiCore no KV v2 do OpenBao (REQ-6).
package openbao

import (
	"context"
	"errors"
	"fmt"

	bao "github.com/openbao/openbao/api/v2"
)

// Segredos reunidos no boot. Nenhum deles sai daqui para log.
type Segredos struct {
	PostgresDSN       string // kv/ominicore/postgres: dsn
	SendGridAPIKey    string // kv/ominicore/sendgrid: api_key
	SendGridRemetente string // kv/ominicore/sendgrid: remetente
}

// Config diz como chegar e entrar no OpenBao. Token é só para o ambiente local
// (bao em modo dev); hml e prod usam JWT + Role (ADR 0002).
type Config struct {
	Endereco string
	Mount    string // padrão "kv"
	Token    string
	JWT      string
	Role     string
}

// ErrOpenBao marca qualquer falha de acesso ao cofre: o processo não inicia.
var ErrOpenBao = errors.New("openbao indisponível ou acesso negado")

// Carregar entra no OpenBao e lê todos os segredos; falta de qualquer um é erro.
func Carregar(ctx context.Context, cfg Config) (Segredos, error) {
	cli, err := entrar(ctx, cfg)
	if err != nil {
		return Segredos{}, fmt.Errorf("%w: %w", ErrOpenBao, err)
	}
	mount := cfg.Mount
	if mount == "" {
		mount = "kv"
	}
	pg, err := ler(ctx, cli, mount, "ominicore/postgres", "dsn")
	if err != nil {
		return Segredos{}, err
	}
	sg, err := ler(ctx, cli, mount, "ominicore/sendgrid", "api_key", "remetente")
	if err != nil {
		return Segredos{}, err
	}
	return Segredos{PostgresDSN: pg["dsn"], SendGridAPIKey: sg["api_key"], SendGridRemetente: sg["remetente"]}, nil
}

func entrar(ctx context.Context, cfg Config) (*bao.Client, error) {
	bcfg := bao.DefaultConfig()
	bcfg.Address = cfg.Endereco
	cli, err := bao.NewClient(bcfg)
	if err != nil {
		return nil, err
	}
	if cfg.Token != "" {
		cli.SetToken(cfg.Token)
		return cli, nil
	}
	if cfg.JWT == "" || cfg.Role == "" {
		return nil, errors.New("sem token nem jwt+role para entrar")
	}
	login, err := cli.Logical().WriteWithContext(ctx, "auth/jwt/login", map[string]any{"role": cfg.Role, "jwt": cfg.JWT})
	if err != nil {
		return nil, fmt.Errorf("login jwt: %w", err)
	}
	if login == nil || login.Auth == nil {
		return nil, errors.New("login jwt sem token")
	}
	cli.SetToken(login.Auth.ClientToken)
	return cli, nil
}

func ler(ctx context.Context, cli *bao.Client, mount, caminho string, campos ...string) (map[string]string, error) {
	s, err := cli.KVv2(mount).Get(ctx, caminho)
	if err != nil {
		return nil, fmt.Errorf("%w: ler %s/%s: %w", ErrOpenBao, mount, caminho, err)
	}
	out := map[string]string{}
	for _, c := range campos {
		v, _ := s.Data[c].(string)
		if v == "" {
			return nil, fmt.Errorf("%w: %s/%s sem o campo %s", ErrOpenBao, mount, caminho, c)
		}
		out[c] = v
	}
	return out, nil
}
