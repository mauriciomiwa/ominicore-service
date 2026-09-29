package openbao_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	})
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t", Prefixo: "ominicore-hml"})
	if err != nil || seg.PostgresDSN != "postgres://hml" || seg.SendGridAPIKey != "SG.hml" {
		t.Fatalf("segredos=%+v err=%v", seg, err)
	}
	// sem prefixo o padrão é prod (ominicore/*), que este cofre não tem
	if _, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t"}); err == nil {
		t.Fatal("leu segredos de hml sem pedir o prefixo de hml")
	}
}
