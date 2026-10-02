// Package dominio guarda as regras do OminiCore sem conhecer HTTP, banco nem
// Provedor. Termos: DOC/Negocio/glossario.md.
package dominio

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Canal é o meio pelo qual a Notificação chega ao Destinatário.
type Canal string

const (
	CanalEmail    Canal = "email"
	CanalSMS      Canal = "sms"
	CanalWhatsApp Canal = "whatsapp"
)

// StatusNotificacao é o estado da Notificação como um todo.
type StatusNotificacao string

const (
	NotificacaoPendente StatusNotificacao = "pendente"
	NotificacaoEnviada  StatusNotificacao = "enviada"
	NotificacaoEntregue StatusNotificacao = "entregue"
	NotificacaoFalhou   StatusNotificacao = "falhou"
)

// StatusTentativa é o estado de um envio por um Canal.
type StatusTentativa string

const (
	TentativaPendente StatusTentativa = "pendente"
	TentativaEnviada  StatusTentativa = "enviada"
	TentativaEntregue StatusTentativa = "entregue"
	TentativaLida     StatusTentativa = "lida"
	TentativaFalhou   StatusTentativa = "falhou"
)

// Destinatario é a pessoa que recebe a Notificação, pelo contato de cada Canal.
type Destinatario struct {
	Email    string
	Telefone string // E.164
	// OptOut marca os Canais em que o contato fez Opt-out (RN-8); carregado
	// pelo repositório, nunca vindo do pedido.
	OptOut map[Canal]bool
}

// Disponivel diz se dá para enviar por c: tem contato e não está em Opt-out.
func (d Destinatario) Disponivel(c Canal) bool {
	return d.Contato(c) != "" && !d.OptOut[c]
}

// Contato devolve o contato do Destinatário no Canal, ou "" se não houver.
func (d Destinatario) Contato(c Canal) string {
	switch c {
	case CanalEmail:
		return d.Email
	case CanalSMS, CanalWhatsApp:
		return d.Telefone
	}
	return ""
}

// PrimeiroComContato devolve o primeiro Canal da ordem, a partir de inicio,
// disponível para o Destinatário (RN-2: Canal sem contato é pulado; RN-8:
// Canal em Opt-out também).
func (d Destinatario) PrimeiroComContato(ordem []Canal, inicio int) (Canal, bool) {
	for _, c := range ordem[min(inicio, len(ordem)):] {
		if d.Disponivel(c) {
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
	Categoria     CategoriaFalha
}

// Erros de validação do pedido (RN-1). O detalhe vai na mensagem.
var (
	ErrTemplateInexistente = errors.New("template inexistente")
	ErrVariavelFaltando    = errors.New("variável faltando")
	ErrSemContato          = errors.New("destinatário sem contato para os canais")
	ErrCanalSemVersao      = errors.New("canal sem versão no template")
	ErrContatoInvalido     = errors.New("contato inválido para o canal")
)

var reE164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

// NormalizarContato põe o contato na forma canônica do Opt-out (REQ-18):
// e-mail sem espaços e em minúsculas; telefone em E.164 (já validado).
// Canal desconhecido, e-mail sem @ ou telefone fora do E.164 = erro.
func NormalizarContato(c Canal, contato string) (string, error) {
	contato = strings.TrimSpace(contato)
	switch c {
	case CanalEmail:
		e := strings.ToLower(contato)
		if at := strings.IndexByte(e, '@'); at <= 0 || at == len(e)-1 || strings.ContainsAny(e, " <>") {
			return "", fmt.Errorf("%w: e-mail", ErrContatoInvalido)
		}
		return e, nil
	case CanalSMS, CanalWhatsApp:
		if !reE164.MatchString(contato) {
			return "", fmt.Errorf("%w: telefone fora do E.164", ErrContatoInvalido)
		}
		return contato, nil
	}
	return "", fmt.Errorf("%w: canal %q desconhecido", ErrContatoInvalido, c)
}
