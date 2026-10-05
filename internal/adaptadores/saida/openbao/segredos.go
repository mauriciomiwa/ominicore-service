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
	PostgresDSN       string            // <mount>/<prefixo>/postgres: dsn
	SendGridAPIKey    string            // <mount>/<prefixo>/sendgrid: api_key
	SendGridRemetente string            // <mount>/<prefixo>/sendgrid: remetente
	Twilio            map[string]string // <mount>/<prefixo>/twilio: CamposTwilio
	WhatsApp          map[string]string // <mount>/<prefixo>/whatsapp: CamposWhatsApp; nil = Canal desligado
	WhatsAppLigado    bool              // o caminho existe: Canal WhatsApp ligado (REQ-20)
	Webhooks          map[string]string // <mount>/<prefixo>/webhooks: CamposWebhooks
	ChaveOptOut       string            // <mount>/<prefixo>/optout: chave (HMAC dos contatos, permanente)
}

// CamposWebhooks validam a assinatura dos Avisos de status (REQ-12).
var CamposWebhooks = []string{"sendgrid_chave_publica", "twilio_auth_token"}

// CamposWebhooksMeta são exigidos em <mount>/<prefixo>/webhooks só com o Canal
// WhatsApp ligado (REQ-6, REQ-20).
var CamposWebhooksMeta = []string{"meta_app_secret", "meta_verify_token"}

// CamposWhatsApp são os campos exigidos em <mount>/<prefixo>/whatsapp (REQ-10).
// Sem o caminho, o Canal fica desligado; com ele incompleto, o boot falha.
var CamposWhatsApp = []string{"token", "phone_number_id"}

// CamposTwilio são os campos exigidos em <mount>/<prefixo>/twilio (REQ-7).
var CamposTwilio = []string{"account_sid", "api_key_sid", "api_key_secret", "messaging_service_sid"}

// Config diz como chegar e entrar no OpenBao. Token é só para o ambiente local
// (bao em modo dev); hml e prod usam JWT + Role (ADR 0002).
type Config struct {
	Endereco string
	Mount    string // padrão "kv"
	Prefixo  string // padrão "ominicore" (prod); hml usa "ominicore-hml" — DOC/Infra/AMBIENTES.md
	Token    string
	JWT      string
	Role     string
	// AuthMount é o mount do auth jwt: um por realm, porque cada mount aceita um
	// só emissor (prod "jwt-ominicore", hml "jwt-ominicore-hml"); padrão "jwt".
	AuthMount string
}

// ErrOpenBao marca qualquer falha de acesso ao cofre: o processo não inicia.
var ErrOpenBao = errors.New("openbao indisponível ou acesso negado")

// Carregar entra no OpenBao e lê todos os segredos; falta de qualquer um é
// erro, salvo o caminho inteiro do WhatsApp, que só desliga o Canal.
func Carregar(ctx context.Context, cfg Config) (Segredos, error) {
	cli, err := entrar(ctx, cfg)
	if err != nil {
		return Segredos{}, fmt.Errorf("%w: %w", ErrOpenBao, err)
	}
	mount := cfg.Mount
	if mount == "" {
		mount = "kv"
	}
	prefixo := cfg.Prefixo
	if prefixo == "" {
		prefixo = "ominicore"
	}
	pg, err := ler(ctx, cli, mount, prefixo+"/postgres", "dsn")
	if err != nil {
		return Segredos{}, err
	}
	sg, err := ler(ctx, cli, mount, prefixo+"/sendgrid", "api_key", "remetente")
	if err != nil {
		return Segredos{}, err
	}
	tw, err := ler(ctx, cli, mount, prefixo+"/twilio", CamposTwilio...)
	if err != nil {
		return Segredos{}, err
	}
	wa, err := ler(ctx, cli, mount, prefixo+"/whatsapp", CamposWhatsApp...)
	waLigado := err == nil
	if errors.Is(err, bao.ErrSecretNotFound) {
		wa, err = nil, nil
	}
	if err != nil {
		return Segredos{}, err
	}
	camposWH := CamposWebhooks
	if waLigado {
		camposWH = append(append([]string{}, CamposWebhooks...), CamposWebhooksMeta...)
	}
	wh, err := ler(ctx, cli, mount, prefixo+"/webhooks", camposWH...)
	if err != nil {
		return Segredos{}, err
	}
	oo, err := ler(ctx, cli, mount, prefixo+"/optout", "chave")
	if err != nil {
		return Segredos{}, err
	}
	return Segredos{
		PostgresDSN: pg["dsn"], SendGridAPIKey: sg["api_key"], SendGridRemetente: sg["remetente"],
		Twilio: tw, WhatsApp: wa, WhatsAppLigado: waLigado, Webhooks: wh, ChaveOptOut: oo["chave"],
	}, nil
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
	mountAuth := cfg.AuthMount
	if mountAuth == "" {
		mountAuth = "jwt"
	}
	login, err := cli.Logical().WriteWithContext(ctx, "auth/"+mountAuth+"/login", map[string]any{"role": cfg.Role, "jwt": cfg.JWT})
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
