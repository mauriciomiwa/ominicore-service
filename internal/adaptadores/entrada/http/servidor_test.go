package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/memoria"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

const corpoValido = `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"}}`

type ambiente struct {
	realm *emissor.Emissor
	repo  *memoria.Repositorio
	h     http.Handler
}

func novo(t *testing.T) ambiente {
	realm := emissor.Novo(t)
	repo := memoria.Novo()
	h := entradahttp.Servidor{
		Auth:      keycloak.Novo(realm.URL, "ominicore", &emissor.Cache{}),
		Enviar:    aplicacao.SendMessageUseCase{Repo: repo},
		Consultar: aplicacao.ConsultarUseCase{Repo: repo},
		OptOut:    aplicacao.OptOutUseCase{Repo: repo},
		Log:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	}.Handler()
	return ambiente{realm: realm, repo: repo, h: h}
}

func (a ambiente) pedir(token, chave, corpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/notificacoes", bytes.NewBufferString(corpo))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if chave != "" {
		req.Header.Set("Idempotency-Key", chave)
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

func TestT1_PedidoSemTokenERecusado(t *testing.T) {
	a := novo(t)
	rec := a.pedir("", "abc-1", corpoValido)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, quero 401", rec.Code)
	}
	if n, _ := a.repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestT3_TokenValidoIdentificaOSistemaDeOrigem(t *testing.T) {
	a := novo(t)
	rec := a.pedir(a.realm.Valido(t, "sistema-a"), "abc-1", corpoValido)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	for _, n := range a.repo.Notificacoes {
		if n.SistemaOrigem != "sistema-a" {
			t.Fatalf("sistema de origem = %q", n.SistemaOrigem)
		}
	}
}

func TestT8_PedidoSemChaveERecusado(t *testing.T) {
	a := novo(t)
	rec := a.pedir(a.realm.Valido(t, "sistema-a"), "", corpoValido)
	if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte("chave de idempotência obrigatória")) {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if n, _ := a.repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestErrosDeValidacaoViram422(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	for nome, corpo := range map[string]string{
		"template inexistente": `{"template_id":"x","destinatario":{"email":"ana@exemplo.com"},"dados":{}}`,
		"variável faltando":    `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{}}`,
	} {
		if rec := a.pedir(tok, "k-"+nome, corpo); rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status = %d", nome, rec.Code)
		}
	}
	if rec := a.pedir(tok, "k-json", `{`); rec.Code != http.StatusBadRequest {
		t.Errorf("json inválido: status = %d", rec.Code)
	}
}

func TestT18_TelefoneForaDoE164ERecusado(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	for _, tel := range []string{"11 99999-8888", "5511999998888", "+0511999998888", "+55"} {
		corpo := `{"template_id":"boas-vindas","destinatario":{"telefone":"` + tel + `"},"dados":{"nome":"Ana"}}`
		rec := a.pedir(tok, "k-"+tel, corpo)
		if rec.Code != http.StatusBadRequest || !bytes.Contains(rec.Body.Bytes(), []byte("E.164")) {
			t.Errorf("%q: status = %d: %s", tel, rec.Code, rec.Body)
		}
	}
	if n, _ := a.repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
	ok := `{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"Ana"}}`
	if rec := a.pedir(tok, "k-valido", ok); rec.Code != http.StatusAccepted {
		t.Fatalf("telefone válido recusado: %d %s", rec.Code, rec.Body)
	}
}

func TestCanaisRepetidosSaoRecusados(t *testing.T) {
	a := novo(t)
	corpo := `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"},"canais":["email","email"]}`
	if rec := a.pedir(a.realm.Valido(t, "sistema-a"), "k-1", corpo); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCanalSemVersaoVira422(t *testing.T) {
	a := novo(t)
	corpo := `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"},"canais":["telegram"]}`
	if rec := a.pedir(a.realm.Valido(t, "sistema-a"), "k-1", corpo); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCanaisVazioERecusado(t *testing.T) {
	a := novo(t)
	corpo := `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"},"canais":[]}`
	if rec := a.pedir(a.realm.Valido(t, "sistema-a"), "k-1", corpo); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func (a ambiente) consultar(token, id string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/notificacoes/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

func TestT43_DonoConsultaStatusETentativas(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	rec := a.pedir(tok, "k-43", `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"}}`)
	var criada struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criada)
	// bounce no Email (contato inválido) e SMS entregue
	a.repo.MarcarAceita(0, "msg-1")
	uc := aplicacao.ReceiveMessageUseCase{Repo: a.repo, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ctx := context.Background()
	_, _ = uc.Receber(ctx, dominio.AvisoStatus{IDProvedor: "msg-1", Status: dominio.TentativaFalhou, Motivo: "sendgrid bounce: 550 ana@exemplo.com", Categoria: dominio.FalhaContatoInvalido})
	a.repo.MarcarAceita(1, "SM1")
	_, _ = uc.Receber(ctx, dominio.AvisoStatus{IDProvedor: "SM1", Status: dominio.TentativaEntregue})

	res := a.consultar(tok, criada.ID)
	corpo := res.Body.String()
	if res.Code != http.StatusOK {
		t.Fatalf("status %d: %s", res.Code, corpo)
	}
	for _, quero := range []string{`"status":"entregue"`, `"canal":"email"`, `"categoria_falha":"contato_invalido"`, `"canal":"sms"`} {
		if !strings.Contains(corpo, quero) {
			t.Errorf("resposta sem %s: %s", quero, corpo)
		}
	}
	for _, proibido := range []string{"ana@exemplo.com", "+5511999998888", "550", "bounce", "motivo"} {
		if strings.Contains(corpo, proibido) {
			t.Errorf("resposta expõe %q: %s", proibido, corpo)
		}
	}
}

func TestT44_OutroSistemaDeOrigemNaoEnxerga(t *testing.T) {
	a := novo(t)
	rec := a.pedir(a.realm.Valido(t, "sistema-a"), "k-44", corpoValido)
	var criada struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criada)
	if res := a.consultar(a.realm.Valido(t, "sistema-b"), criada.ID); res.Code != http.StatusNotFound {
		t.Fatalf("sistema-b: status %d", res.Code)
	}
	if res := a.consultar(a.realm.Valido(t, "sistema-a"), "nao-e-uuid"); res.Code != http.StatusNotFound {
		t.Fatalf("id inválido: status %d", res.Code)
	}
	if res := a.consultar("", criada.ID); res.Code != http.StatusUnauthorized {
		t.Fatalf("sem token: status %d", res.Code)
	}
}

func (a ambiente) optOut(metodo, token, corpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, "/v1/opt-outs", strings.NewReader(corpo))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

const optOutSMS = `{"canal":"sms","contato":"+5511999998888"}`

func TestT45_RegistrarOptOut(t *testing.T) {
	a := novo(t)
	if rec := a.optOut(http.MethodPost, a.realm.Valido(t, "sistema-a"), optOutSMS); rec.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	oo, _ := a.repo.OptOutsDe(context.Background(), dominio.Destinatario{Telefone: "+5511999998888"})
	if !oo[dominio.CanalSMS] || oo[dominio.CanalWhatsApp] {
		t.Fatalf("opt-outs = %v (SMS e WhatsApp são Canais distintos)", oo)
	}
}

func TestT46_RegistroRepetidoEIdempotente(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	for range 2 {
		if rec := a.optOut(http.MethodPost, tok, optOutSMS); rec.Code != http.StatusNoContent {
			t.Fatalf("status %d", rec.Code)
		}
	}
	if n := a.repo.QuantosOptOuts(); n != 1 || len(a.repo.EventosOptOut) != 1 {
		t.Fatalf("opt-outs = %d, eventos = %v", n, a.repo.EventosOptOut)
	}
}

func TestT47_RemocaoGuardaQuemRemoveu(t *testing.T) {
	a := novo(t)
	a.optOut(http.MethodPost, a.realm.Valido(t, "sistema-a"), optOutSMS)
	if rec := a.optOut(http.MethodDelete, a.realm.Valido(t, "sistema-b"), optOutSMS); rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
	oo, _ := a.repo.OptOutsDe(context.Background(), dominio.Destinatario{Telefone: "+5511999998888"})
	if oo[dominio.CanalSMS] || len(a.repo.EventosOptOut) != 2 || a.repo.EventosOptOut[1] != "removido api:sistema-b" {
		t.Fatalf("opt-outs = %v, eventos = %v", oo, a.repo.EventosOptOut)
	}
}

func TestT48_PedidoDeOptOutInvalidoERecusado(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	for _, corpo := range []string{
		`{"canal":"telegram","contato":"x"}`, `{"canal":"email","contato":"sem-arroba"}`,
		`{"canal":"sms","contato":"11 99999-8888"}`, `{"canal":"sms"`, `{"canal":"sms","contato":"+5511999998888","extra":1}`,
	} {
		if rec := a.optOut(http.MethodPost, tok, corpo); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", corpo, rec.Code)
		}
	}
	if n := a.repo.QuantosOptOuts(); n != 0 {
		t.Fatalf("registrou %d opt-outs", n)
	}
	if rec := a.optOut(http.MethodPost, "", optOutSMS); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sem token: %d", rec.Code)
	}
}

func TestT51_OptOutDeSMSFazOrdemSMSEmailSairPorEmail(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	a.optOut(http.MethodPost, tok, optOutSMS)
	rec := a.pedir(tok, "k-51", `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["sms","email"]}`)
	if rec.Code != http.StatusAccepted || a.repo.Historico()[0].Canal != dominio.CanalEmail {
		t.Fatalf("status %d, 1ª Tentativa %s", rec.Code, a.repo.Historico()[0].Canal)
	}
}

func TestT52_TodosOsCanaisEmOptOutRecusamOPedido(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	a.optOut(http.MethodPost, tok, optOutSMS)
	a.optOut(http.MethodPost, tok, `{"canal":"email","contato":"ana@exemplo.com"}`)
	rec := a.pedir(tok, "k-52", `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["sms","email"]}`)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "destinatário sem contato para os canais") {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if n, _ := a.repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}

func TestT54_OptOutValeParaTodosOsSistemasDeOrigem(t *testing.T) {
	a := novo(t)
	a.optOut(http.MethodPost, a.realm.Valido(t, "sistema-a"), `{"canal":"email","contato":"Ana@Exemplo.com"}`)
	rec := a.pedir(a.realm.Valido(t, "sistema-b"), "k-54", `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"},"canais":["email"]}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sistema-b: status %d", rec.Code)
	}
}

func TestT56_ConsultaMostraQuandoFoiAnonimizadaNoHTTP(t *testing.T) {
	a := novo(t)
	tok := a.realm.Valido(t, "sistema-a")
	rec := a.pedir(tok, "k-56", corpoValido)
	var criada struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criada)
	if res := a.consultar(tok, criada.ID); strings.Contains(res.Body.String(), "anonimizada_em") {
		t.Fatalf("anonimizada_em antes da retenção: %s", res.Body)
	}
	a.repo.MarcarAceita(0, "msg-1")
	a.repo.Envelhecer(31 * 24 * time.Hour)
	uc := aplicacao.RetencaoUseCase{Repo: a.repo, Prazo: 30 * 24 * time.Hour, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if n, err := uc.Rodar(context.Background()); err != nil || n != 1 {
		t.Fatalf("retenção: %d %v", n, err)
	}
	res := a.consultar(tok, criada.ID)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"anonimizada_em"`) || !strings.Contains(res.Body.String(), `"tentativas":[{`) {
		t.Fatalf("status %d: %s", res.Code, res.Body)
	}
}

func TestT65_PedidoSoPorWhatsAppDesligadoVira422(t *testing.T) {
	realm := emissor.Novo(t)
	repo := memoria.Novo()
	a := ambiente{realm: realm, repo: repo, h: entradahttp.Servidor{
		Auth:   keycloak.Novo(realm.URL, "ominicore", &emissor.Cache{}),
		Enviar: aplicacao.SendMessageUseCase{Repo: repo, Desligados: map[dominio.Canal]bool{dominio.CanalWhatsApp: true}},
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}.Handler()}
	corpo := `{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["whatsapp"]}`
	rec := a.pedir(a.realm.Valido(t, "sistema-a"), "k-1", corpo)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "canal whatsapp indisponível") {
		t.Fatalf("status = %d corpo = %s", rec.Code, rec.Body.String())
	}
	if n, _ := repo.Contar(); n != 0 {
		t.Fatalf("registrou %d notificações", n)
	}
}
