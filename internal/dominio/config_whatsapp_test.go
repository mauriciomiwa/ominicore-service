package dominio

import (
	"errors"
	"strings"
	"testing"
)

func TestNovaConfigWhatsAppGuardaSoOsCamposDoProvedor(t *testing.T) {
	c, err := NovaConfigWhatsApp("zapi", map[string]string{"instance_id": "i", "token": "t", "client_token": "ct", "phone_number_id": "x", "lixo": "y"})
	if err != nil || c.Validar() != nil {
		t.Fatalf("config = %+v, %v", c, err)
	}
	if len(c.Campos) != 3 || c.Campos["phone_number_id"] != "" {
		t.Fatalf("campos = %v", c.Campos)
	}
	if p := c.Publicos(); len(p) != 1 || p["instance_id"] != "i" {
		t.Fatalf("públicos = %v", p)
	}
	// o segredo do webhook da Z-API entra na lista, só como configurado ou não (REQ-40)
	if s := c.Segredos(); len(s) != 3 || !s["token"] || !s["client_token"] || s[CampoWebhookSegredo] {
		t.Fatalf("segredos = %v", s)
	}
}

func TestValidarCitaOCampoFaltante(t *testing.T) {
	c, _ := NovaConfigWhatsApp("meta", map[string]string{"token": "t", "phone_number_id": "1", "meta_app_secret": "s"})
	if err := c.Validar(); !errors.Is(err, ErrConfigIncompleta) || !strings.Contains(err.Error(), "meta_verify_token") {
		t.Fatalf("erro = %v", err)
	}
	if _, err := NovaConfigWhatsApp("twilio", nil); !errors.Is(err, ErrProvedorWhatsAppDesconhecido) {
		t.Fatalf("erro = %v", err)
	}
	if (ConfigWhatsApp{}).Validar() != nil {
		t.Fatal("desligado não precisa de campos")
	}
}

func TestSegredosSemClientTokenFicaFalso(t *testing.T) {
	c, _ := NovaConfigWhatsApp("zapi", map[string]string{"instance_id": "i", "token": "t"})
	if s := c.Segredos(); !s["token"] || s["client_token"] {
		t.Fatalf("segredos = %v", s)
	}
}
