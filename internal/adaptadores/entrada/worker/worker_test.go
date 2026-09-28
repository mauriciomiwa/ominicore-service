package worker_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"ominicore/internal/adaptadores/entrada/worker"
	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/adaptadores/saida/sendgrid"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/logs"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/memoria"
	"ominicore/internal/testes/sendgridfake"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

const apiKey = "SG.chave-secreta-do-teste"

// fluxo liga API, worker e SendGrid fake sobre o repositório em memória, com
// o logger de produção escrevendo num buffer.
type fluxo struct {
	repo  *memoria.Repositorio
	sg    *sendgridfake.Fake
	realm *emissor.Emissor
	api   http.Handler
	log   *bufferSeguro
}

type bufferSeguro struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *bufferSeguro) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *bufferSeguro) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func novoFluxo(t *testing.T) *fluxo {
	f := &fluxo{repo: memoria.Novo(), sg: sendgridfake.Novo(t), realm: emissor.Novo(t), log: &bufferSeguro{}}
	log := logs.Novo(f.log, slog.LevelDebug)
	f.api = entradahttp.Servidor{
		Auth:   keycloak.Novo(f.realm.URL, "ominicore", &emissor.Cache{}),
		Enviar: aplicacao.SendMessageUseCase{Repo: f.repo},
		Log:    log,
	}.Handler()
	uc := aplicacao.DespacharUseCase{
		Repo:       f.repo,
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{dominio.CanalEmail: sendgrid.Novo(f.sg.URL, apiKey, "no-reply@exemplo.com")},
		Log:        log,
	}
	ctx, cancelar := context.WithCancel(context.Background())
	t.Cleanup(cancelar)
	go worker.Worker{Despachante: uc, Intervalo: 50 * time.Millisecond, Log: log}.Rodar(ctx)
	return f
}

func (f *fluxo) pedir(t *testing.T, corpo string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/notificacoes", strings.NewReader(corpo))
	req.Header.Set("Authorization", "Bearer "+f.realm.Valido(t, "sistema-a"))
	req.Header.Set("Idempotency-Key", "abc-1")
	rec := httptest.NewRecorder()
	f.api.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("pedido: status %d: %s", rec.Code, rec.Body)
	}
}

func (f *fluxo) esperarStatus(t *testing.T, quero dominio.StatusTentativa) {
	t.Helper()
	fim := time.Now().Add(3 * time.Second)
	for time.Now().Before(fim) {
		if f.repo.StatusTentativa(0) == quero {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("tentativa não chegou a %s (está %q)", quero, f.repo.StatusTentativa(0))
}

const pedidoAna = `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com"},"dados":{"nome":"Ana Souza"}}`

func TestT10_TentativaDeEmailEEnviada(t *testing.T) {
	f := novoFluxo(t)
	f.pedir(t, pedidoAna)
	email := f.sg.Esperar(t, 1, 3*time.Second)[0]
	if email.Para != "ana@exemplo.com" || email.Assunto != "Bem-vindo, Ana Souza" || email.Corpo != "Olá Ana Souza, sua conta está pronta." {
		t.Fatalf("e-mail = %+v", email)
	}
	if email.Authorization != "Bearer "+apiKey {
		t.Fatalf("authorization = %q", email.Authorization)
	}
	f.esperarStatus(t, dominio.TentativaEnviada)
}

func TestT14_LogSemCredencialNemDadoPessoal(t *testing.T) {
	f := novoFluxo(t)
	f.sg.Responder(http.StatusBadRequest) // o Provedor recusa: o motivo cita o endereço
	f.pedir(t, pedidoAna)
	f.sg.Esperar(t, 1, 3*time.Second)
	f.esperarStatus(t, dominio.TentativaFalhou)
	saida := f.log.String()
	if !strings.Contains(saida, "tentativa concluída") {
		t.Fatalf("o fluxo não logou a conclusão; log:\n%s", saida)
	}
	for _, proibido := range []string{apiKey, "ana@exemplo.com", "Ana Souza", "sua conta está pronta", "The to address"} {
		if strings.Contains(saida, proibido) {
			t.Errorf("log contém %q:\n%s", proibido, saida)
		}
	}
}

func TestLoggerRedigeCamposSensiveis(t *testing.T) {
	var b bytes.Buffer
	logs.Novo(&b, slog.LevelInfo).Info("x", "email", "ana@exemplo.com", "api_key", apiKey, "canal", "email")
	if strings.Contains(b.String(), "ana@exemplo.com") || strings.Contains(b.String(), apiKey) || !strings.Contains(b.String(), `"canal":"email"`) {
		t.Fatalf("log = %s", b.String())
	}
}
