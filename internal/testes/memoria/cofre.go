package memoria

import (
	"context"
	"sync"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// Cofre guarda a configuração do WhatsApp em memória (CofreWhatsAppPort).
type Cofre struct {
	mu  sync.Mutex
	cfg dominio.ConfigWhatsApp
	// Erro, quando dado, faz toda operação falhar (cofre fora do ar).
	Erro error
}

var _ portas.CofreWhatsAppPort = (*Cofre)(nil)

// NovoCofre começa com a configuração dada.
func NovoCofre(cfg dominio.ConfigWhatsApp) *Cofre { return &Cofre{cfg: cfg} }

func (c *Cofre) LerWhatsApp(context.Context) (dominio.ConfigWhatsApp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg, c.Erro
}

func (c *Cofre) GravarWhatsApp(_ context.Context, cfg dominio.ConfigWhatsApp) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Erro != nil {
		return c.Erro
	}
	c.cfg = cfg
	return nil
}

func (c *Cofre) ApagarWhatsApp(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Erro != nil {
		return c.Erro
	}
	c.cfg = dominio.ConfigWhatsApp{}
	return nil
}
