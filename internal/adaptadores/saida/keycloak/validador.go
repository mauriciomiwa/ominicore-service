// Package keycloak valida o token dos Sistemas de origem (REQ-1) pelo JWKS do
// realm do OminiCore, com as chaves em cache (Redis).
package keycloak

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ominicore/internal/portas"
)

// ErrNaoAutenticado cobre qualquer token recusado; o motivo vai no wrap.
var ErrNaoAutenticado = errors.New("não autenticado")

// Cache guarda o JWKS entre chamadas; o Redis implementa em produção.
type Cache interface {
	Get(ctx context.Context, chave string) ([]byte, bool, error)
	Set(ctx context.Context, chave string, valor []byte, ttl time.Duration) error
}

// Validador confere assinatura (RS256), emissor, público e validade.
type Validador struct {
	Emissor  string // ex.: http://keycloak:8080/realms/ominicore
	Publico  string // "ominicore"
	JWKSURL  string // padrão: <Emissor>/protocol/openid-connect/certs
	Cache    Cache
	HTTP     *http.Client
	TTLCache time.Duration
	// ReleituraMinima limita a releitura do JWKS por kid desconhecido: sem
	// ela, tokens com kid inventado viram uma chamada ao Keycloak cada.
	ReleituraMinima time.Duration

	mu            sync.Mutex
	ultimaLeitura time.Time
}

// Novo monta o Validador com os padrões do Keycloak.
func Novo(emissor, publico string, cache Cache) *Validador {
	return &Validador{
		Emissor: emissor, Publico: publico, Cache: cache,
		JWKSURL:  emissor + "/protocol/openid-connect/certs",
		HTTP:     &http.Client{Timeout: 5 * time.Second},
		TTLCache: 10 * time.Minute,

		ReleituraMinima: time.Minute,
	}
}

// NovoComBackchannel confere `iss` = emissor, mas baixa o JWKS por realmInterno
// (ex.: http://keycloak:8080/realms/ominicore), sem passar pelo host público.
func NovoComBackchannel(emissor, realmInterno, publico string, cache Cache) *Validador {
	v := Novo(emissor, publico, cache)
	v.JWKSURL = realmInterno + "/protocol/openid-connect/certs"
	return v
}

// Identificar valida o token e devolve o client que o pediu (`azp`), a pessoa
// (`preferred_username`, quando o token é de um usuário) e os papéis do realm
// (`realm_access.roles`), que dão permissões como a do Operador (REQ-25).
func (v *Validador) Identificar(ctx context.Context, token string) (portas.Identidade, error) {
	claims := struct {
		jwt.RegisteredClaims
		AZP         string `json:"azp"`
		Usuario     string `json:"preferred_username"`
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
	}{}
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		return v.chave(ctx, kid)
	},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.Emissor),
		jwt.WithAudience(v.Publico),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return portas.Identidade{}, fmt.Errorf("%w: %w", ErrNaoAutenticado, err)
	}
	if claims.AZP == "" {
		return portas.Identidade{}, fmt.Errorf("%w: token sem azp", ErrNaoAutenticado)
	}
	return portas.Identidade{Client: claims.AZP, Usuario: claims.Usuario, Papeis: claims.RealmAccess.Roles}, nil
}

// chave procura o kid no JWKS em cache; kid desconhecido força nova leitura
// (rotação de chave no Keycloak), no máximo uma a cada ReleituraMinima.
func (v *Validador) chave(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	chaveCache := "ominicore:jwks:" + v.JWKSURL
	if bruto, ok, err := v.Cache.Get(ctx, chaveCache); err == nil && ok {
		k, err := procurar(bruto, kid)
		if err == nil || !v.podeReler() {
			return k, err
		}
	}
	bruto, err := v.baixar(ctx)
	if err != nil {
		return nil, err
	}
	_ = v.Cache.Set(ctx, chaveCache, bruto, v.TTLCache) // cache é otimização: falhar aqui não recusa o token
	return procurar(bruto, kid)
}

func (v *Validador) podeReler() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return time.Since(v.ultimaLeitura) >= v.ReleituraMinima
}

func (v *Validador) baixar(ctx context.Context) ([]byte, error) {
	v.mu.Lock()
	v.ultimaLeitura = time.Now()
	v.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.JWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("baixar jwks: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("baixar jwks: status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

type jwks struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		Use string `json:"use"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func procurar(bruto []byte, kid string) (*rsa.PublicKey, error) {
	var conjunto jwks
	if err := json.Unmarshal(bruto, &conjunto); err != nil {
		return nil, fmt.Errorf("jwks inválido: %w", err)
	}
	for _, k := range conjunto.Keys {
		if k.Kid != kid || k.Kty != "RSA" || (k.Use != "" && k.Use != "sig") {
			continue
		}
		n, errN := base64.RawURLEncoding.DecodeString(k.N)
		e, errE := base64.RawURLEncoding.DecodeString(k.E)
		if errN != nil || errE != nil {
			return nil, errors.New("jwks com chave malformada")
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}, nil
	}
	return nil, fmt.Errorf("kid %q fora do jwks", kid)
}
