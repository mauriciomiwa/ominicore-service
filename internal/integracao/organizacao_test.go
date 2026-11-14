package integracao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"ominicore/internal/adaptadores/saida/postgres"
)

const (
	versaoAntesDasOrganizacoes int64 = 12
	versaoOrganizacoes         int64 = 13
)

// gravarAntesDasOrganizacoes volta a base para a 00012 e grava uma Notificação
// com Tentativa como o código de antes gravava (sem Organização).
func gravarAntesDasOrganizacoes(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	if err := postgres.VoltarPara(ctx, pool, versaoAntesDasOrganizacoes); err != nil {
		t.Fatal(err)
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO notificacoes (id, sistema_origem, template_id, destinatario_email, dados, ordem_canais, status)
		VALUES ('11111111-1111-4111-8111-111111111111', 'wbem', 'boas-vindas', 'ana@exemplo.com', '{"nome":"Ana"}', ARRAY['email'], 'entregue');
		INSERT INTO tentativas (notificacao_id, canal, numero, status, id_provedor)
		VALUES ('11111111-1111-4111-8111-111111111111', 'email', 1, 'entregue', 'sg-1')`)
	if err != nil {
		t.Fatal(err)
	}
}

// retrato resume o que não pode mudar na passagem: Notificações, Tentativas e
// Templates.
func retrato(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var s string
	err := pool.QueryRow(context.Background(), `
		SELECT coalesce((SELECT string_agg(n.id || ':' || n.sistema_origem || ':' || n.status || ':' || n.dados::text, ';' ORDER BY n.id) FROM notificacoes n), '')
		    || '|' || coalesce((SELECT string_agg(t.notificacao_id || ':' || t.canal || ':' || t.status || ':' || t.id_provedor, ';' ORDER BY t.id) FROM tentativas t), '')
		    || '|' || coalesce((SELECT string_agg(id, ',' ORDER BY id) FROM templates), '')`).Scan(&s)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestT161_DadosExistentesVaoParaAOrganizacaoWbem(t *testing.T) {
	pool := banco(t)
	gravarAntesDasOrganizacoes(t, pool)
	antes := retrato(t, pool)
	if err := postgres.MigrarAte(context.Background(), pool, versaoOrganizacoes); err != nil {
		t.Fatal(err)
	}
	if depois := retrato(t, pool); depois != antes {
		t.Fatalf("a passagem mudou os dados:\nantes  %s\ndepois %s", antes, depois)
	}
	var foraDaWbem, templatesForaDaWbem int
	var orgDoWbem string
	err := pool.QueryRow(context.Background(), `
		SELECT (SELECT count(*) FROM notificacoes WHERE organizacao <> 'wbem'),
		       (SELECT count(*) FROM templates WHERE organizacao <> 'wbem'),
		       (SELECT organizacao FROM sistemas_origem WHERE client = 'wbem')`).Scan(&foraDaWbem, &templatesForaDaWbem, &orgDoWbem)
	if err != nil {
		t.Fatal(err)
	}
	if foraDaWbem != 0 || templatesForaDaWbem != 0 || orgDoWbem != "wbem" {
		t.Fatalf("notificações fora da wbem = %d, templates fora = %d, organização do client wbem = %q", foraDaWbem, templatesForaDaWbem, orgDoWbem)
	}
}

func TestT162_VoltarAtrasNaoPerdeDados(t *testing.T) {
	pool := banco(t)
	gravarAntesDasOrganizacoes(t, pool)
	antes := retrato(t, pool)
	ctx := context.Background()
	if err := postgres.MigrarAte(ctx, pool, versaoOrganizacoes); err != nil {
		t.Fatal(err)
	}
	if err := postgres.VoltarPara(ctx, pool, versaoAntesDasOrganizacoes); err != nil {
		t.Fatal(err)
	}
	if depois := retrato(t, pool); depois != antes {
		t.Fatalf("voltar atrás mudou os dados:\nantes  %s\ndepois %s", antes, depois)
	}
	var coluna bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'ominicore' AND table_name = 'notificacoes' AND column_name = 'organizacao')`).Scan(&coluna); err != nil || coluna {
		t.Fatalf("coluna organizacao ainda existe (err=%v)", err)
	}
}

