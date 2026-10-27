// Package openbao lê as credenciais do OminiCore no KV v2 do OpenBao (REQ-6).
package openbao

import (
	"context"
	"errors"
	"fmt"
	"sync"

	bao "github.com/openbao/openbao/api/v2"

	"ominicore/internal/dominio"
)

// Segredos reunidos no boot. Nenhum deles sai daqui para log.
type Segredos struct {
	PostgresDSN       string                   // <mount>/<prefixo>/postgres: dsn
	SendGridAPIKey    string                   // <mount>/<prefixo>/sendgrid: api_key
	SendGridRemetente string                   // <mount>/<prefixo>/sendgrid: remetente
	Twilio            map[string]string        // <mount>/<prefixo>/twilio: CamposTwilio
	WhatsApp          map[string]string        // <mount>/<prefixo>/whatsapp: campos do Provedor; nil = Canal desligado
	WhatsAppLigado    bool                     // o caminho existe: Canal WhatsApp ligado (REQ-20)
	ProvedorWhatsApp  dominio.ProvedorWhatsApp // campo `provedor` (ausente = meta); vazio com o Canal desligado
	Webhooks          map[string]string        // <mount>/<prefixo>/webhooks: CamposWebhooks
	ChaveOptOut       string                   // <mount>/<prefixo>/optout: chave (HMAC dos contatos, permanente)
}

// ConfigWhatsApp devolve a configuração do Canal lida no boot.
func (s Segredos) ConfigWhatsApp() dominio.ConfigWhatsApp {
	return dominio.ConfigWhatsApp{Ligado: s.WhatsAppLigado, Provedor: s.ProvedorWhatsApp, Campos: s.WhatsApp}
}

// CamposWebhooks validam a assinatura dos Avisos de status (REQ-12).
var CamposWebhooks = []string{"sendgrid_chave_publica", "twilio_auth_token"}

// CamposWebhooksMeta moram em <mount>/<prefixo>/whatsapp com a Meta; o caminho
// dos webhooks é a alternativa de quem foi configurado antes da API (REQ-28).
var CamposWebhooksMeta = []string{"meta_app_secret", "meta_verify_token"}

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
	// RenovarJWT busca um JWT novo no Keycloak quando a sessão no cofre expira
	// (o token do login vale 1 h); nil = sem re-login (token local).
	RenovarJWT func(context.Context) (string, error)
}

// ErrOpenBao marca qualquer falha de acesso ao cofre: o processo não inicia.
var ErrOpenBao = errors.New("openbao indisponível ou acesso negado")

// Carregar entra no cofre e lê todos os segredos do boot; falha fechado (T-13).
func Carregar(ctx context.Context, cfg Config) (Segredos, error) {
	c, err := Abrir(ctx, cfg)
	if err != nil {
		return Segredos{}, err
	}
	return c.Carregar(ctx)
}

// Cofre é a sessão no OpenBao que fica viva depois do boot: relê e grava a
// configuração do WhatsApp (REQ-22, REQ-24, REQ-26), refazendo o login quando
// a sessão expira.
type Cofre struct {
	cfg     Config
	mount   string
	prefixo string
	mu      sync.Mutex
	cli     *bao.Client
}

// Abrir entra no cofre.
func Abrir(ctx context.Context, cfg Config) (*Cofre, error) {
	c := &Cofre{cfg: cfg, mount: cfg.Mount, prefixo: cfg.Prefixo}
	if c.mount == "" {
		c.mount = "kv"
	}
	if c.prefixo == "" {
		c.prefixo = "ominicore"
	}
	if err := c.entrar(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenBao, err)
	}
	return c, nil
}

