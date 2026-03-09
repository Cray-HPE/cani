package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchToken(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("grant_type") != "password" {
			t.Errorf("expected grant_type=password, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("client_id") != "shasta" {
			t.Errorf("expected client_id=shasta, got %s", r.FormValue("client_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "test-token-123"})
	}))
	defer srv.Close()

	host := srv.URL[len("https://"):]
	tok, err := fetchToken(srv.Client(), host, "admin", "secret")
	if err != nil {
		t.Fatalf("fetchToken: %v", err)
	}
	if tok != "test-token-123" {
		t.Errorf("expected test-token-123, got %s", tok)
	}
}

func TestFetchTokenBadStatus(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	host := srv.URL[len("https://"):]
	_, err := fetchToken(srv.Client(), host, "bad", "creds")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestApplyDefaults(t *testing.T) {
	t.Run("simulation mode", func(t *testing.T) {
		o := Options{UseSimulation: true}
		o.applyDefaults()
		if o.InsecureSkipVerify != true {
			t.Error("expected InsecureSkipVerify=true in simulation mode")
		}
		if o.ProviderHost != "localhost:8443" {
			t.Errorf("expected localhost:8443, got %s", o.ProviderHost)
		}
		if o.BaseURLSLS != "https://localhost:8443/apis/sls/v1" {
			t.Errorf("unexpected SLS URL: %s", o.BaseURLSLS)
		}
		if o.BaseURLHSM != "https://localhost:8443/apis/smd/hsm/v2" {
			t.Errorf("unexpected HSM URL: %s", o.BaseURLHSM)
		}
	})

	t.Run("custom host", func(t *testing.T) {
		o := Options{ProviderHost: "myhost.example.com"}
		o.applyDefaults()
		if o.BaseURLSLS != "https://myhost.example.com/apis/sls/v1" {
			t.Errorf("unexpected SLS URL: %s", o.BaseURLSLS)
		}
	})

	t.Run("explicit URLs preserved", func(t *testing.T) {
		o := Options{
			ProviderHost: "ignored",
			BaseURLSLS:   "http://custom-sls",
			BaseURLHSM:   "http://custom-hsm",
		}
		o.applyDefaults()
		if o.BaseURLSLS != "http://custom-sls" {
			t.Errorf("expected custom SLS URL preserved, got %s", o.BaseURLSLS)
		}
		if o.BaseURLHSM != "http://custom-hsm" {
			t.Errorf("expected custom HSM URL preserved, got %s", o.BaseURLHSM)
		}
	})
}
