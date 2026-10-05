// Package zapi implementa o Canal WhatsApp pela Z-API, um gateway não oficial
// sobre o WhatsApp Web (RN-4, ADR 0004): envia o texto da versão WhatsApp do
// Template (REQ-10).
// https://developer.z-api.io/message/send-text
package zapi

import (
	"bytes"
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

// URLPadrao é a API da Z-API; testes apontam para uma Z-API falsa.
const URLPadrao = "https://api.z-api.io"

// Credenciais vêm do OpenBao (`kv/<prefixo>/whatsapp` com provedor=zapi, REQ-6).
// ClientToken é o token de segurança da conta, opcional.
type Credenciais struct {
	InstanceID  string
	Token       string
	ClientToken string
}

// Remetente envia WhatsApp pela Z-API.
type Remetente struct {
	URL  string
	Cred Credenciais
	HTTP *http.Client
}

var _ portas.MessageSenderPort = Remetente{}

// Novo monta o Remetente com timeout de 10 s por Tentativa.
func Novo(baseURL string, c Credenciais) Remetente {
	if baseURL == "" {
		baseURL = URLPadrao
	}
	return Remetente{URL: baseURL, Cred: c, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Enviar manda o texto preenchido ao telefone sem o "+" do E.164; sem texto ou
// acima do limite é Falha definitiva sem chamar o Provedor (T-77, T-89). A
// Z-API exige o token na URL: nenhum erro da chamada leva a URL ao motivo (T-84).
func (r Remetente) Enviar(ctx context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	if err := dominio.WhatsAppZAPI.ServirTexto(c); err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: err.Error()}
	}
	corpo, err := json.Marshal(map[string]string{"phone": strings.TrimPrefix(para, "+"), "message": c.Corpo})
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar requisição zapi"}
	}
	endpoint := strings.TrimRight(r.URL, "/") + "/instances/" + url.PathEscape(r.Cred.InstanceID) +
		"/token/" + url.PathEscape(r.Cred.Token) + "/send-text"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(corpo))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar requisição zapi"}
	}
	req.Header.Set("Content-Type", "application/json")
	if r.Cred.ClientToken != "" {
		req.Header.Set("Client-Token", r.Cred.ClientToken)
	}
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return provedor.Inacessivel("zapi")
	}
	defer func() { _ = resp.Body.Close() }()
	return r.classificar(resp)
}

type resposta struct {
	MessageID string `json:"messageId"`
	ID        string `json:"id"`
	Error     string `json:"error"`
}

// classificar: 429 e 5xx Falha temporária; campo error ou outro status fora de
// 2xx Falha definitiva; 2xx sem error Aceito, mesmo sem id — tratá-lo como
// falha levaria o fallback a reenviar uma mensagem que pode ter saído.
func (r Remetente) classificar(resp *http.Response) dominio.ResultadoEnvio {
	var corpo resposta
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&corpo)
	motivo := r.mascarar(fmt.Sprintf("zapi %d: %s", resp.StatusCode, corpo.Error))
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return provedor.Classificar(resp.StatusCode, motivo, "")
	}
	if corpo.Error != "" || resp.StatusCode >= 300 {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: motivo}
	}
	id := corpo.MessageID
	if id == "" {
		id = corpo.ID
	}
	return provedor.Classificar(resp.StatusCode, "", id)
}

// mascarar tira do texto do Provedor os tokens que ele possa ecoar (T-84).
func (r Remetente) mascarar(s string) string {
	for _, segredo := range []string{r.Cred.Token, r.Cred.ClientToken} {
		if segredo != "" {
			s = strings.ReplaceAll(s, segredo, "***")
		}
	}
	return s
}
