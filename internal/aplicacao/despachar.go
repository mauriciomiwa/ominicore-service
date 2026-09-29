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
	Router     dominio.NotificationRouter
	Log        *slog.Logger
}

// ProcessarProxima devolve false quando a fila está vazia.
func (uc DespacharUseCase) ProcessarProxima(ctx context.Context) (bool, error) {
	res, ok, err := uc.Repo.ReservarTentativa(ctx)
	if err != nil || !ok {
		return false, err
	}
	resultado := uc.enviar(ctx, res)
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

func (uc DespacharUseCase) enviar(ctx context.Context, res portas.TentativaReservada) dominio.ResultadoEnvio {
	n, canal := res.Notificacao, res.Tentativa.Canal
	remetente, ok := uc.Remetentes[canal]
	if !ok {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: "canal sem provedor configurado: " + string(canal)}
	}
	tpl, err := uc.Repo.BuscarTemplate(ctx, n.TemplateID)
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: err.Error()}
	}
	conteudo, err := tpl.Preencher(canal, n.Dados)
	if err != nil {
		return dominio.ResultadoEnvio{Tipo: dominio.FalhaDefinitiva, Motivo: err.Error()}
	}
	return remetente.Enviar(ctx, n.Destinatario.Contato(canal), conteudo)
}

func descreverProxima(p *dominio.ProximaTentativa) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%s#%d em %s", p.Canal, p.Numero, p.Espera.Round(time.Second))
}
