package dominio

import (
	"math/rand/v2"
	"slices"
	"time"
)

// TipoResultado classifica o que o Provedor respondeu a uma Tentativa.
type TipoResultado int

const (
	Aceito TipoResultado = iota
	FalhaTemporaria
	FalhaDefinitiva
)

// ResultadoEnvio é a resposta do Provedor a uma Tentativa.
type ResultadoEnvio struct {
	Tipo       TipoResultado
	Motivo     string
	IDProvedor string
	Categoria  CategoriaFalha // o adaptador marca contato_invalido quando sabe; vazio = recusado
}

// ProximaTentativa é a Tentativa que a decisão manda criar: Retentativa no
// mesmo Canal (Numero > 1) ou a primeira no próximo Canal da ordem.
type ProximaTentativa struct {
	Canal  Canal
	Numero int
	Espera time.Duration
}

// Decisao é o que fazer depois do envio: o status da Tentativa atual, o da
// Notificação e, se houver, a próxima Tentativa.
type Decisao struct {
	Tentativa   StatusTentativa
	Notificacao StatusNotificacao
	Motivo      string
	Categoria   CategoriaFalha
	Proxima     *ProximaTentativa
}

// esperas antes de cada Retentativa (RN-2): a 1ª após 5 s, a 2ª após 30 s, a
// 3ª após 2 min. Esgotadas, o Canal é tratado como falho.
var esperas = []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute}

// Jitter é a variação aplicada a cada espera, para cima ou para baixo.
const Jitter = 0.2

// NotificationRouter decide o destino da Notificação a partir do resultado de
// cada Tentativa (REQ-8, REQ-9). Sorteio devolve um valor em [0, 1); vazio,
// usa math/rand — os testes injetam um valor fixo.
type NotificationRouter struct {
	Sorteio func() float64
}

// Decidir aplica: aceito → enviada (REQ-4, REQ-7); Falha temporária com
// Retentativa sobrando → mesmo Canal após a espera (REQ-8); senão, próximo
// Canal da ordem com contato (REQ-9); sem Canal restante → falhou (REQ-5).
func (r NotificationRouter) Decidir(t Tentativa, n Notificacao, res ResultadoEnvio) Decisao {
	if res.Tipo == Aceito {
		return Decisao{Tentativa: TentativaEnviada, Notificacao: NotificacaoEnviada}
	}
	d := Decisao{Tentativa: TentativaFalhou, Notificacao: NotificacaoPendente, Motivo: res.Motivo}
	if res.Tipo == FalhaTemporaria && t.Numero <= len(esperas) {
		// Tentativa intermediária: sem categoria, o Canal ainda não falhou
		d.Proxima = &ProximaTentativa{Canal: t.Canal, Numero: t.Numero + 1, Espera: r.espera(t.Numero)}
		return d
	}
	d.Categoria = categoriaDe(res)
	return r.proximoCanal(t, n, d)
}

// proximoCanal completa a decisão de um Canal falho: próximo Canal da ordem
// com contato (REQ-9) ou fim da Notificação.
func (r NotificationRouter) proximoCanal(t Tentativa, n Notificacao, d Decisao) Decisao {
	atual := slices.Index(n.OrdemCanais, t.Canal)
	if proximo, ok := n.Destinatario.PrimeiroComContato(n.OrdemCanais, atual+1); ok && atual >= 0 {
		d.Proxima = &ProximaTentativa{Canal: proximo, Numero: 1}
		return d
	}
	d.Notificacao = NotificacaoFalhou
	return d
}

// espera devolve a espera antes da Retentativa que segue a Tentativa numero,
// com jitter de ±Jitter.
func (r NotificationRouter) espera(numero int) time.Duration {
	sorteio := r.Sorteio
	if sorteio == nil {
		sorteio = rand.Float64 //nolint:gosec // jitter não é segurança
	}
	fator := 1 - Jitter + 2*Jitter*sorteio()
	return time.Duration(float64(esperas[numero-1]) * fator)
}

func categoriaDe(res ResultadoEnvio) CategoriaFalha {
	switch {
	case res.Tipo == FalhaTemporaria:
		return FalhaRetentativasEsgotadas // só chega ao Canal falho depois de esgotar
	case res.Categoria != "":
		return res.Categoria
	default:
		return FalhaRecusadoPeloProvedor
	}
}
