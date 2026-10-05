// Package ultramsgfake simula o endpoint messages/chat do UltraMsg nos testes:
// guarda cada mensagem recebida e responde com o status e o erro configurados.
// O formato da resposta segue a suposição do design da change 08 (a doc não
// traz exemplo): sucesso {"sent":"true","message":"ok","id":N}; erro
// {"error":"..."}.
package ultramsgfake

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Credenciais de teste que o fake confere.
const (
	InstanceID = "instance123"
	Token      = "token-do-teste"
)

// NaoEnviada, passado a Responder, simula 2xx com sent "false" e sem error.
const NaoEnviada = "\x00sent=false"

// Mensagem é o que o fake recebeu num POST /{instance_id}/messages/chat.
type Mensagem struct {
	Para, Texto string
	TokenOK     bool
}

// Fake guarda as mensagens recebidas.
type Fake struct {
	URL    string
	mu     sync.Mutex
	status int
	erro   string
	msgs   []Mensagem
}

// Novo sobe o fake respondendo 200 com id.
func Novo(t *testing.T) *Fake {
	f := &Fake{status: http.StatusOK}
	srv := httptest.NewServer(http.HandlerFunc(f.receber))
	t.Cleanup(srv.Close)
	f.URL = srv.URL
	return f
}

// Responder muda o status HTTP e o campo error dos próximos envios; erro vazio
// com status 2xx é sucesso; erro NaoEnviada responde {"sent":"false"}.
func (f *Fake) Responder(status int, erro string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.erro = status, erro
}

func (f *Fake) receber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/"+InstanceID+"/messages/chat" || r.ParseForm() != nil || r.URL.Query().Get("token") != "" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"instance not found"}`))
		return
	}
	m := Mensagem{Para: r.PostForm.Get("to"), Texto: r.PostForm.Get("body"), TokenOK: r.PostForm.Get("token") == Token}
	f.mu.Lock()
	f.msgs = append(f.msgs, m)
	status, erro := f.status, f.erro
	f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if erro == NaoEnviada {
		_ = json.NewEncoder(w).Encode(map[string]any{"sent": "false", "message": "simulada"})
		return
	}
	if erro != "" || status >= 300 {
		_ = json.NewEncoder(w).Encode(map[string]any{"error": erro})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"sent": "true", "message": "ok", "id": 42})
}

// Recebidas devolve uma cópia das mensagens recebidas.
func (f *Fake) Recebidas() []Mensagem {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Mensagem(nil), f.msgs...)
}

// Esperar aguarda até n mensagens ou o prazo.
func (f *Fake) Esperar(t *testing.T, n int, prazo time.Duration) []Mensagem {
	t.Helper()
	fim := time.Now().Add(prazo)
	for time.Now().Before(fim) {
		if m := f.Recebidas(); len(m) >= n {
			return m
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("esperava %d mensagens em %s, chegaram %d", n, prazo, len(f.Recebidas()))
	return nil
}
