// Package integracao prova, contra um Postgres e um Keycloak reais (testcontainers), o que o
// repositório em memória não alcança: transação, SKIP LOCKED e a idempotência
// sob concorrência. Pula sozinho sem runtime de containers (Docker/Podman);
// com Podman: `make itest`.
package integracao

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	"ominicore/internal/adaptadores/saida/whatsapp"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/metafake"
	"ominicore/internal/testes/sendgridfake"
	"ominicore/internal/testes/twiliofake"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
	"ominicore/internal/adaptadores/entrada/webhooks"
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
	pool, err := postgres.Conectar(ctx, dsn, postgres.SchemaPadrao)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS ominicore CASCADE; DROP TABLE IF EXISTS public.goose_db_version`); err != nil {
		t.Fatal(err)
	}
	if err := postgres.Migrar(ctx, pool, postgres.SchemaPadrao); err != nil {
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
	meta  *metafake.Fake
	realm *emissor.Emissor
	api   http.Handler
}

func novaPonta(t *testing.T) ponta {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	p := ponta{pool: pool, sg: sendgridfake.Novo(t), tw: twiliofake.Novo(t), meta: metafake.Novo(t), realm: emissor.Novo(t)}
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
			dominio.CanalWhatsApp: whatsapp.Novo(p.meta.URL, whatsapp.Credenciais{Token: metafake.Token, PhoneNumberID: metafake.PhoneNumberID}),
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

func TestSchemasDeHmlEProdFicamIsolados(t *testing.T) {
	banco(t) // garante o container e o schema de prod limpo
	ctx := context.Background()
	hml, err := postgres.Conectar(ctx, dsn, "ominicore_hml")
	if err != nil {
		t.Fatal(err)
	}
	defer hml.Close()
	if _, err := hml.Exec(ctx, `DROP SCHEMA IF EXISTS ominicore_hml CASCADE`); err != nil {
		t.Fatal(err)
	}
	if err := postgres.Migrar(ctx, hml, "ominicore_hml"); err != nil {
		t.Fatal(err)
	}
	if _, err := (postgres.Repositorio{Pool: hml}).RegistrarNotificacao(ctx, nova("sistema-a", "k-hml")); err != nil {
		t.Fatal(err)
	}
	prod, err := postgres.Conectar(ctx, dsn, postgres.SchemaPadrao)
	if err != nil {
		t.Fatal(err)
	}
	defer prod.Close()
	var nProd, nHml int
	_ = prod.QueryRow(ctx, `SELECT count(*) FROM notificacoes`).Scan(&nProd)
	_ = hml.QueryRow(ctx, `SELECT count(*) FROM notificacoes`).Scan(&nHml)
	if nProd != 0 || nHml != 1 {
		t.Fatalf("prod=%d hml=%d, quero 0 e 1", nProd, nHml)
	}
}

const pedidoWhatsAppSMS = `{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["whatsapp","sms"]}`

func TestT28_WhatsAppPeloModeloAprovadoPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	id, _ := p.pedirCorpo(t, "k-28", pedidoWhatsAppSMS)
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "enviada" })
	m := p.meta.Recebidas()
	if len(h) != 1 || h[0].Canal != "whatsapp" || len(m) != 1 || m[0].Modelo != "boas_vindas" || m[0].Idioma != "pt_BR" ||
		len(m[0].Parametros) != 1 || m[0].Parametros[0] != "Ana" {
		t.Fatalf("histórico = %+v, meta = %+v", h, m)
	}
}

func TestT29_NumeroSemWhatsAppPassaAoSMSPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	p.meta.Responder(http.StatusBadRequest, 131026)
	id, _ := p.pedirCorpo(t, "k-29", pedidoWhatsAppSMS)
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 2 && h[1].Status == "enviada" })
	if len(h) != 2 || h[0].Canal != "whatsapp" || h[0].Status != "falhou" || h[1].Canal != "sms" || statusNotificacao(t, p.pool, id) != "enviada" {
		t.Fatalf("histórico = %+v", h)
	}
}

const appSecretTeste = "app-secret-do-teste"

func (p ponta) webhookMeta(t *testing.T, corpo string) int {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mux := http.NewServeMux()
	webhooks.Webhooks{
		MetaLigada: true,
		Cred:       webhooks.Credenciais{MetaAppSecret: appSecretTeste, MetaVerifyToken: "v"},
		Receptor:   aplicacao.ReceiveMessageUseCase{Repo: postgres.Repositorio{Pool: p.pool}, Log: log},
		Log:        log,
	}.Registrar(mux)
	m := hmac.New(sha256.New, []byte(appSecretTeste))
	m.Write([]byte(corpo))
	req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", strings.NewReader(corpo))
	req.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(m.Sum(nil)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec.Code
}

func falhaMeta(wamid string) string {
	return `{"entry":[{"changes":[{"value":{"statuses":[{"id":"` + wamid + `","status":"failed","errors":[{"code":131026}]}]}}]}]}`
}

// aceitaWhatsApp pede WhatsApp → SMS e espera a Tentativa de WhatsApp aceita
// pela Meta falsa (wamid.fake); o SMS só nasce se o WhatsApp falhar.
func (p ponta) aceitaWhatsApp(t *testing.T, chave string) string {
	t.Helper()
	id, _ := p.pedirCorpo(t, chave, pedidoWhatsAppSMS)
	esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 1 && h[0].Status == "enviada" })
	return id
}

func TestT41_FalhaTardiaNoWhatsAppPassaAoSMSPontaAPonta(t *testing.T) {
	p := novaPonta(t)
	id := p.aceitaWhatsApp(t, "k-41")
	if code := p.webhookMeta(t, falhaMeta("wamid.fake")); code != http.StatusNoContent {
		t.Fatalf("webhook: %d", code)
	}
	h := esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 2 && h[1].Status == "enviada" })
	if len(h) != 2 || h[0].Status != "falhou" || h[1].Canal != "sms" {
		t.Fatalf("histórico = %+v", h)
	}
	var cat string
	_ = p.pool.QueryRow(context.Background(), `SELECT categoria_falha FROM tentativas WHERE notificacao_id = $1 AND canal = 'whatsapp'`, id).Scan(&cat)
	if cat != "contato_invalido" {
		t.Fatalf("categoria = %q", cat)
	}
}

func TestAvisosDeFalhaSimultaneosCriamUmSoFallback(t *testing.T) {
	p := novaPonta(t)
	id := p.aceitaWhatsApp(t, "k-concorrente")
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); p.webhookMeta(t, falhaMeta("wamid.fake")) }()
	}
	wg.Wait()
	var sms int
	_ = p.pool.QueryRow(context.Background(), `SELECT count(*) FROM tentativas WHERE notificacao_id = $1 AND canal = 'sms'`, id).Scan(&sms)
	if sms != 1 {
		t.Fatalf("8 avisos simultâneos criaram %d Tentativas de SMS, quero 1", sms)
	}
}

func TestT43_ConsultaNoPostgresSemContatoNemMotivo(t *testing.T) {
	p := novaPonta(t)
	id := p.aceitaWhatsApp(t, "k-43")
	p.webhookMeta(t, falhaMeta("wamid.fake"))
	esperarHistorico(t, p.pool, id, func(h []linhaTentativa) bool { return len(h) == 2 && h[1].Status == "enviada" })
	c, ok, err := postgres.Repositorio{Pool: p.pool}.ConsultarNotificacao(context.Background(), "sistema-a", id)
	if err != nil || !ok || len(c.Tentativas) != 2 || c.Tentativas[0].Categoria != dominio.FalhaContatoInvalido || c.CriadaEm.IsZero() {
		t.Fatalf("consulta = %+v ok=%v err=%v", c, ok, err)
	}
	if _, ok, _ := (postgres.Repositorio{Pool: p.pool}).ConsultarNotificacao(context.Background(), "sistema-b", id); ok {
		t.Fatal("sistema-b enxergou a Notificação de sistema-a")
	}
}

var chaveOptOutTeste = []byte("chave-de-teste-com-mais-de-32-caracteres")

func TestT55_TabelaDeOptOutSemContatoLegivel(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool, ChaveOptOut: chaveOptOutTeste}
	ctx := context.Background()
	if err := repo.RegistrarOptOut(ctx, dominio.CanalEmail, "Ana@Exemplo.com", "api:sistema-a"); err != nil {
		t.Fatal(err)
	}
	var despejo string
	if err := pool.QueryRow(ctx, `SELECT coalesce(string_agg(o::text, ' '), '') FROM opt_outs o`).Scan(&despejo); err != nil {
		t.Fatal(err)
	}
	var eventos string
	_ = pool.QueryRow(ctx, `SELECT coalesce(string_agg(e::text, ' '), '') FROM opt_out_eventos e`).Scan(&eventos)
	for _, proibido := range []string{"ana@exemplo.com", "Ana@Exemplo.com", "exemplo"} {
		if strings.Contains(strings.ToLower(despejo+eventos), strings.ToLower(proibido)) {
			t.Fatalf("contato em claro no banco: %s | %s", despejo, eventos)
		}
	}
	oo, err := repo.OptOutsDe(ctx, dominio.Destinatario{Email: "ana@exemplo.com"})
	if err != nil || !oo[dominio.CanalEmail] {
		t.Fatalf("normalização: opt-out = %v err=%v", oo, err)
	}
	// outra chave não enxerga o Opt-out: por isso a chave é permanente
	outra := postgres.Repositorio{Pool: pool, ChaveOptOut: []byte("outra-chave-com-mais-de-32-caracteres!!")}
	if oo, _ := outra.OptOutsDe(ctx, dominio.Destinatario{Email: "ana@exemplo.com"}); oo[dominio.CanalEmail] {
		t.Fatal("Opt-out visível com outra chave")
	}
}

func TestT47_HistoricoDaRemocaoNoPostgres(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool, ChaveOptOut: chaveOptOutTeste}
	ctx := context.Background()
	_ = repo.RegistrarOptOut(ctx, dominio.CanalSMS, "+5511999998888", "api:sistema-a")
	_ = repo.RegistrarOptOut(ctx, dominio.CanalSMS, "+5511999998888", "api:sistema-a") // idempotente
	if err := repo.RemoverOptOut(ctx, dominio.CanalSMS, "+5511999998888", "api:sistema-b"); err != nil {
		t.Fatal(err)
	}
	var historico string
	_ = pool.QueryRow(ctx, `SELECT string_agg(acao || ' ' || por, ', ' ORDER BY id) FROM opt_out_eventos`).Scan(&historico)
	var restantes int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM opt_outs`).Scan(&restantes)
	if historico != "registrado api:sistema-a, removido api:sistema-b" || restantes != 0 {
		t.Fatalf("histórico = %q, opt-outs restantes = %d", historico, restantes)
	}
}

