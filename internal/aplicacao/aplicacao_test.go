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

func pedido() aplicacao.PedidoNotificacao {
	return aplicacao.PedidoNotificacao{
		SistemaOrigem: "sistema-a", ChaveIdempotencia: "abc-1", TemplateID: "boas-vindas",
		Destinatario: dominio.Destinatario{Email: "ana@exemplo.com"}, Dados: map[string]string{"nome": "Ana"},
	}
}

func TestT4_PedidoValidoEAceito(t *testing.T) {
	repo := memoria.Novo()
	reg, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), pedido())
	if err != nil {
		t.Fatal(err)
	}
	if reg.ID == "" || reg.Status != dominio.NotificacaoPendente {
		t.Fatalf("registro = %+v", reg)
	}
	if len(repo.Tentativas) != 1 || repo.Tentativas[0].Canal != dominio.CanalEmail || repo.Tentativas[0].Status != dominio.TentativaPendente {
		t.Fatalf("tentativas = %+v", repo.Tentativas)
	}
}

func TestT5_TemplateInexistenteERecusado(t *testing.T) {
	repo := memoria.Novo()
	p := pedido()
	p.TemplateID = "nao-existe"
	_, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrTemplateInexistente) {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestT6_VariavelFaltandoERecusada(t *testing.T) {
	repo := memoria.Novo()
	p := pedido()
	p.Dados = map[string]string{}
	_, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrVariavelFaltando) || !strings.Contains(err.Error(), "nome") {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestPedidoSemContatoDoCanalERecusado(t *testing.T) {
	p := pedido()
	p.Destinatario.Email = ""
	_, err := aplicacao.SendMessageUseCase{Repo: memoria.Novo()}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrSemContato) {
		t.Fatalf("erro = %v", err)
	}
}

type remetenteFake struct {
	res  dominio.ResultadoEnvio
	para string
	c    dominio.ConteudoCanal
}

func (f *remetenteFake) Enviar(_ context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	f.para, f.c = para, c
	return f.res
}

func TestDespacharFilaVaziaDevolveFalse(t *testing.T) {
	uc := aplicacao.DespacharUseCase{Repo: memoria.Novo(), Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	processou, err := uc.ProcessarProxima(context.Background())
	if processou || err != nil {
		t.Fatalf("processou=%v err=%v", processou, err)
	}
}

func TestDespacharGravaDecisaoDoRouter(t *testing.T) {
	repo := memoria.Novo()
	if _, err := (aplicacao.SendMessageUseCase{Repo: repo}).Pedir(context.Background(), pedido()); err != nil {
		t.Fatal(err)
	}
	fake := &remetenteFake{res: dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "recusado"}}
	uc := aplicacao.DespacharUseCase{
		Repo: repo, Remetentes: map[dominio.Canal]portas.MessageSenderPort{dominio.CanalEmail: fake},
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	if _, err := uc.ProcessarProxima(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fake.para != "ana@exemplo.com" || fake.c.Assunto != "Bem-vindo, Ana" {
		t.Fatalf("envio = %q %+v", fake.para, fake.c)
	}
	if repo.Tentativas[0].Status != dominio.TentativaFalhou || repo.Tentativas[0].Motivo != "recusado" {
		t.Fatalf("tentativa = %+v", repo.Tentativas[0])
	}
}

func TestT15_OrdemDoPedidoSubstituiADoTemplate(t *testing.T) {
	repo := memoria.Novo()
	p := pedido()
	p.Destinatario.Telefone = "+5511999998888"
	p.Canais = []dominio.Canal{dominio.CanalSMS, dominio.CanalEmail}
	reg, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	n := repo.Notificacoes[reg.ID]
	if len(n.OrdemCanais) != 2 || n.OrdemCanais[0] != dominio.CanalSMS || repo.Tentativas[0].Canal != dominio.CanalSMS {
		t.Fatalf("ordem = %v, 1ª Tentativa = %s", n.OrdemCanais, repo.Tentativas[0].Canal)
	}
}

func TestT16_CanalPedidoSemVersaoNoTemplateERecusado(t *testing.T) {
	repo := memoria.Novo()
	p := pedido()
	p.Canais = []dominio.Canal{dominio.CanalEmail, "telegram"}
	_, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrCanalSemVersao) || !strings.Contains(err.Error(), "telegram") {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestT17_DestinatarioSemContatoParaNenhumCanalERecusado(t *testing.T) {
	repo := memoria.Novo()
	p := pedido() // só e-mail
	p.Canais = []dominio.Canal{dominio.CanalSMS}
	_, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrSemContato) || err.Error() != "destinatário sem contato para os canais" {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestPrimeiraTentativaPulaCanalSemContato(t *testing.T) {
	repo := memoria.Novo()
	p := pedido()
	p.Destinatario = dominio.Destinatario{Telefone: "+5511999998888"} // ordem do Template: Email → SMS
	if _, err := (aplicacao.SendMessageUseCase{Repo: repo}).Pedir(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if repo.Tentativas[0].Canal != dominio.CanalSMS {
		t.Fatalf("1ª Tentativa = %s, quero sms", repo.Tentativas[0].Canal)
	}
}

func TestT32_PedidoComWhatsAppNaOrdemEAceito(t *testing.T) {
	repo := memoria.Novo()
	p := pedido()
	p.Destinatario.Telefone = "+5511999998888"
	p.Canais = []dominio.Canal{dominio.CanalWhatsApp, dominio.CanalSMS}
	if _, err := (aplicacao.SendMessageUseCase{Repo: repo}).Pedir(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if repo.Tentativas[0].Canal != dominio.CanalWhatsApp {
		t.Fatalf("1ª Tentativa = %s, quero whatsapp", repo.Tentativas[0].Canal)
	}
}

func TestT33_ParametroDoModeloAprovadoEVariavelExigida(t *testing.T) {
	repo := memoria.Novo()
	// Template só com a versão WhatsApp: "nome" é exigido apenas por ser parâmetro 1
	repo.Templates["so-whatsapp"] = dominio.Template{
		ID: "so-whatsapp", OrdemCanais: []dominio.Canal{dominio.CanalWhatsApp},
		Conteudos: map[dominio.Canal]dominio.ConteudoCanal{
			dominio.CanalWhatsApp: {ModeloAprovado: "boas_vindas", Idioma: "pt_BR", Parametros: []string{"nome"}},
		},
	}
	p := pedido()
	p.TemplateID, p.Dados = "so-whatsapp", map[string]string{}
	p.Destinatario.Telefone = "+5511999998888"
	_, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), p)
	if !errors.Is(err, dominio.ErrVariavelFaltando) || !strings.Contains(err.Error(), "nome") {
		t.Fatalf("erro = %v", err)
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func aceita(t *testing.T, repo *memoria.Repositorio, p aplicacao.PedidoNotificacao, idProvedor string) {
	t.Helper()
	if _, err := (aplicacao.SendMessageUseCase{Repo: repo}).Pedir(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	repo.MarcarAceita(len(repo.Historico())-1, idProvedor)
}

func receber() aplicacao.ReceiveMessageUseCase {
	return aplicacao.ReceiveMessageUseCase{Repo: nil, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestReceberAplicaIgnoraEDesconhece(t *testing.T) {
	repo := memoria.Novo()
	aceita(t, repo, pedido(), "msg-1")
	uc := receber()
	uc.Repo = repo
	ctx := context.Background()
	if r, err := uc.Receber(ctx, dominio.AvisoStatus{IDProvedor: "msg-1", Status: dominio.TentativaEntregue}); err != nil || r != aplicacao.AvisoAplicado {
		t.Fatalf("1º aviso: %v %v", r, err)
	}
	if r, _ := uc.Receber(ctx, dominio.AvisoStatus{IDProvedor: "msg-1", Status: dominio.TentativaEntregue}); r != aplicacao.AvisoIgnorado {
		t.Fatalf("aviso repetido: %v", r)
	}
	if r, _ := uc.Receber(ctx, dominio.AvisoStatus{IDProvedor: "nao-existe", Status: dominio.TentativaEntregue}); r != aplicacao.AvisoDesconhecido {
		t.Fatalf("aviso desconhecido: %v", r)
	}
	if s := repo.StatusNotificacao(repo.Historico()[0].NotificacaoID); s != dominio.NotificacaoEntregue {
		t.Fatalf("notificação = %s", s)
	}
}
