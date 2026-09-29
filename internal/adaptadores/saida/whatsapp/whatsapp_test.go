package whatsapp_test

import (
	"context"
	"net/http"
	"testing"

	"ominicore/internal/adaptadores/saida/whatsapp"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/metafake"
)

var (
	cred     = whatsapp.Credenciais{Token: metafake.Token, PhoneNumberID: metafake.PhoneNumberID}
	conteudo = dominio.ConteudoCanal{ModeloAprovado: "boas_vindas", Idioma: "pt_BR", Parametros: []string{"Ana"}}
)

func TestClassificacaoDosCodigosDaMeta(t *testing.T) {
	casos := []struct {
		nome         string
		status, code int
		quero        dominio.TipoResultado
	}{
		{"aceito", http.StatusOK, 0, dominio.Aceito},
		{"T-29 número sem WhatsApp", http.StatusBadRequest, 131026, dominio.FalhaDefinitiva},
		{"T-31 Modelo não aprovado", http.StatusNotFound, 132001, dominio.FalhaDefinitiva},
		{"T-30 limite de taxa", http.StatusBadRequest, 130429, dominio.FalhaTemporaria},
		{"erro desconhecido", http.StatusInternalServerError, 131000, dominio.FalhaTemporaria},
		{"marketing por usuário (24 h)", http.StatusBadRequest, 131049, dominio.FalhaDefinitiva},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			f := metafake.Novo(t)
			f.Responder(c.status, c.code)
			res := whatsapp.Novo(f.URL, cred).Enviar(context.Background(), "+5511999998888", conteudo)
			if res.Tipo != c.quero {
				t.Fatalf("tipo = %v, quero %v (%s)", res.Tipo, c.quero, res.Motivo)
			}
			if c.quero == dominio.Aceito && res.IDProvedor != "wamid.fake" {
				t.Fatalf("id = %q", res.IDProvedor)
			}
		})
	}
}

func TestVersaoSemModeloAprovadoEFalhaDefinitivaSemChamarAMeta(t *testing.T) {
	f := metafake.Novo(t)
	res := whatsapp.Novo(f.URL, cred).Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{})
	if res.Tipo != dominio.FalhaDefinitiva || len(f.Recebidas()) != 0 {
		t.Fatalf("res = %+v, chamadas = %d", res, len(f.Recebidas()))
	}
}
