package openbao_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/dominio"
)

// kvFalso responde o KV v2 só nos caminhos dados; o resto é 404.
func kvFalso(t *testing.T, dados map[string]map[string]string) string {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, ok := dados[r.URL.Path]
		if !ok {
			http.Error(w, `{"errors":[]}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": d, "metadata": map[string]any{}}})
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestPrefixoSeparaHmlDeProd(t *testing.T) {
	addr := kvFalso(t, map[string]map[string]string{
		"/v1/kv/data/ominicore-hml/postgres": {"dsn": "postgres://hml"},
		"/v1/kv/data/ominicore-hml/sendgrid": {"api_key": "SG.hml", "remetente": "hml@exemplo.com"},
		"/v1/kv/data/ominicore-hml/twilio": {
			"account_sid": "AChml", "api_key_sid": "SKhml", "api_key_secret": "x", "messaging_service_sid": "MGhml",
		},
		"/v1/kv/data/ominicore-hml/whatsapp": {"token": "t", "phone_number_id": "1"},
		"/v1/kv/data/ominicore-hml/webhooks": {"sendgrid_chave_publica": "k", "twilio_auth_token": "t", "meta_app_secret": "s", "meta_verify_token": "v"},
		"/v1/kv/data/ominicore-hml/optout":   {"chave": "chave-de-teste-com-mais-de-32-caracteres"},
	})
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t", Prefixo: "ominicore-hml"})
	if err != nil || seg.PostgresDSN != "postgres://hml" || seg.SendGridAPIKey != "SG.hml" || seg.Twilio["account_sid"] != "AChml" {
		t.Fatalf("segredos=%+v err=%v", seg, err)
	}
	// sem prefixo o padrão é prod (ominicore/*), que este cofre não tem
	if _, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t"}); err == nil {
		t.Fatal("leu segredos de hml sem pedir o prefixo de hml")
	}
}

func TestSemSegredoDoTwilioOProcessoNaoTemCredenciais(t *testing.T) {
	addr := kvFalso(t, map[string]map[string]string{
		"/v1/kv/data/ominicore/postgres": {"dsn": "postgres://x"},
		"/v1/kv/data/ominicore/sendgrid": {"api_key": "SG.x", "remetente": "x@exemplo.com"},
		"/v1/kv/data/ominicore/twilio":   {"account_sid": "AC", "api_key_sid": "SK", "api_key_secret": "x"},
	})
	_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: addr, Token: "t"})
	if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), "messaging_service_sid") {
		t.Fatalf("erro = %v, quero ErrOpenBao citando o campo faltante", err)
	}
}

func TestLoginJWTUsaOMountDoAmbiente(t *testing.T) {
	var loginEm string
	kv := map[string]map[string]string{
		"/v1/kv/data/ominicore-hml/postgres": {"dsn": "postgres://hml"},
		"/v1/kv/data/ominicore-hml/sendgrid": {"api_key": "SG.hml", "remetente": "hml@exemplo.com"},
		"/v1/kv/data/ominicore-hml/twilio":   {"account_sid": "AC", "api_key_sid": "SK", "api_key_secret": "x", "messaging_service_sid": "MG"},
		"/v1/kv/data/ominicore-hml/whatsapp": {"token": "t", "phone_number_id": "1"},
		"/v1/kv/data/ominicore-hml/webhooks": {"sendgrid_chave_publica": "k", "twilio_auth_token": "t", "meta_app_secret": "s", "meta_verify_token": "v"},
		"/v1/kv/data/ominicore-hml/optout":   {"chave": "chave-de-teste-com-mais-de-32-caracteres"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			loginEm = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]any{"client_token": "tok-hml"}})
			return
		}
		d, ok := kv[r.URL.Path]
		if !ok || r.Header.Get("X-Vault-Token") != "tok-hml" {
			http.Error(w, `{"errors":["permission denied"]}`, http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": d, "metadata": map[string]any{}}})
	}))
	defer srv.Close()
	seg, err := openbao.Carregar(context.Background(), openbao.Config{
		Endereco: srv.URL, JWT: "jwt-do-keycloak", Role: "ominicore-hml", AuthMount: "jwt-ominicore-hml", Prefixo: "ominicore-hml",
	})
	if err != nil || seg.PostgresDSN != "postgres://hml" {
		t.Fatalf("segredos=%+v err=%v", seg, err)
	}
	if loginEm != "/v1/auth/jwt-ominicore-hml/login" {
		t.Fatalf("login em %q", loginEm)
	}
}

// cofreSemMeta tem tudo o que o core exige e nada do WhatsApp.
func cofreSemMeta() map[string]map[string]string {
	return map[string]map[string]string{
		"/v1/kv/data/ominicore/postgres": {"dsn": "postgres://x"},
		"/v1/kv/data/ominicore/sendgrid": {"api_key": "SG.x", "remetente": "x@exemplo.com"},
		"/v1/kv/data/ominicore/twilio":   {"account_sid": "AC", "api_key_sid": "SK", "api_key_secret": "x", "messaging_service_sid": "MG"},
		"/v1/kv/data/ominicore/webhooks": {"sendgrid_chave_publica": "k", "twilio_auth_token": "t"},
		"/v1/kv/data/ominicore/optout":   {"chave": "chave-de-teste-com-mais-de-32-caracteres"},
	}
}

func TestT63_SemCredenciaisDaMetaOCanalFicaDesligado(t *testing.T) {
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, cofreSemMeta()), Token: "t"})
	if err != nil {
		t.Fatalf("sem a Meta o boot falhou: %v", err)
	}
	if seg.WhatsAppLigado || seg.WhatsApp != nil || seg.SendGridAPIKey != "SG.x" || seg.Twilio["messaging_service_sid"] != "MG" {
		t.Fatalf("segredos = %+v", seg)
	}
}

func TestT64_CredenciaisDaMetaIncompletasImpedemOInicio(t *testing.T) {
	casos := map[string]struct {
		whatsapp, metaWebhook map[string]string
		campo                 string
	}{
		"sem phone_number_id": {map[string]string{"token": "t"}, map[string]string{"meta_app_secret": "s", "meta_verify_token": "v"}, "phone_number_id"},
		"sem App Secret":      {map[string]string{"token": "t", "phone_number_id": "1"}, map[string]string{"meta_verify_token": "v"}, "meta_app_secret"},
		"sem verify token":    {map[string]string{"token": "t", "phone_number_id": "1"}, map[string]string{"meta_app_secret": "s"}, "meta_verify_token"},
	}
	for nome, c := range casos {
		t.Run(nome, func(t *testing.T) {
			kv := cofreSemMeta()
			kv["/v1/kv/data/ominicore/whatsapp"] = c.whatsapp
			for k, v := range c.metaWebhook {
				kv["/v1/kv/data/ominicore/webhooks"][k] = v
			}
			_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
			if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), c.campo) {
				t.Fatalf("erro = %v, quero ErrOpenBao citando %s", err, c.campo)
			}
		})
	}
}

func TestWhatsAppLigadoComTodasAsCredenciais(t *testing.T) {
	kv := cofreSemMeta()
	kv["/v1/kv/data/ominicore/whatsapp"] = map[string]string{"token": "t", "phone_number_id": "1"}
	kv["/v1/kv/data/ominicore/webhooks"]["meta_app_secret"] = "s"
	kv["/v1/kv/data/ominicore/webhooks"]["meta_verify_token"] = "v"
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
	if err != nil || !seg.WhatsAppLigado || seg.WhatsApp["phone_number_id"] != "1" || seg.Webhooks["meta_verify_token"] != "v" {
		t.Fatalf("segredos = %+v err = %v", seg, err)
	}
}

func TestT70_WhatsAppPeloUltraMsgIniciaSemOsSegredosDaMeta(t *testing.T) {
	kv := cofreSemMeta() // webhooks sem meta_app_secret nem meta_verify_token
	kv["/v1/kv/data/ominicore/whatsapp"] = map[string]string{"provedor": "ultramsg", "instance_id": "instance1", "token": "tk"}
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
	if err != nil {
		t.Fatalf("UltraMsg sem os segredos da Meta não iniciou: %v", err)
	}
	if !seg.WhatsAppLigado || seg.ProvedorWhatsApp != dominio.WhatsAppUltraMsg || seg.WhatsApp["instance_id"] != "instance1" || seg.WhatsApp["token"] != "tk" {
		t.Fatalf("segredos = %+v", seg)
	}
}

func TestT71_CredenciaisDoUltraMsgIncompletasImpedemOInicio(t *testing.T) {
	for campo, whatsapp := range map[string]map[string]string{
		"instance_id": {"provedor": "ultramsg", "token": "tk"},
		"token":       {"provedor": "ultramsg", "instance_id": "instance1"},
	} {
		t.Run("sem "+campo, func(t *testing.T) {
			kv := cofreSemMeta()
			kv["/v1/kv/data/ominicore/whatsapp"] = whatsapp
			_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
			if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), campo) {
				t.Fatalf("erro = %v, quero ErrOpenBao citando %s", err, campo)
			}
		})
	}
}

func TestT72_ProvedorDoWhatsAppDesconhecidoImpedeOInicio(t *testing.T) {
	kv := cofreSemMeta()
	kv["/v1/kv/data/ominicore/whatsapp"] = map[string]string{"provedor": "twilio", "instance_id": "i", "token": "tk"}
	_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
	if !errors.Is(err, openbao.ErrOpenBao) || !errors.Is(err, dominio.ErrProvedorWhatsAppDesconhecido) || !strings.Contains(err.Error(), "twilio") {
		t.Fatalf("erro = %v, quero ErrOpenBao citando o Provedor desconhecido", err)
	}
}

func TestT82_WhatsAppPelaZAPIIniciaSemOsSegredosDaMeta(t *testing.T) {
	kv := cofreSemMeta() // webhooks sem meta_app_secret nem meta_verify_token
	kv["/v1/kv/data/ominicore/whatsapp"] = map[string]string{"provedor": "zapi", "instance_id": "inst1", "token": "tk"}
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
	if err != nil {
		t.Fatalf("Z-API sem client_token nem os segredos da Meta não iniciou: %v", err)
	}
	if !seg.WhatsAppLigado || seg.ProvedorWhatsApp != dominio.WhatsAppZAPI || seg.WhatsApp["instance_id"] != "inst1" || seg.WhatsApp["token"] != "tk" {
		t.Fatalf("segredos = %+v", seg)
	}
	kv["/v1/kv/data/ominicore/whatsapp"]["client_token"] = "ct"
	if seg, err = openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"}); err != nil || seg.WhatsApp["client_token"] != "ct" {
		t.Fatalf("client_token não lido: %+v %v", seg.WhatsApp, err)
	}
}

func TestT83_CredenciaisDaZAPIIncompletasImpedemOInicio(t *testing.T) {
	for campo, whatsapp := range map[string]map[string]string{
		"instance_id": {"provedor": "zapi", "token": "tk", "client_token": "ct"},
		"token":       {"provedor": "zapi", "instance_id": "inst1", "client_token": "ct"},
	} {
		t.Run("sem "+campo, func(t *testing.T) {
			kv := cofreSemMeta()
			kv["/v1/kv/data/ominicore/whatsapp"] = whatsapp
			_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
			if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), campo) {
				t.Fatalf("erro = %v, quero ErrOpenBao citando %s", err, campo)
			}
		})
	}
}

func TestT64_ProvedorMetaExplicitoExigeOsCamposDaMeta(t *testing.T) {
	kv := cofreSemMeta()
	kv["/v1/kv/data/ominicore/whatsapp"] = map[string]string{"provedor": "meta", "token": "t", "phone_number_id": "1"}
	_, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
	if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), "meta_app_secret") {
		t.Fatalf("erro = %v, quero ErrOpenBao citando meta_app_secret", err)
	}
	kv["/v1/kv/data/ominicore/webhooks"]["meta_app_secret"] = "s"
	kv["/v1/kv/data/ominicore/webhooks"]["meta_verify_token"] = "v"
	seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
	if err != nil || seg.ProvedorWhatsApp != dominio.WhatsAppMeta {
		t.Fatalf("segredos = %+v err = %v", seg, err)
	}
}
