package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/memoria"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

// ambienteConfig sobe a API com a configuração do WhatsApp em memória.
type ambienteConfig struct {
	realm *emissor.Emissor
	repo  *memoria.Repositorio
	cofre *memoria.Cofre
	atual *aplicacao.ConfigWhatsAppAtual
	log   *bytes.Buffer
	h     http.Handler
}

func novoConfig(t *testing.T, inicial dominio.ConfigWhatsApp) ambienteConfig {
	a := ambienteConfig{realm: emissor.Novo(t), repo: memoria.Novo(), cofre: memoria.NovoCofre(inicial), log: &bytes.Buffer{}}
	a.atual = aplicacao.NovaConfigWhatsAppAtual(inicial, func(dominio.ConfigWhatsApp) portas.MessageSenderPort { return nil })
	log := slog.New(slog.NewTextHandler(a.log, nil))
	a.h = entradahttp.Servidor{
		Auth:          keycloak.Novo(a.realm.URL, "ominicore", &emissor.Cache{}),
		Enviar:        aplicacao.SendMessageUseCase{Repo: a.repo, Estado: a.atual},
		Config:        &aplicacao.ConfigurarUseCase{Cofre: a.cofre, Registro: a.repo, Atual: a.atual, Log: log},
		ConsoleOrigem: origemConsole,
		ConsoleClient: clientConsole,
		Log:           log,
	}.Handler()
	return a
}

// operador emite o token de um client com o papel de configuração.
func (a ambienteConfig) operador(t *testing.T, azp string) string {
	return a.realm.Token(t, jwt.MapClaims{
		"iss": a.realm.URL, "aud": []string{"ominicore"}, "azp": azp,
		"realm_access": map[string]any{"roles": []string{entradahttp.PapelOperador}},
		"exp":          time.Now().Add(5 * time.Minute).Unix(),
	})
}

func (a ambienteConfig) chamar(metodo, token, corpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, "/v1/config/whatsapp", strings.NewReader(corpo))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

func ultraMsg() dominio.ConfigWhatsApp {
	c, _ := dominio.NovaConfigWhatsApp("ultramsg", map[string]string{"instance_id": "instance1", "token": "tk-ultra"})
	return c
}

const corpoZAPI = `{"provedor":"zapi","instance_id":"inst1","token":"tk-zapi-secreto","client_token":"ct-zapi-secreto"}`

func TestT92_OperadorGravaAZAPI(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	rec := a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpoZAPI)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	cfg, _ := a.cofre.LerWhatsApp(t.Context())
	if cfg.Provedor != dominio.WhatsAppZAPI || cfg.Campos["token"] != "tk-zapi-secreto" || cfg.Campos["client_token"] != "ct-zapi-secreto" {
		t.Fatalf("cofre = %+v", cfg)
	}
	consulta := a.chamar(http.MethodGet, a.operador(t, "ops-a"), "")
	if !strings.Contains(consulta.Body.String(), `"provedor":"zapi"`) || a.atual.ProvedorWhatsApp() != dominio.WhatsAppZAPI {
		t.Fatalf("consulta = %s, em uso = %s", consulta.Body, a.atual.ProvedorWhatsApp())
	}
}

func TestT93_CampoExigidoFaltandoERecusado(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	rec := a.chamar(http.MethodPut, a.operador(t, "ops-a"), `{"provedor":"ultramsg","instance_id":"instance2"}`)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "token") {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if cfg, _ := a.cofre.LerWhatsApp(t.Context()); !cfg.Igual(ultraMsg()) || len(a.repo.Eventos()) != 0 {
		t.Fatalf("cofre mudou: %+v", cfg)
	}
}

func TestT94_ProvedorDesconhecidoERecusado(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	rec := a.chamar(http.MethodPut, a.operador(t, "ops-a"), `{"provedor":"twilio","instance_id":"i","token":"t"}`)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "twilio") {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if cfg, _ := a.cofre.LerWhatsApp(t.Context()); !cfg.Igual(ultraMsg()) {
		t.Fatalf("cofre mudou: %+v", cfg)
	}
}

func TestT95_MetaExigeEGravaOsSegredosDoWebhook(t *testing.T) {
	a := novoConfig(t, dominio.ConfigWhatsApp{})
	sem := a.chamar(http.MethodPut, a.operador(t, "ops-a"), `{"provedor":"meta","token":"t","phone_number_id":"1","meta_verify_token":"v"}`)
	if sem.Code != http.StatusUnprocessableEntity || !strings.Contains(sem.Body.String(), "meta_app_secret") {
		t.Fatalf("sem App Secret: status = %d: %s", sem.Code, sem.Body)
	}
	rec := a.chamar(http.MethodPut, a.operador(t, "ops-a"), `{"provedor":"meta","token":"t","phone_number_id":"1","meta_app_secret":"s","meta_verify_token":"v"}`)
	cfg, _ := a.cofre.LerWhatsApp(t.Context())
	if rec.Code != http.StatusOK || cfg.Campos["meta_app_secret"] != "s" || cfg.Campos["meta_verify_token"] != "v" || cfg.Campos["phone_number_id"] != "1" {
		t.Fatalf("status = %d, cofre = %+v", rec.Code, cfg)
	}
}

