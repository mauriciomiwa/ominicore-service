// Package twilio implementa o Canal SMS pela API de Messages do Twilio, com
// API Key e Messaging Service (REQ-7).
// https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource
package twilio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ominicore/internal/adaptadores/saida/provedor"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// URLPadrao é a API pública do Twilio; testes apontam para um fake.
const URLPadrao = "https://api.twilio.com"

// Credenciais vêm do OpenBao (`kv/<prefixo>/twilio`, REQ-6).
type Credenciais struct {
	AccountSID          string
	APIKeySID           string
	APIKeySecret        string
	MessagingServiceSID string
}

// Remetente envia SMS pelo Twilio.
type Remetente struct {
	URL  string
	Cred Credenciais
	HTTP *http.Client
	// StatusCallback é a URL pública do webhook de status (REQ-12); vazia no
	// ambiente local, onde o Twilio não alcança o OminiCore.
	StatusCallback string
}

var _ portas.MessageSenderPort = Remetente{}

// Novo monta o Remetente com timeout de 10 s por Tentativa.
func Novo(baseURL string, c Credenciais) Remetente {
	if baseURL == "" {
		baseURL = URLPadrao
	}
	return Remetente{URL: baseURL, Cred: c, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Enviar manda o corpo da versão SMS; o Messaging Service escolhe o número.
func (r Remetente) Enviar(ctx context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	form := url.Values{"To": {para}, "MessagingServiceSid": {r.Cred.MessagingServiceSID}, "Body": {c.Corpo}}
	if r.StatusCallback != "" {
		form.Set("StatusCallback", r.StatusCallback)
	}
	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", r.URL, url.PathEscape(r.Cred.AccountSID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar requisição: " + err.Error()}
	}
	req.SetBasicAuth(r.Cred.APIKeySID, r.Cred.APIKeySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return provedor.Inacessivel("twilio")
	}
	defer func() { _ = resp.Body.Close() }()
	return classificar(resp)
}

// resposta cobre o sucesso (sid) e o erro (code, message) do Twilio.
type resposta struct {
	SID     string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func classificar(resp *http.Response) dominio.ResultadoEnvio {
	var r resposta
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&r)
	motivo := ""
	if resp.StatusCode >= 300 {
		motivo = fmt.Sprintf("twilio %d/%d: %s", resp.StatusCode, r.Code, r.Message)
	}
	return provedor.Classificar(resp.StatusCode, motivo, r.SID)
}
