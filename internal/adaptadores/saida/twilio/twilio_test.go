package twilio_test

import (
	"context"
	"net/http"
	"testing"

	"ominicore/internal/adaptadores/saida/twilio"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/twiliofake"
)

func cred() twilio.Credenciais {
	return twilio.Credenciais{AccountSID: twiliofake.AccountSID, APIKeySID: twiliofake.APIKeySID, APIKeySecret: twiliofake.APIKeySecret, MessagingServiceSID: twiliofake.MessagingServiceSID}
}

func TestClassificacaoDaRespostaDoTwilio(t *testing.T) {
	casos := []struct {
		status, code int
		quero        dominio.TipoResultado
	}{
		{http.StatusCreated, 0, dominio.Aceito},
		{http.StatusBadRequest, 21211, dominio.FalhaDefinitiva},
		{http.StatusBadRequest, 21610, dominio.FalhaDefinitiva},
		{http.StatusTooManyRequests, 20429, dominio.FalhaTemporaria},
		{http.StatusServiceUnavailable, 0, dominio.FalhaTemporaria},
	}
	for _, c := range casos {
		f := twiliofake.Novo(t)
		f.Responder(c.status, c.code)
		res := twilio.Novo(f.URL, cred()).Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{Corpo: "oi"})
		if res.Tipo != c.quero {
			t.Errorf("%d/%d: tipo = %v, quero %v (%s)", c.status, c.code, res.Tipo, c.quero, res.Motivo)
		}
		if c.quero == dominio.Aceito && res.IDProvedor != "SMfake" {
			t.Errorf("sid = %q", res.IDProvedor)
		}
	}
}

func TestTwilioInacessivelEFalhaTemporaria(t *testing.T) {
	res := twilio.Novo("http://127.0.0.1:1", cred()).Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{Corpo: "oi"})
	if res.Tipo != dominio.FalhaTemporaria {
		t.Fatalf("tipo = %v", res.Tipo)
	}
}
