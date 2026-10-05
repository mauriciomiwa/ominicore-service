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
	"ominicore/internal/adaptadores/saida/twilio"
	"ominicore/internal/adaptadores/saida/ultramsg"
	"ominicore/internal/adaptadores/saida/whatsapp"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/logs"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/memoria"
	"ominicore/internal/testes/metafake"
	"ominicore/internal/testes/sendgridfake"
	"ominicore/internal/testes/twiliofake"
	"ominicore/internal/testes/ultramsgfake"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

const apiKey = "SG.chave-secreta-do-teste"

// fluxo liga API, worker e SendGrid fake sobre o repositório em memória, com
// o logger de produção escrevendo num buffer.
type fluxo struct {
	repo  *memoria.Repositorio
	sg    *sendgridfake.Fake
	tw    *twiliofake.Fake
	meta  *metafake.Fake
	um    *ultramsgfake.Fake
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

func novoFluxo(t *testing.T) *fluxo { return novoFluxoCom(t, dominio.WhatsAppMeta) }

// novoFluxoCom liga o Canal WhatsApp ao Provedor dado (REQ-10).
func novoFluxoCom(t *testing.T, prov dominio.ProvedorWhatsApp) *fluxo {
	f := &fluxo{repo: memoria.Novo(), sg: sendgridfake.Novo(t), tw: twiliofake.Novo(t), meta: metafake.Novo(t), um: ultramsgfake.Novo(t), realm: emissor.Novo(t), log: &bufferSeguro{}}
	log := logs.Novo(f.log, slog.LevelDebug)
	f.api = entradahttp.Servidor{
		Auth:   keycloak.Novo(f.realm.URL, "ominicore", &emissor.Cache{}),
		Enviar: aplicacao.SendMessageUseCase{Repo: f.repo, ProvedorWhatsApp: prov},
		Log:    log,
	}.Handler()
	var wa portas.MessageSenderPort = whatsapp.Novo(f.meta.URL, whatsapp.Credenciais{Token: metafake.Token, PhoneNumberID: metafake.PhoneNumberID})
	if prov == dominio.WhatsAppUltraMsg {
		wa = ultramsg.Novo(f.um.URL, ultramsg.Credenciais{InstanceID: ultramsgfake.InstanceID, Token: ultramsgfake.Token})
	}
	uc := aplicacao.DespacharUseCase{
		Repo: f.repo,
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{
			dominio.CanalEmail: sendgrid.Novo(f.sg.URL, apiKey, "no-reply@exemplo.com"),
			dominio.CanalSMS: twilio.Novo(f.tw.URL, twilio.Credenciais{
				AccountSID: twiliofake.AccountSID, APIKeySID: twiliofake.APIKeySID,
				APIKeySecret: twiliofake.APIKeySecret, MessagingServiceSID: twiliofake.MessagingServiceSID,
			}),
			dominio.CanalWhatsApp: wa,
		},
		Log: log,
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

const pedidoSMS = `{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"Ana Souza"},"canais":["sms"]}`

func TestT19_TentativaDeSMSEEnviada(t *testing.T) {
	f := novoFluxo(t)
	f.pedir(t, pedidoSMS)
	sms := f.tw.Esperar(t, 1, 3*time.Second)[0]
	if sms.Para != "+5511999998888" || sms.Corpo != "Olá Ana Souza, sua conta está pronta." ||
		sms.MessagingService != twiliofake.MessagingServiceSID || !sms.AuthOK {
		t.Fatalf("sms = %+v", sms)
	}
	f.esperarStatus(t, dominio.TentativaEnviada)
}

func TestT20_NumeroInvalidoEFalhaDefinitiva(t *testing.T) {
	f := novoFluxo(t)
	f.tw.Responder(http.StatusBadRequest, 21211)
	f.pedir(t, pedidoSMS)
	f.tw.Esperar(t, 1, 3*time.Second)
	f.esperarStatus(t, dominio.TentativaFalhou)
	time.Sleep(200 * time.Millisecond) // o worker teria tempo de pegar uma Retentativa imediata
	h := f.repo.Historico()
	if len(h) != 1 || !strings.Contains(h[0].Motivo, "21211") {
		t.Fatalf("quero 1 Tentativa falhou com 21211, sem Retentativa: %+v", h)
	}
	if len(f.tw.Recebidos()) != 1 {
		t.Fatalf("o Twilio recebeu %d envios", len(f.tw.Recebidos()))
	}
}

func TestFallbackEmailParaSMSPontaAPontaEmMemoria(t *testing.T) {
	f := novoFluxo(t)
	f.sg.Responder(http.StatusBadRequest)
	f.pedir(t, `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"}}`)
	f.tw.Esperar(t, 1, 3*time.Second)
	fim := time.Now().Add(3 * time.Second)
	for time.Now().Before(fim) && f.repo.StatusTentativa(1) != dominio.TentativaEnviada {
		time.Sleep(20 * time.Millisecond)
	}
	h := f.repo.Historico()
	if len(h) != 2 || h[0].Canal != dominio.CanalEmail || h[0].Status != dominio.TentativaFalhou ||
		h[1].Canal != dominio.CanalSMS || h[1].Status != dominio.TentativaEnviada {
		t.Fatalf("histórico = %+v", h)
	}
}

const pedidoWhatsApp = `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana Souza"},"canais":["whatsapp","sms"]}`

func TestT28_TentativaDeWhatsAppEEnviadaPeloModeloAprovado(t *testing.T) {
	f := novoFluxo(t)
	f.pedir(t, pedidoWhatsApp)
	m := f.meta.Esperar(t, 1, 3*time.Second)[0]
	if m.Modelo != "boas_vindas" || m.Idioma != "pt_BR" || len(m.Parametros) != 1 || m.Parametros[0] != "Ana Souza" ||
		m.Para != "+5511999998888" || !m.AuthOK {
		t.Fatalf("mensagem = %+v", m)
	}
	f.esperarStatus(t, dominio.TentativaEnviada)
	if len(f.tw.Recebidos()) != 0 {
		t.Fatal("sucesso no WhatsApp não deveria chegar ao SMS")
	}
}

func TestT29_NumeroSemWhatsAppPassaAoSMS(t *testing.T) {
	f := novoFluxo(t)
	f.meta.Responder(http.StatusBadRequest, 131026)
	f.pedir(t, pedidoWhatsApp)
	f.tw.Esperar(t, 1, 3*time.Second)
	fim := time.Now().Add(3 * time.Second)
	for time.Now().Before(fim) && f.repo.StatusTentativa(1) != dominio.TentativaEnviada {
		time.Sleep(20 * time.Millisecond)
	}
	h := f.repo.Historico()
	if len(h) != 2 || h[0].Canal != dominio.CanalWhatsApp || h[0].Status != dominio.TentativaFalhou ||
		h[1].Canal != dominio.CanalSMS || h[1].Status != dominio.TentativaEnviada || len(f.meta.Recebidas()) != 1 {
		t.Fatalf("histórico = %+v", h)
	}
}

func TestT30_LimiteDeTaxaAgendaRetentativaNoWhatsApp(t *testing.T) {
	f := novoFluxo(t)
	f.meta.Responder(http.StatusBadRequest, 130429)
	f.pedir(t, pedidoWhatsApp)
	f.meta.Esperar(t, 1, 3*time.Second)
	f.esperarStatus(t, dominio.TentativaFalhou)
	fim := time.Now().Add(time.Second)
	for time.Now().Before(fim) && len(f.repo.Historico()) < 2 {
		time.Sleep(20 * time.Millisecond)
	}
	h := f.repo.Historico()
	if len(h) != 2 || h[1].Canal != dominio.CanalWhatsApp || h[1].Numero != 2 || h[1].Status != dominio.TentativaPendente {
		t.Fatalf("quero Retentativa pendente no WhatsApp: %+v", h)
	}
}

func TestT31_ModeloNaoAprovadoEFalhaDefinitiva(t *testing.T) {
	f := novoFluxo(t)
	f.meta.Responder(http.StatusNotFound, 132001)
	f.pedir(t, `{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["whatsapp"]}`)
	f.meta.Esperar(t, 1, 3*time.Second)
	f.esperarStatus(t, dominio.TentativaFalhou)
	time.Sleep(200 * time.Millisecond)
	h := f.repo.Historico()
	if len(h) != 1 || !strings.Contains(h[0].Motivo, "132001") || len(f.meta.Recebidas()) != 1 {
		t.Fatalf("quero 1 Tentativa falhou com 132001, sem Retentativa: %+v", h)
	}
}

func TestT50_STOPNoTwilioRegistraOptOutDeSMSESegueAOrdem(t *testing.T) {
	f := novoFluxo(t)
	f.tw.Responder(http.StatusBadRequest, 21610)
	f.pedir(t, `{"template_id":"boas-vindas","destinatario":{"email":"ana@exemplo.com","telefone":"+5511999998888"},"dados":{"nome":"Ana"},"canais":["sms","email"]}`)
	f.sg.Esperar(t, 1, 3*time.Second) // seguiu para o Email
	h := f.repo.Historico()
	oo, _ := f.repo.OptOutsDe(context.Background(), dominio.Destinatario{Telefone: "+5511999998888"})
	if len(h) != 2 || h[0].Canal != dominio.CanalSMS || h[0].Categoria != dominio.FalhaOptOut || h[1].Canal != dominio.CanalEmail || !oo[dominio.CanalSMS] {
		t.Fatalf("histórico = %+v, opt-out = %v", h, oo)
	}
	if len(f.tw.Recebidos()) != 1 {
		t.Fatalf("Twilio chamado %d vezes: STOP não tem Retentativa", len(f.tw.Recebidos()))
	}
}

// esperarHistorico aguarda até n Tentativas no histórico, a última fora de pendente.
func (f *fluxo) esperarHistorico(t *testing.T, n int) []dominio.Tentativa {
	t.Helper()
	fim := time.Now().Add(3 * time.Second)
	for time.Now().Before(fim) {
		if h := f.repo.Historico(); len(h) >= n && h[n-1].Status != dominio.TentativaPendente {
			return h
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("histórico não chegou a %d Tentativas concluídas: %+v", n, f.repo.Historico())
	return nil
}

func TestT73_TentativaDeWhatsAppEEnviadaComoTextoPeloUltraMsg(t *testing.T) {
	f := novoFluxoCom(t, dominio.WhatsAppUltraMsg)
	f.pedir(t, pedidoWhatsApp)
	m := f.um.Esperar(t, 1, 3*time.Second)[0]
	if m.Texto != "Olá Ana Souza, sua conta está pronta." || m.Para != "+5511999998888" || !m.TokenOK {
		t.Fatalf("mensagem = %+v", m)
	}
	f.esperarStatus(t, dominio.TentativaEnviada)
	if h := f.repo.Historico(); len(h) != 1 || f.repo.IDProvedor(0) != "42" || len(f.meta.Recebidas()) != 0 || len(f.tw.Recebidos()) != 0 {
		t.Fatalf("histórico = %+v", h)
	}
}

func TestT74_RecusaDoUltraMsgPassaAoSMS(t *testing.T) {
	f := novoFluxoCom(t, dominio.WhatsAppUltraMsg)
	f.um.Responder(http.StatusOK, "Wrong token")
	f.pedir(t, pedidoWhatsApp)
	f.tw.Esperar(t, 1, 3*time.Second)
	h := f.esperarHistorico(t, 2)
	if h[0].Canal != dominio.CanalWhatsApp || h[0].Status != dominio.TentativaFalhou || !strings.Contains(h[0].Motivo, "Wrong token") ||
		h[1].Canal != dominio.CanalSMS || len(f.um.Recebidas()) != 1 {
		t.Fatalf("histórico = %+v", h)
	}
}

func TestT75_UltraMsgIndisponivelAgendaRetentativa(t *testing.T) {
	f := novoFluxoCom(t, dominio.WhatsAppUltraMsg)
	f.um.Responder(http.StatusServiceUnavailable, "")
	f.pedir(t, pedidoWhatsApp)
	f.um.Esperar(t, 1, 3*time.Second)
	f.esperarStatus(t, dominio.TentativaFalhou)
	fim := time.Now().Add(time.Second)
	for time.Now().Before(fim) && len(f.repo.Historico()) < 2 {
		time.Sleep(20 * time.Millisecond)
	}
	h := f.repo.Historico()
	if len(h) != 2 || h[1].Canal != dominio.CanalWhatsApp || h[1].Numero != 2 || h[1].Status != dominio.TentativaPendente {
		t.Fatalf("quero Retentativa pendente no WhatsApp: %+v", h)
	}
}

func TestT76_TextoLongoDemaisNaoChegaAoUltraMsgESegueAOrdem(t *testing.T) {
	f := novoFluxoCom(t, dominio.WhatsAppUltraMsg)
	f.pedir(t, `{"template_id":"boas-vindas","destinatario":{"telefone":"+5511999998888"},"dados":{"nome":"`+strings.Repeat("a", 4100)+`"},"canais":["whatsapp","sms"]}`)
	f.tw.Esperar(t, 1, 3*time.Second)
	h := f.esperarHistorico(t, 2)
	if h[0].Status != dominio.TentativaFalhou || h[1].Canal != dominio.CanalSMS || len(f.um.Recebidas()) != 0 {
		t.Fatalf("histórico = %+v, chamadas ao UltraMsg = %d", h, len(f.um.Recebidas()))
	}
}

func TestT77_TentativaPendenteSemAFormaDoProvedorAtivoSegueAOrdem(t *testing.T) {
	f := novoFluxoCom(t, dominio.WhatsAppUltraMsg)
	// o pedido foi aceito com a Meta; o Provedor virou UltraMsg antes do envio
	// e a versão WhatsApp não tem texto
	tpl := f.repo.Templates["boas-vindas"]
	tpl.Conteudos[dominio.CanalWhatsApp] = dominio.ConteudoCanal{ModeloAprovado: "boas_vindas", Idioma: "pt_BR", Parametros: []string{"nome"}}
	if _, err := (aplicacao.SendMessageUseCase{Repo: f.repo, ProvedorWhatsApp: dominio.WhatsAppMeta}).Pedir(context.Background(), aplicacao.PedidoNotificacao{
		SistemaOrigem: "sistema-a", ChaveIdempotencia: "k-77", TemplateID: "boas-vindas",
		Destinatario: dominio.Destinatario{Telefone: "+5511999998888"}, Dados: map[string]string{"nome": "Ana"},
		Canais: []dominio.Canal{dominio.CanalWhatsApp, dominio.CanalSMS},
	}); err != nil {
		t.Fatal(err)
	}
	f.tw.Esperar(t, 1, 3*time.Second)
	h := f.esperarHistorico(t, 2)
	if h[0].Status != dominio.TentativaFalhou || h[0].Motivo != "versão whatsapp sem texto" || h[1].Canal != dominio.CanalSMS || len(f.um.Recebidas()) != 0 {
		t.Fatalf("histórico = %+v", h)
	}
}
