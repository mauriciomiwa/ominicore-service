package zapi_test

import (
	"strings"
	"testing"

	"ominicore/internal/adaptadores/saida/zapi"
	"ominicore/internal/dominio"
	"ominicore/internal/testes/zapifake"
)

func configZAPI(t *testing.T) dominio.ConfigWhatsApp {
	t.Helper()
	cfg, err := dominio.NovaConfigWhatsApp("zapi", map[string]string{
		"instance_id": zapifake.InstanceID, "token": zapifake.Token, "client_token": zapifake.ClientToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestCadastroDosWebhooksPoeAMesmaURLNosDoisAvisos(t *testing.T) {
	f := zapifake.Novo(t)
	url := "https://notify.flipapi.com.br/webhooks/zapi/segredo-do-teste"
	if err := zapi.NovoWebhooks(f.URL).Cadastrar(t.Context(), configZAPI(t), url); err != nil {
		t.Fatal(err)
	}
	if w := f.Webhooks(); w["update-webhook-delivery"] != url || w["update-webhook-message-status"] != url {
		t.Fatalf("URLs cadastradas = %v", w)
	}
}

func TestFalhaNoCadastroNaoLevaTokenNemSegredoAoErro(t *testing.T) {
	f := zapifake.Novo(t)
	f.FalharCadastro(true)
	url := "https://notify.flipapi.com.br/webhooks/zapi/segredo-do-teste"
	err := zapi.NovoWebhooks(f.URL).Cadastrar(t.Context(), configZAPI(t), url)
	if err == nil {
		t.Fatal("o cadastro devia falhar")
	}
	for _, s := range []string{zapifake.Token, zapifake.ClientToken, "segredo-do-teste"} {
		if strings.Contains(err.Error(), s) {
			t.Fatalf("o erro traz um segredo: %v", err)
		}
	}
}

func TestCadastroSemZAPIERecusado(t *testing.T) {
	cfg, _ := dominio.NovaConfigWhatsApp("ultramsg", map[string]string{"instance_id": "i", "token": "t"})
	if err := zapi.NovoWebhooks("http://nao-chamar").Cadastrar(t.Context(), cfg, "https://x"); err == nil {
		t.Fatal("cadastro com UltraMsg devia ser recusado")
	}
}
