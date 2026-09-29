// Package whatsapp implementa o Canal WhatsApp pela Cloud API da Meta, só por
// Modelo aprovado (REQ-10, RN-4).
// https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages
package whatsapp

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

// URLPadrao é a Graph API; testes apontam para uma Meta falsa.
const URLPadrao = "https://graph.facebook.com"

// VersaoGraph é a versão da Graph API conferida em 2026-09-29 (v26.0, lançada
// em 2026-07-29, sem data de desativação). Revisar antes da desativação
// (https://developers.facebook.com/docs/graph-api/changelog/versions).
const VersaoGraph = "v26.0"

// Credenciais vêm do OpenBao (`kv/<prefixo>/whatsapp`, REQ-6).
type Credenciais struct {
	Token         string // token permanente de System User
	PhoneNumberID string
}

// Remetente envia WhatsApp pela Cloud API.
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

type parametro struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type componente struct {
	Type       string      `json:"type"`
	Parameters []parametro `json:"parameters"`
}

type mensagem struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Template         struct {
		Name     string `json:"name"`
		Language struct {
			Code string `json:"code"`
		} `json:"language"`
		Components []componente `json:"components,omitempty"`
	} `json:"template"`
}

func montar(para string, c dominio.ConteudoCanal) mensagem {
	m := mensagem{MessagingProduct: "whatsapp", To: para, Type: "template"}
	m.Template.Name = c.ModeloAprovado
	m.Template.Language.Code = c.Idioma
	if len(c.Parametros) > 0 {
		ps := make([]parametro, len(c.Parametros))
		for i, v := range c.Parametros {
			ps[i] = parametro{Type: "text", Text: v}
		}
		m.Template.Components = []componente{{Type: "body", Parameters: ps}}
	}
	return m
}

// Enviar manda o Modelo aprovado com os parâmetros na ordem.
func (r Remetente) Enviar(ctx context.Context, para string, c dominio.ConteudoCanal) dominio.ResultadoEnvio {
	if c.ModeloAprovado == "" {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "versão whatsapp sem modelo aprovado"}
	}
	corpo, err := json.Marshal(montar(para, c))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar mensagem: " + err.Error()}
	}
	endpoint := fmt.Sprintf("%s/%s/%s/messages", r.URL, VersaoGraph, r.Cred.PhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(corpo))
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "montar requisição: " + err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+r.Cred.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return provedor.Inacessivel("whatsapp")
	}
	defer func() { _ = resp.Body.Close() }()
	return classificar(resp)
}

type resposta struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Códigos da Meta que decidem pelo conteúdo, não pelo status HTTP
// (https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes).
var (
	definitivos = map[int]bool{131026: true, 132001: true} // número sem WhatsApp; Modelo não aprovado
	temporarios = map[int]bool{130429: true, 131000: true} // limite de taxa; erro desconhecido
)

func classificar(resp *http.Response) dominio.ResultadoEnvio {
	var r resposta
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&r)
	if resp.StatusCode < 300 {
		id := ""
		if len(r.Messages) > 0 {
			id = r.Messages[0].ID
		}
		return provedor.Classificar(resp.StatusCode, "", id)
	}
	motivo := fmt.Sprintf("whatsapp %d/%d: %s", resp.StatusCode, r.Error.Code, r.Error.Message)
	switch {
	case definitivos[r.Error.Code]:
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: motivo}
	case temporarios[r.Error.Code]:
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaTemporaria, Motivo: motivo}
	}
	return provedor.Classificar(resp.StatusCode, motivo, "")
}
