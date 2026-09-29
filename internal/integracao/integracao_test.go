// Package integracao prova, contra um Postgres real (testcontainers), o que o
// repositório em memória não alcança: transação, SKIP LOCKED e a idempotência
// sob concorrência. Pula sozinho sem runtime de containers (Docker/Podman);
// com Podman: `make itest`.
package integracao

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"ominicore/internal/adaptadores/entrada/worker"
	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/adaptadores/saida/postgres"
	"ominicore/internal/adaptadores/saida/sendgrid"
	"ominicore/internal/adaptadores/saida/twilio"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/sendgridfake"
	"ominicore/internal/testes/twiliofake"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

var (
	umaVez sync.Once
	dsn    string
	erroPG error
)

// banco sobe um Postgres por pacote e devolve um pool com o schema limpo.
func banco(t *testing.T) *pgxpool.Pool {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	umaVez.Do(func() {
		c, err := tcpostgres.Run(ctx, "docker.io/library/postgres:17-alpine",
			tcpostgres.WithDatabase("ominicore"), tcpostgres.WithUsername("ominicore"), tcpostgres.WithPassword("ominicore"),
			tcpostgres.BasicWaitStrategies())
		if err != nil {
			erroPG = err
			return
		}
		dsn, erroPG = c.ConnectionString(ctx, "sslmode=disable")
	})
	if erroPG != nil {
		t.Fatalf("subir postgres: %v", erroPG)
	}
	pool, err := postgres.Conectar(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS ominicore CASCADE; DROP TABLE IF EXISTS public.goose_db_version`); err != nil {
		t.Fatal(err)
	}
	if err := postgres.Migrar(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func nova(origem, chave string) portas.NovaNotificacao {
	return portas.NovaNotificacao{
		SistemaOrigem: origem, ChaveIdempotencia: chave, TemplateID: "boas-vindas",
		Destinatario: dominio.Destinatario{Email: "ana@exemplo.com"}, Dados: map[string]string{"nome": "Ana"},
		OrdemCanais: []dominio.Canal{dominio.CanalEmail},
	}
}

func contar(t *testing.T, pool *pgxpool.Pool) (notificacoes, tentativas int) {
	t.Helper()
	err := pool.QueryRow(context.Background(), `SELECT (SELECT count(*) FROM notificacoes), (SELECT count(*) FROM tentativas)`).Scan(&notificacoes, &tentativas)
	if err != nil {
		t.Fatal(err)
	}
	return notificacoes, tentativas
}

func TestT7_ReenvioComAMesmaChaveNaoDuplica(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	ctx := context.Background()
	primeiro, err := repo.RegistrarNotificacao(ctx, nova("sistema-a", "abc-1"))
	if err != nil {
		t.Fatal(err)
	}
	// reenvios concorrentes, como timeouts do chamador em sequência
	var wg sync.WaitGroup
	ids := make([]string, 10)
	for i := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reg, err := repo.RegistrarNotificacao(ctx, nova("sistema-a", "abc-1"))
			if err != nil {
				t.Error(err)
				return
			}
			ids[i] = reg.ID
		}()
	}
	wg.Wait()
	for _, id := range ids {
		if id != primeiro.ID {
			t.Fatalf("id = %s, quero %s", id, primeiro.ID)
		}
	}
	if n, tent := contar(t, pool); n != 1 || tent != 1 {
		t.Fatalf("notificações=%d tentativas=%d, quero 1 e 1", n, tent)
	}
}

func TestT9_AChaveValePorSistemaDeOrigem(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	a, errA := repo.RegistrarNotificacao(context.Background(), nova("sistema-a", "abc-1"))
	b, errB := repo.RegistrarNotificacao(context.Background(), nova("sistema-b", "abc-1"))
	if errA != nil || errB != nil {
		t.Fatal(errA, errB)
	}
	if a.ID == b.ID {
		t.Fatal("as duas origens receberam a mesma Notificação")
	}
	if n, _ := contar(t, pool); n != 2 {
		t.Fatalf("notificações = %d, quero 2", n)
	}
}

func TestChaveVencidaApos24hGeraNovaNotificacao(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	ctx := context.Background()
	antiga, err := repo.RegistrarNotificacao(ctx, nova("sistema-a", "abc-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE idempotencia SET expira_em = now() - interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	nova, err := repo.RegistrarNotificacao(ctx, nova("sistema-a", "abc-1"))
	if err != nil || nova.Reutilizada || nova.ID == antiga.ID {
		t.Fatalf("depois de 24 h: %+v err=%v", nova, err)
	}
}

// ponta liga API real, worker real e Postgres real ao SendGrid fake.
type ponta struct {
	pool  *pgxpool.Pool
	sg    *sendgridfake.Fake
	tw    *twiliofake.Fake
	realm *emissor.Emissor
	api   http.Handler
}

func novaPonta(t *testing.T) ponta {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	p := ponta{pool: pool, sg: sendgridfake.Novo(t), tw: twiliofake.Novo(t), realm: emissor.Novo(t)}
	p.api = entradahttp.Servidor{
		Auth: keycloak.Novo(p.realm.URL, "ominicore", &emissor.Cache{}), Enviar: aplicacao.SendMessageUseCase{Repo: repo}, Log: log,
	}.Handler()
	uc := aplicacao.DespacharUseCase{
		Repo: repo, Log: log,
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{
			dominio.CanalEmail: sendgrid.Novo(p.sg.URL, "SG.teste", "no-reply@exemplo.com"),
			dominio.CanalSMS: twilio.Novo(p.tw.URL, twilio.Credenciais{
				AccountSID: twiliofake.AccountSID, APIKeySID: twiliofake.APIKeySID,
				APIKeySecret: twiliofake.APIKeySecret, MessagingServiceSID: twiliofake.MessagingServiceSID,
			}),
		},
	}
	ctx, cancelar := context.WithCancel(context.Background())
	t.Cleanup(cancelar)
	go worker.Worker{Despachante: uc, Log: log}.Rodar(ctx) // intervalo de produção (500 ms)
	return p
}

func (p ponta) pedir(t *testing.T, chave string) (id string, respondidoEm time.Time) {
	t.Helper()
	return p.pedirCorpo(t, chave, `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"}}`)
}

func (p ponta) pedirCorpo(t *testing.T, chave, corpoJSON string) (id string, respondidoEm time.Time) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/notificacoes", strings.NewReader(corpoJSON))
	req.Header.Set("Authorization", "Bearer "+p.realm.Valido(t, "sistema-a"))
	req.Header.Set("Idempotency-Key", chave)
	rec := httptest.NewRecorder()
	p.api.ServeHTTP(rec, req)
	respondidoEm = time.Now()
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var corpo struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil || corpo.ID == "" {
		t.Fatalf("resposta sem id: %s", rec.Body)
	}
	return corpo.ID, respondidoEm
}

