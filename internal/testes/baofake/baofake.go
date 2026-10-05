// Package baofake simula o KV v2 do OpenBao nos testes: lê, grava e apaga
// segredos por caminho (ex.: /v1/kv/data/ominicore/whatsapp).
package baofake

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// KV guarda os segredos por caminho da API.
type KV struct {
	URL   string
	mu    sync.Mutex
	dados map[string]map[string]any
}

// Novo sobe o KV com os segredos dados.
func Novo(t *testing.T, dados map[string]map[string]any) *KV {
	k := &KV{dados: dados}
	srv := httptest.NewServer(http.HandlerFunc(k.servir))
	t.Cleanup(srv.Close)
	k.URL = srv.URL
	return k
}

// Por troca um segredo por fora da API do OminiCore (edição manual).
func (k *KV) Por(caminho string, d map[string]any) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.dados[caminho] = d
}

// Ler devolve o segredo do caminho, ou nil.
func (k *KV) Ler(caminho string) map[string]any {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.dados[caminho]
}

func (k *KV) servir(w http.ResponseWriter, r *http.Request) {
	k.mu.Lock()
	defer k.mu.Unlock()
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
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"version": 1, "created_time": "2026-10-05T00:00:00Z"}})
	case http.MethodDelete:
		delete(k.dados, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}
}
