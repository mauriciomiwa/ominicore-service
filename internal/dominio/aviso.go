package dominio

// CategoriaFalha é a classificação genérica de uma Tentativa falhou, a única
// que a consulta expõe (REQ-14): o texto do motivo pode citar o contato.
type CategoriaFalha string

const (
	FalhaContatoInvalido       CategoriaFalha = "contato_invalido"
	FalhaRecusadoPeloProvedor  CategoriaFalha = "recusado_pelo_provedor"
	FalhaRetentativasEsgotadas CategoriaFalha = "retentativas_esgotadas"
	FalhaOptOut                CategoriaFalha = "opt_out"
)

// AvisoStatus é o Aviso de status de um Provedor já traduzido (REQ-12).
type AvisoStatus struct {
	IDProvedor string
	Status     StatusTentativa // entregue, lida ou falhou
	Motivo     string
	Categoria  CategoriaFalha // só para falhou
}

// ordem de avanço da Tentativa depois do aceite; falhou é final.
var ordemStatus = map[StatusTentativa]int{TentativaEnviada: 1, TentativaEntregue: 2, TentativaLida: 3}

// Avanca diz se o aviso faz a Tentativa avançar: enviada → entregue → lida;
// falhou só a partir de enviada (falha de entrega depois do aceite). Repetido,
// atrasado ou sobre Tentativa final é ignorado (idempotência).
func (s StatusTentativa) Avanca(novo StatusTentativa) bool {
	atual, ok := ordemStatus[s]
	if !ok {
		return false // pendente ainda não foi aceita; falhou é final
	}
	if novo == TentativaFalhou {
		return s == TentativaEnviada
	}
	proximo, ok := ordemStatus[novo]
	return ok && proximo > atual
}

// AtualizacaoAviso é o que gravar depois de um aviso que avança.
type AtualizacaoAviso struct {
	Tentativa   StatusTentativa
	Motivo      string
	Categoria   CategoriaFalha
	Notificacao StatusNotificacao // "" = não muda
	Proxima     *ProximaTentativa
}

// DecidirAviso aplica REQ-12 e REQ-13 a um aviso que avança: entrega marca a
// Notificação entregue; falha tardia é Falha definitiva e segue para o
// próximo Canal, salvo Notificação já entregue.
func (r NotificationRouter) DecidirAviso(t Tentativa, n Notificacao, a AvisoStatus) AtualizacaoAviso {
	switch a.Status {
	case TentativaEntregue, TentativaLida:
		return AtualizacaoAviso{Tentativa: a.Status, Notificacao: NotificacaoEntregue}
	}
	categoria := a.Categoria
	if categoria == "" {
		categoria = FalhaRecusadoPeloProvedor
	}
	if n.Status == NotificacaoEntregue {
		return AtualizacaoAviso{Tentativa: TentativaFalhou, Motivo: a.Motivo, Categoria: categoria}
	}
	d := r.Decidir(t, n, ResultadoEnvio{Tipo: FalhaDefinitiva, Motivo: a.Motivo, Categoria: categoria})
	return AtualizacaoAviso{Tentativa: TentativaFalhou, Motivo: d.Motivo, Categoria: d.Categoria, Notificacao: d.Notificacao, Proxima: d.Proxima}
}
