// Package dominio guarda as regras do OminiCore sem conhecer HTTP, banco nem
// Provedor. Termos: DOC/Negocio/glossario.md.
package dominio

import (
	"errors"
	"time"
)

// Canal é o meio pelo qual a Notificação chega ao Destinatário.
type Canal string

const (
	CanalEmail Canal = "email"
	CanalSMS   Canal = "sms"
)

// StatusNotificacao é o estado da Notificação como um todo.
type StatusNotificacao string

const (
	NotificacaoPendente StatusNotificacao = "pendente"
	NotificacaoEnviada  StatusNotificacao = "enviada"
	NotificacaoFalhou   StatusNotificacao = "falhou"
)

// StatusTentativa é o estado de um envio por um Canal.
type StatusTentativa string

const (
	TentativaPendente StatusTentativa = "pendente"
	TentativaEnviada  StatusTentativa = "enviada"
	TentativaFalhou   StatusTentativa = "falhou"
)

// Destinatario é a pessoa que recebe a Notificação, pelo contato de cada Canal.
type Destinatario struct {
	Email    string
	Telefone string // E.164
}

// Contato devolve o contato do Destinatário no Canal, ou "" se não houver.
func (d Destinatario) Contato(c Canal) string {
	switch c {
	case CanalEmail:
		return d.Email
	case CanalSMS:
		return d.Telefone
	}
	return ""
}

// PrimeiroComContato devolve o primeiro Canal da ordem, a partir de inicio,
// para o qual o Destinatário tem contato (RN-2: Canal sem contato é pulado).
func (d Destinatario) PrimeiroComContato(ordem []Canal, inicio int) (Canal, bool) {
	for _, c := range ordem[min(inicio, len(ordem)):] {
		if d.Contato(c) != "" {
			return c, true
		}
	}
	return "", false
}

// Notificacao é o pedido de um Sistema de origem para avisar um Destinatário.
type Notificacao struct {
	ID            string
	SistemaOrigem string
	TemplateID    string
	Destinatario  Destinatario
	Dados         map[string]string
	OrdemCanais   []Canal
	Status        StatusNotificacao
	CriadaEm      time.Time
}

// Tentativa é um envio da Notificação por um Canal.
type Tentativa struct {
	ID            int64
	NotificacaoID string
	Canal         Canal
	Numero        int
	Status        StatusTentativa
	Motivo        string
}

// Erros de validação do pedido (RN-1). O detalhe vai na mensagem.
var (
	ErrTemplateInexistente = errors.New("template inexistente")
	ErrVariavelFaltando    = errors.New("variável faltando")
	ErrSemContato          = errors.New("destinatário sem contato para os canais")
	ErrCanalSemVersao      = errors.New("canal sem versão no template")
)
