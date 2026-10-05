package aplicacao_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/memoria"
)

// pedidoSo pede o Template dado com a ordem WhatsApp → SMS.
func pedidoSo(repo *memoria.Repositorio, whatsapp dominio.ConteudoCanal) aplicacao.PedidoNotificacao {
	repo.Templates["t"] = dominio.Template{ID: "t", Conteudos: map[dominio.Canal]dominio.ConteudoCanal{
		dominio.CanalWhatsApp: whatsapp,
		dominio.CanalSMS:      {Corpo: "Olá {{nome}}"},
	}}
	p := pedido()
	p.TemplateID = "t"
	p.Destinatario.Telefone = "+5511999998888"
	p.Canais = []dominio.Canal{dominio.CanalWhatsApp, dominio.CanalSMS}
	return p
}

func recusadoSemRegistro(t *testing.T, repo *memoria.Repositorio, err error, motivo string) {
	t.Helper()
	if !errors.Is(err, dominio.ErrVersaoWhatsAppSemForma) || err.Error() != motivo {
		t.Fatalf("erro = %v, quero %q", err, motivo)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestT78_VersaoSemTextoERecusadaComOUltraMsgAtivo(t *testing.T) {
	repo := memoria.Novo()
	p := pedidoSo(repo, dominio.ConteudoCanal{ModeloAprovado: "boas_vindas", Idioma: "pt_BR", Parametros: []string{"nome"}})
	_, err := aplicacao.SendMessageUseCase{Repo: repo, ProvedorWhatsApp: dominio.WhatsAppUltraMsg}.Pedir(context.Background(), p)
	recusadoSemRegistro(t, repo, err, "versão whatsapp sem texto")
}

func TestT79_VersaoSemModeloAprovadoERecusadaComAMetaAtiva(t *testing.T) {
	repo := memoria.Novo()
	p := pedidoSo(repo, dominio.ConteudoCanal{Corpo: "Olá {{nome}}"})
	_, err := aplicacao.SendMessageUseCase{Repo: repo, ProvedorWhatsApp: dominio.WhatsAppMeta}.Pedir(context.Background(), p)
	recusadoSemRegistro(t, repo, err, "versão whatsapp sem modelo aprovado")
	// com o Canal desligado a forma não importa: ele é pulado (REQ-20)
	desligado := aplicacao.SendMessageUseCase{Repo: repo, Desligados: map[dominio.Canal]bool{dominio.CanalWhatsApp: true}}
	if _, err := desligado.Pedir(context.Background(), p); err != nil {
		t.Fatalf("Canal desligado recusou pela forma: %v", err)
	}
}

func TestT80_VariavelDoTextoDoWhatsAppEExigida(t *testing.T) {
	repo := memoria.Novo()
	repo.Templates["so-texto"] = dominio.Template{
		ID: "so-texto", OrdemCanais: []dominio.Canal{dominio.CanalWhatsApp},
		Conteudos: map[dominio.Canal]dominio.ConteudoCanal{dominio.CanalWhatsApp: {Corpo: "Olá {{nome}}"}},
	}
	p := pedido()
	p.TemplateID, p.Dados = "so-texto", map[string]string{}
	p.Destinatario.Telefone = "+5511999998888"
	_, err := aplicacao.SendMessageUseCase{Repo: repo, ProvedorWhatsApp: dominio.WhatsAppUltraMsg}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrVariavelFaltando) || !strings.Contains(err.Error(), "nome") {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}