func TestT96_ConsultaMostraOProvedorSemOsSegredos(t *testing.T) {
	zapi, _ := dominio.NovaConfigWhatsApp("zapi", map[string]string{"instance_id": "inst1", "token": "tk-zapi-secreto", "client_token": "ct-zapi-secreto"})
	a := novoConfig(t, zapi)
	rec := a.chamar(http.MethodGet, a.operador(t, "ops-a"), "")
	var v struct {
		Ligado   bool              `json:"ligado"`
		Provedor string            `json:"provedor"`
		Campos   map[string]string `json:"campos"`
		Segredos map[string]bool   `json:"segredos"`
	}
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &v) != nil {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if !v.Ligado || v.Provedor != "zapi" || v.Campos["instance_id"] != "inst1" || !v.Segredos["token"] || !v.Segredos["client_token"] || len(v.Campos) != 1 {
		t.Fatalf("consulta = %+v", v)
	}
	for _, segredo := range []string{"tk-zapi", "ct-zapi", "secreto"} {
		if strings.Contains(rec.Body.String(), segredo) {
			t.Fatalf("consulta expõe segredo: %s", rec.Body)
		}
	}
}

func TestT97_ConsultaComOCanalDesligado(t *testing.T) {
	a := novoConfig(t, dominio.ConfigWhatsApp{})
	rec := a.chamar(http.MethodGet, a.operador(t, "ops-a"), "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"ligado":false}` {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
}

func TestT98_OperadorDesligaOWhatsApp(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	if rec := a.chamar(http.MethodDelete, a.operador(t, "ops-a"), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if cfg, _ := a.cofre.LerWhatsApp(t.Context()); cfg.Ligado {
		t.Fatalf("cofre = %+v", cfg)
	}
	if rec := a.chamar(http.MethodGet, a.operador(t, "ops-a"), ""); !strings.Contains(rec.Body.String(), `"ligado":false`) {
		t.Fatalf("consulta = %s", rec.Body)
	}
	pedido := httptest.NewRequest(http.MethodPost, "/v1/notificacoes", strings.NewReader(
		`{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["whatsapp"]}`))
	pedido.Header.Set("Authorization", "Bearer "+a.realm.Valido(t, "sistema-a"))
	pedido.Header.Set("Idempotency-Key", "k-98")
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, pedido)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "canal whatsapp indisponível") {
		t.Fatalf("pedido só WhatsApp: status = %d: %s", rec.Code, rec.Body)
	}
	if rec := a.chamar(http.MethodDelete, a.operador(t, "ops-a"), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("desligar de novo: status = %d", rec.Code)
	}
}

func TestT99_SistemaDeOrigemSemOPapelERecusado(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	token := a.realm.Valido(t, "sistema-a")
	for _, c := range []struct{ metodo, corpo string }{{http.MethodGet, ""}, {http.MethodPut, corpoZAPI}, {http.MethodDelete, ""}} {
		if rec := a.chamar(c.metodo, token, c.corpo); rec.Code != http.StatusForbidden {
			t.Fatalf("%s: status = %d, quero 403", c.metodo, rec.Code)
		}
	}
	if cfg, _ := a.cofre.LerWhatsApp(t.Context()); !cfg.Igual(ultraMsg()) || len(a.repo.Eventos()) != 0 {
		t.Fatalf("mudou: cofre = %+v eventos = %v", cfg, a.repo.Eventos())
	}
}

func TestT100_SemTokenAConfiguracaoERecusada(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	for _, token := range []string{"", "nao-e-um-jwt"} {
		for _, metodo := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			if rec := a.chamar(metodo, token, corpoZAPI); rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s com token %q: status = %d, quero 401", metodo, token, rec.Code)
			}
		}
	}
	if cfg, _ := a.cofre.LerWhatsApp(t.Context()); !cfg.Igual(ultraMsg()) {
		t.Fatalf("cofre mudou: %+v", cfg)
	}
}

func TestT104_TrocaFicaRegistradaSemSegredo(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpoZAPI)
	a.chamar(http.MethodDelete, a.operador(t, "ops-a"), "")
	ev := a.repo.Eventos()
	if len(ev) != 2 || ev[0].Acao != "gravada" || ev[0].Provedor != dominio.WhatsAppZAPI || ev[0].Por != "ops-a" ||
		ev[1].Acao != "apagada" || ev[1].Provedor != "" || ev[1].Por != "ops-a" || ev[0].Canal != dominio.CanalWhatsApp {
		t.Fatalf("eventos = %+v", ev)
	}
	if strings.Contains(a.log.String(), "secreto") || !strings.Contains(a.log.String(), "ops-a") {
		t.Fatalf("log = %s", a.log)
	}
}

func TestEventoQueFalhaViraErroComOLogComoRegistro(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	a.repo.FalharEventoConfig = errors.New("banco fora")
	rec := a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpoZAPI)
	if rec.Code != http.StatusInternalServerError || !strings.Contains(a.log.String(), "configuração trocada sem registro") {
		t.Fatalf("status = %d log = %s", rec.Code, a.log)
	}
	if cfg, _ := a.cofre.LerWhatsApp(t.Context()); cfg.Provedor != dominio.WhatsAppZAPI {
		t.Fatalf("a troca valeu e não deve ser desfeita: %+v", cfg)
	}
}

func TestT120_TrocaPeloConsoleRegistraAPessoa(t *testing.T) {
	a := novoConfig(t, ultraMsg())
	rec := a.chamar(http.MethodPut, a.realm.DeUsuario(t, clientConsole, "ana", entradahttp.PapelOperador), corpoZAPI)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if ev := a.repo.Eventos(); len(ev) != 1 || ev[0].Por != "ana" {
		t.Fatalf("eventos = %+v, quero o Operador ana", ev)
	}
}
