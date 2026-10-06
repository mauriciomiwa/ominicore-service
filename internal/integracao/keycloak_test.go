package integracao

import (
	"context"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

// raizDoRepo é a raiz do repositório, a partir deste arquivo.
func raizDoRepo() string {
	_, arquivo, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(arquivo), "..", "..")
}

// keycloakReal sobe o Keycloak da casa (mesma versão do aviva2.0) com o realm
// local e a biblioteca do Console em /tmp/lib-console.sh.
func keycloakReal(t *testing.T) *testcontainers.DockerContainer {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	raiz := raizDoRepo()
	c, err := testcontainers.Run(context.Background(), "quay.io/keycloak/keycloak:26.3",
		testcontainers.WithEnv(map[string]string{"KC_BOOTSTRAP_ADMIN_USERNAME": "admin", "KC_BOOTSTRAP_ADMIN_PASSWORD": "admin"}),
		testcontainers.WithCmd("start-dev", "--import-realm"),
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithFiles(
			testcontainers.ContainerFile{HostFilePath: filepath.Join(raiz, "deploy/local/keycloak/ominicore-realm.json"), ContainerFilePath: "/opt/keycloak/data/import/ominicore-realm.json", FileMode: 0o644},
			testcontainers.ContainerFile{HostFilePath: filepath.Join(raiz, "deploy/aviva2/keycloak/lib-console.sh"), ContainerFilePath: "/tmp/lib-console.sh", FileMode: 0o644},
		),
		testcontainers.WithWaitStrategyAndDeadline(3*time.Minute, wait.ForHTTP("/realms/ominicore").WithPort("8080/tcp")),
	)
	testcontainers.CleanupContainer(t, c)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// kcadm roda um trecho de bash no container com KC=(kcadm.sh) já autenticado
// no realm master e a biblioteca do Console carregada.
func kcadm(t *testing.T, c *testcontainers.DockerContainer, script string) string {
	t.Helper()
	completo := `set -euo pipefail; KC=(/opt/keycloak/bin/kcadm.sh); . /tmp/lib-console.sh
"${KC[@]}" config credentials --server http://localhost:8080 --realm master --user admin --password admin >/dev/null
` + script
	codigo, saida, err := c.Exec(context.Background(), []string{"bash", "-c", completo}, tcexec.Multiplexed())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(saida)
	if codigo != 0 {
		t.Fatalf("script saiu com %d:\n%s", codigo, b)
	}
	return string(b)
}

func TestT118_OperadorNovoConfiguraOOTPNoPrimeiroLogin(t *testing.T) {
	c := keycloakReal(t)
	// duas vezes cada: os scripts são idempotentes
	kcadm(t, c, `configurar_console ominicore https://ominicore.flipapi.com.br http://localhost:5173
configurar_console ominicore https://ominicore.flipapi.com.br http://localhost:5173
criar_operador ominicore ana ana@exemplo.com
criar_operador ominicore ana ana@exemplo.com`)

	usuario := kcadm(t, c, `ID=$("${KC[@]}" get users -r ominicore -q username=ana -q exact=true --fields id --format csv --noquotes)
"${KC[@]}" get users/$ID -r ominicore --fields requiredActions --format csv --noquotes
"${KC[@]}" get users/$ID/role-mappings/realm -r ominicore --fields name --format csv --noquotes`)
	for _, quero := range []string{"CONFIGURE_TOTP", "UPDATE_PASSWORD", "ominicore-config"} {
		if !strings.Contains(usuario, quero) {
			t.Fatalf("ana sem %s:\n%s", quero, usuario)
		}
	}

	fluxo := kcadm(t, c, `"${KC[@]}" get authentication/flows/navegador-otp/executions -r ominicore --fields providerId,requirement,level --format csv --noquotes`)
	if !strings.Contains(fluxo, "auth-otp-form,REQUIRED,2") {
		t.Fatalf("OTP não é obrigatório no fluxo do Console:\n%s", fluxo)
	}
	// o OTP vem depois da senha: ele precisa do usuário já identificado
	if senha, otp := strings.Index(fluxo, "auth-username-password-form"), strings.Index(fluxo, "auth-otp-form"); senha < 0 || otp < senha {
		t.Fatalf("o OTP precisa vir depois do formulário de senha:\n%s", fluxo)
	}
	if strings.Count(fluxo, "conditional-user-configured,DISABLED") != 1 {
		t.Fatalf("só a condição do subfluxo do OTP deve ser desligada:\n%s", fluxo)
	}

	client := kcadm(t, c, `ID=$(id_do_client ominicore ominicore-console)
FLUXO=$("${KC[@]}" get authentication/flows -r ominicore --fields id,alias --format csv --noquotes | grep ',navegador-otp$' | cut -d, -f1)
echo "fluxo=$FLUXO"
"${KC[@]}" get clients/$ID -r ominicore`)
	for _, quero := range []string{
		`"publicClient" : true`, `"directAccessGrantsEnabled" : false`, `"pkce.code.challenge.method" : "S256"`,
		`"https://ominicore.flipapi.com.br/*"`, `"included.custom.audience" : "ominicore"`,
	} {
		if !strings.Contains(client, quero) {
			t.Fatalf("client do Console sem %s:\n%s", quero, client)
		}
	}
	fluxoID := strings.TrimPrefix(strings.SplitN(client, "\n", 2)[0], "fluxo=")
	if fluxoID == "" || !strings.Contains(client, `"browser" : "`+fluxoID+`"`) {
		t.Fatalf("override do login não aponta para navegador-otp (%q):\n%s", fluxoID, client)
	}
	if strings.Count(client, "audience-ominicore") != 1 {
		t.Fatalf("mapper de audiência duplicado:\n%s", client)
	}
}
