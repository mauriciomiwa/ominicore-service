// Package webhooks é o adaptador de entrada dos Avisos de status dos
// Provedores (REQ-12, REQ-13), a única rota pública do OminiCore (ADR 0003).
// A assinatura é validada sobre o corpo cru antes de qualquer interpretação;
// assinatura inválida = 401 sem efeito; aviso válido = 2xx sempre.
package webhooks

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
)

// Receptor aplica um Aviso de status (ReceiveMessageUseCase).
type Receptor interface {
	Receber(ctx context.Context, a dominio.AvisoStatus) (aplicacao.ResultadoAviso, error)
}

// Credenciais validam as assinaturas; vêm de kv/<prefixo>/webhooks (REQ-6).
type Credenciais struct {
	SendGridChavePublica *ecdsa.PublicKey
	TwilioAuthToken      string
	MetaAppSecret        string
	MetaVerifyToken      string
	// URLPublica é a origem que o Twilio assina (OMINICORE_WEBHOOK_URL): atrás
	// do Traefik o host da requisição não serve para isso.
	URLPublica string
}

// RegistradorOptOut registra o descadastro avisado pelo Provedor (REQ-16).
type RegistradorOptOut interface {
	RegistrarPeloProvedor(ctx context.Context, c dominio.Canal, contato, provedor string) error
}

// Webhooks liga as rotas /webhooks/* ao Receptor.
type Webhooks struct {
	Cred Credenciais
	// WhatsAppLigado registra as rotas da Meta; desligado, elas respondem 404 (REQ-20).
	WhatsAppLigado bool
	Receptor       Receptor
	OptOut         RegistradorOptOut // nil = descadastro do Provedor ignorado
	Log            *slog.Logger
}

// Registrar põe as rotas no mux.
func (w Webhooks) Registrar(mux *http.ServeMux) {
	mux.HandleFunc("POST /webhooks/sendgrid", w.sendgrid)
	mux.HandleFunc("POST /webhooks/twilio", w.twilio)
	if w.WhatsAppLigado {
		mux.HandleFunc("GET /webhooks/whatsapp", w.verificarMeta)
		mux.HandleFunc("POST /webhooks/whatsapp", w.whatsapp)
	}
}

// ChavePublicaSendGrid aceita a chave do painel em PEM ou em base64 (DER).
func ChavePublicaSendGrid(texto string) (*ecdsa.PublicKey, error) {
	der := []byte(texto)
	if b, _ := pem.Decode([]byte(texto)); b != nil {
		der = b.Bytes
	} else if d, err := base64.StdEncoding.DecodeString(strings.TrimSpace(texto)); err == nil {
		der = d
	}
	k, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("chave pública do SendGrid inválida: %w", err)
	}
	ec, ok := k.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("chave pública do SendGrid não é ECDSA")
	}
	return ec, nil
}

const limiteCorpo = 1 << 20

func lerCorpo(r *http.Request) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r.Body, limiteCorpo))
}

// aplicar entrega cada aviso ao Receptor; erro de infraestrutura vira 500
// para o Provedor reenviar (o efeito é idempotente).
func (w Webhooks) aplicar(rw http.ResponseWriter, r *http.Request, provedor string, avisos []dominio.AvisoStatus) {
	for _, a := range avisos {
		if _, err := w.Receptor.Receber(r.Context(), a); err != nil {
			w.Log.ErrorContext(r.Context(), "aviso de status falhou", "provedor", provedor, "erro", err.Error())
			http.Error(rw, "erro interno", http.StatusInternalServerError)
			return
		}
	}
	rw.WriteHeader(http.StatusNoContent)
}

func (w Webhooks) recusar(rw http.ResponseWriter, r *http.Request, provedor string) {
	w.Log.WarnContext(r.Context(), "aviso com assinatura inválida", "provedor", provedor)
	http.Error(rw, "assinatura inválida", http.StatusUnauthorized)
}