func (p ponta) esperar(t *testing.T, id string, notificacao dominio.StatusNotificacao) (tentativa, motivo string) {
	t.Helper()
	fim := time.Now().Add(5 * time.Second)
	for time.Now().Before(fim) {
		var status string
		err := p.pool.QueryRow(context.Background(),
			`SELECT n.status, t.status, t.motivo FROM notificacoes n JOIN tentativas t ON t.notificacao_id = n.id WHERE n.id = $1`, id).
			Scan(&status, &tentativa, &motivo)
		if err == nil && status == string(notificacao) {
			return tentativa, motivo
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("notificação %s não chegou a %s", id, notificacao)
	return "", ""
}

func TestT4_PedidoValidoFicaPendenteComTentativaNoEmail(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	reg, err := aplicacao.SendMessageUseCase{Repo: repo}.Pedir(context.Background(), aplicacao.PedidoNotificacao{
		SistemaOrigem: "sistema-a", ChaveIdempotencia: "abc-1", TemplateID: "boas-vindas",
		Destinatario: dominio.Destinatario{Email: "ana@exemplo.com"}, Dados: map[string]string{"nome": "Ana"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var status, canal, tentativa string
	err = pool.QueryRow(context.Background(),
		`SELECT n.status, t.canal, t.status FROM notificacoes n JOIN tentativas t ON t.notificacao_id = n.id WHERE n.id = $1`, reg.ID).
		Scan(&status, &canal, &tentativa)
	if err != nil || status != "pendente" || canal != "email" || tentativa != "pendente" {
		t.Fatalf("status=%s canal=%s tentativa=%s err=%v", status, canal, tentativa, err)
	}
}

func TestT10_TentativaDeEmailEEnviadaPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	id, _ := p.pedir(t, "abc-1")
	if tentativa, _ := p.esperar(t, id, dominio.NotificacaoEnviada); tentativa != "enviada" {
		t.Fatalf("tentativa = %s", tentativa)
	}
	if e := p.sg.Emails(); len(e) != 1 || e[0].Assunto != "Bem-vindo, Ana" {
		t.Fatalf("e-mails = %+v", e)
	}
}

func TestT11_PrimeiraTentativaEmAte5Segundos(t *testing.T) {
	p := novaPonta(t)
	_, respondidoEm := p.pedir(t, "abc-1")
	email := p.sg.Esperar(t, 1, 5*time.Second)[0]
	if atraso := email.RecebidoEm.Sub(respondidoEm); atraso > 5*time.Second {
		t.Fatalf("primeira Tentativa levou %s", atraso)
	}
}

func TestT12_RecusaDoProvedorEncerraSemOutroCanalPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	p.sg.Responder(http.StatusBadRequest)
	id, _ := p.pedir(t, "abc-1")
	tentativa, motivo := p.esperar(t, id, dominio.NotificacaoFalhou)
	if tentativa != "falhou" || !strings.Contains(motivo, "sendgrid 400") {
		t.Fatalf("tentativa=%s motivo=%q", tentativa, motivo)
	}
}

// linhaTentativa é uma Tentativa como está no banco.
type linhaTentativa struct {
	Canal, Status string
	Numero        int
	Espera        time.Duration // disponivel_em - criada_em
}

func historico(t *testing.T, pool *pgxpool.Pool, id string) []linhaTentativa {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT canal, status, numero, extract(epoch FROM disponivel_em - criada_em) FROM tentativas WHERE notificacao_id = $1 ORDER BY id`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []linhaTentativa
	for rows.Next() {
		var l linhaTentativa
		var seg float64
		if err := rows.Scan(&l.Canal, &l.Status, &l.Numero, &seg); err != nil {
			t.Fatal(err)
		}
		l.Espera = time.Duration(seg * float64(time.Second))
		out = append(out, l)
	}
	return out
}

// esperarHistorico aguarda até o histórico satisfazer ok (5 s).
func esperarHistorico(t *testing.T, pool *pgxpool.Pool, id string, ok func([]linhaTentativa) bool) []linhaTentativa {
	t.Helper()
	fim := time.Now().Add(5 * time.Second)
	for {
		h := historico(t, pool, id)
		if ok(h) || time.Now().After(fim) {
			return h
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func statusNotificacao(t *testing.T, pool *pgxpool.Pool, id string) string {
	t.Helper()
	var s string
	if err := pool.QueryRow(context.Background(), `SELECT status FROM notificacoes WHERE id = $1`, id).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

const comTelefone = `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"}}`

func TestT15_OrdemDoPedidoGravadaEPrimeiraTentativaNoSMS(t *testing.T) {
	p := novaPonta(t)
	id, _ := p.pedirCorpo(t, "k-15", `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["sms","email"]}`)
	var ordem []string
	if err := p.pool.QueryRow(context.Background(), `SELECT ordem_canais FROM notificacoes WHERE id = $1`, id).Scan(&ordem); err != nil {
		t.Fatal(err)
	}
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) > 0 && h[0].Status == "enviada" })
	if len(ordem) != 2 || ordem[0] != "sms" || h[0].Canal != "sms" {
		t.Fatalf("ordem = %v, histórico = %+v", ordem, h)
	}
}

func TestT21_RetentativaAgendadaEntre4e6Segundos(t *testing.T) {
	p := novaPonta(t)
	p.sg.Responder(http.StatusServiceUnavailable)
	id, _ := p.pedirCorpo(t, "k-21", comTelefone)
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) >= 2 })
	if len(h) != 2 || h[0].Status != "falhou" || h[1].Canal != "email" || h[1].Numero != 2 || h[1].Status != "pendente" {
		t.Fatalf("histórico = %+v", h)
	}
	// a espera conta do commit da conclusão; a criada_em da Retentativa é o mesmo instante
	if h[1].Espera < 4*time.Second || h[1].Espera > 6*time.Second {
		t.Fatalf("Retentativa disponível após %s, quero entre 4 s e 6 s", h[1].Espera)
	}
	if s := statusNotificacao(t, p.pool, id); s != "pendente" {
		t.Fatalf("notificação = %s, quero pendente", s)
	}
}

