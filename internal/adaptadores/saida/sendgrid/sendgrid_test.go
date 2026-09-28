package sendgrid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ominicore/internal/adaptadores/saida/sendgrid"
	"ominicore/internal/dominio"
)

func TestClassificacaoDaRespostaDoSendGrid(t *testing.T) {
	casos := []struct {
		status int
		quero  dominio.TipoResultado
	}{
		{http.StatusAccepted, dominio.Aceito},
		{http.StatusBadRequest, dominio.FalhaDefinitiva},
		{http.StatusUnauthorized, dominio.FalhaDefinitiva},
		{http.StatusTooManyRequests, dominio.FalhaTemporaria},
		{http.StatusBadGateway, dominio.FalhaTemporaria},
	}
	for _, c := range casos {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("X-Message-Id", "msg-1")
			w.WriteHeader(c.status)
			_, _ = w.Write([]byte(`{"errors":[{"message":"detalhe do provedor"}]}`))
		}))
		res := sendgrid.Novo(srv.URL, "SG.chave", "no-reply@exemplo.com").Enviar(context.Background(), "ana@exemplo.com", dominio.ConteudoCanal{Assunto: "a", Corpo: "b"})
		srv.Close()
		if res.Tipo != c.quero {
			t.Errorf("status %d: tipo = %v, quero %v", c.status, res.Tipo, c.quero)
		}
		if c.quero == dominio.FalhaDefinitiva && res.Motivo == "" {
			t.Errorf("status %d: falha sem motivo", c.status)
		}
	}
}

func TestSendGridInacessivelEFalhaTemporaria(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	res := sendgrid.Novo(srv.URL, "SG.chave", "no-reply@exemplo.com").Enviar(context.Background(), "ana@exemplo.com", dominio.ConteudoCanal{})
	if res.Tipo != dominio.FalhaTemporaria {
		t.Fatalf("tipo = %v", res.Tipo)
	}
}
