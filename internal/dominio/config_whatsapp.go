package dominio

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// CamposWhatsApp são os campos exigidos de cada Provedor do WhatsApp, no boot,
// na releitura e na gravação pela API (REQ-6, REQ-22, REQ-28).
var CamposWhatsApp = map[ProvedorWhatsApp][]string{
	WhatsAppMeta:     {"token", "phone_number_id", "meta_app_secret", "meta_verify_token"},
	WhatsAppUltraMsg: {"instance_id", "token"},
	WhatsAppZAPI:     {"instance_id", "token"},
}

// CamposOpcionaisWhatsApp entram quando existem: o `client_token` da Z-API só
// existe se a conta ativou o token de segurança; os campos do webhook da Z-API
// (REQ-37, REQ-40) são escritos só pelo OminiCore.
var CamposOpcionaisWhatsApp = map[ProvedorWhatsApp][]string{
	WhatsAppZAPI: {"client_token", CampoWebhookSegredo, CampoWebhooksEstado, CampoWebhooksEm},
}

// Campos do webhook da Z-API, gerados e gravados pelo OminiCore: o segredo da
// URL do aviso e o estado do cadastro na Z-API (REQ-40, ADR 0007).
const (
	CampoWebhookSegredo = "webhook_segredo"
	CampoWebhooksEstado = "webhooks_estado"
	CampoWebhooksEm     = "webhooks_em"
	WebhooksCadastrados = "cadastrados"
	WebhooksFalhou      = "falhou"
)

// CamposDoOminiCore não vêm do Operador: se chegarem na gravação, são
// descartados antes de gerar os valores (REQ-40).
var CamposDoOminiCore = []string{CampoWebhookSegredo, CampoWebhooksEstado, CampoWebhooksEm}

// segredosWhatsApp nunca saem pela API nem para log (REQ-23); os demais campos
// (instance_id, phone_number_id, estado dos webhooks) são públicos.
var segredosWhatsApp = map[string]bool{
	"token": true, "client_token": true, "meta_app_secret": true, "meta_verify_token": true,
	CampoWebhookSegredo: true,
}

// ErrConfigIncompleta: falta um campo exigido pelo Provedor (REQ-22, REQ-26).
var ErrConfigIncompleta = errors.New("configuração whatsapp incompleta")

// ConfigWhatsApp é a configuração do Canal WhatsApp no ambiente (RN-11):
// desligado, ou ligado por um Provedor com os campos dele.
type ConfigWhatsApp struct {
	Ligado   bool
	Provedor ProvedorWhatsApp
	Campos   map[string]string
}

// NovaConfigWhatsApp lê o Provedor e guarda só os campos conhecidos dele; o
// resto é descartado. Valide com Validar.
func NovaConfigWhatsApp(provedor string, campos map[string]string) (ConfigWhatsApp, error) {
	p, err := LerProvedorWhatsApp(provedor)
	if err != nil {
		return ConfigWhatsApp{}, err
	}
	c := ConfigWhatsApp{Ligado: true, Provedor: p, Campos: map[string]string{}}
	for _, k := range slices.Concat(CamposWhatsApp[p], CamposOpcionaisWhatsApp[p]) {
		if v := campos[k]; v != "" {
			c.Campos[k] = v
		}
	}
	return c, nil
}

// Validar exige os campos do Provedor; o erro cita o primeiro que falta.
func (c ConfigWhatsApp) Validar() error {
	if !c.Ligado {
		return nil
	}
	for _, k := range CamposWhatsApp[c.Provedor] {
		if c.Campos[k] == "" {
			return fmt.Errorf("%w: sem o campo %s", ErrConfigIncompleta, k)
		}
	}
	return nil
}

// Publicos são os campos que não são segredo, para a consulta (REQ-23).
func (c ConfigWhatsApp) Publicos() map[string]string {
	out := map[string]string{}
	for k, v := range c.Campos {
		if !segredosWhatsApp[k] {
			out[k] = v
		}
	}
	return out
}

// Segredos diz, para cada segredo do Provedor, se está configurado — nunca o
// valor (REQ-23).
func (c ConfigWhatsApp) Segredos() map[string]bool {
	out := map[string]bool{}
	for _, k := range slices.Concat(CamposWhatsApp[c.Provedor], CamposOpcionaisWhatsApp[c.Provedor]) {
		if segredosWhatsApp[k] {
			out[k] = c.Campos[k] != ""
		}
	}
	return out
}

// Igual compara duas configurações, para a releitura só trocar quando mudou.
func (c ConfigWhatsApp) Igual(o ConfigWhatsApp) bool {
	return c.Ligado == o.Ligado && c.Provedor == o.Provedor && maps.Equal(c.Campos, o.Campos)
}
