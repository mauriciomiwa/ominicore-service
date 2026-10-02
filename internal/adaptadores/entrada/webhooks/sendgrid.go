package webhooks

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"ominicore/internal/dominio"
)

// sendgrid: Event Webhook, lote JSON assinado com ECDSA/SHA-256 sobre
// timestamp + corpo cru.
// https://www.twilio.com/docs/sendgrid/for-developers/tracking-events/getting-started-event-webhook-security-features
func (w Webhooks) sendgrid(rw http.ResponseWriter, r *http.Request) {
	corpo, err := lerCorpo(r)
	if err != nil || !assinaturaSendGridValida(w.Cred.SendGridChavePublica, r, corpo) {
		w.recusar(rw, r, "sendgrid")
		return
	}
	var eventos []struct {
		Event       string `json:"event"`
		Type        string `json:"type"`
		SgMessageID string `json:"sg_message_id"`
		Reason      string `json:"reason"`
		Email       string `json:"email"`
	}
	if err := json.Unmarshal(corpo, &eventos); err != nil {
		http.Error(rw, "lote inválido", http.StatusBadRequest)
		return
	}
	var avisos []dominio.AvisoStatus
	var descadastros []string
	for _, e := range eventos {
		// sg_message_id = <X-Message-Id do envio>.<sufixo> (validar em hml)
		id, _, _ := strings.Cut(e.SgMessageID, ".")
		switch {
		case e.Event == "delivered":
			avisos = append(avisos, dominio.AvisoStatus{IDProvedor: id, Status: dominio.TentativaEntregue})
		case e.Event == "bounce" && e.Type != "blocked":
			avisos = append(avisos, dominio.AvisoStatus{
				IDProvedor: id, Status: dominio.TentativaFalhou,
				Motivo: "sendgrid bounce: " + e.Reason, Categoria: dominio.FalhaContatoInvalido,
			})
		case e.Event == "unsubscribe" || e.Event == "spamreport": // Opt-out (REQ-16)
			descadastros = append(descadastros, e.Email)
		case e.Event == "dropped":
			avisos = append(avisos, dominio.AvisoStatus{
				IDProvedor: id, Status: dominio.TentativaFalhou,
				Motivo: "sendgrid dropped: " + e.Reason, Categoria: dominio.FalhaRecusadoPeloProvedor,
			})
		}
		// processed, deferred e blocked (temporários) e engajamento: ignorados
	}
	if w.OptOut != nil {
		for _, email := range descadastros {
			if err := w.OptOut.RegistrarPeloProvedor(r.Context(), dominio.CanalEmail, email, "sendgrid"); err != nil {
				w.Log.ErrorContext(r.Context(), "opt-out do sendgrid falhou", "erro", err.Error())
				http.Error(rw, "erro interno", http.StatusInternalServerError)
				return
			}
		}
	}
	w.aplicar(rw, r, "sendgrid", avisos)
}

func assinaturaSendGridValida(chave *ecdsa.PublicKey, r *http.Request, corpo []byte) bool {
	sig, err := base64.StdEncoding.DecodeString(r.Header.Get("X-Twilio-Email-Event-Webhook-Signature"))
	ts := r.Header.Get("X-Twilio-Email-Event-Webhook-Timestamp")
	if err != nil || len(sig) == 0 || ts == "" || chave == nil {
		return false
	}
	h := sha256.Sum256(append([]byte(ts), corpo...))
	return ecdsa.VerifyASN1(chave, h[:], sig)
}
