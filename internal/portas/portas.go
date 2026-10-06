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
	// OptOutsDe diz em quais Canais os contatos do Destinatário fizeram Opt-out.
	OptOutsDe(ctx context.Context, d dominio.Destinatario) (map[dominio.Canal]bool, error)
	// RegistrarOptOut e RemoverOptOut são idempotentes e guardam quem fez (REQ-15).
	RegistrarOptOut(ctx context.Context, c dominio.Canal, contato, por string) error
	RemoverOptOut(ctx context.Context, c dominio.Canal, contato, por string) error
	// AnonimizarLote anonimiza até lote Notificações finalizadas criadas antes
	// de antes (REQ-19) e devolve quantas; PendentesVencidas conta as que
	// passaram do prazo ainda pendentes (só para o alerta).
	AnonimizarLote(ctx context.Context, antes time.Time, lote int) (int, error)
	PendentesVencidas(ctx context.Context, antes time.Time) (int, error)
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
	ID            string
	TemplateID    string
	Status        dominio.StatusNotificacao
	CriadaEm      time.Time
	AtualizadaEm  time.Time
	AnonimizadaEm *time.Time // nil = ainda não anonimizada (REQ-14, REQ-19)
	Tentativas    []TentativaConsultada
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

// EstadoCanais é o estado atual dos Canais no ambiente, que pode mudar sem
// reiniciar (REQ-26): quais estão desligados e o Provedor do WhatsApp.
type EstadoCanais interface {
	Desligados() map[dominio.Canal]bool
	ProvedorWhatsApp() dominio.ProvedorWhatsApp
}

// CofreWhatsAppPort lê e grava a configuração do Canal WhatsApp no cofre
// (REQ-22, REQ-24, REQ-26).
type CofreWhatsAppPort interface {
	LerWhatsApp(ctx context.Context) (dominio.ConfigWhatsApp, error)
	GravarWhatsApp(ctx context.Context, cfg dominio.ConfigWhatsApp) error
	ApagarWhatsApp(ctx context.Context) error
}

// EventoConfig é uma troca de configuração de um Canal pela API, sem segredo
// (REQ-27). Acao é "gravada" ou "apagada"; Provedor fica vazio ao apagar.
type EventoConfig struct {
	Canal    dominio.Canal
	Acao     string
	Provedor dominio.ProvedorWhatsApp
	Por      string
}

// RegistroConfigPort guarda as trocas de configuração (REQ-27).
type RegistroConfigPort interface {
	RegistrarEventoConfig(ctx context.Context, e EventoConfig) error
}

// Identidade é quem chamou a API, lida do token do Keycloak: o client (`azp`),
// a pessoa (`preferred_username`, vazio em token de client) e os papéis do
// realm (REQ-1, REQ-25, REQ-36).
type Identidade struct {
	Client  string
	Usuario string
	Papeis  []string
}