func TestT53_OptOutCarregadoNaReservaDaFila(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool, ChaveOptOut: chaveOptOutTeste}
	ctx := context.Background()
	p := nova("sistema-a", "k-53")
	p.Destinatario.Telefone = "+5511999998888"
	p.OrdemCanais = []dominio.Canal{dominio.CanalSMS, dominio.CanalEmail}
	p.PrimeiroCanal = dominio.CanalSMS
	if _, err := repo.RegistrarNotificacao(ctx, p); err != nil {
		t.Fatal(err)
	}
	_ = repo.RegistrarOptOut(ctx, dominio.CanalSMS, "+5511999998888", "api:sistema-a")
	res, ok, err := repo.ReservarTentativa(ctx)
	if err != nil || !ok {
		t.Fatalf("reserva: ok=%v err=%v", ok, err)
	}
	if !res.Notificacao.Destinatario.OptOut[dominio.CanalSMS] || res.Notificacao.Destinatario.OptOut[dominio.CanalEmail] {
		t.Fatalf("opt-out na reserva = %v", res.Notificacao.Destinatario.OptOut)
	}
	_ = res.Concluir(ctx, dominio.Decisao{Tentativa: dominio.TentativaFalhou, Notificacao: dominio.NotificacaoFalhou}, dominio.ResultadoEnvio{})
}

