package webhooks_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/entrada/webhooks"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/memoria"
)

const (
	authToken   = "auth-token-do-teste"
	appSecret   = "app-secret-do-teste"
	verifyToken = "verify-token-do-teste"
	urlPublica  = "https://notify-hml.flipapi.com.br"
)

type ambiente struct {
	repo  *memoria.Repositorio
	chave *ecdsa.PrivateKey
	h     http.Handler
}

func novo(t *testing.T) *ambiente {
	chave, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(&chave.PublicKey)
	pub, err := webhooks.ChavePublicaSendGrid(string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})))
	if err != nil {
		t.Fatal(err)
	}
	a := &ambiente{repo: memoria.Novo(), chave: chave}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	webhooks.Webhooks{
		Cred:     webhooks.Credenciais{SendGridChavePublica: pub, TwilioAuthToken: authToken, MetaAppSecret: appSecret, MetaVerifyToken: verifyToken, URLPublica: urlPublica},
		Receptor: aplicacao.ReceiveMessageUseCase{Repo: a.repo, Log: log},
		Log:      log,
	}.Registrar(mux)
	a.h = mux
	return a
}

// aceita registra uma Notificação (ordem do pedido) com a 1ª Tentativa aceita.
func (a *ambiente) aceita(t *testing.T, canais []dominio.Canal, idProvedor string) {
	t.Helper()
	p := aplicacao.PedidoNotificacao{
		SistemaOrigem: "sistema-a", ChaveIdempotencia: idProvedor, TemplateID: "boas-vindas", Canais: canais,
		Destinatario: dominio.Destinatario{Email: "ana@exemplo.com", Telefone: "+5511999998888"}, Dados: map[string]string{"nome": "Ana"},
	}
	if _, err := (aplicacao.SendMessageUseCase{Repo: a.repo}).Pedir(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	a.repo.MarcarAceita(len(a.repo.Historico())-1, idProvedor)
}

func (a *ambiente) enviar(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

func (a *ambiente) sendgrid(t *testing.T, corpo string, assinar bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/sendgrid", strings.NewReader(corpo))
	if assinar {
		const ts = "1700000000"
		h := sha256.Sum256([]byte(ts + corpo))
		sig, err := ecdsa.SignASN1(rand.Reader, a.chave, h[:])
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Twilio-Email-Event-Webhook-Signature", base64.StdEncoding.EncodeToString(sig))
		req.Header.Set("X-Twilio-Email-Event-Webhook-Timestamp", ts)
	}
	return a.enviar(req)
}

func (a *ambiente) twilio(form url.Values, assinar bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/twilio", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if assinar {
		req.Header.Set("X-Twilio-Signature", base64.StdEncoding.EncodeToString(webhooks.AssinarTwilio(authToken, urlPublica+"/webhooks/twilio", form)))
	}
	return a.enviar(req)
}

func (a *ambiente) meta(corpo string, segredo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", bytes.NewBufferString(corpo))
	if segredo != "" {
		m := hmac.New(sha256.New, []byte(segredo))
		m.Write([]byte(corpo))
		req.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(m.Sum(nil)))
	}
	return a.enviar(req)
}

func statusMeta(id, status string) string {
	return `{"object":"whatsapp_business_account","entry":[{"changes":[{"value":{"statuses":[{"id":"` + id + `","status":"` + status + `"}]}}]}]}`
}

func (a *ambiente) statusTentativa(i int) dominio.StatusTentativa { return a.repo.StatusTentativa(i) }

func (a *ambiente) statusNotificacao() dominio.StatusNotificacao {
	return a.repo.StatusNotificacao(a.repo.Historico()[0].NotificacaoID)
}

func TestT34_EntregaConfirmadaPeloEmail(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalEmail}, "msg-sg-1")
	rec := a.sendgrid(t, `[{"event":"processed","sg_message_id":"msg-sg-1.filter-1"},{"event":"delivered","sg_message_id":"msg-sg-1.filter-1"}]`, true)
	if rec.Code != http.StatusNoContent || a.statusTentativa(0) != dominio.TentativaEntregue || a.statusNotificacao() != dominio.NotificacaoEntregue {
		t.Fatalf("status %d, tentativa %s, notificação %s", rec.Code, a.statusTentativa(0), a.statusNotificacao())
	}
}

func TestT35_EntregaConfirmadaPeloSMS(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalSMS}, "SM123")
	rec := a.twilio(url.Values{"MessageSid": {"SM123"}, "MessageStatus": {"delivered"}, "AccountSid": {"AC1"}}, true)
	if rec.Code != http.StatusNoContent || a.statusTentativa(0) != dominio.TentativaEntregue || a.statusNotificacao() != dominio.NotificacaoEntregue {
		t.Fatalf("status %d, tentativa %s", rec.Code, a.statusTentativa(0))
	}
}

func TestT36_LeituraConfirmadaPeloWhatsApp(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "wamid.1")
	a.meta(statusMeta("wamid.1", "delivered"), appSecret)
	rec := a.meta(statusMeta("wamid.1", "read"), appSecret)
	if rec.Code != http.StatusNoContent || a.statusTentativa(0) != dominio.TentativaLida || a.statusNotificacao() != dominio.NotificacaoEntregue {
		t.Fatalf("status %d, tentativa %s, notificação %s", rec.Code, a.statusTentativa(0), a.statusNotificacao())
	}
}

