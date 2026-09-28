// Package logs monta o slog JSON do OminiCore com redação de campos sensíveis:
// credencial, contato do Destinatário e conteúdo nunca saem em log (REQ-6).
package logs

import (
	"io"
	"log/slog"
	"strings"
)

// Redigido substitui o valor de um campo sensível.
const Redigido = "[redigido]"

var sensiveis = map[string]bool{
	"authorization": true, "token": true, "jwt": true, "api_key": true, "apikey": true,
	"secret": true, "senha": true, "password": true, "dsn": true,
	"email": true, "destinatario": true, "telefone": true, "contato": true,
	"dados": true, "conteudo": true, "assunto": true, "corpo": true, "motivo": true,
}

// Sensivel diz se o nome de campo é redigido.
func Sensivel(chave string) bool {
	return sensiveis[strings.ToLower(chave)]
}

// Novo devolve o logger JSON com redação.
func Novo(w io.Writer, nivel slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: nivel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if Sensivel(a.Key) {
				return slog.String(a.Key, Redigido)
			}
			return a
		},
	}))
}
