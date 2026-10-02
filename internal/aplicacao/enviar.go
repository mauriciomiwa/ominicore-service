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
	Canais            []dominio.Canal // Ordem de Canais do pedido; vazia = a do Template
}

// SendMessageUseCase aceita pedidos de Notificação (REQ-2, REQ-3).
type SendMessageUseCase struct {
	Repo portas.MessageRepositoryPort
}

// Pedir valida o pedido contra o Template e registra a Notificação com a
// primeira Tentativa no primeiro Canal da Ordem de Canais com contato do
// Destinatário. A ordem é a do pedido ou, sem ela, a do Template.
func (uc SendMessageUseCase) Pedir(ctx context.Context, p PedidoNotificacao) (portas.Registro, error) {
	tpl, err := uc.Repo.BuscarTemplate(ctx, p.TemplateID)
	if err != nil {
		return portas.Registro{}, err
	}
	if err := tpl.ValidarDados(p.Dados); err != nil {
		return portas.Registro{}, err
	}
	ordem := tpl.OrdemCanais
	if len(p.Canais) > 0 {
		ordem = p.Canais
	}
	for _, c := range ordem {
		if _, ok := tpl.Conteudos[c]; !ok {
			return portas.Registro{}, fmt.Errorf("%w: %s", dominio.ErrCanalSemVersao, c)
		}
	}
	optOut, err := uc.Repo.OptOutsDe(ctx, p.Destinatario)
	if err != nil {
		return portas.Registro{}, err
	}
	p.Destinatario.OptOut = optOut
	primeiro, ok := p.Destinatario.PrimeiroComContato(ordem, 0)
	if !ok {
		return portas.Registro{}, dominio.ErrSemContato
	}
	return uc.Repo.RegistrarNotificacao(ctx, portas.NovaNotificacao{
		SistemaOrigem:     p.SistemaOrigem,
		ChaveIdempotencia: p.ChaveIdempotencia,
		TemplateID:        tpl.ID,
		Destinatario:      p.Destinatario,
		Dados:             p.Dados,
		OrdemCanais:       ordem,
		PrimeiroCanal:     primeiro,
	})
}
