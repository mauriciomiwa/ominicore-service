package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ominicore/internal/dominio"
)

// verificarMeta responde a verificação do webhook: hub.challenge quando o
// verify token confere (T-40).
// https://developers.facebook.com/docs/graph-api/webhooks/getting-started
func (w Webhooks) verificarMeta(rw http.ResponseWriter, r *http.Request, m MetaAtual) {
	q := r.URL.Query()
	token := q.Get("hub.verify_token")
	if q.Get("hub.mode") != "subscribe" || m.VerifyToken == "" ||
		subtle.ConstantTimeCompare([]byte(token), []byte(m.VerifyToken)) != 1 {
		http.Error(rw, "verificação recusada", http.StatusForbidden)
		return
	}
	desafio, err := strconv.ParseUint(q.Get("hub.challenge"), 10, 64)
	if err != nil { // a Meta manda um inteiro; nada além dele é refletido
		http.Error(rw, "desafio inválido", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	_, _ = rw.Write([]byte(strconv.FormatUint(desafio, 10)))
}

// whatsapp: statuses da Cloud API, corpo assinado com HMAC-SHA256 (App Secret)
// em X-Hub-Signature-256: sha256=<hex>.
func (w Webhooks) whatsapp(rw http.ResponseWriter, r *http.Request, m MetaAtual) {
	corpo, err := lerCorpo(r)
	if err != nil || !assinaturaMetaValida(m.AppSecret, r.Header.Get("X-Hub-Signature-256"), corpo) {
		w.recusar(rw, r, "whatsapp")
		return
	}
	var aviso struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Statuses []struct {
						ID     string `json:"id"`
						Status string `json:"status"`
						Errors []struct {
							Code int `json:"code"`
						} `json:"errors"`
					} `json:"statuses"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(corpo, &aviso); err != nil {
		http.Error(rw, "aviso inválido", http.StatusBadRequest)
		return
	}
	var avisos []dominio.AvisoStatus
	for _, e := range aviso.Entry {
		for _, c := range e.Changes {
			for _, s := range c.Value.Statuses {
				if a, ok := avisoMeta(s.ID, s.Status, s.Errors); ok {
					avisos = append(avisos, a)
				}
			}
		}
	}
	w.aplicar(rw, r, "whatsapp", avisos)
}

func avisoMeta(id, status string, erros []struct {
	Code int `json:"code"`
},
) (dominio.AvisoStatus, bool) {
	switch status {
	case "delivered":
		return dominio.AvisoStatus{IDProvedor: id, Status: dominio.TentativaEntregue}, true
	case "read":
		return dominio.AvisoStatus{IDProvedor: id, Status: dominio.TentativaLida}, true
	case "failed":
		codigo, cat := 0, dominio.FalhaRecusadoPeloProvedor
		if len(erros) > 0 {
			codigo = erros[0].Code
		}
		if codigo == 131026 {
			cat = dominio.FalhaContatoInvalido
		}
		return dominio.AvisoStatus{
			IDProvedor: id, Status: dominio.TentativaFalhou,
			Motivo: "whatsapp failed/" + strconv.Itoa(codigo), Categoria: cat,
		}, true
	}
	return dominio.AvisoStatus{}, false // sent: já registrado no aceite
}

func assinaturaMetaValida(segredo, cabecalho string, corpo []byte) bool {
	hexa, ok := strings.CutPrefix(cabecalho, "sha256=")
	recebida, err := hex.DecodeString(hexa)
	if !ok || err != nil || segredo == "" {
		return false
	}
	m := hmac.New(sha256.New, []byte(segredo))
	m.Write(corpo)
	return hmac.Equal(recebida, m.Sum(nil))
}
