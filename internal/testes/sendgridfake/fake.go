// Package sendgridfake simula a API v3 do SendGrid nos testes: guarda cada
// e-mail recebido e responde com o status configurado.
package sendgridfake

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Email é o que o fake recebeu num POST /v3/mail/send.
type Email struct {
	Para          string
	Assunto       string
	Corpo         string
	Authorization string
	RecebidoEm    time.Time
}

// Fake guarda os e-mails recebidos.
type Fake struct {
	URL    string
	Status int
	mu     sync.Mutex
	emails []Email
}

// Novo sobe o fake respondendo 202.
func Novo(t *testing.T) *Fake {
	f := &Fake{Status: http.StatusAccepted}
	srv := httptest.NewServer(http.HandlerFunc(f.receber))
	t.Cleanup(srv.Close)
	f.URL = srv.URL
	return f
}

func (f *Fake) receber(w http.ResponseWriter, r *http.Request) {
	var corpo struct {
		Personalizations []struct {
			To []struct {
				Email string `json:"email"`
			} `json:"to"`
		} `json:"personalizations"`
		Subject string `json:"subject"`
		Content []struct {
			Value string `json:"value"`
		} `json:"content"`
	}
	if r.URL.Path != "/v3/mail/send" || json.NewDecoder(r.Body).Decode(&corpo) != nil ||
		len(corpo.Personalizations) == 0 || len(corpo.Personalizations[0].To) == 0 || len(corpo.Content) == 0 {
		http.Error(w, `{"errors":[{"message":"corpo inválido"}]}`, http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.emails = append(f.emails, Email{
		Para: corpo.Personalizations[0].To[0].Email, Assunto: corpo.Subject, Corpo: corpo.Content[0].Value,
		Authorization: r.Header.Get("Authorization"), RecebidoEm: time.Now(),
	})
	status := f.Status
	f.mu.Unlock()
	w.Header().Set("X-Message-Id", "msg-fake")
	w.WriteHeader(status)
	if status >= 400 {
		_, _ = w.Write([]byte(`{"errors":[{"message":"The to address is not valid"}]}`))
	}
}

// Responder muda o status devolvido nos próximos envios.
func (f *Fake) Responder(status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Status = status
}

// Emails devolve uma cópia dos e-mails recebidos.
func (f *Fake) Emails() []Email {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Email(nil), f.emails...)
}

// Esperar aguarda até n e-mails ou o prazo.
func (f *Fake) Esperar(t *testing.T, n int, prazo time.Duration) []Email {
	t.Helper()
	fim := time.Now().Add(prazo)
	for time.Now().Before(fim) {
		if e := f.Emails(); len(e) >= n {
			return e
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("esperava %d e-mails em %s, chegaram %d", n, prazo, len(f.Emails()))
	return nil
}
