package aplicacao

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ominicore/internal/portas"
)

// LoteRetencao é o tamanho de cada transação da rodada de retenção.
const LoteRetencao = 1000

// RetencaoUseCase anonimiza as Notificações finalizadas que passaram do prazo
// (REQ-19). Agora é injetado para os testes.
type RetencaoUseCase struct {
	Repo  portas.MessageRepositoryPort
	Prazo time.Duration
	Agora func() time.Time
	Log   *slog.Logger
}

// Rodar anonimiza em lotes até não sobrar nada elegível e devolve o total.
// Idempotente: o que já foi anonimizado não é elegível de novo.
func (uc RetencaoUseCase) Rodar(ctx context.Context) (int, error) {
	agora := time.Now
	if uc.Agora != nil {
		agora = uc.Agora
	}
	antes := agora().Add(-uc.Prazo)
	total := 0
	for {
		n, err := uc.Repo.AnonimizarLote(ctx, antes, LoteRetencao)
		if err != nil {
			return total, err
		}
		total += n
		if n < LoteRetencao {
			break
		}
	}
	pendentes, err := uc.Repo.PendentesVencidas(ctx, antes)
	if err != nil {
		return total, fmt.Errorf("contar pendentes vencidas: %w", err)
	}
	if pendentes > 0 { // anomalia: Retentativas acabam em minutos; fica para a próxima rodada
		uc.Log.WarnContext(ctx, "notificações pendentes além do prazo de retenção", "quantidade", pendentes)
	}
	uc.Log.InfoContext(ctx, "rodada de retenção", "anonimizadas", total)
	return total, nil
}