// finalizada cria uma Notificação com e-mail, telefone, `nome` e uma Tentativa
// falhou com motivo, no status dado, criada há `dias` dias.
func finalizada(t *testing.T, pool *pgxpool.Pool, chave, status string, dias int) string {
	t.Helper()
	ctx := context.Background()
	repo := postgres.Repositorio{Pool: pool, ChaveOptOut: chaveOptOutTeste}
	n := nova("sistema-a", chave)
	n.Destinatario.Telefone = "+5511999998888"
	n.PrimeiroCanal = dominio.CanalEmail
	reg, err := repo.RegistrarNotificacao(ctx, n)
	if err != nil {
		t.Fatal(err)
	}
	statusTentativa := "falhou"
	if status == "pendente" {
		statusTentativa = "pendente"
	}
	_, err = pool.Exec(ctx, `UPDATE tentativas SET status = $2, motivo = 'sendgrid 400: ana@exemplo.com inválido', categoria_falha = 'contato_invalido' WHERE notificacao_id = $1`, reg.ID, statusTentativa)
	if err == nil {
		_, err = pool.Exec(ctx, `UPDATE notificacoes SET status = $2, criada_em = now() - make_interval(days => $3) WHERE id = $1`, reg.ID, status, dias)
	}
	if err != nil {
		t.Fatal(err)
	}
	return reg.ID
}

