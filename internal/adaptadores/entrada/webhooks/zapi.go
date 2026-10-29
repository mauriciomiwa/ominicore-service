package webhooks

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"ominicore/internal/dominio"
)

// ZAPIAtual diz se o WhatsApp está ligado pela Z-API, com qual instância e qual
// segredo da URL dos avisos (REQ-37, ADR 0007).
type ZAPIAtual struct {
	Ligada     bool
	InstanceID string
	Segredo    string
}

// avisoZAPI cobre os dois avisos usados: "Ao enviar" (DeliveryCallback) e
// "Status" (MessageStatusCallback).
// https://developer.z-api.io/webhooks/on-message-send
// https://developer.z-api.io/webhooks/on-whatsapp-message-status-changes
type avisoZAPI struct {
	Type       string   `json:"type"`
	InstanceID string   `json:"instanceId"`
	MessageID  string   `json:"messageId"`
	Error      string   `json:"error"`
	Status     string   `json:"status"`
	IDs        []string `json:"ids"`
}

// limiteMotivoZAPI corta o texto do erro da Z-API guardado no motivo.
const limiteMotivoZAPI = 200

// avisosZAPI traduz o aviso da Z-API (REQ-38, REQ-39): RECEIVED → entregue,
// READ/PLAYED → lida; "Ao enviar" com error → falhou (número inexistente ou
// inválido = contato inválido; o resto = recusado pelo Provedor). SENT,
// READ_BY_ME, "Ao enviar" sem error e outros tipos não geram aviso.
func avisosZAPI(a avisoZAPI) []dominio.AvisoStatus {
	switch a.Type {
	case "MessageStatusCallback":
		status, ok := map[string]dominio.StatusTentativa{
			"RECEIVED": dominio.TentativaEntregue, "READ": dominio.TentativaLida, "PLAYED": dominio.TentativaLida,
		}[a.Status]
		if !ok {
			return nil
		}
		var avisos []dominio.AvisoStatus
		for _, id := range a.IDs {
			if id != "" {
				avisos = append(avisos, dominio.AvisoStatus{IDProvedor: id, Status: status})
			}
		}
		return avisos
	case "DeliveryCallback":
		if a.Error == "" || a.MessageID == "" {
			return nil
		}
		categoria := dominio.FalhaRecusadoPeloProvedor
		erro := strings.ToLower(a.Error)
		if strings.Contains(erro, "phone number does not exist") || strings.Contains(erro, "invalid phone number") {
			categoria = dominio.FalhaContatoInvalido
		}
		motivo := "zapi delivery: " + a.Error
		if len(motivo) > limiteMotivoZAPI {
			motivo = motivo[:limiteMotivoZAPI]
		}
		return []dominio.AvisoStatus{{IDProvedor: a.MessageID, Status: dominio.TentativaFalhou, Motivo: motivo, Categoria: categoria}}
	}
	return nil
}

// zapi recebe os avisos em /webhooks/zapi/{segredo}: só com a Z-API ativa, o
// segredo atual e a instância configurada; qualquer outro caso é 404, sem
// efeito nem distinção do motivo (T-122 a T-124). Aviso autêntico responde
// sempre 200 {"value": true}, como a Z-API espera (T-125).
func (w Webhooks) zapi(rw http.ResponseWriter, r *http.Request) {
	z := ZAPIAtual{}
	if w.ZAPI != nil {
		z = w.ZAPI()
	}
	segredo := r.PathValue("segredo")
	if !z.Ligada || z.Segredo == "" || subtle.ConstantTimeCompare([]byte(segredo), []byte(z.Segredo)) != 1 {
		http.NotFound(rw, r)
		return
	}
	corpo, err := io.ReadAll(io.LimitReader(r.Body, 64<<10))
	var aviso avisoZAPI
	if err != nil || json.Unmarshal(corpo, &aviso) != nil || aviso.InstanceID != z.InstanceID {
		http.NotFound(rw, r)
		return
	}
	for _, a := range avisosZAPI(aviso) {
		if _, err := w.Receptor.Receber(r.Context(), a); err != nil {
			w.Log.ErrorContext(r.Context(), "aviso de status falhou", "provedor", "zapi", "erro", err.Error())
			http.Error(rw, "erro interno", http.StatusInternalServerError)
			return
		}
	}
	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write([]byte(`{"value":true}`))
}
