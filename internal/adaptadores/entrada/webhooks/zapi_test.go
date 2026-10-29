package webhooks_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/entrada/webhooks"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/memoria"
)

const (
	segredoZAPI   = "segredo-atual-do-webhook-com-32-bytes-xx"
	instanciaZAPI = "inst1"
)

// ambienteZAPI monta a rota /webhooks/zapi com a Z-API trocável por teste.
type ambienteZAPI struct {
	*ambiente
	atual webhooks.ZAPIAtual
}

func novoZAPI(t *testing.T) *ambienteZAPI {
	a := &ambienteZAPI{ambiente: &ambiente{repo: memoria.Novo()}, atual: webhooks.ZAPIAtual{Ligada: true, InstanceID: instanciaZAPI, Segredo: segredoZAPI}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	webhooks.Webhooks{
		ZAPI:     func() webhooks.ZAPIAtual { return a.atual },
		Receptor: aplicacao.ReceiveMessageUseCase{Repo: a.repo, Log: log},
		Log:      log,
	}.Registrar(mux)
	a.h = mux
	return a
}

func (a *ambienteZAPI) aviso(segredo, corpo string) *httptest.ResponseRecorder {
	return a.enviar(httptest.NewRequest(http.MethodPost, "/webhooks/zapi/"+segredo, strings.NewReader(corpo)))
}

func statusZAPI(instancia, status, id string) string {
	return `{"type":"MessageStatusCallback","instanceId":"` + instancia + `","status":"` + status + `","ids":["` + id + `"],"phone":"5511999998888","momment":1777494009341}`
}

func entregaZAPI(id, erro string) string {
	e := ""
	if erro != "" {
		e = `,"error":"` + erro + `"`
	}
	return `{"type":"DeliveryCallback","instanceId":"` + instanciaZAPI + `","messageId":"` + id + `","zaapId":"Z1","phone":"5511999998888"` + e + `}`
}

func sucessoZAPI(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"value":true}` {
		t.Fatalf("resposta = %d %s, quero 200 {\"value\":true}", rec.Code, rec.Body)
	}
}

func TestT121_AvisoComOSegredoEAInstanciaCertosEAceito(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "RECEIVED", "MSG1")))
	if a.statusTentativa(0) != dominio.TentativaEntregue || a.statusNotificacao() != dominio.NotificacaoEntregue {
		t.Fatalf("tentativa %s, notificação %s", a.statusTentativa(0), a.statusNotificacao())
	}
}

func TestT122_SegredoErradoERecusadoComoRotaInexistente(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	anterior := a.atual.Segredo
	a.atual.Segredo = "segredo-novo-depois-da-regravacao-xxxxxxx" // regravação trocou o segredo
	for _, segredo := range []string{"outro", anterior} {
		if rec := a.aviso(segredo, statusZAPI(instanciaZAPI, "RECEIVED", "MSG1")); rec.Code != http.StatusNotFound {
			t.Fatalf("segredo %q: status %d, quero 404", segredo, rec.Code)
		}
	}
	if a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("a Tentativa mudou: %s", a.statusTentativa(0))
	}
}

func TestT123_AvisoDeOutraInstanciaERecusado(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	if rec := a.aviso(segredoZAPI, statusZAPI("outra-instancia", "RECEIVED", "MSG1")); rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, quero 404", rec.Code)
	}
	if a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("a Tentativa mudou: %s", a.statusTentativa(0))
	}
}

func TestT124_SemAZAPIAtivaARotaNaoExiste(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	for nome, z := range map[string]webhooks.ZAPIAtual{
		"UltraMsg ou Meta ativos": {Ligada: false, InstanceID: instanciaZAPI, Segredo: segredoZAPI},
		"WhatsApp desligado":      {},
	} {
		a.atual = z
		if rec := a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "RECEIVED", "MSG1")); rec.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d, quero 404", nome, rec.Code)
		}
		if rec := a.aviso("", statusZAPI(instanciaZAPI, "RECEIVED", "MSG1")); rec.Code != http.StatusNotFound {
			t.Fatalf("%s sem segredo: status %d, quero 404", nome, rec.Code)
		}
	}
	if a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("a Tentativa mudou: %s", a.statusTentativa(0))
	}
}

func TestT125_AvisoAutenticoSemEfeitoRespondeSucesso(t *testing.T) {
	a := novoZAPI(t)
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "RECEIVED", "DESCONHECIDA")))
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "READ", "MSG1")))
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "RECEIVED", "MSG1"))) // atrasado
	if a.statusTentativa(0) != dominio.TentativaLida {
		t.Fatalf("regrediu: %s", a.statusTentativa(0))
	}
	sucessoZAPI(t, a.aviso(segredoZAPI, `{"type":"PresenceChatCallback","instanceId":"`+instanciaZAPI+`"}`))
}

func TestT126_LeituraConfirmadaPelaZAPI(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "RECEIVED", "MSG1")))
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "READ", "MSG1")))
	if a.statusTentativa(0) != dominio.TentativaLida || a.statusNotificacao() != dominio.NotificacaoEntregue {
		t.Fatalf("tentativa %s, notificação %s", a.statusTentativa(0), a.statusNotificacao())
	}
}

func TestT127_StatusSENTNaoMudaNada(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "SENT", "MSG1")))
	sucessoZAPI(t, a.aviso(segredoZAPI, statusZAPI(instanciaZAPI, "READ_BY_ME", "MSG1")))
	if a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("tentativa %s, quero enviada", a.statusTentativa(0))
	}
}

func TestT128_NumeroSemWhatsAppPassaAoSMS(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp, dominio.CanalSMS}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, entregaZAPI("MSG1", "Phone number does not exist")))
	h := a.repo.Historico()
	if len(h) != 2 || h[0].Status != dominio.TentativaFalhou || h[0].Categoria != dominio.FalhaContatoInvalido ||
		h[1].Canal != dominio.CanalSMS || h[1].Status != dominio.TentativaPendente {
		t.Fatalf("histórico %+v", h)
	}
	if strings.Contains(h[0].Motivo, "5511999998888") {
		t.Fatalf("o motivo traz o telefone: %q", h[0].Motivo)
	}
}

func TestT129_ShadowBanERecusaDoProvedor(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, entregaZAPI("MSG1", "The message was not sent due to a likely shadow ban")))
	h := a.repo.Historico()
	if h[0].Status != dominio.TentativaFalhou || h[0].Categoria != dominio.FalhaRecusadoPeloProvedor {
		t.Fatalf("tentativa %+v", h[0])
	}
}

func TestT130_AoEnviarSemErroNaoMudaNada(t *testing.T) {
	a := novoZAPI(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "MSG1")
	sucessoZAPI(t, a.aviso(segredoZAPI, entregaZAPI("MSG1", "")))
	if a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("tentativa %s, quero enviada", a.statusTentativa(0))
	}
}

// erro de banco: 500 para a Z-API reenviar (o efeito é idempotente).
func TestAvisoDaZAPIComErroNoBancoResponde500(t *testing.T) {
	a := novoZAPI(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	webhooks.Webhooks{ZAPI: func() webhooks.ZAPIAtual { return a.atual }, Receptor: receptorQueFalha{}, Log: log}.Registrar(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/zapi/"+segredoZAPI, strings.NewReader(statusZAPI(instanciaZAPI, "READ", "MSG1"))))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, quero 500", rec.Code)
	}
}

type receptorQueFalha struct{}

func (receptorQueFalha) Receber(context.Context, dominio.AvisoStatus) (aplicacao.ResultadoAviso, error) {
	return aplicacao.AvisoIgnorado, errors.New("banco fora")
}
