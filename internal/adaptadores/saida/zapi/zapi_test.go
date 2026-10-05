package zapi_test

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/saida/zapi"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/zapifake"
)

var (
	cred  = zapi.Credenciais{InstanceID: zapifake.InstanceID, Token: zapifake.Token}
	texto = dominio.ConteudoCanal{Corpo: "Olá Ana, sua conta está pronta."}
)

func semSegredo(t *testing.T, motivo string) {
	t.Helper()
	if strings.Contains(motivo, zapifake.Token) || strings.Contains(motivo, zapifake.ClientToken) {
		t.Fatalf("motivo com token: %q", motivo)
	}
}

func TestT85_EnviaOTextoAoTelefoneSemMaisEGuardaOMessageID(t *testing.T) {
	f := zapifake.Novo(t)
	res := zapi.Novo(f.URL, cred).Enviar(context.Background(), "+5511999990000", texto)
	if res.Tipo != dominio.Aceito || res.IDProvedor != zapifake.MessageID {
		t.Fatalf("res = %+v", res)
	}
	m := f.Recebidas()
	if len(m) != 1 || m[0].Telefone != "5511999990000" || m[0].Texto != texto.Corpo || m[0].ClientToken != "" {
		t.Fatalf("recebidas = %+v", m)
	}
}

func TestT86_ClientTokenVaiNoHeaderQuandoConfigurado(t *testing.T) {
	f := zapifake.Novo(t)
	c := cred
	c.ClientToken = zapifake.ClientToken
	if res := zapi.Novo(f.URL, c).Enviar(context.Background(), "+5511999990000", texto); res.Tipo != dominio.Aceito {
		t.Fatalf("res = %+v", res)
	}
	if m := f.Recebidas(); len(m) != 1 || m[0].ClientToken != zapifake.ClientToken {
		t.Fatalf("recebidas = %+v", m)
	}
}

func TestClassificacaoDasRespostasDaZAPI(t *testing.T) {
	casos := []struct {
		nome   string
		status int
		erro   string
		quero  dominio.TipoResultado
	}{
		{"T-87 client-token ausente", http.StatusBadRequest, "null not allowed", dominio.FalhaDefinitiva},
		{"T-87 erro com 200", http.StatusOK, "instance not found", dominio.FalhaDefinitiva},
		{"T-87 token recusado sem error", http.StatusUnauthorized, "", dominio.FalhaDefinitiva},
		{"T-84 recusa que ecoa o token", http.StatusBadRequest, "token " + zapifake.Token + " invalid", dominio.FalhaDefinitiva},
		{"T-88 limite de taxa", http.StatusTooManyRequests, "too many requests", dominio.FalhaTemporaria},
		{"T-88 erro de servidor", http.StatusBadGateway, "", dominio.FalhaTemporaria},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			f := zapifake.Novo(t)
			f.Responder(c.status, c.erro)
			res := zapi.Novo(f.URL, cred).Enviar(context.Background(), "+5511999990000", texto)
			if res.Tipo != c.quero {
				t.Fatalf("tipo = %v, quero %v (%s)", res.Tipo, c.quero, res.Motivo)
			}
			semSegredo(t, res.Motivo)
		})
	}
}

func TestT84_ZAPISemRespostaNaoExpoeOToken(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	c := cred
	c.ClientToken = zapifake.ClientToken
	res := zapi.Novo("http://"+addr, c).Enviar(context.Background(), "+5511999990000", texto)
	if res.Tipo != dominio.FalhaTemporaria {
		t.Fatalf("T-88: res = %+v", res)
	}
	semSegredo(t, res.Motivo)
}

func TestT89_TextoAcimaDe4096NaoChegaAZAPI(t *testing.T) {
	f := zapifake.Novo(t)
	res := zapi.Novo(f.URL, cred).Enviar(context.Background(), "+5511999990000", dominio.ConteudoCanal{Corpo: strings.Repeat("ç", dominio.LimiteTextoWhatsApp+1)})
	if res.Tipo != dominio.FalhaDefinitiva || len(f.Recebidas()) != 0 {
		t.Fatalf("res = %+v, chamadas = %d", res, len(f.Recebidas()))
	}
}

func TestT77_VersaoSemTextoEFalhaDefinitivaSemChamarAZAPI(t *testing.T) {
	f := zapifake.Novo(t)
	res := zapi.Novo(f.URL, cred).Enviar(context.Background(), "+5511999990000", dominio.ConteudoCanal{ModeloAprovado: "boas_vindas"})
	if res.Tipo != dominio.FalhaDefinitiva || res.Motivo != "versão whatsapp sem texto" || len(f.Recebidas()) != 0 {
		t.Fatalf("res = %+v, chamadas = %d", res, len(f.Recebidas()))
	}
}
