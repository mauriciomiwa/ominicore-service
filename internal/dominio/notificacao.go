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
	Email string
}

// Contato devolve o contato do Destinatário no Canal, ou "" se não houver.
func (d Destinatario) Contato(c Canal) string {
	if c == CanalEmail {
		return d.Email
	}
	return ""
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
	ErrSemContato          = errors.New("destinatário sem contato para o canal")
)
