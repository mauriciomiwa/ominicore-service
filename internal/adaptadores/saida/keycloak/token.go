package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenDeServico pede ao Keycloak um token client_credentials para o próprio
// OminiCore (client `ominicore-app`), usado no login do OpenBao (ADR 0002).
func TokenDeServico(ctx context.Context, emissor, clientID, clientSecret string) (string, error) {
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {clientID}, "client_secret": {clientSecret}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, emissor+"/protocol/openid-connect/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("pedir token ao keycloak: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pedir token ao keycloak: status %d", resp.StatusCode)
	}
	var corpo struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&corpo); err != nil || corpo.AccessToken == "" {
		return "", fmt.Errorf("pedir token ao keycloak: resposta sem access_token")
	}
	return corpo.AccessToken, nil
}
