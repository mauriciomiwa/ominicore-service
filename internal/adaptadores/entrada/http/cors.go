package http

import "net/http"

// cors autoriza só a origem do Console do Operador (REQ-34) e só nas rotas de
// configuração. O token vai no Authorization, não em cookie: sem
// Allow-Credentials. Origem vazia (sem Console) = nada é autorizado.
func (s Servidor) cors(prox http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if s.origemDoConsole(r) {
			w.Header().Set("Access-Control-Allow-Origin", s.ConsoleOrigem)
		}
		prox.ServeHTTP(w, r)
	})
}

// preflight responde o OPTIONS antes da autenticação: o navegador não manda o
// token no preflight (T-116).
func (s Servidor) preflight(metodos string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { s.responderPreflight(w, r, metodos) }
}

func (s Servidor) responderPreflight(w http.ResponseWriter, r *http.Request, metodos string) {
	w.Header().Add("Vary", "Origin")
	if !s.origemDoConsole(r) {
		problema(w, http.StatusForbidden, "origem não permitida", "")
		return
	}
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", s.ConsoleOrigem)
	h.Set("Access-Control-Allow-Methods", metodos)
	h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	h.Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}

func (s Servidor) origemDoConsole(r *http.Request) bool {
	return s.ConsoleOrigem != "" && r.Header.Get("Origin") == s.ConsoleOrigem
}
