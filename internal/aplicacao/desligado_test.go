package aplicacao_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/memoria"
)

// whatsappDesligado é o ambiente sem credenciais da Meta (REQ-20).
var whatsappDesligado = map[dominio.Canal]bool{dominio.CanalWhatsApp: true}

func pedidoComTelefone(canais ...dominio.Canal) aplicacao.PedidoNotificacao {
	p := pedido()
	p.Destinatario.Telefone = "+5511999998888"
	p.Canais = canais
	return p
}

func TestT65_PedidoSoPorWhatsAppComOCanalDesligadoERecusado(t *testing.T) {
	repo := memoria.Novo()
	p := pedidoComTelefone(dominio.CanalWhatsApp)
	p.Destinatario.Email = ""
	_, err := aplicacao.SendMessageUseCase{Repo: repo, Desligados: whatsappDesligado}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrCanalIndisponivel) || err.Error() != "canal whatsapp indisponível" {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestT66_CanalDesligadoNoComecoDaOrdemEPulado(t *testing.T) {
	repo := memoria.Novo()
	p := pedidoComTelefone(dominio.CanalWhatsApp, dominio.CanalSMS)
	reg, err := aplicacao.SendMessageUseCase{Repo: repo, Desligados: whatsappDesligado}.Pedir(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	n := repo.Notificacoes[reg.ID]
	if len(n.OrdemCanais) != 2 || n.OrdemCanais[0] != dominio.CanalWhatsApp || repo.Tentativas[0].Canal != dominio.CanalSMS {
		t.Fatalf("ordem = %v, 1ª Tentativa = %s", n.OrdemCanais, repo.Tentativas[0].Canal)
	}
}

func TestT67_FallbackPulaOCanalDesligado(t *testing.T) {
	repo := memoria.Novo()
	p := pedidoComTelefone(dominio.CanalEmail, dominio.CanalWhatsApp, dominio.CanalSMS)
	if _, err := (aplicacao.SendMessageUseCase{Repo: repo, Desligados: whatsappDesligado}).Pedir(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	uc := aplicacao.DespacharUseCase{
		Repo: repo, Desligados: whatsappDesligado, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{
			dominio.CanalEmail: &remetenteFake{res: dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "recusado"}},
			dominio.CanalSMS:   &contaEnvios{},
		},
	}
	if _, err := uc.ProcessarProxima(context.Background()); err != nil {
		t.Fatal(err)
	}
	h := repo.Historico()
	if len(h) != 2 || h[0].Canal != dominio.CanalEmail || h[1].Canal != dominio.CanalSMS {
		t.Fatalf("histórico = %+v, quero Email → SMS sem WhatsApp", h)
	}
}

func TestT68_TentativaPendenteNoCanalDesligadoSegueAOrdem(t *testing.T) {
	repo := memoria.Novo()
	// aceita com o WhatsApp ainda ligado; o worker sobe já sem as credenciais
	p := pedidoComTelefone(dominio.CanalWhatsApp, dominio.CanalSMS)
	if _, err := (aplicacao.SendMessageUseCase{Repo: repo}).Pedir(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	sms := &contaEnvios{}
	uc := aplicacao.DespacharUseCase{
		Repo: repo, Desligados: whatsappDesligado, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{dominio.CanalSMS: sms},
	}
	for range 2 {
		if _, err := uc.ProcessarProxima(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	h := repo.Historico()
	if len(h) != 2 || h[0].Canal != dominio.CanalWhatsApp || h[0].Status != dominio.TentativaFalhou ||
		!strings.Contains(h[0].Motivo, "sem provedor") || h[1].Canal != dominio.CanalSMS || sms.n != 1 {
		t.Fatalf("envios SMS = %d, histórico = %+v", sms.n, h)
	}
}