func TestT37_AssinaturaInvalidaERecusada(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalEmail, dominio.CanalSMS, dominio.CanalWhatsApp}, "id-1")
	casos := map[string]*httptest.ResponseRecorder{
		"sendgrid sem assinatura": a.sendgrid(t, `[{"event":"delivered","sg_message_id":"id-1.x"}]`, false),
		"twilio sem assinatura":   a.twilio(url.Values{"MessageSid": {"id-1"}, "MessageStatus": {"delivered"}}, false),
		"meta com outro segredo":  a.meta(statusMeta("id-1", "delivered"), "outro-segredo"),
		"meta sem assinatura":     a.meta(statusMeta("id-1", "delivered"), ""),
	}
	// Twilio com assinatura de outra URL (host da requisição não conta)
	form := url.Values{"MessageSid": {"id-1"}, "MessageStatus": {"delivered"}}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/twilio", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Twilio-Signature", base64.StdEncoding.EncodeToString(webhooks.AssinarTwilio(authToken, "http://atacante/webhooks/twilio", form)))
	casos["twilio com outra URL"] = a.enviar(req)
	for nome, rec := range casos {
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d, quero 401", nome, rec.Code)
		}
	}
	if a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("aviso inválido mudou a Tentativa: %s", a.statusTentativa(0))
	}
}

func TestT38_StatusNaoRegride(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp}, "wamid.2")
	a.meta(statusMeta("wamid.2", "read"), appSecret)
	rec := a.meta(statusMeta("wamid.2", "delivered"), appSecret)
	if rec.Code != http.StatusNoContent || a.statusTentativa(0) != dominio.TentativaLida {
		t.Fatalf("status %d, tentativa %s", rec.Code, a.statusTentativa(0))
	}
}

func TestT39_AvisoDeTentativaDesconhecida(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalEmail}, "msg-conhecido")
	rec := a.sendgrid(t, `[{"event":"delivered","sg_message_id":"msg-desconhecido.x"}]`, true)
	if rec.Code != http.StatusNoContent || a.statusTentativa(0) != dominio.TentativaEnviada {
		t.Fatalf("status %d, tentativa %s", rec.Code, a.statusTentativa(0))
	}
}

func TestT40_VerificacaoDoWebhookDaMeta(t *testing.T) {
	a := novo(t)
	ok := a.enviar(httptest.NewRequest(http.MethodGet, "/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token="+verifyToken+"&hub.challenge=12345", nil))
	if ok.Code != http.StatusOK || ok.Body.String() != "12345" {
		t.Fatalf("token certo: %d %q", ok.Code, ok.Body.String())
	}
	if html := a.enviar(httptest.NewRequest(http.MethodGet, "/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token="+verifyToken+"&hub.challenge=%3Cscript%3E", nil)); html.Code != http.StatusBadRequest {
		t.Fatalf("desafio não numérico: %d", html.Code)
	}
	errado := a.enviar(httptest.NewRequest(http.MethodGet, "/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=errado&hub.challenge=12345", nil))
	if errado.Code != http.StatusForbidden || strings.Contains(errado.Body.String(), "12345") {
		t.Fatalf("token errado: %d %q", errado.Code, errado.Body.String())
	}
}

func TestT41_BounceNoEmailPassaAoSMS(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalEmail, dominio.CanalSMS}, "msg-sg-2")
	rec := a.sendgrid(t, `[{"event":"bounce","type":"bounce","sg_message_id":"msg-sg-2.x","reason":"550 mailbox unavailable"}]`, true)
	h := a.repo.Historico()
	if rec.Code != http.StatusNoContent || len(h) != 2 || h[0].Status != dominio.TentativaFalhou || h[0].Categoria != dominio.FalhaContatoInvalido ||
		h[1].Canal != dominio.CanalSMS || h[1].Status != dominio.TentativaPendente {
		t.Fatalf("status %d, histórico %+v", rec.Code, h)
	}
}

func TestT42_FalhaTardiaNaoReabreNotificacaoEntregue(t *testing.T) {
	a := novo(t)
	a.aceita(t, []dominio.Canal{dominio.CanalWhatsApp, dominio.CanalSMS}, "wamid.3")
	notificacao := a.repo.Historico()[0].NotificacaoID
	a.repo.AdicionarAceita(notificacao, dominio.CanalSMS, "SM9") // outra Tentativa ainda aberta
	a.meta(statusMeta("wamid.3", "delivered"), appSecret)        // Notificação entregue pelo WhatsApp
	rec := a.twilio(url.Values{"MessageSid": {"SM9"}, "MessageStatus": {"undelivered"}, "ErrorCode": {"30003"}}, true)
	h := a.repo.Historico()
	if rec.Code != http.StatusNoContent || len(h) != 2 || h[1].Status != dominio.TentativaFalhou ||
		h[1].Categoria != dominio.FalhaContatoInvalido || a.statusNotificacao() != dominio.NotificacaoEntregue {
		t.Fatalf("status %d, histórico %+v, notificação %s", rec.Code, h, a.statusNotificacao())
	}
}

func TestChavePublicaSendGridAceitaPEMEBase64(t *testing.T) {
	chave, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalPKIXPublicKey(&chave.PublicKey)
	for nome, texto := range map[string]string{
		"pem":        string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})),
		"base64 der": base64.StdEncoding.EncodeToString(der),
	} {
		if _, err := webhooks.ChavePublicaSendGrid(texto); err != nil {
			t.Errorf("%s: %v", nome, err)
		}
	}
	if _, err := webhooks.ChavePublicaSendGrid("local-invalida"); err == nil {
		t.Error("chave inválida aceita")
	}
}
