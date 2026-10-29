package aplicacao

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"time"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// ErrTrocaSemRegistro: a configuração foi gravada (ou apagada) no cofre, mas o
// registro da troca falhou; o log de erro fica como registro (REQ-27).
var ErrTrocaSemRegistro = errors.New("configuração trocada sem registro")

// ErrRecadastroSoZAPI: o recadastro dos webhooks só vale com a Z-API ativa (T-135).
var ErrRecadastroSoZAPI = errors.New("recadastro dos webhooks só vale para a Z-API")

// ConfigurarUseCase deixa o Operador consultar, gravar e apagar a configuração
// do Canal WhatsApp (RN-11, REQ-22 a REQ-24, REQ-27).
type ConfigurarUseCase struct {
	Cofre    portas.CofreWhatsAppPort
	Registro portas.RegistroConfigPort
	Atual    *ConfigWhatsAppAtual // a troca vale na hora neste processo (REQ-26)
	Log      *slog.Logger
	// Cadastro cadastra na Z-API as URLs dos avisos (REQ-40); nil = não cadastra.
	Cadastro portas.CadastroWebhooksPort
	// URLPublica é a origem pública dos webhooks (OMINICORE_WEBHOOK_URL); vazia
	// (local) = o cadastro não é tentado e o estado fica "falhou".
	URLPublica string
	Agora      func() time.Time       // nil = time.Now
	Segredo    func() (string, error) // nil = 32 bytes de crypto/rand em base64url
}

// Consultar lê a configuração no cofre; quem chama expõe só Publicos e
// Segredos (REQ-23).
func (uc ConfigurarUseCase) Consultar(ctx context.Context) (dominio.ConfigWhatsApp, error) {
	return uc.Cofre.LerWhatsApp(ctx)
}

// Gravar valida e substitui a configuração; Provedor desconhecido ou campo
// faltando não grava nada (T-93, T-94).
// Com a Z-API, gera um segredo novo para a URL dos avisos (o anterior deixa de
// valer) e cadastra as URLs na Z-API; falha no cadastro não desfaz a gravação
// (REQ-40, T-131, T-132).
func (uc ConfigurarUseCase) Gravar(ctx context.Context, provedor string, campos map[string]string, por string) (dominio.ConfigWhatsApp, error) {
	campos = maps.Clone(campos)
	for _, k := range dominio.CamposDoOminiCore {
		delete(campos, k) // vêm só do OminiCore, nunca do Operador
	}
	cfg, err := dominio.NovaConfigWhatsApp(provedor, campos)
	if err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	if err := cfg.Validar(); err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	if cfg.Provedor == dominio.WhatsAppZAPI {
		segredo, err := uc.novoSegredo()
		if err != nil {
			return dominio.ConfigWhatsApp{}, err
		}
		cfg.Campos[dominio.CampoWebhookSegredo] = segredo
		cfg.Campos[dominio.CampoWebhooksEstado] = dominio.WebhooksFalhou
	}
	if err := uc.gravar(ctx, cfg); err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	errRegistro := uc.registrar(ctx, portas.EventoConfig{Canal: dominio.CanalWhatsApp, Acao: "gravada", Provedor: cfg.Provedor, Por: por})
	if cfg.Provedor == dominio.WhatsAppZAPI {
		cfg = uc.cadastrar(ctx, cfg)
	}
	return cfg, errRegistro
}

// Recadastrar repete o cadastro dos webhooks na Z-API com o segredo atual, sem
// trocá-lo (REQ-41); sem a Z-API ativa, recusa sem efeito (T-135). Configuração
// gravada antes da fatia 13 ainda não tem segredo: ganha um.
func (uc ConfigurarUseCase) Recadastrar(ctx context.Context, por string) (dominio.ConfigWhatsApp, error) {
	cfg, err := uc.Cofre.LerWhatsApp(ctx)
	if err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	if !cfg.Ligado || cfg.Provedor != dominio.WhatsAppZAPI {
		return dominio.ConfigWhatsApp{}, ErrRecadastroSoZAPI
	}
	cfg.Campos = maps.Clone(cfg.Campos)
	if cfg.Campos[dominio.CampoWebhookSegredo] == "" {
		segredo, err := uc.novoSegredo()
		if err != nil {
			return dominio.ConfigWhatsApp{}, err
		}
		cfg.Campos[dominio.CampoWebhookSegredo] = segredo
	}
	cfg = uc.cadastrar(ctx, cfg)
	return cfg, uc.registrar(ctx, portas.EventoConfig{Canal: dominio.CanalWhatsApp, Acao: "webhooks", Provedor: cfg.Provedor, Por: por})
}

