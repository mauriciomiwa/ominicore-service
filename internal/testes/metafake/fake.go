// Package metafake simula a Cloud API do WhatsApp nos testes: guarda cada
// mensagem recebida e responde com o status e o código de erro configurados.
package metafake

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Credenciais de teste que o fake confere.
const (
	Token         = "token-do-teste"
	PhoneNumberID = "123456789"
)

// Mensagem é o que o fake recebeu num POST /vNN.0/{phone_number_id}/messages.
type Mensagem struct {
	Para, Modelo, Idioma string
	Parametros           []string
	AuthOK               bool
}

// Fake guarda as mensagens recebidas.
type Fake struct {
	URL    string
	mu     sync.Mutex
	status int
	code   int
	msgs   []Mensagem
}

// Novo sobe o fake respondendo 200.
func Novo(t *testing.T) *Fake {
	f := &Fake{status: http.StatusOK}
	srv := httptest.NewServer(http.HandlerFunc(f.receber))
	t.Cleanup(srv.Close)
	f.URL = srv.URL
	return f
}

// Responder muda o status HTTP e o error.code dos próximos envios.
func (f *Fake) Responder(status, code int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.code = status, code
}

func (f *Fake) receber(w http.ResponseWriter, r *http.Request) {
	var corpo struct {
		MessagingProduct string `json:"messaging_product"`
		To               string `json:"to"`
		Type             string `json:"type"`
		Template         struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
			Components []struct {
				Type       string `json:"type"`
				Parameters []struct {
					Text string `json:"text"`
				} `json:"parameters"`
			} `json:"components"`
		} `json:"template"`
	}
	if !strings.HasSuffix(r.URL.Path, "/"+PhoneNumberID+"/messages") || json.NewDecoder(r.Body).Decode(&corpo) != nil ||
		corpo.MessagingProduct != "whatsapp" || corpo.Type != "template" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":100,"message":"invalid parameter"}}`))
		return
	}
	m := Mensagem{
		Para: corpo.To, Modelo: corpo.Template.Name, Idioma: corpo.Template.Language.Code,
		AuthOK: r.Header.Get("Authorization") == "Bearer "+Token,
	}
	for _, c := range corpo.Template.Components {
		for _, p := range c.Parameters {
			m.Parametros = append(m.Parametros, p.Text)
		}
	}
	f.mu.Lock()
	f.msgs = append(f.msgs, m)
	status, code := f.status, f.code
	f.mu.Unlock()
	w.WriteHeader(status)
	if status >= 300 {
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": code, "message": "erro simulado"}})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"messaging_product": "whatsapp", "messages": []map[string]string{{"id": "wamid.fake"}}})
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
