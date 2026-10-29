package integracao

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"

	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/portas"
	"ominicore/internal/testes/emissor"
	"ominicore/internal/testes/memoria"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

// openbaoReal sobe o OpenBao da casa (mesma versão do aviva2.0) em modo dev,
// com o KV v2 montado em kv/ como em hml e prod, e devolve o cofre do OminiCore.
func openbaoReal(t *testing.T) *openbao.Cofre {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	c, err := testcontainers.Run(ctx, "quay.io/openbao/openbao:2.6.3",
		testcontainers.WithCmd("server", "-dev", "-dev-root-token-id=dev-root", "-dev-listen-address=0.0.0.0:8200"),
		testcontainers.WithExposedPorts("8200/tcp"),
		testcontainers.WithWaitStrategyAndDeadline(time.Minute, wait.ForHTTP("/v1/sys/health").WithPort("8200/tcp")),
	)
	testcontainers.CleanupContainer(t, c)
	if err != nil {
		t.Fatal(err)
	}
	codigo, saida, err := c.Exec(ctx, []string{"bao", "secrets", "enable", "-address=http://127.0.0.1:8200", "-path=kv", "kv-v2"},
		tcexec.Multiplexed(), tcexec.WithEnv([]string{"BAO_TOKEN=dev-root"}))
	if err != nil || codigo != 0 {
		b, _ := io.ReadAll(saida)
		t.Fatalf("montar kv/: %d %v %s", codigo, err, b)
	}
	endereco, err := c.PortEndpoint(ctx, "8200/tcp", "http")
	if err != nil {
		t.Fatal(err)
	}
	cofre, err := openbao.Abrir(ctx, openbao.Config{Endereco: endereco, Token: "dev-root", Prefixo: "ominicore"})
	if err != nil {
		t.Fatal(err)
	}
	return cofre
}

// T-98 contra o OpenBao real: o KV v2 apaga só a versão (soft delete), e a
// leitura precisa tratar a versão apagada como Canal desligado (REQ-24).
func TestT98_OperadorDesligaOWhatsAppNoOpenBaoReal(t *testing.T) {
	cofre := openbaoReal(t)
	realm := emissor.Novo(t)
	repo := memoria.Novo()
	atual := aplicacao.NovaConfigWhatsAppAtual(dominio.ConfigWhatsApp{}, func(dominio.ConfigWhatsApp) portas.MessageSenderPort { return nil })
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := entradahttp.Servidor{
		Auth:   keycloak.Novo(realm.URL, "ominicore", &emissor.Cache{}),
		Enviar: aplicacao.SendMessageUseCase{Repo: repo, Estado: atual},
		Config: &aplicacao.ConfigurarUseCase{Cofre: cofre, Registro: repo, Atual: atual, Log: log},
		Log:    log,
	}.Handler()
	operador := realm.Token(t, jwt.MapClaims{
		"iss": realm.URL, "aud": []string{"ominicore"}, "azp": "ops-a",
		"realm_access": map[string]any{"roles": []string{entradahttp.PapelOperador}},
		"exp":          time.Now().Add(5 * time.Minute).Unix(),
	})
	chamar := func(metodo, corpo string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(metodo, "/v1/config/whatsapp", strings.NewReader(corpo))
		req.Header.Set("Authorization", "Bearer "+operador)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	zapi := `{"provedor":"zapi","instance_id":"inst1","token":"tk-zapi-secreto"}`

	for passo, quero := range []struct {
		metodo, corpo string
		status        int
		trecho        string
	}{
		{http.MethodPut, zapi, http.StatusOK, `"provedor":"zapi"`},
		{http.MethodDelete, "", http.StatusNoContent, ""},
		{http.MethodGet, "", http.StatusOK, `"ligado":false`},
		{http.MethodDelete, "", http.StatusNoContent, ""},      // desligar de novo é idempotente
		{http.MethodPut, zapi, http.StatusOK, `"ligado":true`}, // gravar depois de desligar religa
	} {
		rec := chamar(quero.metodo, quero.corpo)
		if rec.Code != quero.status || !strings.Contains(rec.Body.String(), quero.trecho) {
			t.Fatalf("passo %d %s: status = %d, quero %d com %q: %s", passo, quero.metodo, rec.Code, quero.status, quero.trecho, rec.Body)
		}
		if passo == 2 {
			// a releitura (api e worker, REQ-26) passa pelo mesmo LerWhatsApp
			if cfg, err := cofre.LerWhatsApp(t.Context()); err != nil || cfg.Ligado {
				t.Fatalf("releitura depois de desligar: cfg = %+v err = %v", cfg, err)
			}
		}
	}
}
