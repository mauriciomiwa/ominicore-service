package aplicacao

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// MontarRemetenteWhatsApp liga a configuração ao Remetente do Provedor dela.
type MontarRemetenteWhatsApp func(dominio.ConfigWhatsApp) portas.MessageSenderPort

// ConfigWhatsAppAtual guarda a configuração do Canal WhatsApp em uso e o
// Remetente dela, trocáveis sem reiniciar (REQ-26). É também o Remetente do
// WhatsApp: delega ao do retrato atual.
type ConfigWhatsAppAtual struct {
	montar  MontarRemetenteWhatsApp
	retrato atomic.Pointer[retratoWhatsApp]
	// troca serializa "ler ou gravar no cofre + Aplicar": sem ela, uma
	// releitura que leu antes de um PUT aplicaria a configuração velha depois.
	troca sync.Mutex
}

// Trocar roda f com as outras trocas em espera.
func (a *ConfigWhatsAppAtual) Trocar(f func()) {
	a.troca.Lock()
	defer a.troca.Unlock()
	f()
}

type retratoWhatsApp struct {
	cfg       dominio.ConfigWhatsApp
	remetente portas.MessageSenderPort
}

var (
	_ portas.EstadoCanais      = (*ConfigWhatsAppAtual)(nil)
	_ portas.MessageSenderPort = (*ConfigWhatsAppAtual)(nil)
)

// NovaConfigWhatsAppAtual começa com a configuração lida no boot.
func NovaConfigWhatsAppAtual(cfg dominio.ConfigWhatsApp, montar MontarRemetenteWhatsApp) *ConfigWhatsAppAtual {
	a := &ConfigWhatsAppAtual{montar: montar}
	a.Aplicar(cfg)
	return a
}

// Aplicar troca a configuração em uso; devolve false quando ela não mudou.
func (a *ConfigWhatsAppAtual) Aplicar(cfg dominio.ConfigWhatsApp) bool {
	if atual := a.retrato.Load(); atual != nil && atual.cfg.Igual(cfg) {
		return false
	}
	r := &retratoWhatsApp{cfg: cfg}
	if cfg.Ligado {
		r.remetente = a.montar(cfg)
	}
	a.retrato.Store(r)
	return true
}

// Config devolve a configuração em uso.
func (a *ConfigWhatsAppAtual) Config() dominio.ConfigWhatsApp { return a.retrato.Load().cfg }

// Desligados: o WhatsApp sem configuração (REQ-20).
func (a *ConfigWhatsAppAtual) Desligados() map[dominio.Canal]bool {
	if a.Config().Ligado {
		return nil
	}
	return map[dominio.Canal]bool{dominio.CanalWhatsApp: true}
}

// ProvedorWhatsApp em uso; vazio com o Canal desligado.
func (a *ConfigWhatsAppAtual) ProvedorWhatsApp() dominio.ProvedorWhatsApp { return a.Config().Provedor }

// Enviar usa o Remetente do Provedor em uso; Canal desligado é Falha
// definitiva, como Canal sem Remetente (T-68).
func (a *ConfigWhatsAppAtual) Enviar(ctx context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	r := a.retrato.Load()
	if r.remetente == nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "canal sem provedor configurado: " + string(dominio.CanalWhatsApp)}
	}
	return r.remetente.Enviar(ctx, para, c)
}

// desligadosDe lê o estado atual quando há um; senão, o mapa fixo.
func desligadosDe(e portas.EstadoCanais, fixo map[dominio.Canal]bool) map[dominio.Canal]bool {
	if e != nil {
		return e.Desligados()
	}
	return fixo
}

// ReleituraConfig relê a configuração do WhatsApp no cofre e a aplica quando
// mudou (REQ-26); inválida ou ilegível, mantém a atual e loga o erro (T-102).
type ReleituraConfig struct {
	Cofre portas.CofreWhatsAppPort
	Atual *ConfigWhatsAppAtual
	Log   *slog.Logger
}

// Reler faz uma releitura.
func (uc ReleituraConfig) Reler(ctx context.Context) {
	var (
		cfg   dominio.ConfigWhatsApp
		err   error
		mudou bool
	)
	uc.Atual.Trocar(func() {
		if cfg, err = uc.Cofre.LerWhatsApp(ctx); err == nil {
			mudou = uc.Atual.Aplicar(cfg)
		}
	})
	if err != nil {
		uc.Log.ErrorContext(ctx, "releitura da configuração whatsapp: mantida a anterior", "erro", err.Error())
		return
	}
	if !mudou {
		return
	}
	if cfg.Ligado {
		uc.Log.InfoContext(ctx, "whatsapp reconfigurado: "+string(cfg.Provedor), "canal", dominio.CanalWhatsApp, "provedor", cfg.Provedor)
		return
	}
	uc.Log.InfoContext(ctx, "whatsapp desligado", "canal", dominio.CanalWhatsApp, "motivo", "sem configuração no openbao")
}

// Rodar relê a cada intervalo até o contexto acabar.
func (uc ReleituraConfig) Rodar(ctx context.Context, intervalo time.Duration) {
	t := time.NewTicker(intervalo)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			uc.Reler(ctx)
		}
	}
}