type linhaRetencao struct {
	Email, Telefone, Dados, Status, Motivo, Categoria string
	Anonimizada                                       bool
}

func retencaoDe(t *testing.T, pool *pgxpool.Pool, id string) linhaRetencao {
	t.Helper()
	var l linhaRetencao
	err := pool.QueryRow(context.Background(), `
		SELECT n.destinatario_email, n.destinatario_telefone, n.dados::text, n.status, t.motivo, t.categoria_falha, n.anonimizada_em IS NOT NULL
		FROM notificacoes n JOIN tentativas t ON t.notificacao_id = n.id WHERE n.id = $1`, id).
		Scan(&l.Email, &l.Telefone, &l.Dados, &l.Status, &l.Motivo, &l.Categoria, &l.Anonimizada)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func rodarRetencao(t *testing.T, pool *pgxpool.Pool, dias int) int {
	t.Helper()
	uc := aplicacao.RetencaoUseCase{
		Repo:  postgres.Repositorio{Pool: pool, ChaveOptOut: chaveOptOutTeste},
		Prazo: time.Duration(dias) * 24 * time.Hour, Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	n, err := uc.Rodar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestT57_NotificacaoFinalizadaAlemDoPrazoEAnonimizada(t *testing.T) {
	pool := banco(t)
	id := finalizada(t, pool, "k-57", "entregue", 31)
	antes := retencaoDe(t, pool, id)
	if n := rodarRetencao(t, pool, 30); n != 1 {
		t.Fatalf("anonimizadas = %d", n)
	}
	l := retencaoDe(t, pool, id)
	if l.Email != "" || l.Telefone != "" || l.Dados != "{}" || l.Motivo != "" || !l.Anonimizada {
		t.Fatalf("depois: %+v", l)
	}
	if l.Status != antes.Status || l.Categoria != antes.Categoria {
		t.Fatalf("histórico de entrega mudou: antes %+v, depois %+v", antes, l)
	}
}

func TestT58_NotificacaoDentroDoPrazoNaoMuda(t *testing.T) {
	pool := banco(t)
	id := finalizada(t, pool, "k-58", "entregue", 29)
	antes := retencaoDe(t, pool, id)
	rodarRetencao(t, pool, 30)
	if depois := retencaoDe(t, pool, id); depois != antes {
		t.Fatalf("antes %+v, depois %+v", antes, depois)
	}
}

func TestT59_NotificacaoPendenteNaoEAnonimizada(t *testing.T) {
	pool := banco(t)
	id := finalizada(t, pool, "k-59", "pendente", 31)
	if n := rodarRetencao(t, pool, 30); n != 0 {
		t.Fatalf("anonimizadas = %d", n)
	}
	if l := retencaoDe(t, pool, id); l.Email == "" || l.Anonimizada {
		t.Fatalf("pendente anonimizada: %+v", l)
	}
	if n, _ := (postgres.Repositorio{Pool: pool}).PendentesVencidas(context.Background(), time.Now().Add(-30*24*time.Hour)); n != 1 {
		t.Fatalf("pendentes vencidas = %d", n)
	}
}

func TestT60_PrazoConfiguravel(t *testing.T) {
	pool := banco(t)
	id := finalizada(t, pool, "k-60", "falhou", 8)
	if n := rodarRetencao(t, pool, 30); n != 0 {
		t.Fatalf("prazo de 30 dias anonimizou %d", n)
	}
	if n := rodarRetencao(t, pool, 7); n != 1 || !retencaoDe(t, pool, id).Anonimizada {
		t.Fatalf("prazo de 7 dias: %d", n)
	}
}

func TestT61_RodadaRepetidaEIdempotente(t *testing.T) {
	pool := banco(t)
	id := finalizada(t, pool, "k-61", "entregue", 31)
	rodarRetencao(t, pool, 30)
	var primeira time.Time
	_ = pool.QueryRow(context.Background(), `SELECT anonimizada_em FROM notificacoes WHERE id = $1`, id).Scan(&primeira)
	if n := rodarRetencao(t, pool, 30); n != 0 {
		t.Fatalf("segunda rodada anonimizou %d", n)
	}
	var segunda time.Time
	_ = pool.QueryRow(context.Background(), `SELECT anonimizada_em FROM notificacoes WHERE id = $1`, id).Scan(&segunda)
	if !primeira.Equal(segunda) {
		t.Fatalf("anonimizada_em mudou: %s → %s", primeira, segunda)
	}
}

func TestT62_OptOutSobreviveAAnonimizacao(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool, ChaveOptOut: chaveOptOutTeste}
	ctx := context.Background()
	_ = repo.RegistrarOptOut(ctx, dominio.CanalSMS, "+5511999998888", "api:sistema-a")
	finalizada(t, pool, "k-62", "entregue", 31)
	rodarRetencao(t, pool, 30)
	oo, err := repo.OptOutsDe(ctx, dominio.Destinatario{Telefone: "+5511999998888"})
	if err != nil || !oo[dominio.CanalSMS] {
		t.Fatalf("opt-out depois da retenção = %v err=%v", oo, err)
	}
}

func TestT56_ConsultaMostraQuandoFoiAnonimizada(t *testing.T) {
	pool := banco(t)
	id := finalizada(t, pool, "k-56", "entregue", 31)
	repo := postgres.Repositorio{Pool: pool}
	c, _, _ := repo.ConsultarNotificacao(context.Background(), "sistema-a", id)
	if c.AnonimizadaEm != nil {
		t.Fatal("anonimizada_em antes da rodada")
	}
	rodarRetencao(t, pool, 30)
	c, ok, err := repo.ConsultarNotificacao(context.Background(), "sistema-a", id)
	if err != nil || !ok || c.AnonimizadaEm == nil || len(c.Tentativas) != 1 || c.Status != dominio.NotificacaoEntregue {
		t.Fatalf("consulta = %+v ok=%v err=%v", c, ok, err)
	}
}

func TestT73_TextoDaVersaoWhatsAppNoSeedDoPostgres(t *testing.T) {
	repo := postgres.Repositorio{Pool: banco(t)}
	tpl, err := repo.BuscarTemplate(context.Background(), "boas-vindas")
	if err != nil {
		t.Fatal(err)
	}
	c, err := tpl.Preencher(dominio.CanalWhatsApp, map[string]string{"nome": "Ana"})
	if err != nil || c.Corpo != "Olá Ana, sua conta está pronta." || c.ModeloAprovado != "boas_vindas" {
		t.Fatalf("versão whatsapp = %+v err = %v", c, err)
	}
	if dominio.WhatsAppUltraMsg.Servir(c) != nil || dominio.WhatsAppMeta.Servir(c) != nil {
		t.Fatal("o seed não serve aos dois Provedores")
	}
}

func TestT104_TrocasDeConfiguracaoNoPostgresSemSegredo(t *testing.T) {
	pool := banco(t)
	repo := postgres.Repositorio{Pool: pool}
	ctx := context.Background()
	for _, e := range []portas.EventoConfig{
		{Canal: dominio.CanalWhatsApp, Acao: "gravada", Provedor: dominio.WhatsAppZAPI, Por: "ops-a"},
		{Canal: dominio.CanalWhatsApp, Acao: "apagada", Por: "ops-a"},
		{Canal: dominio.CanalWhatsApp, Acao: "webhooks", Provedor: dominio.WhatsAppZAPI, Por: "ana"}, // recadastro (REQ-41, migration 00010)
	} {
		if err := repo.RegistrarEventoConfig(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := pool.Query(ctx, `SELECT canal, acao, coalesce(provedor, ''), por, em IS NOT NULL FROM config_eventos ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var canal, acao, prov, por string
		var comHora bool
		if err := rows.Scan(&canal, &acao, &prov, &por, &comHora); err != nil || !comHora {
			t.Fatalf("linha: %v", err)
		}
		got = append(got, canal+" "+acao+" "+prov+" "+por)
	}
	if strings.Join(got, "|") != "whatsapp gravada zapi ops-a|whatsapp apagada  ops-a|whatsapp webhooks zapi ana" {
		t.Fatalf("eventos = %q", got)
	}
	if err := repo.RegistrarEventoConfig(ctx, portas.EventoConfig{Canal: dominio.CanalWhatsApp, Acao: "trocada", Por: "x"}); err == nil {
		t.Fatal("ação fora de gravada/apagada/webhooks aceita")
	}
}

func TestT139_TemplatesDoWbemPreenchemOTextoDoWhatsApp(t *testing.T) {
	repo := postgres.Repositorio{Pool: banco(t)}
	casos := []struct{ id, variavel, valor string }{
		{"wbem-link-tomador", "link", "https://app.wbem.com.br/j/acesso/?t=abc"},
		{"wbem-otp-tomador", "codigo", "482915"},
	}
	for _, caso := range casos {
		tpl, err := repo.BuscarTemplate(context.Background(), caso.id)
		if err != nil {
			t.Fatalf("%s: %v", caso.id, err)
		}
		if len(tpl.OrdemCanais) != 1 || tpl.OrdemCanais[0] != dominio.CanalWhatsApp || len(tpl.Conteudos) != 1 {
			t.Fatalf("%s: ordem %v e %d versões, quer só WhatsApp", caso.id, tpl.OrdemCanais, len(tpl.Conteudos))
		}
		c, err := tpl.Preencher(dominio.CanalWhatsApp, map[string]string{caso.variavel: caso.valor})
		if err != nil || !strings.Contains(c.Corpo, caso.valor) || !strings.HasPrefix(c.Corpo, "WBem: ") {
			t.Fatalf("%s: corpo = %q err = %v", caso.id, c.Corpo, err)
		}
		if dominio.WhatsAppZAPI.Servir(c) != nil || dominio.WhatsAppUltraMsg.Servir(c) != nil {
			t.Fatalf("%s: o texto não serve à Z-API e ao UltraMsg", caso.id)
		}
	}
}

func TestT140_PedidoSemAVariavelOuPorOutroCanalERecusado(t *testing.T) {
	pool := banco(t)
	uc := aplicacao.SendMessageUseCase{Repo: postgres.Repositorio{Pool: pool}, ProvedorWhatsApp: dominio.WhatsAppZAPI}
	tel := dominio.Destinatario{Telefone: "+5511999998888"}
	_, err := uc.Pedir(context.Background(), aplicacao.PedidoNotificacao{
		SistemaOrigem: "wbem", ChaveIdempotencia: "t140-a", TemplateID: "wbem-otp-tomador",
		Destinatario: tel, Dados: map[string]string{},
	})
	if !errors.Is(err, dominio.ErrVariavelFaltando) {
		t.Fatalf("sem codigo: err = %v, quer variável faltando", err)
	}
	_, err = uc.Pedir(context.Background(), aplicacao.PedidoNotificacao{
		SistemaOrigem: "wbem", ChaveIdempotencia: "t140-b", TemplateID: "wbem-otp-tomador",
		Destinatario: tel, Dados: map[string]string{"codigo": "482915"}, Canais: []dominio.Canal{dominio.CanalSMS},
	})
	if !errors.Is(err, dominio.ErrCanalSemVersao) {
		t.Fatalf("por SMS: err = %v, quer canal sem versão", err)
	}
	if n, ten := contar(t, pool); n != 0 || ten != 0 {
		t.Fatalf("nada devia ser registrado: %d notificações, %d tentativas", n, ten)
	}
}