func TestT23_QuatroFalhasTemporariasPassamAoSMS(t *testing.T) {
	p := novaPonta(t)
	p.sg.Responder(http.StatusServiceUnavailable)
	id, _ := p.pedirCorpo(t, "k-23", comTelefone)
	// adianta o relógio das Retentativas para o teste não esperar 2min35s
	ctx, parar := context.WithCancel(context.Background())
	defer parar()
	go func() {
		for ctx.Err() == nil {
			_, _ = p.pool.Exec(ctx, `UPDATE tentativas SET disponivel_em = now() WHERE notificacao_id = $1 AND status = 'pendente'`, id)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 5 && h[4].Status == "enviada" })
	parar()
	for i := range 4 {
		if h[i].Canal != "email" || h[i].Numero != i+1 || h[i].Status != "falhou" {
			t.Fatalf("Tentativa %d = %+v (histórico %+v)", i+1, h[i], h)
		}
	}
	if len(h) != 5 || h[4].Canal != "sms" || h[4].Numero != 1 {
		t.Fatalf("depois de 4 Tentativas no Email quero o SMS: %+v", h)
	}
	if len(p.sg.Emails()) != 4 {
		t.Fatalf("SendGrid recebeu %d envios, quero 4", len(p.sg.Emails()))
	}
}

