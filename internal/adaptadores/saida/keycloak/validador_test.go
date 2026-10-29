package keycloak_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/testes/emissor"
)

func TestT2_TokenDeOutroEmissorOuPublicoERecusado(t *testing.T) {
	realm := emissor.Novo(t)
	outro := emissor.Novo(t)
	v := keycloak.Novo(realm.URL, "ominicore", &emissor.Cache{})
	agora := time.Now()
	casos := map[string]string{
		"expirado":     realm.Token(t, jwt.MapClaims{"iss": realm.URL, "aud": "ominicore", "azp": "sistema-a", "exp": agora.Add(-time.Hour).Unix()}),
		"outro realm":  realm.Token(t, jwt.MapClaims{"iss": outro.URL, "aud": "ominicore", "azp": "sistema-a", "exp": agora.Add(time.Hour).Unix()}),
		"sem público":  realm.Token(t, jwt.MapClaims{"iss": realm.URL, "aud": "account", "azp": "sistema-a", "exp": agora.Add(time.Hour).Unix()}),
		"sem exp":      realm.Token(t, jwt.MapClaims{"iss": realm.URL, "aud": "ominicore", "azp": "sistema-a"}),
		"sem azp":      realm.Token(t, jwt.MapClaims{"iss": realm.URL, "aud": "ominicore", "exp": agora.Add(time.Hour).Unix()}),
		"outra chave":  outro.Token(t, jwt.MapClaims{"iss": realm.URL, "aud": "ominicore", "azp": "sistema-a", "exp": agora.Add(time.Hour).Unix()}),
		"não é um jwt": "abc.def.ghi",
	}
	for nome, token := range casos {
		t.Run(nome, func(t *testing.T) {
			if _, err := sistemaOrigem(v, context.Background(), token); !errors.Is(err, keycloak.ErrNaoAutenticado) {
				t.Fatalf("erro = %v, quero ErrNaoAutenticado", err)
			}
		})
	}
}

func TestTokenValidoDevolveAzpEGuardaJWKSNoCache(t *testing.T) {
	realm := emissor.Novo(t)
	cache := &emissor.Cache{}
	v := keycloak.Novo(realm.URL, "ominicore", cache)
	origem, err := sistemaOrigem(v, context.Background(), realm.Valido(t, "sistema-a"))
	if err != nil || origem != "sistema-a" {
		t.Fatalf("origem=%q err=%v", origem, err)
	}
	if _, ok, _ := cache.Get(context.Background(), "ominicore:jwks:"+v.JWKSURL); !ok {
		t.Fatal("jwks não foi para o cache")
	}
}

func TestKidDesconhecidoNaoReleJWKSACadaToken(t *testing.T) {
	realm := emissor.Novo(t)
	v := keycloak.Novo(realm.URL, "ominicore", &emissor.Cache{})
	if _, err := sistemaOrigem(v, context.Background(), realm.Valido(t, "sistema-a")); err != nil {
		t.Fatal(err)
	}
	claims := jwt.MapClaims{"iss": realm.URL, "aud": "ominicore", "azp": "x", "exp": time.Now().Add(time.Hour).Unix()}
	for i := range 20 {
		forjado := realm.TokenComKid(t, fmt.Sprintf("kid-inventado-%d", i), claims)
		if _, err := sistemaOrigem(v, context.Background(), forjado); !errors.Is(err, keycloak.ErrNaoAutenticado) {
			t.Fatalf("kid inventado aceito: %v", err)
		}
	}
	if n := realm.Leituras(); n != 1 {
		t.Fatalf("jwks baixado %d vezes, quero 1", n)
	}
}

func TestBackchannelConfereOEmissorPublicoEBaixaOJWKSPorDentro(t *testing.T) {
	realm := emissor.Novo(t) // faz o papel do http://keycloak:8080/realms/ominicore
	const publico = "https://auth.flipapi.com.br/realms/ominicore"
	v := keycloak.NovoComBackchannel(publico, realm.URL, "ominicore", &emissor.Cache{})
	tok := realm.Token(t, jwt.MapClaims{"iss": publico, "aud": "ominicore", "azp": "sistema-a", "exp": time.Now().Add(time.Hour).Unix()})
	if origem, err := sistemaOrigem(v, context.Background(), tok); err != nil || origem != "sistema-a" {
		t.Fatalf("origem=%q err=%v", origem, err)
	}
	if _, err := sistemaOrigem(v, context.Background(), realm.Valido(t, "sistema-a")); !errors.Is(err, keycloak.ErrNaoAutenticado) {
		t.Fatalf("token com iss do backchannel foi aceito: %v", err)
	}
}

func TestIdentificarDevolveOsPapeisDoRealm(t *testing.T) {
	e := emissor.Novo(t)
	v := keycloak.Novo(e.URL, "ominicore", &emissor.Cache{})
	tok := e.Token(t, jwt.MapClaims{
		"iss": e.URL, "aud": []string{"ominicore"}, "azp": "ops-a",
		"realm_access": map[string]any{"roles": []string{"default-roles-ominicore", "ominicore-config"}},
		"exp":          time.Now().Add(time.Minute).Unix(),
	})
	id, err := v.Identificar(context.Background(), tok)
	if err != nil || id.Client != "ops-a" || id.Usuario != "" || len(id.Papeis) != 2 || id.Papeis[1] != "ominicore-config" {
		t.Fatalf("identidade = %+v err = %v", id, err)
	}
	if id, err := v.Identificar(context.Background(), e.Valido(t, "sistema-a")); err != nil || len(id.Papeis) != 0 {
		t.Fatalf("sem realm_access: identidade = %+v err = %v", id, err)
	}
}

func TestIdentificarDevolveAPessoaDoTokenDeUsuario(t *testing.T) {
	e := emissor.Novo(t)
	v := keycloak.Novo(e.URL, "ominicore", &emissor.Cache{})
	id, err := v.Identificar(context.Background(), e.DeUsuario(t, "ominicore-console", "ana", "ominicore-config"))
	if err != nil || id.Client != "ominicore-console" || id.Usuario != "ana" || len(id.Papeis) != 1 {
		t.Fatalf("identidade = %+v err = %v", id, err)
	}
}

// sistemaOrigem é o azp do token validado (REQ-1).
func sistemaOrigem(v *keycloak.Validador, ctx context.Context, token string) (string, error) {
	id, err := v.Identificar(ctx, token)
	return id.Client, err
}
