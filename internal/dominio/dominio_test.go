package dominio

import (
	"errors"
	"testing"
)

func boasVindas() Template {
	return Template{
		ID:          "boas-vindas",
		OrdemCanais: []Canal{CanalEmail},
		Conteudos:   map[Canal]ConteudoCanal{CanalEmail: {Assunto: "Bem-vindo, {{nome}}", Corpo: "Olá {{ nome }}, código {{codigo}}."}},
	}
}

func TestTemplateVariaveisSemRepeticao(t *testing.T) {
	got := boasVindas().Variaveis()
	if len(got) != 2 || got[0] != "codigo" || got[1] != "nome" {
		t.Fatalf("variáveis = %v, quero [codigo nome]", got)
	}
}

func TestTemplateValidarDadosCitaAsFaltantes(t *testing.T) {
	err := boasVindas().ValidarDados(map[string]string{})
	if !errors.Is(err, ErrVariavelFaltando) || err.Error() != "variável faltando: codigo, nome" {
		t.Fatalf("erro = %v", err)
	}
}

func TestTemplatePreencher(t *testing.T) {
	c, err := boasVindas().Preencher(CanalEmail, map[string]string{"nome": "Ana", "codigo": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Assunto != "Bem-vindo, Ana" || c.Corpo != "Olá Ana, código 42." {
		t.Fatalf("conteúdo = %+v", c)
	}
}

func TestT12_RecusaDoProvedorEncerraSemOutroCanal(t *testing.T) {
	d := NotificationRouter{}.Decidir(ResultadoEnvio{Tipo: FalhaDefinitiva, Motivo: "sendgrid 400: endereço inválido"})
	if d.Tentativa != TentativaFalhou || d.Notificacao != NotificacaoFalhou || d.Motivo != "sendgrid 400: endereço inválido" {
		t.Fatalf("decisão = %+v", d)
	}
}

func TestRouterAceitoMarcaEnviada(t *testing.T) {
	d := NotificationRouter{}.Decidir(ResultadoEnvio{Tipo: Aceito})
	if d.Tentativa != TentativaEnviada || d.Notificacao != NotificacaoEnviada {
		t.Fatalf("decisão = %+v", d)
	}
}
