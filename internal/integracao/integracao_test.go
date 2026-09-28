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
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/sendgridfake"

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
	realm *emissor.Emissor
	api   http.Handler
}

func novaPonta(t *testing.T) ponta {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	p := ponta{pool: pool, sg: sendgridfake.Novo(t), realm: emissor.Novo(t)}
	p.api = entradahttp.Servidor{
		Auth: keycloak.Novo(p.realm.URL, "ominicore", &emissor.Cache{}), Enviar: aplicacao.SendMessageUseCase{Repo: repo}, Log: log,
	}.Handler()
	uc := aplicacao.DespacharUseCase{
		Repo: repo, Log: log,
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{dominio.CanalEmail: sendgrid.Novo(p.sg.URL, "SG.teste", "no-reply@exemplo.com")},
	}
	ctx, cancelar := context.WithCancel(context.Background())
	t.Cleanup(cancelar)
	go worker.Worker{Despachante: uc, Log: log}.Rodar(ctx) // intervalo de produção (500 ms)
	return p
}

func (p ponta) pedir(t *testing.T, chave string) (id string, respondidoEm time.Time) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/notificacoes",
		strings.NewReader(`{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana"}}`))
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
