package openbao_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/saida/openbao"
)

// kvFalso responde o KV v2 só nos caminhos dados; o resto é 404.
func kvFalso(t *testing.T, dados map[string]map[string]string) string {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, ok := dados[r.URL.Path]
		if !ok {
			http.Error(w, `{"errors":[]}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": d, "metadata": map[string]any{}}})
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestPrefixoSeparaHmlDeProd(t *testing.T) {
	addr := kvFalso(t, map[string]map[string]string{
		"/v1/kv/data/ominicore-hml/postgres": {"dsn": "postgres://hml"},
		"/v1/kv/data/ominicore-hml/sendgrid": {"api_key": "SG.hml", "remetente": "hml@exemplo.com"},
		"/v1/kv/data/ominicore-hml/twilio": {
			"account_sid": "AChml", "api_key_sid": "SKhml", "api_key_secret": "x", "messaging_service_sid": "MGhml",
		},
		"/v1/kv/data/ominicore-hml/whatsapp": {"token": "t", "phone_number_id": "1"},
	})
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t", Prefixo: "ominicore-hml"})
	if err != nil || seg.PostgresDSN != "postgres://hml" || seg.SendGridAPIKey != "SG.hml" || seg.Twilio["account_sid"] != "AChml" {
		t.Fatalf("segredos=%+v err=%v", seg, err)
	}
	// sem prefixo o padrão é prod (ominicore/*), que este cofre não tem
	if _, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t"}); err == nil {
		t.Fatal("leu segredos de hml sem pedir o prefixo de hml")
	}
}

func TestSemSegredoDoTwilioOProcessoNaoTemCredenciais(t *testing.T) {
	addr := kvFalso(t, map[string]map[string]string{
		"/v1/kv/data/ominicore/postgres": {"dsn": "postgres://x"},
		"/v1/kv/data/ominicore/sendgrid": {"api_key": "SG.x", "remetente": "x@exemplo.com"},
		"/v1/kv/data/ominicore/twilio":   {"account_sid": "AC", "api_key_sid": "SK", "api_key_secret": "x"},
	})
	_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t"})
	if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), "messaging_service_sid") {
		t.Fatalf("erro = %v, quero ErrOpenBao citando o campo faltante", err)
	}
}

func TestLoginJWTUsaOMountDoAmbiente(t *testing.T) {
	var loginEm string
	kv := map[string]map[string]string{
		"/v1/kv/data/ominicore-hml/postgres": {"dsn": "postgres://hml"},
		"/v1/kv/data/ominicore-hml/sendgrid": {"api_key": "SG.hml", "remetente": "hml@exemplo.com"},
		"/v1/kv/data/ominicore-hml/twilio":   {"account_sid": "AC", "api_key_sid": "SK", "api_key_secret": "x", "messaging_service_sid": "MG"},
		"/v1/kv/data/ominicore-hml/whatsapp": {"token": "t", "phone_number_id": "1"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			loginEm = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]any{"client_token": "tok-hml"}})
			return
		}
		d, ok := kv[r.URL.Path]
		if !ok || r.Header.Get("X-Vault-Token") != "tok-hml" {
			http.Error(w, `{"errors":["permission denied"]}`, http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": d, "metadata": map[string]any{}}})
	}))
	defer srv.Close()
	seg, err := openbao.Carregar(context.Background(), openbao.Config{
		Endereco: srv.URL, JWT: "jwt-do-keycloak", Role: "ominicore-hml", AuthMount: "jwt-ominicore-hml", Prefixo: "ominicore-hml",
	})
	if err != nil || seg.PostgresDSN != "postgres://hml" {
		t.Fatalf("segredos=%+v err=%v", seg, err)
	}
	if loginEm != "/v1/auth/jwt-ominicore-hml/login" {
		t.Fatalf("login em %q", loginEm)
	}
}
