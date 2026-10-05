package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ominicore/internal/adaptadores/entrada/webhooks"
	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/adaptadores/saida/ultramsg"
	"ominicore/internal/adaptadores/saida/whatsapp"
	"ominicore/internal/adaptadores/saida/zapi"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/logs"
	"ominicore/internal/testes/memoria"
)

func envCom(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

// enderecoFechado devolve um endereço em que ninguém escuta.
func enderecoFechado(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return "http://" + addr
}

func TestT13_SemOpenBaoOProcessoNaoInicia(t *testing.T) {
	negando := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"errors":["permission denied"]}`, http.StatusForbidden)
	}))
	defer negando.Close()
	casos := map[string]map[string]string{
		"indisponível":   {"OMINICORE_BAO_ADDR": enderecoFechado(t), "OMINICORE_BAO_TOKEN": "dev"},
		"nega a leitura": {"OMINICORE_BAO_ADDR": negando.URL, "OMINICORE_BAO_TOKEN": "dev"},
		"sem endereço":   {"OMINICORE_BAO_TOKEN": "dev"},
	}
	for nome, vars := range casos {
		for _, sub := range []string{"api", "worker"} {
			t.Run(nome+"/"+sub, func(t *testing.T) {
				ctx, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancelar()
				vars["OMINICORE_ADDR"] = "127.0.0.1:0"
				var saida bytes.Buffer
				err := run(ctx, []string{sub}, envCom(vars), &saida)
				if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(strings.ToLower(err.Error()), "openbao") {
					t.Fatalf("erro = %v, quero ErrOpenBao citando o OpenBao", err)
				}
				if strings.Contains(saida.String(), "no ar") {
					t.Fatalf("o processo subiu mesmo sem segredos: %s", saida.String())
				}
			})
		}
	}
}

func TestRealmInternoTrocaOHostPublicoPeloBackchannel(t *testing.T) {
	casos := map[config]string{
		{KCEmissor: "https://auth.flipapi.com.br/realms/ominicore-hml", KCURL: "http://keycloak:8080/"}: "http://keycloak:8080/realms/ominicore-hml",
		{KCEmissor: "http://localhost:8081/realms/ominicore"}:                                           "http://localhost:8081/realms/ominicore",
	}
	for cfg, quero := range casos {
		if got := cfg.realmInterno(); got != quero {
			t.Errorf("realmInterno(%+v) = %q, quero %q", cfg, got, quero)
		}
	}
}

func TestT63_SemCredenciaisDaMetaSobeComEmailESMSEAvisaNoLog(t *testing.T) {
	seg := openbao.Segredos{SendGridAPIKey: "SG.x", Twilio: map[string]string{"account_sid": "AC"}} // WhatsAppLigado = false
	r := remetentes(config{}, seg)
	if _, ok := r[dominio.CanalEmail]; !ok {
		t.Fatal("sem Remetente de Email")
	}
	if _, ok := r[dominio.CanalSMS]; !ok {
		t.Fatal("sem Remetente de SMS")
	}
	if _, ok := r[dominio.CanalWhatsApp]; ok {
		t.Fatal("Remetente de WhatsApp registrado sem credenciais")
	}
	if !canaisDesligados(seg)[dominio.CanalWhatsApp] {
		t.Fatal("WhatsApp não ficou desligado")
	}
	var saida bytes.Buffer
	anunciarCanais(logs.Novo(&saida, slog.LevelInfo), seg)
	if !strings.Contains(saida.String(), "whatsapp desligado") {
		t.Fatalf("log do boot = %q", saida.String())
	}
	seg.WhatsAppLigado = true
	if _, ok := remetentes(config{}, seg)[dominio.CanalWhatsApp]; !ok || len(canaisDesligados(seg)) != 0 {
		t.Fatal("WhatsApp ligado sem Remetente")
	}
}

func TestT70_UltraMsgLigadoRegistraORemetenteEAvisaOProvedorNoLog(t *testing.T) {
	seg := openbao.Segredos{
		SendGridAPIKey: "SG.x", Twilio: map[string]string{"account_sid": "AC"},
		WhatsAppLigado: true, ProvedorWhatsApp: dominio.WhatsAppUltraMsg,
		WhatsApp: map[string]string{"instance_id": "instance1", "token": "tk-secreto"},
	}
	if _, ok := remetentes(config{}, seg)[dominio.CanalWhatsApp].(ultramsg.Remetente); !ok {
		t.Fatal("Remetente do WhatsApp não é o UltraMsg")
	}
	if len(canaisDesligados(seg)) != 0 || metaLigada(seg) {
		t.Fatal("UltraMsg ligado deixou o Canal desligado ou as rotas da Meta ligadas")
	}
	var saida bytes.Buffer
	anunciarCanais(logs.Novo(&saida, slog.LevelInfo), seg)
	if !strings.Contains(saida.String(), "whatsapp ligado: ultramsg") || strings.Contains(saida.String(), "tk-secreto") {
		t.Fatalf("log do boot = %q", saida.String())
	}
	seg.ProvedorWhatsApp = dominio.WhatsAppMeta
	if _, ok := remetentes(config{}, seg)[dominio.CanalWhatsApp].(whatsapp.Remetente); !ok || !metaLigada(seg) {
		t.Fatal("com a Meta, o Remetente ou as rotas não são da Meta")
	}
}

func TestT82_ZAPILigadaRegistraORemetenteEAvisaOProvedorNoLog(t *testing.T) {
	seg := openbao.Segredos{
		SendGridAPIKey: "SG.x", Twilio: map[string]string{"account_sid": "AC"},
		WhatsAppLigado: true, ProvedorWhatsApp: dominio.WhatsAppZAPI,
		WhatsApp: map[string]string{"instance_id": "inst1", "token": "tk-secreto", "client_token": "ct-secreto"},
	}
	r, ok := remetentes(config{}, seg)[dominio.CanalWhatsApp].(zapi.Remetente)
	if !ok || r.Cred.ClientToken != "ct-secreto" {
		t.Fatalf("Remetente do WhatsApp não é a Z-API com o client_token: %+v", r)
	}
	if len(canaisDesligados(seg)) != 0 || metaLigada(seg) {
		t.Fatal("Z-API ligada deixou o Canal desligado ou as rotas da Meta ligadas")
	}
	var saida bytes.Buffer
	anunciarCanais(logs.Novo(&saida, slog.LevelInfo), seg)
	if l := saida.String(); !strings.Contains(l, "whatsapp ligado: zapi") || strings.Contains(l, "tk-secreto") || strings.Contains(l, "ct-secreto") {
		t.Fatalf("log do boot = %q", l)
	}
}

func TestT81_WebhookDaMetaComOUltraMsgAtivoResponde404(t *testing.T) {
	webhookDaMetaResponde404(t, dominio.WhatsAppUltraMsg)
}

func TestT91_WebhookDaMetaComAZAPIAtivaResponde404(t *testing.T) {
	webhookDaMetaResponde404(t, dominio.WhatsAppZAPI)
}

// webhookDaMetaResponde404: com um Provedor de texto, as rotas da Meta não
// existem (REQ-21).
func webhookDaMetaResponde404(t *testing.T, prov dominio.ProvedorWhatsApp) {
	t.Helper()
	seg := openbao.Segredos{WhatsAppLigado: true, ProvedorWhatsApp: prov}
	mux := http.NewServeMux()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := memoria.Novo()
	webhooks.Webhooks{
		Cred:       webhooks.Credenciais{MetaAppSecret: "s", MetaVerifyToken: "v"},
		MetaLigada: metaLigada(seg),
		Receptor:   aplicacao.ReceiveMessageUseCase{Repo: repo, Log: log},
		Log:        log,
	}.Registrar(mux)
	for _, r := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=v&hub.challenge=1", nil),
		httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", strings.NewReader(`{"entry":[{"changes":[{"value":{"statuses":[{"id":"42","status":"read"}]}}]}]}`)),
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, r)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s = %d, quero 404", r.Method, r.URL.Path, rec.Code)
		}
	}
	if len(repo.Historico()) != 0 {
		t.Fatal("aviso mudou alguma Tentativa")
	}
}
