package aplicacao_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"

	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/memoria"
)

// Gravações e releituras concorrentes terminam com a configuração do cofre
// em uso: a releitura nunca aplica uma leitura anterior à última gravação.
func TestReleituraEGravacaoConcorrentesTerminamComOCofre(t *testing.T) {
	ultra, _ := dominio.NovaConfigWhatsApp("ultramsg", map[string]string{"instance_id": "i", "token": "t"})
	cofre := memoria.NovoCofre(ultra)
	atual := aplicacao.NovaConfigWhatsAppAtual(ultra, func(dominio.ConfigWhatsApp) portas.MessageSenderPort { return nil })
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	conf := aplicacao.ConfigurarUseCase{Cofre: cofre, Registro: memoria.Novo(), Atual: atual, Log: log}
	rel := aplicacao.ReleituraConfig{Cofre: cofre, Atual: atual, Log: log}
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func() { defer wg.Done(); rel.Reler(context.Background()) }()
		go func() {
			defer wg.Done()
			prov := []string{"zapi", "ultramsg"}[i%2]
			_, _ = conf.Gravar(context.Background(), prov, map[string]string{"instance_id": "i", "token": "t"}, "ops-a")
		}()
	}
	wg.Wait()
	noCofre, _ := cofre.LerWhatsApp(context.Background())
	if !atual.Config().Igual(noCofre) {
		t.Fatalf("em uso = %s, no cofre = %s", atual.ProvedorWhatsApp(), noCofre.Provedor)
	}
}
