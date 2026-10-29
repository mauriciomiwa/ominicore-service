// Package zapifake simula o endpoint send-text da Z-API nos testes: guarda cada
// mensagem recebida (com o header Client-Token) e responde com o status e o
// erro configurados. Também simula o cadastro das URLs dos webhooks
// (PUT .../update-webhook-delivery e .../update-webhook-message-status). Sucesso segue a doc
// (https://developer.z-api.io/message/send-text): {"zaapId","messageId","id"};
// erro: {"error":"..."}.
package zapifake

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
	InstanceID  = "inst-teste"
	Token       = "token-do-teste"
	ClientToken = "client-token-do-teste"
	MessageID   = "D241XXXX732339502B68"
)

// Mensagem é o que o fake recebeu num POST /instances/{id}/token/{token}/send-text.
type Mensagem struct {
	Telefone, Texto string
	ClientToken     string
}

// Fake guarda as mensagens recebidas.
type Fake struct {
	URL       string
	mu        sync.Mutex
	status    int
	erro      string
	msgs      []Mensagem
	webhooks  map[string]string // caminho (update-webhook-...) → URL cadastrada
	falharCad bool
}

// Novo sobe o fake respondendo 200 com messageId.
func Novo(t *testing.T) *Fake {
	f := &Fake{status: http.StatusOK, webhooks: map[string]string{}}
	srv := httptest.NewServer(http.HandlerFunc(f.receber))
	t.Cleanup(srv.Close)
	f.URL = srv.URL
	return f
}

// Responder muda o status HTTP e o campo error dos próximos envios; erro vazio
// com status 2xx é sucesso.
func (f *Fake) Responder(status int, erro string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status, f.erro = status, erro
}

// FalharCadastro faz o cadastro de webhooks responder erro (T-132).
func (f *Fake) FalharCadastro(falhar bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.falharCad = falhar
}

// Webhooks devolve as URLs cadastradas por caminho.
func (f *Fake) Webhooks() map[string]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]string{}
	for k, v := range f.webhooks {
		out[k] = v
	}
	return out
}

func (f *Fake) cadastrar(w http.ResponseWriter, r *http.Request, caminho string) {
	var corpo struct {
		Value string `json:"value"`
	}
	w.Header().Set("Content-Type", "application/json")
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.falharCad || json.NewDecoder(r.Body).Decode(&corpo) != nil || corpo.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"webhook não aceito"}`))
		return
	}
	f.webhooks[caminho] = corpo.Value
	_, _ = w.Write([]byte(`{"value":true}`))
}

func (f *Fake) receber(w http.ResponseWriter, r *http.Request) {
	base := "/instances/" + InstanceID + "/token/" + Token + "/"
	for _, caminho := range []string{"update-webhook-delivery", "update-webhook-message-status"} {
		if r.Method == http.MethodPut && r.URL.Path == base+caminho {
			f.cadastrar(w, r, caminho)
			return
		}
	}
	var corpo struct {
		Phone   string `json:"phone"`
		Message string `json:"message"`
	}
	if r.Method != http.MethodPost || r.URL.Path != "/instances/"+InstanceID+"/token/"+Token+"/send-text" ||
		r.Header.Get("Content-Type") != "application/json" || json.NewDecoder(r.Body).Decode(&corpo) != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"instance not found"}`))
		return
	}
	f.mu.Lock()
	f.msgs = append(f.msgs, Mensagem{Telefone: corpo.Phone, Texto: corpo.Message, ClientToken: r.Header.Get("Client-Token")})
	status, erro := f.status, f.erro
	f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if erro != "" || status >= 300 {
		_ = json.NewEncoder(w).Encode(map[string]any{"error": erro})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"zaapId": "3999984263738042930CD6ECDE9VDWSA", "messageId": MessageID, "id": MessageID})
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
