package aplicacao

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// DespacharUseCase envia a próxima Tentativa da fila e grava a decisão do
// NotificationRouter (REQ-4, REQ-5, REQ-7 a REQ-9).
type DespacharUseCase struct {
	Repo       portas.MessageRepositoryPort
	Remetentes map[dominio.Canal]portas.MessageSenderPort
	Desligados map[dominio.Canal]bool // o fallback pula esses Canais (REQ-20)
	Router     dominio.NotificationRouter
	Log        *slog.Logger
}

// ProcessarProxima devolve false quando a fila está vazia.
func (uc DespacharUseCase) ProcessarProxima(ctx context.Context) (bool, error) {
	res, ok, err := uc.Repo.ReservarTentativa(ctx)
	if err != nil || !ok {
		return false, err
	}
	res.Notificacao.Destinatario.Desligados = uc.Desligados
	resultado, consultouProvedor := uc.enviar(ctx, res)
	if consultouProvedor && resultado.Categoria == dominio.FalhaOptOut { // STOP no Provedor (REQ-16)
		canal := res.Tentativa.Canal
		if err := uc.Repo.RegistrarOptOut(ctx, canal, res.Notificacao.Destinatario.Contato(canal), "provedor:"+string(canal)); err != nil {
			return true, fmt.Errorf("registrar opt-out do provedor: %w", err)
		}
	}
	decisao := uc.Router.Decidir(res.Tentativa, res.Notificacao, resultado)
	if err := res.Concluir(ctx, decisao, resultado); err != nil {
		return true, fmt.Errorf("concluir tentativa %d: %w", res.Tentativa.ID, err)
	}
	uc.Log.InfoContext(ctx, "tentativa concluída",
		"notificacao_id", res.Notificacao.ID,
		"tentativa_id", res.Tentativa.ID,
		"canal", res.Tentativa.Canal,
		"status", decisao.Tentativa, // motivo fica só no banco: pode citar o contato (REQ-6)
		"notificacao_status", decisao.Notificacao,
		"proxima", descreverProxima(decisao.Proxima))
	return true, nil
}

// enviar devolve o resultado e se o Provedor chegou a ser chamado.
func (uc DespacharUseCase) enviar(ctx context.Context, res portas.TentativaReservada) (dominio.ResultadoEnvio, bool) {
	n, canal := res.Notificacao, res.Tentativa.Canal
	if n.Destinatario.OptOut[canal] { // Opt-out depois do pedido: não chama o Provedor (REQ-17)
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "opt-out", Categoria: dominio.FalhaOptOut}, false
	}
	remetente, ok := uc.Remetentes[canal]
	if !ok {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "canal sem provedor configurado: " + string(canal)}, false
	}
	tpl, err := uc.Repo.BuscarTemplate(ctx, n.TemplateID)
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: err.Error()}, false
	}
	conteudo, err := tpl.Preencher(canal, n.Dados)
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: err.Error()}, false
	}
	return remetente.Enviar(ctx, n.Destinatario.Contato(canal), conteudo), true
}

func descreverProxima(p *dominio.ProximaTentativa) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%s#%d em %s", p.Canal, p.Numero, p.Espera.Round(time.Second))
}
