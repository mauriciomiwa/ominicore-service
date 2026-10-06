package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
)

// PapelOperador dá acesso à configuração dos Canais (RN-11, REQ-25).
const PapelOperador = "ominicore-config"

type chavePapeis struct{}

// exigirPapel recusa com 403, sem efeito, quem não tem o papel (T-99).
func exigirPapel(papel string, prox http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		papeis, _ := r.Context().Value(chavePapeis{}).([]string)
		if !slices.Contains(papeis, papel) {
			problema(w, http.StatusForbidden, "sem permissão", "requer o papel "+papel)
			return
		}
		prox.ServeHTTP(w, r)
	})
}

// configJSON é a visão sem segredos (REQ-23).
type configJSON struct {
	Ligado   bool              `json:"ligado"`
	Provedor string            `json:"provedor,omitempty"`
	Campos   map[string]string `json:"campos,omitempty"`
	Segredos map[string]bool   `json:"segredos,omitempty"`
}

func visao(c dominio.ConfigWhatsApp) configJSON {
	if !c.Ligado {
		return configJSON{}
	}
	return configJSON{Ligado: true, Provedor: string(c.Provedor), Campos: c.Publicos(), Segredos: c.Segredos()}
}

func (s Servidor) consultarConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.Config.Consultar(r.Context())
	if err != nil {
		s.Log.ErrorContext(r.Context(), "consultar configuração", "erro", err.Error())
		problema(w, http.StatusServiceUnavailable, "cofre indisponível", "não foi possível ler a configuração")
		return
	}
	escreverJSON(w, http.StatusOK, visao(cfg))
}

func (s Servidor) gravarConfig(w http.ResponseWriter, r *http.Request) {
	var corpo map[string]string
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&corpo); err != nil {
		problema(w, http.StatusBadRequest, "corpo inválido", "esperado um objeto JSON com campos de texto")
		return
	}
	cfg, err := s.Config.Gravar(r.Context(), corpo["provedor"], corpo, r.Context().Value(chaveQuem{}).(string))
	if s.falhaDeConfig(w, r, err) {
		return
	}
	escreverJSON(w, http.StatusOK, visao(cfg))
}

func (s Servidor) apagarConfig(w http.ResponseWriter, r *http.Request) {
	if s.falhaDeConfig(w, r, s.Config.Apagar(r.Context(), r.Context().Value(chaveQuem{}).(string))) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// falhaDeConfig responde o erro da troca; false quando não houve erro.
func (s Servidor) falhaDeConfig(w http.ResponseWriter, r *http.Request, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, dominio.ErrProvedorWhatsAppDesconhecido), errors.Is(err, dominio.ErrConfigIncompleta):
		problema(w, http.StatusUnprocessableEntity, "configuração recusada", err.Error())
	case errors.Is(err, aplicacao.ErrTrocaSemRegistro):
		problema(w, http.StatusInternalServerError, "gravada sem registro", "a troca valeu, mas o registro falhou; veja o log")
	default:
		s.Log.ErrorContext(r.Context(), "trocar configuração", "erro", err.Error())
		problema(w, http.StatusServiceUnavailable, "cofre indisponível", "a configuração não foi trocada")
	}
	return true
}
