// Package http é o adaptador de entrada REST: API de envio (/v1/*), contrato
// em api/openapi.yaml.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
)

// Autenticador devolve o Sistema de origem do token (REQ-1).
type Autenticador interface {
	SistemaOrigem(ctx context.Context, token string) (string, error)
}

// Servidor liga as rotas aos casos de uso.
type Servidor struct {
	Auth   Autenticador
	Enviar aplicacao.SendMessageUseCase
	Log    *slog.Logger
}

// Handler devolve o roteador HTTP.
func (s Servidor) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.Handle("POST /v1/notificacoes", s.autenticar(http.HandlerFunc(s.pedirNotificacao)))
	return mux
}

type chaveOrigem struct{}

// autenticar recusa com 401, antes de qualquer efeito, pedido sem token válido.
func (s Servidor) autenticar(prox http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			problema(w, http.StatusUnauthorized, "não autenticado", "token Bearer ausente")
			return
		}
		origem, err := s.Auth.SistemaOrigem(r.Context(), token)
		if err != nil {
			s.Log.InfoContext(r.Context(), "token recusado", "erro", err.Error())
			problema(w, http.StatusUnauthorized, "não autenticado", "token inválido")
			return
		}
		prox.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), chaveOrigem{}, origem)))
	})
}

type pedidoJSON struct {
	TemplateID   string `json:"template_id"`
	Destinatario struct {
		Email    string `json:"email"`
		Telefone string `json:"telefone"`
	} `json:"destinatario"`
	Dados  map[string]string `json:"dados"`
	Canais []string          `json:"canais"`
}

type respostaJSON struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// reE164 é o formato E.164: + código do país + número, até 15 dígitos.
var reE164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

func (s Servidor) pedirNotificacao(w http.ResponseWriter, r *http.Request) {
	chave := r.Header.Get("Idempotency-Key")
	if chave == "" || len(chave) > 255 {
		problema(w, http.StatusBadRequest, "chave de idempotência obrigatória", "envie o header Idempotency-Key (1 a 255 caracteres)")
		return
	}
	var p pedidoJSON
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil || p.TemplateID == "" || p.Dados == nil {
		problema(w, http.StatusBadRequest, "pedido inválido", "JSON com template_id, destinatario e dados")
		return
	}
	if detalhe := validarFormato(p); detalhe != "" {
		problema(w, http.StatusBadRequest, "pedido inválido", detalhe)
		return
	}
	canais := make([]dominio.Canal, len(p.Canais))
	for i, c := range p.Canais {
		canais[i] = dominio.Canal(c)
	}
	reg, err := s.Enviar.Pedir(r.Context(), aplicacao.PedidoNotificacao{
		SistemaOrigem:     r.Context().Value(chaveOrigem{}).(string),
		ChaveIdempotencia: chave,
		TemplateID:        p.TemplateID,
		Destinatario:      dominio.Destinatario{Email: p.Destinatario.Email, Telefone: p.Destinatario.Telefone},
		Dados:             p.Dados,
		Canais:            canais,
	})
	if err != nil {
		s.responderErro(w, r, err)
		return
	}
	if reg.Reutilizada {
		w.Header().Set("Idempotent-Replayed", "true")
	}
	escreverJSON(w, http.StatusAccepted, respostaJSON{ID: reg.ID, Status: string(reg.Status)})
}

// validarFormato confere o formato dos contatos e da ordem; "" = válido. A
// regra de negócio (Canal do Template, contato para algum Canal) é do caso de uso.
func validarFormato(p pedidoJSON) string {
	if p.Destinatario.Email != "" {
		if _, err := mail.ParseAddress(p.Destinatario.Email); err != nil {
			return "destinatario.email inválido"
		}
	}
	if p.Destinatario.Telefone != "" && !reE164.MatchString(p.Destinatario.Telefone) {
		return "destinatario.telefone fora do formato E.164 (ex.: +5511999998888)"
	}
	if p.Canais != nil && len(p.Canais) == 0 {
		return "canais vazio: omita o campo para usar a ordem do Template"
	}
	vistos := map[string]bool{}
	for _, c := range p.Canais {
		if vistos[c] {
			return "canais repetido: " + c
		}
		vistos[c] = true
	}
	return ""
}

func (s Servidor) responderErro(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, dominio.ErrTemplateInexistente):
		problema(w, http.StatusUnprocessableEntity, "template inexistente", err.Error())
	case errors.Is(err, dominio.ErrVariavelFaltando), errors.Is(err, dominio.ErrSemContato), errors.Is(err, dominio.ErrCanalSemVersao):
		problema(w, http.StatusUnprocessableEntity, "pedido incompleto", err.Error())
	default:
		s.Log.ErrorContext(r.Context(), "pedido de notificação falhou", "erro", err.Error())
		problema(w, http.StatusInternalServerError, "erro interno", "tente de novo com a mesma Idempotency-Key")
	}
}

func problema(w http.ResponseWriter, status int, titulo, detalhe string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "title": titulo, "detail": detalhe})
}

func escreverJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
