// Package sendgrid implementa o Canal Email pela API v3 do SendGrid.
// https://www.twilio.com/docs/sendgrid/api-reference/mail-send/mail-send
package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ominicore/internal/adaptadores/saida/provedor"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// URLPadrao é a API pública do SendGrid; testes apontam para um fake.
const URLPadrao = "https://api.sendgrid.com"

// Remetente envia e-mails pelo SendGrid.
type Remetente struct {
	URL    string
	APIKey string // vem do OpenBao (REQ-6)
	De     string // e-mail remetente verificado no SendGrid
	HTTP   *http.Client
}

var _ portas.MessageSenderPort = Remetente{}

// Novo monta o Remetente com timeout de 10 s por Tentativa.
func Novo(url, apiKey, de string) Remetente {
	if url == "" {
		url = URLPadrao
	}
	return Remetente{URL: url, APIKey: apiKey, De: de, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

type endereco struct {
	Email string `json:"email"`
}

type corpoEnvio struct {
	Personalizations []struct {
		To []endereco `json:"to"`
	} `json:"personalizations"`
	From    endereco `json:"from"`
	Subject string   `json:"subject"`
	Content []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"content"`
}

func montar(de, para string, c dominio.ConteudoCanal) corpoEnvio {
	var b corpoEnvio
	b.Personalizations = make([]struct {
		To []endereco `json:"to"`
	}, 1)
	b.Personalizations[0].To = []endereco{{Email: para}}
	b.From = endereco{Email: de}
	b.Subject = c.Assunto
	b.Content = []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{{Type: "text/plain", Value: c.Corpo}}
	return b
}

// Enviar classifica a resposta pela regra comum (provedor.Classificar); erro
// de rede é Falha temporária.
func (r Remetente) Enviar(ctx context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	corpo, err := json.Marshal(montar(r.De, para, c))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar e-mail: " + err.Error()}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.URL+"/v3/mail/send", bytes.NewReader(corpo))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar requisição: " + err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return provedor.Inacessivel("sendgrid")
	}
	defer func() { _ = resp.Body.Close() }()
	return classificar(resp)
}

func classificar(resp *http.Response) dominio.ResultadoEnvio {
	motivo := ""
	if resp.StatusCode >= 300 {
		motivo = fmt.Sprintf("sendgrid %d: %s", resp.StatusCode, mensagemDeErro(resp.Body))
	}
	return provedor.Classificar(resp.StatusCode, motivo, resp.Header.Get("X-Message-Id"))
}

// mensagemDeErro extrai errors[0].message do corpo de erro do SendGrid.
func mensagemDeErro(r io.Reader) string {
	var e struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(io.LimitReader(r, 64<<10)).Decode(&e); err != nil || len(e.Errors) == 0 {
		return "sem detalhe"
	}
	return e.Errors[0].Message
}
