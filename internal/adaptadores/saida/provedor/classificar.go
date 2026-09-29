// Package provedor guarda a regra comum de todos os Provedores: como uma
// resposta HTTP vira aceite, Falha temporária ou Falha definitiva (RN-2).
package provedor

import (
	"net/http"

	"ominicore/internal/dominio"
)

// Classificar: 2xx aceito; 429 e 5xx Falha temporária; demais 4xx Falha
// definitiva (contato inválido, recusa, credencial errada).
func Classificar(status int, motivo, idProvedor string) dominio.ResultadoEnvio {
	switch {
	case status >= 200 && status < 300:
		return dominio.ResultadoEnvio{Tipo: dominio.Aceito, IDProvedor: idProvedor}
	case status == http.StatusTooManyRequests || status >= 500:
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaTemporaria, Motivo: motivo}
	default:
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: motivo}
	}
}

// Inacessivel é o resultado de erro de rede ou timeout: Falha temporária.
func Inacessivel(nome string) dominio.ResultadoEnvio {
	return dominio.ResultadoEnvio{Tipo: dominio.FalhaTemporaria, Motivo: nome + " inacessível"}
}
