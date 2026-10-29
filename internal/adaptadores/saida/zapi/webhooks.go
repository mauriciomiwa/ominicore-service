package zapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// Caminhos da API da Z-API que cadastram a URL de cada aviso: "Ao enviar"
// (DeliveryCallback) e "Status" (MessageStatusCallback). A doc mostra só o
// corpo {"value": url} e os headers; método e caminho conferidos com a
// instância real (tarefa 2.1 da change 13 — DOC/Padroes/fontes-externas.md).
var caminhosWebhook = []string{"update-webhook-delivery", "update-webhook-message-status"}

const metodoWebhook = http.MethodPut

// Webhooks cadastra na Z-API as URLs dos Avisos de status (REQ-40).
type Webhooks struct {
	URL  string
	HTTP *http.Client
}

var _ portas.CadastroWebhooksPort = Webhooks{}

// NovoWebhooks monta o cadastro com timeout de 10 s por chamada.
func NovoWebhooks(baseURL string) Webhooks {
	if baseURL == "" {
		baseURL = URLPadrao
	}
	return Webhooks{URL: baseURL, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Cadastrar põe a mesma URL nos dois avisos; o primeiro erro interrompe. O
// erro nunca carrega a URL cadastrada (segredo) nem os tokens (T-132).
func (w Webhooks) Cadastrar(ctx context.Context, cfg dominio.ConfigWhatsApp, urlAviso string) error {
	if !cfg.Ligado || cfg.Provedor != dominio.WhatsAppZAPI {
		return errors.New("cadastro de webhooks só vale para a Z-API")
	}
	cred := Credenciais{InstanceID: cfg.Campos["instance_id"], Token: cfg.Campos["token"], ClientToken: cfg.Campos["client_token"]}
	corpo, err := json.Marshal(map[string]string{"value": urlAviso})
	if err != nil {
		return fmt.Errorf("montar cadastro zapi: %w", err)
	}
	for _, caminho := range caminhosWebhook {
		if err := w.chamar(ctx, cred, caminho, corpo, urlAviso); err != nil {
			return err
		}
	}
	return nil
}

func (w Webhooks) chamar(ctx context.Context, cred Credenciais, caminho string, corpo []byte, urlAviso string) error {
	endpoint := strings.TrimRight(w.URL, "/") + "/instances/" + url.PathEscape(cred.InstanceID) +
		"/token/" + url.PathEscape(cred.Token) + "/" + caminho
	req, err := http.NewRequestWithContext(ctx, metodoWebhook, endpoint, bytes.NewReader(corpo))
	if err != nil {
		return fmt.Errorf("zapi %s: montar requisição", caminho)
	}
	req.Header.Set("Content-Type", "application/json")
	if cred.ClientToken != "" {
		req.Header.Set("Client-Token", cred.ClientToken)
	}
	resp, err := w.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("zapi %s: inacessível", caminho) // o erro do client cita a URL com o token
	}
	defer func() { _ = resp.Body.Close() }()
	var r struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&r)
	if resp.StatusCode >= 300 || r.Error != "" {
		motivo := strings.ReplaceAll(Remetente{Cred: cred}.mascarar(r.Error), urlAviso, "***")
		return fmt.Errorf("zapi %s: %d %s", caminho, resp.StatusCode, motivo)
	}
	return nil
}
