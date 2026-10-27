// Package baofake simula o KV v2 do OpenBao nos testes: lê, grava e apaga
// segredos por caminho (ex.: /v1/kv/data/ominicore/whatsapp). Como o real, o
// DELETE só marca a versão como apagada (soft delete): a leitura seguinte
// responde 404 com `data: null` e a `deletion_time` preenchida.
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
	URL      string
	mu       sync.Mutex
	dados    map[string]map[string]any
	apagados map[string]bool
	versoes  map[string]int
}

// Novo sobe o KV com os segredos dados.
func Novo(t *testing.T, dados map[string]map[string]any) *KV {
	k := &KV{dados: dados, apagados: map[string]bool{}, versoes: map[string]int{}}
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

// Ler devolve o segredo do caminho, ou nil se não existe ou está apagado.
func (k *KV) Ler(caminho string) map[string]any {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.apagados[caminho] {
		return nil
	}
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
		if k.apagados[r.URL.Path] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": nil, "metadata": map[string]any{
				"version": k.versoes[r.URL.Path], "deletion_time": "2026-10-27T12:00:00Z", "destroyed": false,
			}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": d, "metadata": map[string]any{}}})
	case http.MethodPut, http.MethodPost:
		var corpo struct {
			Data map[string]any `json:"data"`
		}
		_ = json.NewDecoder(r.Body).Decode(&corpo)
		k.dados[r.URL.Path] = corpo.Data
		k.apagados[r.URL.Path] = false
		k.versoes[r.URL.Path]++
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"version": k.versoes[r.URL.Path], "created_time": "2026-10-05T00:00:00Z"}})
	case http.MethodDelete:
		if _, ok := k.dados[r.URL.Path]; ok {
			k.apagados[r.URL.Path] = true
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
