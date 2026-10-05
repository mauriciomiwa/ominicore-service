package ultramsg_test

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/saida/ultramsg"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/ultramsgfake"
)

var (
	cred  = ultramsg.Credenciais{InstanceID: ultramsgfake.InstanceID, Token: ultramsgfake.Token}
	texto = dominio.ConteudoCanal{Corpo: "Olá Ana, sua conta está pronta."}
)

func TestT73_EnviaOTextoAoTelefoneEGuardaOID(t *testing.T) {
	f := ultramsgfake.Novo(t)
	res := ultramsg.Novo(f.URL, cred).Enviar(context.Background(), "+5511999998888", texto)
	if res.Tipo != dominio.Aceito || res.IDProvedor != "42" {
		t.Fatalf("res = %+v", res)
	}
	m := f.Recebidas()
	if len(m) != 1 || m[0].Para != "+5511999998888" || m[0].Texto != texto.Corpo || !m[0].TokenOK {
		t.Fatalf("recebidas = %+v", m)
	}
}

func TestClassificacaoDasRespostasDoUltraMsg(t *testing.T) {
	casos := []struct {
		nome   string
		status int
		erro   string
		quero  dominio.TipoResultado
	}{
		{"T-74 token inválido com 200", http.StatusOK, "Wrong token. Please provide token as a GET parameter.", dominio.FalhaDefinitiva},
		{"T-74 instância recusada", http.StatusUnauthorized, "instance not authorized", dominio.FalhaDefinitiva},
		{"4xx sem error", http.StatusBadRequest, "", dominio.FalhaDefinitiva},
		{"T-74 sent false com 200", http.StatusOK, ultramsgfake.NaoEnviada, dominio.FalhaDefinitiva},
		{"T-75 limite de taxa", http.StatusTooManyRequests, "too many requests", dominio.FalhaTemporaria},
		{"T-75 erro de servidor", http.StatusBadGateway, "", dominio.FalhaTemporaria},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			f := ultramsgfake.Novo(t)
			f.Responder(c.status, c.erro)
			res := ultramsg.Novo(f.URL, cred).Enviar(context.Background(), "+5511999998888", texto)
			if res.Tipo != c.quero {
				t.Fatalf("tipo = %v, quero %v (%s)", res.Tipo, c.quero, res.Motivo)
			}
			if strings.Contains(res.Motivo, ultramsgfake.Token) {
				t.Fatalf("motivo com o token: %q", res.Motivo)
			}
		})
	}
}

func TestT75_UltraMsgSemRespostaEFalhaTemporaria(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	res := ultramsg.Novo("http://"+addr, cred).Enviar(context.Background(), "+5511999998888", texto)
	if res.Tipo != dominio.FalhaTemporaria || strings.Contains(res.Motivo, ultramsgfake.InstanceID) {
		t.Fatalf("res = %+v", res)
	}
}

func TestT76_TextoAcimaDe4096NaoChegaAoUltraMsg(t *testing.T) {
	f := ultramsgfake.Novo(t)
	r := ultramsg.Novo(f.URL, cred)
	if res := r.Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{Corpo: strings.Repeat("ç", ultramsg.LimiteTexto+1)}); res.Tipo != dominio.FalhaDefinitiva {
		t.Fatalf("res = %+v", res)
	}
	if len(f.Recebidas()) != 0 {
		t.Fatal("o Provedor foi chamado")
	}
	// 4096 caracteres multibyte ainda passam: o limite é em caracteres, não em bytes
	if res := r.Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{Corpo: strings.Repeat("ç", ultramsg.LimiteTexto)}); res.Tipo != dominio.Aceito {
		t.Fatalf("res = %+v", res)
	}
}

func TestT77_VersaoSemTextoEFalhaDefinitivaSemChamarOUltraMsg(t *testing.T) {
	f := ultramsgfake.Novo(t)
	res := ultramsg.Novo(f.URL, cred).Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{ModeloAprovado: "boas_vindas"})
	if res.Tipo != dominio.FalhaDefinitiva || res.Motivo != "versão whatsapp sem texto" || len(f.Recebidas()) != 0 {
		t.Fatalf("res = %+v, chamadas = %d", res, len(f.Recebidas()))
	}
}
