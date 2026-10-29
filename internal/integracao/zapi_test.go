package integracao

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ominicore/internal/adaptadores/entrada/webhooks"
	"ominicore/internal/adaptadores/saida/postgres"
	"ominicore/internal/aplicacao"
)

// webhookZAPI entrega um aviso da Z-API pela rota /webhooks/zapi com o
// segredo e a instância certos, aplicando no Postgres.
func (p ponta) webhookZAPI(t *testing.T, corpo string) int {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	webhooks.Webhooks{
		ZAPI: func() webhooks.ZAPIAtual {
			return webhooks.ZAPIAtual{Ligada: true, InstanceID: "inst1", Segredo: "segredo-de-teste"}
		},
		Receptor: aplicacao.ReceiveMessageUseCase{Repo: postgres.Repositorio{Pool: p.pool}, Log: log},
		Log:      log,
	}.Registrar(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/zapi/segredo-de-teste", strings.NewReader(corpo)))
	return rec.Code
}

// T-128 no Postgres: "Ao enviar" da Z-API com número inexistente falha a
// Tentativa de WhatsApp como contato inválido e cria a de SMS (REQ-39, REQ-13).
func TestT128_NumeroSemWhatsAppNaZAPIPassaAoSMSNoPostgres(t *testing.T) {
	p := novaPonta(t)
	id := p.aceitaWhatsApp(t, "k-128") // a Tentativa de WhatsApp aceita guarda o id "wamid.fake"
	corpo := `{"type":"DeliveryCallback","instanceId":"inst1","messageId":"wamid.fake","zaapId":"Z1","phone":"5511999998888","error":"Phone number does not exist"}`
	if code := p.webhookZAPI(t, corpo); code != http.StatusOK {
		t.Fatalf("webhook: %d", code)
	}
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 2 && h[1].Status == "enviada" })
	if len(h) != 2 || h[0].Status != "falhou" || h[1].Canal != "sms" {
		t.Fatalf("histórico = %+v", h)
	}
	var cat, motivo string
	_ = p.pool.QueryRow(context.Background(), `SELECT categoria_falha, motivo FROM tentativas WHERE notificacao_id = $1 AND canal = 'whatsapp'`, id).Scan(&cat, &motivo)
	if cat != "contato_invalido" || strings.Contains(motivo, "5511999998888") {
		t.Fatalf("categoria = %q, motivo = %q", cat, motivo)
	}
}