// Carregar lê todos os segredos do boot.
func (c *Cofre) Carregar(ctx context.Context) (Segredos, error) {
	pg, err := c.ler(ctx, "postgres", "dsn")
	if err != nil {
		return Segredos{}, err
	}
	sg, err := c.ler(ctx, "sendgrid", "api_key", "remetente")
	if err != nil {
		return Segredos{}, err
	}
	tw, err := c.ler(ctx, "twilio", CamposTwilio...)
	if err != nil {
		return Segredos{}, err
	}
	wa, err := c.LerWhatsApp(ctx)
	if err != nil {
		return Segredos{}, err
	}
	wh, err := c.ler(ctx, "webhooks", CamposWebhooks...)
	if err != nil {
		return Segredos{}, err
	}
	oo, err := c.ler(ctx, "optout", "chave")
	if err != nil {
		return Segredos{}, err
	}
	return Segredos{
		PostgresDSN: pg["dsn"], SendGridAPIKey: sg["api_key"], SendGridRemetente: sg["remetente"],
		Twilio: tw, WhatsApp: wa.Campos, WhatsAppLigado: wa.Ligado, ProvedorWhatsApp: wa.Provedor, Webhooks: wh, ChaveOptOut: oo["chave"],
	}, nil
}

// LerWhatsApp lê a configuração do Canal: caminho ausente = desligado. Com a
// Meta, os segredos do webhook que faltarem no caminho do WhatsApp vêm do
// caminho dos webhooks (REQ-28). Configuração incompleta é erro que cita o campo.
func (c *Cofre) LerWhatsApp(ctx context.Context) (dominio.ConfigWhatsApp, error) {
	var dados map[string]any
	apagada := false
	err := c.comSessao(ctx, func(cli *bao.Client) error {
		s, err := cli.KVv2(c.mount).Get(ctx, c.prefixo+"/whatsapp")
		if err == nil {
			dados = s.Data
			apagada = s.VersionMetadata != nil && !s.VersionMetadata.DeletionTime.IsZero()
		}
		return err
	})
	// Desligar é soft delete no KV v2 (REQ-24): a versão apagada continua lá,
	// sem dados e com deletion_time — é Canal desligado, não configuração
	// incompleta. Gravar de novo cria uma versão nova e religa.
	if errors.Is(err, bao.ErrSecretNotFound) || (err == nil && (apagada || len(dados) == 0)) {
		return dominio.ConfigWhatsApp{}, nil
	}
	if err != nil {
		return dominio.ConfigWhatsApp{}, fmt.Errorf("%w: ler %s: %w", ErrOpenBao, c.caminho("whatsapp"), err)
	}
	campos := texto(dados)
	cfg, err := dominio.NovaConfigWhatsApp(campos["provedor"], campos)
	if err != nil {
		return dominio.ConfigWhatsApp{}, fmt.Errorf("%w: %s: %w", ErrOpenBao, c.caminho("whatsapp"), err)
	}
	if cfg.Provedor == dominio.WhatsAppMeta && (cfg.Campos["meta_app_secret"] == "" || cfg.Campos["meta_verify_token"] == "") {
		if err := c.completarMeta(ctx, cfg.Campos); err != nil {
			return dominio.ConfigWhatsApp{}, err
		}
	}
	if err := cfg.Validar(); err != nil {
		return dominio.ConfigWhatsApp{}, fmt.Errorf("%w: %s: %w", ErrOpenBao, c.caminho("whatsapp"), err)
	}
	return cfg, nil
}

// completarMeta busca no caminho dos webhooks os segredos da Meta ausentes.
func (c *Cofre) completarMeta(ctx context.Context, campos map[string]string) error {
	var dados map[string]any
	err := c.comSessao(ctx, func(cli *bao.Client) error {
		s, err := cli.KVv2(c.mount).Get(ctx, c.prefixo+"/webhooks")
		if err == nil {
			dados = s.Data
		}
		return err
	})
	if err != nil && !errors.Is(err, bao.ErrSecretNotFound) {
		return fmt.Errorf("%w: ler %s: %w", ErrOpenBao, c.caminho("webhooks"), err)
	}
	wh := texto(dados)
	for _, k := range CamposWebhooksMeta {
		if campos[k] == "" && wh[k] != "" {
			campos[k] = wh[k]
		}
	}
	return nil
}

// GravarWhatsApp substitui a configuração do Canal pela dada, já validada
// (REQ-22).
func (c *Cofre) GravarWhatsApp(ctx context.Context, cfg dominio.ConfigWhatsApp) error {
	dados := map[string]any{"provedor": string(cfg.Provedor)}
	for k, v := range cfg.Campos {
		dados[k] = v
	}
	err := c.comSessao(ctx, func(cli *bao.Client) error {
		_, err := cli.KVv2(c.mount).Put(ctx, c.prefixo+"/whatsapp", dados)
		return err
	})
	if err != nil {
		return fmt.Errorf("%w: gravar %s: %w", ErrOpenBao, c.caminho("whatsapp"), err)
	}
	return nil
}

