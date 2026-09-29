// Package portas declara as portas de saída: o que o núcleo precisa do mundo,
// sem saber quem implementa.
package portas

import (
	"context"
	"time"

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
	PrimeiroCanal     dominio.Canal // 1º Canal da ordem com contato (REQ-2)
}

// Registro é o resultado de registrar uma Notificação: Reutilizada indica que a
// Chave de idempotência já existia (REQ-3) e nada novo foi gravado.
type Registro struct {
	ID          string
	Status      dominio.StatusNotificacao
	Reutilizada bool
}

// TentativaReservada é uma Tentativa tirada da fila, com o necessário para o
// envio. Concluir grava a decisão — inclusive a próxima Tentativa, na mesma
// transação — e libera a reserva; se Concluir falhar, a Tentativa volta para a
// fila.
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
	// ReservarPorIDProvedor trava a Tentativa (e a Notificação) do aviso;
	// ok=false quando o id do Provedor não existe (REQ-12).
	ReservarPorIDProvedor(ctx context.Context, idProvedor string) (a AvisoReservado, ok bool, err error)
	// ConsultarNotificacao devolve a Notificação só se for do Sistema de origem.
	ConsultarNotificacao(ctx context.Context, sistemaOrigem, id string) (c Consulta, ok bool, err error)
}

// AvisoReservado é a Tentativa travada para aplicar um Aviso de status.
// Aplicar grava e libera; Liberar desiste sem gravar (aviso que não avança).
type AvisoReservado struct {
	Tentativa   dominio.Tentativa
	Notificacao dominio.Notificacao
	Aplicar     func(context.Context, dominio.AtualizacaoAviso) error
	Liberar     func(context.Context)
}

// Consulta é a visão da Notificação para o Sistema de origem (REQ-14): sem
// contato do Destinatário nem texto do motivo.
type Consulta struct {
	ID           string
	TemplateID   string
	Status       dominio.StatusNotificacao
	CriadaEm     time.Time
	AtualizadaEm time.Time
	Tentativas   []TentativaConsultada
}

// TentativaConsultada é uma Tentativa na Consulta.
type TentativaConsultada struct {
	Canal        dominio.Canal
	Numero       int
	Status       dominio.StatusTentativa
	Categoria    dominio.CategoriaFalha
	CriadaEm     time.Time
	AtualizadaEm time.Time
}

// MessageSenderPort envia uma Tentativa pelo Provedor de um Canal.
type MessageSenderPort interface {
	Enviar(ctx context.Context, para string, conteudo dominio.ConteudoCanal) dominio.ResultadoEnvio
}
