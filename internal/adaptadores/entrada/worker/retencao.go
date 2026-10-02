package worker

import (
	"context"
	"log/slog"
	"time"
)

// Executor roda uma rodada de retenção (RetencaoUseCase).
type Executor interface {
	Rodar(ctx context.Context) (int, error)
}

// Retencao roda a anonimização no boot e a cada Intervalo (padrão 1 h), ao
// lado do laço de despacho (REQ-19).
type Retencao struct {
	Executor  Executor
	Intervalo time.Duration
	Log       *slog.Logger
}

// Rodar bloqueia até o ctx ser cancelado.
func (r Retencao) Rodar(ctx context.Context) {
	intervalo := r.Intervalo
	if intervalo <= 0 {
		intervalo = time.Hour
	}
	t := time.NewTicker(intervalo)
	defer t.Stop()
	for {
		if _, err := r.Executor.Rodar(ctx); err != nil && ctx.Err() == nil {
			r.Log.ErrorContext(ctx, "rodada de retenção falhou", "erro", err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
