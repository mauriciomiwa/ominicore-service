// Package portas declara as portas de saída: o que o núcleo precisa do mundo,
// sem saber quem implementa.
package portas

import (
	"context"

	"ominicore/internal/dominio"
)

// NovaNotificacao é o que o caso de uso de envio pede para registrar.
type NovaNotificacao struct {
	SistemaOrigem     string
	ChaveIdempotencia string
	TemplateID        string
	Destinatario      dominio.Destinatario
	Dados             map[string]string
	OrdemCanais       []dominio.Canal
}

// Registro é o resultado de registrar uma Notificação: Reutilizada indica que a
// Chave de idempotência já existia (REQ-3) e nada novo foi gravado.
type Registro struct {
	ID          string
	Status      dominio.StatusNotificacao
	Reutilizada bool
}

// TentativaReservada é uma Tentativa tirada da fila, com o necessário para o
// envio. Concluir grava a decisão e libera a reserva; se Concluir falhar, a
// Tentativa volta para a fila.
type TentativaReservada struct {
	Tentativa   dominio.Tentativa
	Notificacao dominio.Notificacao
	Concluir    func(context.Context, dominio.Decisao, dominio.ResultadoEnvio) error
}

// MessageRepositoryPort guarda Notificações, Tentativas e Templates.
type MessageRepositoryPort interface {
	BuscarTemplate(ctx context.Context, id string) (dominio.Template, error)
	// RegistrarNotificacao grava a Notificação e a primeira Tentativa numa
	// transação, respeitando a Chave de idempotência (24 h por Sistema de origem).
	RegistrarNotificacao(ctx context.Context, n NovaNotificacao) (Registro, error)
	// ReservarTentativa tira da fila a próxima Tentativa pendente; ok=false
	// quando a fila está vazia.
	ReservarTentativa(ctx context.Context) (t TentativaReservada, ok bool, err error)
}

// MessageSenderPort envia uma Tentativa pelo Provedor de um Canal.
type MessageSenderPort interface {
	Enviar(ctx context.Context, para string, conteudo dominio.ConteudoCanal) dominio.ResultadoEnvio
}
