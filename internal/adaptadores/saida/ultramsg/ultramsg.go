// Package ultramsg implementa o Canal WhatsApp pelo UltraMsg, um gateway não
// oficial sobre o WhatsApp Web (RN-4, ADR 0004): envia o texto da versão
// WhatsApp do Template (REQ-10).
// https://docs.ultramsg.com/api/post/messages/chat
package ultramsg

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
	"unicode/utf8"

	"ominicore/internal/adaptadores/saida/provedor"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// URLPadrao é a API do UltraMsg; testes apontam para um UltraMsg falso.
const URLPadrao = "https://api.ultramsg.com"

// LimiteTexto é o tamanho máximo do body aceito pelo UltraMsg, em caracteres.
const LimiteTexto = 4096

// Credenciais vêm do OpenBao (`kv/<prefixo>/whatsapp` com provedor=ultramsg, REQ-6).
type Credenciais struct {
	InstanceID string
	Token      string
}

// Remetente envia WhatsApp pelo UltraMsg.
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

// Enviar manda o texto preenchido; sem texto ou acima do limite é Falha
// definitiva sem chamar o Provedor (T-76, T-77). O token vai no corpo, nunca
// na URL.
func (r Remetente) Enviar(ctx context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	if err := dominio.WhatsAppUltraMsg.Servir(c); err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: err.Error()}
	}
	if n := utf8.RuneCountInString(c.Corpo); n > LimiteTexto {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: fmt.Sprintf("texto whatsapp com %d caracteres (limite %d)", n, LimiteTexto)}
	}
	form := url.Values{"token": {r.Cred.Token}, "to": {para}, "body": {c.Corpo}}
	endpoint := strings.TrimRight(r.URL, "/") + "/" + url.PathEscape(r.Cred.InstanceID) + "/messages/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar requisição: " + err.Error()}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return provedor.Inacessivel("ultramsg")
	}
	defer func() { _ = resp.Body.Close() }()
	return classificar(resp)
}

// resposta tolera o formato não documentado: id pode vir como número ou texto,
// error como texto ou objeto e sent como "true"/"false" ou booleano.
type resposta struct {
	ID      json.RawMessage `json:"id"`
	Error   json.RawMessage `json:"error"`
	Sent    json.RawMessage `json:"sent"`
	Message string          `json:"message"`
}

// classificar: 429 e 5xx Falha temporária; campo error, sent falso ou outro
// 4xx Falha definitiva; 2xx sem error Aceito, mesmo sem id — tratá-lo como falha levaria
// o fallback a reenviar uma mensagem que pode ter saído.
func classificar(resp *http.Response) dominio.ResultadoEnvio {
	var r resposta
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&r)
	erro := texto(r.Error)
	if erro == "" && strings.EqualFold(strings.Trim(string(r.Sent), `"`), "false") {
		erro = "não enviada: " + r.Message
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return provedor.Classificar(resp.StatusCode, fmt.Sprintf("ultramsg %d: %s", resp.StatusCode, erro), "")
	}
	if erro != "" || resp.StatusCode >= 300 {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: fmt.Sprintf("ultramsg %d: %s", resp.StatusCode, erro)}
	}
	return provedor.Classificar(resp.StatusCode, "", texto(r.ID))
}

// texto devolve string JSON sem aspas e qualquer outro valor como veio;
// ausente, null, "" ou false viram vazio.
func texto(v json.RawMessage) string {
	var s string
	if json.Unmarshal(v, &s) == nil {
		return s
	}
	switch t := strings.TrimSpace(string(v)); t {
	case "", "null", "false":
		return ""
	default:
		return t
	}
}
