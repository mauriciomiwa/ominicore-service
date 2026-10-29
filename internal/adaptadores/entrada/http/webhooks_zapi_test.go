package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ominicore/internal/adaptadores/entrada/webhooks"
	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/adaptadores/saida/zapi"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/memoria"
	"ominicore/internal/testes/zapifake"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

const urlPublicaWebhooks = "https://notify.flipapi.com.br"

var corpoZAPIFake = `{"provedor":"zapi","instance_id":"` + zapifake.InstanceID + `","token":"` + zapifake.Token + `","client_token":"` + zapifake.ClientToken + `"}`

// ambienteWebhooks sobe a API de configuração com o cadastro na Z-API falsa e
// a rota /webhooks/zapi lendo a configuração atual.
type ambienteWebhooks struct {
	ambienteConfig
	zapi     *zapifake.Fake
	segredos []string
}

func novoWebhooks(t *testing.T, inicial dominio.ConfigWhatsApp) *ambienteWebhooks {
	a := &ambienteWebhooks{zapi: zapifake.Novo(t)}
	a.ambienteConfig = ambienteConfig{realm: emissor.Novo(t), repo: memoria.Novo(), cofre: memoria.NovoCofre(inicial), log: &bytes.Buffer{}}
	a.atual = aplicacao.NovaConfigWhatsAppAtual(inicial, func(dominio.ConfigWhatsApp) portas.MessageSenderPort { return nil })
	log := slog.New(slog.NewTextHandler(a.log, nil))
	n := 0
	a.h = entradahttp.Servidor{
		Auth: keycloak.Novo(a.realm.URL, "ominicore", &emissor.Cache{}),
		Config: &aplicacao.ConfigurarUseCase{
			Cofre: a.cofre, Registro: a.repo, Atual: a.atual, Log: log,
			Cadastro: zapi.NovoWebhooks(a.zapi.URL), URLPublica: urlPublicaWebhooks,
			Agora: func() time.Time { return time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC) },
			Segredo: func() (string, error) {
				n++
				s := strings.Repeat("s", 40) + string(rune('a'+n)) // um segredo novo por gravação
				a.segredos = append(a.segredos, s)
				return s, nil
			},
		},
		Webhooks: webhooks.Webhooks{
			ZAPI: func() webhooks.ZAPIAtual {
				c := a.atual.Config()
				return webhooks.ZAPIAtual{Ligada: c.Ligado && c.Provedor == dominio.WhatsAppZAPI, InstanceID: c.Campos["instance_id"], Segredo: c.Campos[dominio.CampoWebhookSegredo]}
			},
			Receptor: aplicacao.ReceiveMessageUseCase{Repo: a.repo, Log: slog.New(slog.NewTextHandler(io.Discard, nil))},
			Log:      log,
		},
		ConsoleOrigem: origemConsole,
		ConsoleClient: clientConsole,
		Log:           log,
	}.Handler()
	return a
}

func (a *ambienteWebhooks) recadastrar(token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/config/whatsapp/webhooks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, req)
	return rec
}

func (a *ambienteWebhooks) avisoZAPI(segredo string) int {
	rec := httptest.NewRecorder()
	corpo := `{"type":"MessageStatusCallback","instanceId":"` + zapifake.InstanceID + `","status":"RECEIVED","ids":["X"]}`
	a.h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/zapi/"+segredo, strings.NewReader(corpo)))
	return rec.Code
}

