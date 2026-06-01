// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
)

func generateTestJWK(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	jwk := jose.JSONWebKey{Key: key, KeyID: "test-kid", Algorithm: "RS256", Use: "sig"}
	data, err := json.Marshal(jwk)
	if err != nil {
		t.Fatalf("marshal JWK: %v", err)
	}
	return data
}

func writeTestJWKFile(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jwk")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write JWK file: %v", err)
	}
	return path
}

func generateTestJWKBase64(t *testing.T) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString(generateTestJWK(t))
}

func createMockJWTTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.FormValue("grant_type") != "client_credentials" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "unsupported_grant_type"})
			return
		}

		assertionType := r.FormValue("client_assertion_type")
		if assertionType != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_client"})
			return
		}

		assertion := r.FormValue("client_assertion")
		if assertion == "" || !strings.Contains(assertion, ".") {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_client"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "sa-access-token-12345",
			"token_type":   "Bearer",
			"expires_in":   300,
		})
	}))
}

func TestJWTTokenSource_TokenFromFile(t *testing.T) {
	server := createMockJWTTokenServer(t)
	defer server.Close()

	jwkFile := writeTestJWKFile(t, generateTestJWK(t))

	ts, err := newJWTTokenSource(context.Background(), "sa1_my-project", server.URL, jwkFile, server.Client())
	if err != nil {
		t.Fatalf("newJWTTokenSource: %v", err)
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token(): %v", err)
	}

	if token.AccessToken != "sa-access-token-12345" {
		t.Errorf("got access_token=%q, want %q", token.AccessToken, "sa-access-token-12345")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("got token_type=%q, want %q", token.TokenType, "Bearer")
	}
}

func TestJWTTokenSource_TokenFromBase64(t *testing.T) {
	server := createMockJWTTokenServer(t)
	defer server.Close()

	b64 := generateTestJWKBase64(t)

	ts, err := newJWTTokenSource(context.Background(), "sa1_my-project", server.URL, b64, server.Client())
	if err != nil {
		t.Fatalf("newJWTTokenSource: %v", err)
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token(): %v", err)
	}

	if token.AccessToken != "sa-access-token-12345" {
		t.Errorf("got access_token=%q, want %q", token.AccessToken, "sa-access-token-12345")
	}
}

func TestJWTTokenSource_Caching(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "cached-token",
			"token_type":   "Bearer",
			"expires_in":   300,
		})
	}))
	defer server.Close()

	jwkFile := writeTestJWKFile(t, generateTestJWK(t))
	ts, err := newJWTTokenSource(context.Background(), "sa1_proj", server.URL, jwkFile, server.Client())
	if err != nil {
		t.Fatalf("newJWTTokenSource: %v", err)
	}

	tok1, err := ts.Token()
	if err != nil {
		t.Fatalf("first Token(): %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 server call after first Token(), got %d", calls.Load())
	}

	tok2, err := ts.Token()
	if err != nil {
		t.Fatalf("second Token(): %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 server call after second Token() (cached), got %d", calls.Load())
	}
	if tok1.AccessToken != tok2.AccessToken {
		t.Errorf("cached token mismatch: %q != %q", tok1.AccessToken, tok2.AccessToken)
	}
}

func TestJWTTokenSource_CacheExpiry(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": fmt.Sprintf("token-%d", n),
			"token_type":   "Bearer",
			"expires_in":   300,
		})
	}))
	defer server.Close()

	jwkFile := writeTestJWKFile(t, generateTestJWK(t))
	ts, err := newJWTTokenSource(context.Background(), "sa1_proj", server.URL, jwkFile, server.Client())
	if err != nil {
		t.Fatalf("newJWTTokenSource: %v", err)
	}

	_, err = ts.Token()
	if err != nil {
		t.Fatalf("first Token(): %v", err)
	}

	ts.mu.Lock()
	ts.token = &oauth2.Token{
		AccessToken: "expired",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(-time.Minute),
	}
	ts.mu.Unlock()

	tok, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() after expiry: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 server calls after expiry, got %d", calls.Load())
	}
	if tok.AccessToken != "token-2" {
		t.Errorf("got access_token=%q, want %q", tok.AccessToken, "token-2")
	}
}

func TestJWTTokenSource_InvalidSecret(t *testing.T) {
	_, err := newJWTTokenSource(context.Background(), "sa1_proj", "https://example.com/token", "not-valid-base64-!!!", &http.Client{})
	if err == nil {
		t.Fatal("expected error for invalid secret, got nil")
	}
}

func TestJWTTokenSource_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer server.Close()

	jwkFile := writeTestJWKFile(t, generateTestJWK(t))
	ts, err := newJWTTokenSource(context.Background(), "sa1_proj", server.URL, jwkFile, server.Client())
	if err != nil {
		t.Fatalf("newJWTTokenSource: %v", err)
	}

	_, err = ts.Token()
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention 401, got: %v", err)
	}
}

func TestLoadJWK_File(t *testing.T) {
	jwkFile := writeTestJWKFile(t, generateTestJWK(t))
	jwk, err := loadJWK(jwkFile)
	if err != nil {
		t.Fatalf("loadJWK with file: %v", err)
	}
	if jwk.KeyID != "test-kid" {
		t.Errorf("got kid=%q, want %q", jwk.KeyID, "test-kid")
	}
}

func TestLoadJWK_Base64(t *testing.T) {
	b64 := generateTestJWKBase64(t)
	jwk, err := loadJWK(b64)
	if err != nil {
		t.Fatalf("loadJWK with base64: %v", err)
	}
	if jwk.KeyID != "test-kid" {
		t.Errorf("got kid=%q, want %q", jwk.KeyID, "test-kid")
	}
}

func TestLoadJWK_NonexistentFile(t *testing.T) {
	_, err := loadJWK("/nonexistent/path/key.jwk")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
