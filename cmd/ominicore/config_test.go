package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/entrada/webhooks"
	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/logs"
	"ominicore/internal/portas"
	"ominicore/internal/testes/baofake"
	"ominicore/internal/testes/memoria"
	"ominicore/internal/testes/ultramsgfake"
	"ominicore/internal/testes/zapifake"
)

const caminhoWhatsApp = "/v1/kv/data/ominicore/whatsapp"

// processo é um api ou worker já no ar: o cofre aberto e a configuração em uso.
type processo struct {
	cofre *openbao.Cofre
	atual *aplicacao.ConfigWhatsAppAtual
	log   *bytes.Buffer
}

func subir(t *testing.T, kv *baofake.KV, cfg config) processo {
	t.Helper()
	cofre, err := openbao.Abrir(context.Background(), openbao.Config{Endereco: kv.URL, Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	inicial, err := cofre.LerWhatsApp(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p := processo{cofre: cofre, log: &bytes.Buffer{}}
	p.atual = aplicacao.NovaConfigWhatsAppAtual(inicial, func(c dominio.ConfigWhatsApp) portas.MessageSenderPort { return remetenteWhatsApp(cfg, c) })
	return p
}

func (p processo) reler() {
	aplicacao.ReleituraConfig{Cofre: p.cofre, Atual: p.atual, Log: logs.Novo(p.log, slog.LevelInfo)}.Reler(context.Background())
}

func TestT101_TrocaDoUltraMsgPelaZAPISemReinicio(t *testing.T) {
	um, za := ultramsgfake.Novo(t), zapifake.Novo(t)
	cfg := config{UltraMsgURL: um.URL, ZAPIURL: za.URL}
	kv := baofake.Novo(t, map[string]map[string]any{
		caminhoWhatsApp: {"provedor": "ultramsg", "instance_id": ultramsgfake.InstanceID, "token": ultramsgfake.Token},
	})
	api, worker := subir(t, kv, cfg), subir(t, kv, cfg)
	conf := aplicacao.ConfigurarUseCase{Cofre: api.cofre, Registro: memoria.Novo(), Atual: api.atual, Log: slog.New(slog.NewTextHandler(api.log, nil))}
	if _, err := conf.Gravar(context.Background(), "zapi", map[string]string{"instance_id": zapifake.InstanceID, "token": zapifake.Token}, "ops-a"); err != nil {
		t.Fatal(err)
	}
	if worker.atual.ProvedorWhatsApp() != dominio.WhatsAppUltraMsg {
		t.Fatal("o worker trocou antes da releitura")
	}
	worker.reler()
	res := worker.atual.Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{Corpo: "Olá Ana"})
	if res.Tipo != dominio.Aceito || len(za.Recebidas()) != 1 || len(um.Recebidas()) != 0 {
		t.Fatalf("res = %+v, z-api = %d, ultramsg = %d", res, len(za.Recebidas()), len(um.Recebidas()))
	}
	if !strings.Contains(worker.log.String(), "whatsapp reconfigurado: zapi") {
		t.Fatalf("log = %s", worker.log)
	}
}

func TestT102_ConfiguracaoInvalidaNaReleituraMantemAAnterior(t *testing.T) {
	za := zapifake.Novo(t)
	kv := baofake.Novo(t, map[string]map[string]any{
		caminhoWhatsApp: {"provedor": "zapi", "instance_id": zapifake.InstanceID, "token": zapifake.Token, "client_token": zapifake.ClientToken},
	})
	worker := subir(t, kv, config{ZAPIURL: za.URL})
	kv.Por(caminhoWhatsApp, map[string]any{"provedor": "zapi", "instance_id": "outra", "client_token": "ct-novo-secreto"})
	worker.reler()
	if c := worker.atual.Config(); c.Provedor != dominio.WhatsAppZAPI || c.Campos["token"] != zapifake.Token {
		t.Fatalf("em uso = %+v", c)
	}
	if res := worker.atual.Enviar(context.Background(), "+5511999998888", dominio.ConteudoCanal{Corpo: "Olá"}); res.Tipo != dominio.Aceito {
		t.Fatalf("res = %+v", res)
	}
	l := worker.log.String()
	if !strings.Contains(l, "mantida a anterior") || !strings.Contains(l, "token") {
		t.Fatalf("log = %s", l)
	}
	for _, segredo := range []string{zapifake.Token, zapifake.ClientToken, "ct-novo-secreto"} {
		if strings.Contains(l, segredo) {
			t.Fatalf("log com credencial: %s", l)
		}
	}
}

func TestT103_RotasDaMetaAcompanhamOProvedorAtivo(t *testing.T) {
	kv := baofake.Novo(t, map[string]map[string]any{
		caminhoWhatsApp: {"provedor": "zapi", "instance_id": "i", "token": "t"},
	})
	api := subir(t, kv, config{})
	mux := http.NewServeMux()
	webhooks.Webhooks{Meta: metaAtual(api.atual), Log: slog.New(slog.NewTextHandler(api.log, nil))}.Registrar(mux)
	verificar := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=vt-gravado&hub.challenge=77", nil))
		return rec
	}
	if rec := verificar(); rec.Code != http.StatusNotFound {
		t.Fatalf("com a Z-API: status = %d, quero 404", rec.Code)
	}
	meta, _ := dominio.NovaConfigWhatsApp("meta", map[string]string{"token": "t", "phone_number_id": "1", "meta_app_secret": "s", "meta_verify_token": "vt-gravado"})
	if err := api.cofre.GravarWhatsApp(context.Background(), meta); err != nil {
		t.Fatal(err)
	}
	if rec := verificar(); rec.Code != http.StatusNotFound {
		t.Fatalf("antes da releitura: status = %d, quero 404", rec.Code)
	}
	api.reler()
	if rec := verificar(); rec.Code != http.StatusOK || rec.Body.String() != "77" {
		t.Fatalf("com a Meta: status = %d corpo = %q", rec.Code, rec.Body)
	}
}
