package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ominicore/internal/adaptadores/saida/openbao"
)

func envCom(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

// enderecoFechado devolve um endereço em que ninguém escuta.
func enderecoFechado(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return "http://" + addr
}

func TestT13_SemOpenBaoOProcessoNaoInicia(t *testing.T) {
	negando := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"errors":["permission denied"]}`, http.StatusForbidden)
	}))
	defer negando.Close()
	casos := map[string]map[string]string{
		"indisponível":   {"OMINICORE_BAO_ADDR": enderecoFechado(t), "OMINICORE_BAO_TOKEN": "dev"},
		"nega a leitura": {"OMINICORE_BAO_ADDR": negando.URL, "OMINICORE_BAO_TOKEN": "dev"},
		"sem endereço":   {"OMINICORE_BAO_TOKEN": "dev"},
	}
	for nome, vars := range casos {
		for _, sub := range []string{"api", "worker"} {
			t.Run(nome+"/"+sub, func(t *testing.T) {
				ctx, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancelar()
				vars["OMINICORE_ADDR"] = "127.0.0.1:0"
				var saida bytes.Buffer
				err := run(ctx, []string{sub}, envCom(vars), &saida)
				if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(strings.ToLower(err.Error()), "openbao") {
					t.Fatalf("erro = %v, quero ErrOpenBao citando o OpenBao", err)
				}
				if strings.Contains(saida.String(), "no ar") {
					t.Fatalf("o processo subiu mesmo sem segredos: %s", saida.String())
				}
			})
		}
	}
}
