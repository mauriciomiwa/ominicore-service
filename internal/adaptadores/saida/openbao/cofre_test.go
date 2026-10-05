package openbao_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/dominio"
)

// kvMutavel é um KV v2 falso que aceita gravar e apagar, e pode negar com 403
// as próximas leituras até um novo login (sessão expirada).
type kvMutavel struct {
	mu       sync.Mutex
	dados    map[string]map[string]any
	expirada bool
	logins   int
	token    string
}

func novoKVMutavel(t *testing.T, dados map[string]map[string]any) (*kvMutavel, string) {
	k := &kvMutavel{dados: dados, token: "sessao-1"}
	srv := httptest.NewServer(http.HandlerFunc(k.servir))
	t.Cleanup(srv.Close)
	return k, srv.URL
}

func (k *kvMutavel) servir(w http.ResponseWriter, r *http.Request) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if strings.HasSuffix(r.URL.Path, "/login") {
		k.logins++
		k.expirada = false
		_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]any{"client_token": "sessao-nova"}})
		return
	}
	if k.expirada {
		http.Error(w, `{"errors":["permission denied"]}`, http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		d, ok := k.dados[r.URL.Path]
		if !ok {
			http.Error(w, `{"errors":[]}`, http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": d, "metadata": map[string]any{}}})
	case http.MethodPut, http.MethodPost:
		var corpo struct {
			Data map[string]any `json:"data"`
		}
		_ = json.NewDecoder(r.Body).Decode(&corpo)
		k.dados[r.URL.Path] = corpo.Data
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"version": 2, "created_time": "2026-10-05T00:00:00Z"}})
	case http.MethodDelete:
		delete(k.dados, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (k *kvMutavel) ler(caminho string) map[string]any {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.dados[caminho]
}

func TestGravarELerAConfiguracaoDoWhatsApp(t *testing.T) {
	kv, addr := novoKVMutavel(t, map[string]map[string]any{})
	c, err := openbao.Abrir(context.Background(), openbao.Config{Endereco: addr, Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg, err := c.LerWhatsApp(context.Background()); err != nil || cfg.Ligado {
		t.Fatalf("sem caminho = desligado: %+v %v", cfg, err)
	}
	zapi, _ := dominio.NovaConfigWhatsApp("zapi", map[string]string{"instance_id": "inst1", "token": "tk", "client_token": "ct"})
	if err := c.GravarWhatsApp(context.Background(), zapi); err != nil {
		t.Fatal(err)
	}
	if d := kv.ler("/v1/kv/data/ominicore/whatsapp"); d["provedor"] != "zapi" || d["token"] != "tk" || d["client_token"] != "ct" {
		t.Fatalf("gravado = %v", d)
	}
	if lido, err := c.LerWhatsApp(context.Background()); err != nil || !lido.Igual(zapi) {
		t.Fatalf("lido = %+v %v", lido, err)
	}
	if err := c.ApagarWhatsApp(context.Background()); err != nil {
		t.Fatal(err)
	}
	if cfg, err := c.LerWhatsApp(context.Background()); err != nil || cfg.Ligado {
		t.Fatalf("depois de apagar = %+v %v", cfg, err)
	}
	if err := c.ApagarWhatsApp(context.Background()); err != nil {
		t.Fatalf("apagar de novo não é erro: %v", err)
	}
}

func TestSessaoExpiradaRefazOLoginUmaVez(t *testing.T) {
	kv, addr := novoKVMutavel(t, map[string]map[string]any{
		"/v1/kv/data/ominicore/whatsapp": {"provedor": "ultramsg", "instance_id": "i", "token": "tk"},
	})
	jwts := 0
	c, err := openbao.Abrir(context.Background(), openbao.Config{
		Endereco: addr, Role: "ominicore",
		RenovarJWT: func(context.Context) (string, error) { jwts++; return "jwt", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	kv.mu.Lock()
	kv.expirada = true
	kv.mu.Unlock()
	cfg, err := c.LerWhatsApp(context.Background())
	if err != nil || cfg.Provedor != dominio.WhatsAppUltraMsg {
		t.Fatalf("cfg = %+v err = %v", cfg, err)
	}
	if jwts != 2 || kv.logins != 2 {
		t.Fatalf("jwts = %d logins = %d, quero 2 e 2 (boot + renovação)", jwts, kv.logins)
	}
}

func TestT105_SegredosDaMetaNoCaminhoDoWhatsAppOuNoAntigo(t *testing.T) {
	meta := map[string]string{"provedor": "meta", "token": "t", "phone_number_id": "1"}
	segredos := map[string]string{"meta_app_secret": "s", "meta_verify_token": "v"}
	junto := func(a, b map[string]string) map[string]string {
		out := map[string]string{}
		for k, v := range a {
			out[k] = v
		}
		for k, v := range b {
			out[k] = v
		}
		return out
	}
	casos := []struct {
		nome               string
		whatsapp, webhooks map[string]string
		ok                 bool
	}{
		{"no caminho do whatsapp", junto(meta, segredos), map[string]string{}, true},
		{"no caminho antigo", meta, segredos, true},
		{"em nenhum dos dois", meta, map[string]string{}, false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			kv := cofreSemMeta()
			kv["/v1/kv/data/ominicore/whatsapp"] = c.whatsapp
			for k, v := range c.webhooks {
				kv["/v1/kv/data/ominicore/webhooks"][k] = v
			}
			seg, err := openbao.Carregar(context.Background(), openbao.Config{Endereco: kvFalso(t, kv), Token: "t"})
			if !c.ok {
				if !errors.Is(err, openbao.ErrOpenBao) || !strings.Contains(err.Error(), "meta_app_secret") {
					t.Fatalf("erro = %v, quero ErrOpenBao citando meta_app_secret", err)
				}
				return
			}
			if err != nil || seg.WhatsApp["meta_app_secret"] != "s" || seg.WhatsApp["meta_verify_token"] != "v" {
				t.Fatalf("segredos = %+v err = %v", seg.WhatsApp, err)
			}
		})
	}
}
