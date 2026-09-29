// Package twiliofake simula a API de Messages do Twilio nos testes: guarda
// cada SMS recebido e responde com o status configurado.
package twiliofake

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Credenciais de teste que o fake confere no Basic auth.
const (
	AccountSID          = "ACteste"
	APIKeySID           = "SKteste"
	APIKeySecret        = "segredo-do-teste"
	MessagingServiceSID = "MGteste"
)

// SMS é o que o fake recebeu num POST .../Messages.json.
type SMS struct {
	Para, Corpo, MessagingService, StatusCallback string
	AuthOK                                        bool
}

// Fake guarda os SMS recebidos.
type Fake struct {
	URL    string
	mu     sync.Mutex
	status int
	code   int
	sms    []SMS
}

// Novo sobe o fake respondendo 201.
func Novo(t *testing.T) *Fake {
	f := &Fake{status: http.StatusCreated}
	srv := httptest.NewServer(http.HandlerFunc(f.receber))
	t.Cleanup(srv.Close)
	f.URL = srv.URL
	return f
}

// Responder muda o status (e o código de erro do Twilio) dos próximos envios.
func (f *Fake) Responder(status, code int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.code = status, code
}

func (f *Fake) receber(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/Accounts/"+AccountSID+"/Messages.json") || r.ParseForm() != nil {
		http.Error(w, `{"code":20404,"message":"not found"}`, http.StatusNotFound)
		return
	}
	u, p, _ := r.BasicAuth()
	f.mu.Lock()
	f.sms = append(f.sms, SMS{
		Para: r.PostForm.Get("To"), Corpo: r.PostForm.Get("Body"),
		MessagingService: r.PostForm.Get("MessagingServiceSid"), StatusCallback: r.PostForm.Get("StatusCallback"),
		AuthOK: u == APIKeySID && p == APIKeySecret,
	})
	status, code := f.status, f.code
	f.mu.Unlock()
	w.WriteHeader(status)
	if status >= 300 {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": "The 'To' number is not a valid phone number.", "status": status})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"sid": "SMfake"})
}

// Recebidos devolve uma cópia dos SMS recebidos.
func (f *Fake) Recebidos() []SMS {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]SMS(nil), f.sms...)
}

// Esperar aguarda até n SMS ou o prazo.
func (f *Fake) Esperar(t *testing.T, n int, prazo time.Duration) []SMS {
	t.Helper()
	fim := time.Now().Add(prazo)
	for time.Now().Before(fim) {
		if s := f.Recebidos(); len(s) >= n {
			return s
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("esperava %d SMS em %s, chegaram %d", n, prazo, len(f.Recebidos()))
	return nil
}