func estadoWebhooks(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var v struct {
		Campos map[string]string `json:"campos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("resposta %d: %s", rec.Code, rec.Body)
	}
	return v.Campos[dominio.CampoWebhooksEstado]
}

func TestT131_GravarAZAPICadastraOsWebhooksComSegredoNovo(t *testing.T) {
	a := novoWebhooks(t, ultraMsg())
	op := a.operador(t, "ops-a")
	if rec := a.chamar(http.MethodPut, op, corpoZAPIFake); rec.Code != http.StatusOK || estadoWebhooks(t, rec) != dominio.WebhooksCadastrados {
		t.Fatalf("1ª gravação: %d %s", rec.Code, rec.Body)
	}
	primeiro := a.segredos[0]
	urls := a.zapi.Webhooks()
	quero := urlPublicaWebhooks + "/webhooks/zapi/" + primeiro
	if urls["update-webhook-delivery"] != quero || urls["update-webhook-message-status"] != quero {
		t.Fatalf("URLs cadastradas = %v, quero as duas com %q", urls, quero)
	}
	if code := a.avisoZAPI(primeiro); code != http.StatusOK {
		t.Fatalf("aviso com o segredo atual: %d", code)
	}
	if rec := a.chamar(http.MethodPut, op, corpoZAPIFake); rec.Code != http.StatusOK {
		t.Fatalf("2ª gravação: %d", rec.Code)
	}
	if a.segredos[1] == primeiro || a.zapi.Webhooks()["update-webhook-delivery"] != urlPublicaWebhooks+"/webhooks/zapi/"+a.segredos[1] {
		t.Fatalf("a regravação não trocou o segredo: %v", a.segredos)
	}
	if code := a.avisoZAPI(primeiro); code != http.StatusNotFound {
		t.Fatalf("aviso com o segredo anterior: %d, quero 404", code)
	}
	if rec := a.chamar(http.MethodGet, op, ""); estadoWebhooks(t, rec) != dominio.WebhooksCadastrados {
		t.Fatalf("consulta: %s", rec.Body)
	}
}

func TestT132_FalhaNoCadastroNaoDesfazAGravacao(t *testing.T) {
	a := novoWebhooks(t, ultraMsg())
	a.zapi.FalharCadastro(true)
	rec := a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpoZAPIFake)
	if rec.Code != http.StatusOK || estadoWebhooks(t, rec) != dominio.WebhooksFalhou {
		t.Fatalf("gravação: %d %s", rec.Code, rec.Body)
	}
	if cfg := a.atual.Config(); cfg.Provedor != dominio.WhatsAppZAPI {
		t.Fatalf("a Z-API devia estar ativa: %+v", cfg.Provedor)
	}
	if rec := a.chamar(http.MethodGet, a.operador(t, "ops-a"), ""); estadoWebhooks(t, rec) != dominio.WebhooksFalhou {
		t.Fatalf("consulta: %s", rec.Body)
	}
	log := a.log.String()
	if !strings.Contains(log, "webhooks da z-api não cadastrados") {
		t.Fatalf("o log não registra a falha: %s", log)
	}
	for _, segredo := range []string{a.segredos[0], zapifake.Token, zapifake.ClientToken} {
		if strings.Contains(log, segredo) {
			t.Fatalf("o log traz um segredo: %s", log)
		}
	}
}

func TestT133_ConsultaNaoMostraOSegredoDoWebhook(t *testing.T) {
	a := novoWebhooks(t, ultraMsg())
	a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpoZAPIFake)
	rec := a.chamar(http.MethodGet, a.operador(t, "ops-a"), "")
	if !strings.Contains(rec.Body.String(), `"webhook_segredo":true`) || strings.Contains(rec.Body.String(), a.segredos[0]) {
		t.Fatalf("consulta = %s", rec.Body)
	}
	if estadoWebhooks(t, rec) != dominio.WebhooksCadastrados {
		t.Fatalf("estado = %s", rec.Body)
	}
}

func TestT134_RecadastroComOSegredoAtual(t *testing.T) {
	a := novoWebhooks(t, ultraMsg())
	a.zapi.FalharCadastro(true)
	a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpoZAPIFake)
	a.zapi.FalharCadastro(false)
	rec := a.recadastrar(a.operador(t, "ops-a"))
	if rec.Code != http.StatusOK || estadoWebhooks(t, rec) != dominio.WebhooksCadastrados {
		t.Fatalf("recadastro: %d %s", rec.Code, rec.Body)
	}
	if len(a.segredos) != 1 || a.zapi.Webhooks()["update-webhook-message-status"] != urlPublicaWebhooks+"/webhooks/zapi/"+a.segredos[0] {
		t.Fatalf("o recadastro devia manter o segredo: segredos %v, urls %v", a.segredos, a.zapi.Webhooks())
	}
	if code := a.avisoZAPI(a.segredos[0]); code != http.StatusOK {
		t.Fatalf("o segredo atual devia continuar valendo: %d", code)
	}
	if ev := a.repo.Eventos(); ev[len(ev)-1].Acao != "webhooks" || ev[len(ev)-1].Por != "ops-a" {
		t.Fatalf("eventos = %+v", ev)
	}
}

func TestT135_RecadastroSemAZAPIAtivaERecusado(t *testing.T) {
	a := novoWebhooks(t, ultraMsg())
	rec := a.recadastrar(a.operador(t, "ops-a"))
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "Z-API") {
		t.Fatalf("recadastro com UltraMsg: %d %s", rec.Code, rec.Body)
	}
	if len(a.zapi.Webhooks()) != 0 || len(a.repo.Eventos()) != 0 {
		t.Fatalf("nada devia ser chamado: urls %v, eventos %v", a.zapi.Webhooks(), a.repo.Eventos())
	}
}

func TestCamposDoWebhookNoCorpoDoPUTSaoDescartados(t *testing.T) {
	a := novoWebhooks(t, ultraMsg())
	corpo := strings.TrimSuffix(corpoZAPIFake, "}") + `,"webhook_segredo":"meu-segredo","webhooks_estado":"cadastrados"}`
	a.zapi.FalharCadastro(true)
	a.chamar(http.MethodPut, a.operador(t, "ops-a"), corpo)
	cfg := a.atual.Config()
	if cfg.Campos[dominio.CampoWebhookSegredo] != a.segredos[0] || cfg.Campos[dominio.CampoWebhooksEstado] != dominio.WebhooksFalhou {
		t.Fatalf("campos do OminiCore vindos do corpo foram aceitos: %+v", cfg.Campos)
	}
}
