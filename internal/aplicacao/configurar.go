package aplicacao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// ErrTrocaSemRegistro: a configuração foi gravada (ou apagada) no cofre, mas o
// registro da troca falhou; o log de erro fica como registro (REQ-27).
var ErrTrocaSemRegistro = errors.New("configuração trocada sem registro")

// ConfigurarUseCase deixa o Operador consultar, gravar e apagar a configuração
// do Canal WhatsApp (RN-11, REQ-22 a REQ-24, REQ-27).
type ConfigurarUseCase struct {
	Cofre    portas.CofreWhatsAppPort
	Registro portas.RegistroConfigPort
	Atual    *ConfigWhatsAppAtual // a troca vale na hora neste processo (REQ-26)
	Log      *slog.Logger
}

// Consultar lê a configuração no cofre; quem chama expõe só Publicos e
// Segredos (REQ-23).
func (uc ConfigurarUseCase) Consultar(ctx context.Context) (dominio.ConfigWhatsApp, error) {
	return uc.Cofre.LerWhatsApp(ctx)
}

// Gravar valida e substitui a configuração; Provedor desconhecido ou campo
// faltando não grava nada (T-93, T-94).
func (uc ConfigurarUseCase) Gravar(ctx context.Context, provedor string, campos map[string]string, por string) (dominio.ConfigWhatsApp, error) {
	cfg, err := dominio.NovaConfigWhatsApp(provedor, campos)
	if err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	if err := cfg.Validar(); err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	uc.Atual.Trocar(func() {
		if err = uc.Cofre.GravarWhatsApp(ctx, cfg); err == nil {
			uc.Atual.Aplicar(cfg)
		}
	})
	if err != nil {
		return dominio.ConfigWhatsApp{}, err
	}
	return cfg, uc.registrar(ctx, portas.EventoConfig{Canal: dominio.CanalWhatsApp, Acao: "gravada", Provedor: cfg.Provedor, Por: por})
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