// cadastrar chama a Z-API e grava o estado do cadastro; a falha fica no log,
// sem o segredo nem o token.
func (uc ConfigurarUseCase) cadastrar(ctx context.Context, cfg dominio.ConfigWhatsApp) dominio.ConfigWhatsApp {
	estado := dominio.WebhooksFalhou
	switch {
	case uc.Cadastro == nil || uc.URLPublica == "":
		uc.Log.WarnContext(ctx, "webhooks da z-api não cadastrados: sem URL pública (OMINICORE_WEBHOOK_URL)")
	default:
		url := uc.URLPublica + "/webhooks/zapi/" + cfg.Campos[dominio.CampoWebhookSegredo]
		if err := uc.Cadastro.Cadastrar(ctx, cfg, url); err != nil {
			uc.Log.ErrorContext(ctx, "webhooks da z-api não cadastrados", "erro", err.Error())
		} else {
			estado = dominio.WebhooksCadastrados
			uc.Log.InfoContext(ctx, "webhooks da z-api cadastrados")
		}
	}
	cfg.Campos = maps.Clone(cfg.Campos)
	cfg.Campos[dominio.CampoWebhooksEstado] = estado
	cfg.Campos[dominio.CampoWebhooksEm] = uc.agora().UTC().Format(time.RFC3339)
	if err := uc.gravar(ctx, cfg); err != nil {
		uc.Log.ErrorContext(ctx, "estado dos webhooks não gravado no cofre", "erro", err.Error())
	}
	return cfg
}

// gravar grava no cofre e aplica no processo, serializado com a releitura.
func (uc ConfigurarUseCase) gravar(ctx context.Context, cfg dominio.ConfigWhatsApp) error {
	var err error
	uc.Atual.Trocar(func() {
		if err = uc.Cofre.GravarWhatsApp(ctx, cfg); err == nil {
			uc.Atual.Aplicar(cfg)
		}
	})
	return err
}

func (uc ConfigurarUseCase) agora() time.Time {
	if uc.Agora != nil {
		return uc.Agora()
	}
	return time.Now()
}

// novoSegredo: 32 bytes aleatórios em base64url, seguro para caminho de URL.
func (uc ConfigurarUseCase) novoSegredo() (string, error) {
	if uc.Segredo != nil {
		return uc.Segredo()
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("gerar segredo do webhook: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Apagar desliga o Canal (T-98); idempotente.
func (uc ConfigurarUseCase) Apagar(ctx context.Context, por string) error {
	var err error
	uc.Atual.Trocar(func() {
		if err = uc.Cofre.ApagarWhatsApp(ctx); err == nil {
			uc.Atual.Aplicar(dominio.ConfigWhatsApp{})
		}
	})
	if err != nil {
		return err
	}
	return uc.registrar(ctx, portas.EventoConfig{Canal: dominio.CanalWhatsApp, Acao: "apagada", Por: por})
}

// registrar grava o evento e o loga; sem o evento, o log de erro é o registro.
func (uc ConfigurarUseCase) registrar(ctx context.Context, e portas.EventoConfig) error {
	attrs := []any{"canal", e.Canal, "acao", e.Acao, "provedor", e.Provedor, "por", e.Por}
	if err := uc.Registro.RegistrarEventoConfig(ctx, e); err != nil {
		uc.Log.ErrorContext(ctx, "configuração trocada sem registro", append(attrs, "erro", err.Error())...)
		return fmt.Errorf("%w: %w", ErrTrocaSemRegistro, err)
	}
	uc.Log.InfoContext(ctx, "configuração trocada", attrs...)
	return nil
}