func TestT24_FalhaDefinitivaNoEmailPassaAoSMSPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	p.sg.Responder(http.StatusBadRequest)
	id, _ := p.pedirCorpo(t, "k-24", comTelefone)
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 2 && h[1].Status == "enviada" })
	if len(h) != 2 || h[0].Canal != "email" || h[0].Status != "falhou" || h[1].Canal != "sms" {
		t.Fatalf("histórico = %+v", h)
	}
	if s := statusNotificacao(t, p.pool, id); s != "enviada" {
		t.Fatalf("notificação = %s", s)
	}
}

func TestT25_SemTelefoneOSMSEPuladoPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	p.sg.Responder(http.StatusBadRequest)
	id, _ := p.pedir(t, "k-25") // só e-mail; ordem do Template Email → SMS
	esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "falhou" })
	time.Sleep(700 * time.Millisecond) // mais que um ciclo do worker
	if h := historico(t, p.pool, id); len(h) != 1 || statusNotificacao(t, p.pool, id) != "falhou" || len(p.tw.Recebidos()) != 0 {
		t.Fatalf("histórico = %+v, SMS = %d", h, len(p.tw.Recebidos()))
	}
}

func TestT26_TodosOsCanaisFalhamPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	p.sg.Responder(http.StatusBadRequest)
	p.tw.Responder(http.StatusBadRequest, 21211)
	id, _ := p.pedirCorpo(t, "k-26", comTelefone)
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 2 && h[1].Status == "falhou" })
	if len(h) != 2 || h[0].Status != "falhou" || statusNotificacao(t, p.pool, id) != "falhou" {
		t.Fatalf("histórico = %+v", h)
	}
}

func TestT27_SucessoNoEmailNaoCriaSMS(t *testing.T) {
	p := novaPonta(t)
	id, _ := p.pedirCorpo(t, "k-27", comTelefone)
	esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "enviada" })
	time.Sleep(700 * time.Millisecond)
	if h := historico(t, p.pool, id); len(h) != 1 || statusNotificacao(t, p.pool, id) != "enviada" || len(p.tw.Recebidos()) != 0 {
		t.Fatalf("histórico = %+v", h)
	}
}
