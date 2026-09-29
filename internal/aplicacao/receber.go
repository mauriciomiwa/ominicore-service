package aplicacao

import (
	"context"
	"fmt"
	"log/slog"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// ResultadoAviso diz o que aconteceu com um Aviso de status válido; qualquer
// um deles responde sucesso ao Provedor (REQ-12).
type ResultadoAviso int

const (
	AvisoAplicado     ResultadoAviso = iota
	AvisoIgnorado                    // repetido, atrasado ou sobre Tentativa final
	AvisoDesconhecido                // id do Provedor que não existe
)

// ReceiveMessageUseCase aplica Avisos de status (REQ-12, REQ-13).
type ReceiveMessageUseCase struct {
	Repo   portas.MessageRepositoryPort
	Router dominio.NotificationRouter
	Log    *slog.Logger
}

// Receber aplica o aviso só se ele fizer a Tentativa avançar.
func (uc ReceiveMessageUseCase) Receber(ctx context.Context, a dominio.AvisoStatus) (ResultadoAviso, error) {
	res, ok, err := uc.Repo.ReservarPorIDProvedor(ctx, a.IDProvedor)
	if err != nil {
		return AvisoIgnorado, err
	}
	if !ok {
		uc.Log.InfoContext(ctx, "aviso de tentativa desconhecida", "id_provedor", a.IDProvedor)
		return AvisoDesconhecido, nil
	}
	if !res.Tentativa.Status.Avanca(a.Status) {
		res.Liberar(ctx)
		return AvisoIgnorado, nil
	}
	atualizacao := uc.Router.DecidirAviso(res.Tentativa, res.Notificacao, a)
	if err := res.Aplicar(ctx, atualizacao); err != nil {
		return AvisoIgnorado, fmt.Errorf("aplicar aviso na tentativa %d: %w", res.Tentativa.ID, err)
	}
	uc.Log.InfoContext(ctx, "aviso de status aplicado",
		"notificacao_id", res.Notificacao.ID, "tentativa_id", res.Tentativa.ID, "canal", res.Tentativa.Canal,
		"status", atualizacao.Tentativa, "proxima", descreverProxima(atualizacao.Proxima))
	return AvisoAplicado, nil
}

// ConsultarUseCase devolve a Notificação ao Sistema de origem dono (REQ-14).
type ConsultarUseCase struct {
	Repo portas.MessageRepositoryPort
}

// Consultar devolve ok=false para Notificação inexistente ou de outro dono.
func (uc ConsultarUseCase) Consultar(ctx context.Context, sistemaOrigem, id string) (portas.Consulta, bool, error) {
	return uc.Repo.ConsultarNotificacao(ctx, sistemaOrigem, id)
}
