package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func (a ambienteConfig) comOrigem(metodo, caminho, origem, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, caminho, nil)
	req.Header.Set("Origin", origem)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if metodo == http.MethodOptions {
		req.Header.Set("Access-Control-Request-Method", http.MethodPut)
		req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

func TestT116_PreflightDaOrigemDoConsoleEAceito(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	rec := a.comOrigem(http.MethodOptions, "/v1/config/whatsapp", origemConsole, "")
	h := rec.Header()
	if rec.Code != http.StatusNoContent || h.Get("Access-Control-Allow-Origin") != origemConsole ||
		h.Get("Access-Control-Allow-Methods") != "GET, PUT, DELETE" || h.Get("Access-Control-Allow-Headers") != "Authorization, Content-Type" {
		t.Fatalf("status = %d cabeçalhos = %v", rec.Code, h)
	}
	if h.Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("o token vai no Authorization: sem Allow-Credentials")
	}
	get := a.comOrigem(http.MethodGet, "/v1/config/whatsapp", origemConsole, a.operador(t, "ops-a"))
	if get.Code != http.StatusOK || get.Header().Get("Access-Control-Allow-Origin") != origemConsole {
		t.Fatalf("GET: status = %d cabeçalhos = %v", get.Code, get.Header())
	}
}

func TestT117_OutraOrigemNaoEAutorizada(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	respostas := map[string]*httptest.ResponseRecorder{
		"preflight de outra origem":     a.comOrigem(http.MethodOptions, "/v1/config/whatsapp", "https://exemplo.com", ""),
		"GET de outra origem":           a.comOrigem(http.MethodGet, "/v1/config/whatsapp", "https://exemplo.com", a.operador(t, "ops-a")),
		"envio com a origem do Console": a.comOrigem(http.MethodPost, "/v1/notificacoes", origemConsole, a.operador(t, "ops-a")),
		"preflight do envio":            a.comOrigem(http.MethodOptions, "/v1/notificacoes", origemConsole, ""),
	}
	for nome, rec := range respostas {
		if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
			t.Fatalf("%s: Access-Control-Allow-Origin = %q", nome, v)
		}
	}
}
