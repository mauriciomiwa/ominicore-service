package dominio

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
}

// Decisao é o que fazer com a Tentativa e a Notificação depois do envio.
type Decisao struct {
	Tentativa   StatusTentativa
	Notificacao StatusNotificacao
	Motivo      string
}

// NotificationRouter decide o destino da Notificação a partir do resultado de
// cada Tentativa. Nesta fatia (01) toda falha encerra a Notificação; retry de
// Falha temporária e fallback para o próximo Canal chegam na fatia 02 (RN-2).
type NotificationRouter struct{}

// Decidir aplica REQ-4 (aceito → enviada) e REQ-5 (falha → falhou).
func (NotificationRouter) Decidir(r ResultadoEnvio) Decisao {
	if r.Tipo == Aceito {
		return Decisao{Tentativa: TentativaEnviada, Notificacao: NotificacaoEnviada}
	}
	return Decisao{Tentativa: TentativaFalhou, Notificacao: NotificacaoFalhou, Motivo: r.Motivo}
}
