// Package worker é o adaptador de entrada que puxa Tentativas da fila no
// Postgres e as entrega ao caso de uso de despacho.
package worker

import (
	"context"
	"log/slog"
	"time"
)

// Despachante processa uma Tentativa por vez; false = fila vazia.
type Despachante interface {
	ProcessarProxima(ctx context.Context) (bool, error)
}

// Worker consulta a fila a cada Intervalo quando ela está vazia e esvazia a
// fila sem espera quando há trabalho. 500 ms mantém a 1ª Tentativa bem abaixo
// dos 5 s da meta (T-11).
type Worker struct {
	Despachante Despachante
	Intervalo   time.Duration
	Log         *slog.Logger
}

// Rodar bloqueia até o ctx ser cancelado.
func (w Worker) Rodar(ctx context.Context) {
	intervalo := w.Intervalo
	if intervalo <= 0 {
		intervalo = 500 * time.Millisecond
	}
	for ctx.Err() == nil {
		processou, err := w.Despachante.ProcessarProxima(ctx)
		if err != nil && ctx.Err() == nil {
			w.Log.ErrorContext(ctx, "despacho falhou", "erro", err.Error())
		}
		if processou && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
		case <-time.After(intervalo):
		}
	}
}
