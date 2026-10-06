// Package emissor simula o realm do Keycloak nos testes: chave RSA, JWKS em
// httptest e emissão de tokens com as claims pedidas.
package emissor

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Emissor é um realm falso; URL é o `iss` esperado.
type Emissor struct {
	URL      string
	chave    *rsa.PrivateKey
	mu       sync.Mutex
	leituras int
}

// Leituras conta quantas vezes o JWKS foi baixado.
func (e *Emissor) Leituras() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.leituras
}

// Novo sobe o JWKS em <URL>/protocol/openid-connect/certs.
func Novo(t *testing.T) *Emissor {
	t.Helper()
	chave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	e := &Emissor{chave: chave}
	jwks := map[string]any{"keys": []map[string]string{{
		"kid": "k1", "kty": "RSA", "use": "sig", "alg": "RS256",
		"n": base64.RawURLEncoding.EncodeToString(chave.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(chave.E)).Bytes()),
	}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/protocol/openid-connect/certs" {
			http.NotFound(w, r)
			return
		}
		e.mu.Lock()
		e.leituras++
		e.mu.Unlock()
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)
	e.URL = srv.URL
	return e
}

// Valido emite um token aceito: iss certo, aud ominicore, azp dado, 5 min.
func (e *Emissor) Valido(t *testing.T, azp string) string {
	return e.Token(t, jwt.MapClaims{"iss": e.URL, "aud": []string{"ominicore"}, "azp": azp, "exp": time.Now().Add(5 * time.Minute).Unix()})
}

// DeUsuario emite o token de uma pessoa que entrou por um client público
// (Console do Operador): azp do client, preferred_username e papéis do realm.
func (e *Emissor) DeUsuario(t *testing.T, azp, usuario string, papeis ...string) string {
	return e.Token(t, jwt.MapClaims{
		"iss": e.URL, "aud": []string{"ominicore"}, "azp": azp, "preferred_username": usuario,
		"realm_access": map[string]any{"roles": papeis}, "exp": time.Now().Add(5 * time.Minute).Unix(),
	})
}

// Token assina as claims com a chave do realm (kid k1).
func (e *Emissor) Token(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	return e.TokenComKid(t, "k1", claims)
}

// TokenComKid assina com a chave do realm, mas anuncia o kid dado.
func (e *Emissor) TokenComKid(t *testing.T, kid string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(e.chave)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Cache é um cache em memória com a interface do adaptador Redis.
type Cache struct {
	mu sync.Mutex
	m  map[string][]byte
}

func (c *Cache) Get(_ context.Context, k string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[k]
	return v, ok, nil
}

func (c *Cache) Set(_ context.Context, k string, v []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = map[string][]byte{}
	}
	c.m[k] = v
	return nil
}
