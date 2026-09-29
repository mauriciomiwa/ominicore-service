package webhooks

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // HMAC-SHA1 é o algoritmo que o Twilio usa na assinatura
	"encoding/base64"
	"net/http"
	"sort"

	"ominicore/internal/dominio"
)

// errosDeContato são os ErrorCode do Twilio de destino inalcançável.
var errosDeContato = map[string]bool{"30003": true, "30005": true, "30006": true}

// twilio: status callback do SMS, assinado com HMAC-SHA1 (Auth Token) sobre a
// URL completa + parâmetros POST em ordem alfabética.
// https://www.twilio.com/docs/usage/webhooks/webhooks-security
func (w Webhooks) twilio(rw http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(rw, r.Body, limiteCorpo)
	if err := r.ParseForm(); err != nil || !assinaturaTwilioValida(w.Cred, r) {
		w.recusar(rw, r, "twilio")
		return
	}
	id, status, codigo := r.PostForm.Get("MessageSid"), r.PostForm.Get("MessageStatus"), r.PostForm.Get("ErrorCode")
	var avisos []dominio.AvisoStatus
	switch status {
	case "delivered":
		avisos = append(avisos, dominio.AvisoStatus{IDProvedor: id, Status: dominio.TentativaEntregue})
	case "undelivered", "failed":
		cat := dominio.FalhaRecusadoPeloProvedor
		if errosDeContato[codigo] {
			cat = dominio.FalhaContatoInvalido
		}
		avisos = append(avisos, dominio.AvisoStatus{
			IDProvedor: id, Status: dominio.TentativaFalhou,
			Motivo: "twilio " + status + "/" + codigo, Categoria: cat,
		})
	}
	w.aplicar(rw, r, "twilio", avisos)
}

func assinaturaTwilioValida(c Credenciais, r *http.Request) bool {
	recebida, err := base64.StdEncoding.DecodeString(r.Header.Get("X-Twilio-Signature"))
	if err != nil || len(recebida) == 0 || c.TwilioAuthToken == "" || c.URLPublica == "" {
		return false
	}
	return hmac.Equal(recebida, AssinarTwilio(c.TwilioAuthToken, c.URLPublica+r.URL.RequestURI(), r.PostForm))
}

// AssinarTwilio calcula a assinatura que o Twilio manda (também usada pelos testes).
func AssinarTwilio(authToken, url string, params map[string][]string) []byte {
	chaves := make([]string, 0, len(params))
	for k := range params {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)
	m := hmac.New(sha1.New, []byte(authToken))
	m.Write([]byte(url))
	for _, k := range chaves {
		for _, v := range params[k] {
			m.Write([]byte(k + v))
		}
	}
	return m.Sum(nil)
}