// TestT163_BarreiraDoBancoEntreOrganizacoes lê com um role comum (o do app em
// hml e prod não é superusuário, que ignoraria a política), com a Organização
// fixada como a API fixa, e busca por id sem filtrar pelo Sistema de origem.
func TestT163_BarreiraDoBancoEntreOrganizacoes(t *testing.T) {
	pool := banco(t)
	ctx := context.Background()
	repo := postgres.Repositorio{Pool: pool}
	reg, err := repo.RegistrarNotificacao(ctx, nova("wbem", "k-163"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'leitor_api_teste') THEN
				CREATE ROLE leitor_api_teste NOLOGIN;
			END IF;
		END $$;
		GRANT USAGE ON SCHEMA ominicore TO leitor_api_teste;
		GRANT SELECT ON notificacoes TO leitor_api_teste`)
	if err != nil {
		t.Fatal(err)
	}
	ler := func(org string) int {
		t.Helper()
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		var n int
		err = tx.QueryRow(ctx, `
			SELECT set_config('ominicore.organizacao', $1, true);`, org).Scan(new(string))
		if err == nil {
			_, err = tx.Exec(ctx, `SET LOCAL ROLE leitor_api_teste`)
		}
		if err == nil {
			err = tx.QueryRow(ctx, `SELECT count(*) FROM notificacoes WHERE id = $1`, reg.ID).Scan(&n)
		}
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := ler("org-b"); n != 0 {
		t.Fatalf("leitura fixada na org-b encontrou %d Notificação(ões) da wbem", n)
	}
	if n := ler("wbem"); n != 1 {
		t.Fatalf("controle: leitura fixada na wbem encontrou %d, quero 1", n)
	}
}

// pedirComo pede uma Notificação pela API em nome do client dado.
func (p ponta) pedirComo(t *testing.T, client, chave, corpoJSON string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/notificacoes", strings.NewReader(corpoJSON))
	req.Header.Set("Authorization", "Bearer "+p.realm.Valido(t, client))
	req.Header.Set("Idempotency-Key", chave)
	rec := httptest.NewRecorder()
	p.api.ServeHTTP(rec, req)
	return rec
}

func idDe(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var corpo struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil || corpo.ID == "" {
		t.Fatalf("resposta sem id: %s", rec.Body)
	}
	return corpo.ID
}

func entregaMeta(wamid string) string {
	return `{"entry":[{"changes":[{"value":{"statuses":[{"id":"` + wamid + `","status":"delivered"}]}}]}]}`
}

func TestT165_DespachoEAvisosSeguemParaTodasAsOrganizacoes(t *testing.T) {
	p := novaPonta(t)
	recWbem := p.pedirComo(t, "wbem", "k-165-wbem", `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"}}`)
	recB := p.pedirComo(t, "sistema-b", "k-165-b", pedidoWhatsAppSMS)
	if recWbem.Code != http.StatusAccepted || recB.Code != http.StatusAccepted {
		t.Fatalf("pedidos: wbem %d %s, org-b %d %s", recWbem.Code, recWbem.Body, recB.Code, recB.Body)
	}
	idWbem, idB := idDe(t, recWbem), idDe(t, recB)
	esperarHistorico(t, p.pool, idWbem, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "enviada" })
	esperarHistorico(t, p.pool, idB, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "enviada" })
	if code := p.webhookMeta(t, entregaMeta("wamid.fake")); code != http.StatusNoContent {
		t.Fatalf("webhook: %d", code)
	}
	h := esperarHistorico(t, p.pool, idB, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "entregue" })
	if h[0].Canal != "whatsapp" {
		t.Fatalf("histórico da org-b = %+v", h)
	}
	var orgs string
	if err := p.pool.QueryRow(context.Background(), `SELECT string_agg(organizacao, ',' ORDER BY organizacao) FROM notificacoes WHERE id IN ($1, $2)`, idWbem, idB).Scan(&orgs); err != nil || orgs != "org-b,wbem" {
		t.Fatalf("organizações = %q (err=%v), quero org-b,wbem", orgs, err)
	}
}
