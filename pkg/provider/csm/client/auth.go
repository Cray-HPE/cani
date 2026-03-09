package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// tokenResponse is the minimal JSON shape returned by Keycloak.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

// fetchToken performs an OAuth2 password-credentials grant against
// the Keycloak instance at the given host. It returns the raw
// access token string.
func fetchToken(httpClient *http.Client, host, username, password string) (string, error) {
	tokenURL := fmt.Sprintf("https://%s/keycloak/realms/shasta/protocol/openid-connect/token", host)

	form := url.Values{
		"grant_type": {"password"},
		"client_id":  {"shasta"},
		"scope":      {"openid"},
		"username":   {username},
		"password":   {password},
	}

	resp, err := httpClient.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("requesting token from %s: %w", host, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response from %s", host)
	}

	return tok.AccessToken, nil
}