// ApagarWhatsApp desliga o Canal: apaga a versão atual (as anteriores ficam no
// histórico do KV v2); idempotente (REQ-24).
func (c *Cofre) ApagarWhatsApp(ctx context.Context) error {
	err := c.comSessao(ctx, func(cli *bao.Client) error {
		return cli.KVv2(c.mount).Delete(ctx, c.prefixo+"/whatsapp")
	})
	if err != nil && !errors.Is(err, bao.ErrSecretNotFound) {
		return fmt.Errorf("%w: apagar %s: %w", ErrOpenBao, c.caminho("whatsapp"), err)
	}
	return nil
}

func (c *Cofre) caminho(nome string) string { return c.mount + "/" + c.prefixo + "/" + nome }

// comSessao roda f e, se o cofre negar por sessão expirada (403) e houver como
// renovar o JWT, refaz o login uma vez e tenta de novo.
func (c *Cofre) comSessao(ctx context.Context, f func(*bao.Client) error) error {
	c.mu.Lock()
	cli := c.cli
	c.mu.Unlock()
	err := f(cli)
	var re *bao.ResponseError
	if err == nil || c.cfg.RenovarJWT == nil || !errors.As(err, &re) || re.StatusCode != 403 {
		return err
	}
	jwt, errJWT := c.cfg.RenovarJWT(ctx)
	if errJWT != nil {
		return fmt.Errorf("renovar jwt: %w", errJWT)
	}
	c.mu.Lock()
	c.cfg.JWT = jwt
	errLogin := c.entrarSemTrava(ctx)
	cli = c.cli
	c.mu.Unlock()
	if errLogin != nil {
		return errLogin
	}
	return f(cli)
}

func (c *Cofre) entrar(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Token == "" && c.cfg.JWT == "" && c.cfg.RenovarJWT != nil {
		jwt, err := c.cfg.RenovarJWT(ctx)
		if err != nil {
			return fmt.Errorf("obter jwt: %w", err)
		}
		c.cfg.JWT = jwt
	}
	return c.entrarSemTrava(ctx)
}

func (c *Cofre) entrarSemTrava(ctx context.Context) error {
	bcfg := bao.DefaultConfig()
	bcfg.Address = c.cfg.Endereco
	cli, err := bao.NewClient(bcfg)
	if err != nil {
		return err
	}
	if c.cfg.Token != "" {
		cli.SetToken(c.cfg.Token)
		c.cli = cli
		return nil
	}
	if c.cfg.JWT == "" || c.cfg.Role == "" {
		return errors.New("sem token nem jwt+role para entrar")
	}
	mountAuth := c.cfg.AuthMount
	if mountAuth == "" {
		mountAuth = "jwt"
	}
	login, err := cli.Logical().WriteWithContext(ctx, "auth/"+mountAuth+"/login", map[string]any{"role": c.cfg.Role, "jwt": c.cfg.JWT})
	if err != nil {
		return fmt.Errorf("login jwt: %w", err)
	}
	if login == nil || login.Auth == nil {
		return errors.New("login jwt sem token")
	}
	cli.SetToken(login.Auth.ClientToken)
	c.cli = cli
	return nil
}

func (c *Cofre) ler(ctx context.Context, nome string, campos ...string) (map[string]string, error) {
	var dados map[string]any
	err := c.comSessao(ctx, func(cli *bao.Client) error {
		s, err := cli.KVv2(c.mount).Get(ctx, c.prefixo+"/"+nome)
		if err == nil {
			dados = s.Data
		}
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("%w: ler %s: %w", ErrOpenBao, c.caminho(nome), err)
	}
	out := map[string]string{}
	for _, k := range campos {
		v, _ := dados[k].(string)
		if v == "" {
			return nil, fmt.Errorf("%w: %s sem o campo %s", ErrOpenBao, c.caminho(nome), k)
		}
		out[k] = v
	}
	return out, nil
}

// texto copia os campos de texto do segredo.
func texto(dados map[string]any) map[string]string {
	out := map[string]string{}
	for k, v := range dados {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}
	return out
}
