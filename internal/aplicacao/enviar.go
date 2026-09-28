// Package aplicacao guarda os casos de uso (portas de entrada) do OminiCore.
package aplicacao

import (
	"context"
	"fmt"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// PedidoNotificacao é o pedido de um Sistema de origem já autenticado (REQ-1).
type PedidoNotificacao struct {
	SistemaOrigem     string
	ChaveIdempotencia string
	TemplateID        string
	Destinatario      dominio.Destinatario
	Dados             map[string]string
}

// SendMessageUseCase aceita pedidos de Notificação (REQ-2, REQ-3).
type SendMessageUseCase struct {
	Repo portas.MessageRepositoryPort
}

// Pedir valida o pedido contra o Template e registra a Notificação com a
// primeira Tentativa no primeiro Canal da Ordem de Canais do Template.
func (uc SendMessageUseCase) Pedir(ctx context.Context, p PedidoNotificacao) (portas.Registro, error) {
	tpl, err := uc.Repo.BuscarTemplate(ctx, p.TemplateID)
	if err != nil {
		return portas.Registro{}, err
	}
	if err := tpl.ValidarDados(p.Dados); err != nil {
		return portas.Registro{}, err
	}
	primeiro := tpl.OrdemCanais[0]
	if p.Destinatario.Contato(primeiro) == "" {
		return portas.Registro{}, fmt.Errorf("%w %s", dominio.ErrSemContato, primeiro)
	}
	return uc.Repo.RegistrarNotificacao(ctx, portas.NovaNotificacao{
		SistemaOrigem:     p.SistemaOrigem,
		ChaveIdempotencia: p.ChaveIdempotencia,
		TemplateID:        tpl.ID,
		Destinatario:      p.Destinatario,
		Dados:             p.Dados,
		OrdemCanais:       tpl.OrdemCanais,
	})
}
