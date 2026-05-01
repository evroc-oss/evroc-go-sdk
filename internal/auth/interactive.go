// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// OIDCConfig contains configuration for OIDC authentication.
type OIDCConfig struct {
	TokenURL string
	AuthURL  string
	ClientID string
	Scopes   []string
}

// LoginResult contains the result of an OIDC login flow.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

// GetAuthorizationURL returns the authorization URL for manual browser-based login
// User opens this URL in their browser, completes authentication, and then
// provides the resulting callback URL to NewClientFromManualCallback.
func GetAuthorizationURL(cfg OIDCConfig, verifier, state string) string {
	redirectURI := "http://localhost:8000"

	oauth2Config := &oauth2.Config{
		ClientID:    cfg.ClientID,
		Endpoint:    oauth2.Endpoint{AuthURL: cfg.AuthURL},
		RedirectURL: redirectURI,
		Scopes:      cfg.Scopes,
	}
	return oauth2Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}

// NewClientWithHTTPServer creates a new auth client using an HTTP server to catch the OAuth callback.
// It starts a local HTTP server on localhost:8000, waits for the OAuth redirect, then exchanges
// the authorization code for tokens. Returns an error if the server cannot start (e.g., port in use).
//
// The expectedState parameter should match the state used in GetAuthorizationURL for CSRF protection.
func NewClientWithHTTPServer(ctx context.Context, cfg OIDCConfig, verifier, expectedState string) (*Client, *LoginResult, error) {
	if cfg.ClientID == "" || cfg.TokenURL == "" {
		return nil, nil, fmt.Errorf("OIDC configuration required: provide ClientID and TokenURL")
	}

	redirectURI := "http://localhost:8000"
	oauth2Config := &oauth2.Config{
		ClientID:    cfg.ClientID,
		Endpoint:    oauth2.Endpoint{TokenURL: cfg.TokenURL, AuthURL: cfg.AuthURL},
		RedirectURL: redirectURI,
		Scopes:      cfg.Scopes,
	}

	// Create channels for result communication
	type result struct {
		client *Client
		login  *LoginResult
		err    error
	}
	resultChan := make(chan result, 1)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Check for OAuth errors from the server
		if errorCode := query.Get("error"); errorCode != "" {
			errorDesc := query.Get("error_description")
			errorMsg := fmt.Sprintf("OAuth error: %s", errorCode)
			if errorDesc != "" {
				errorMsg = fmt.Sprintf("OAuth error: %s - %s", errorCode, errorDesc)
			}
			http.Error(w, errorMsg, http.StatusBadRequest)
			resultChan <- result{err: fmt.Errorf("%s", errorMsg)}
			return
		}

		code := query.Get("code")
		state := query.Get("state")

		// Validate state for CSRF protection
		if state != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			resultChan <- result{err: fmt.Errorf("invalid state parameter (CSRF check failed)")}
			return
		}

		if code == "" {
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			resultChan <- result{err: fmt.Errorf("no authorization code in callback")}
			return
		}

		// Exchange code for token
		token, err := oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
			resultChan <- result{err: fmt.Errorf("failed to exchange authorization code: %w", err)}
			return
		}

		// Create login result
		loginResult := &LoginResult{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		}

		// Create auth client
		client := &Client{
			oauth2Config: oauth2Config,
			token:        token,
			tokenSource:  oauth2Config.TokenSource(ctx, token),
			httpClient:   oauth2.NewClient(ctx, oauth2Config.TokenSource(ctx, token)),
		}

		// Send success response to browser
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // Writing HTML response to browser, error not actionable
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Authentication Successful</title>
</head>
<body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
	<h1 style="color: #4CAF50;">✓ Authentication Successful</h1>
	<p>You have successfully authenticated with evroc Cloud.</p>
	<p>You can close this window and return to your terminal.</p>
</body>
</html>`)

		// Send result
		resultChan <- result{client: client, login: loginResult}
	})

	// Try to start server on localhost:8000
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", "localhost:8000")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start HTTP server on localhost:8000 (port may be in use): %w", err)
	}

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in background
	go func() {
		_ = server.Serve(listener) //nolint:errcheck // Server error is not relevant, we close it explicitly
	}()

	// Wait for callback or context cancellation
	select {
	case res := <-resultChan:
		// Shutdown server gracefully (use fresh context since parent may be cancelled)
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx) //nolint:errcheck // Best effort shutdown, error not relevant

		if res.err != nil {
			return nil, nil, res.err
		}
		return res.client, res.login, nil

	case <-ctx.Done():
		// Shutdown server gracefully (use fresh context since parent is cancelled)
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx) //nolint:errcheck // Best effort shutdown, error not relevant
		return nil, nil, fmt.Errorf("authentication cancelled: %w", ctx.Err())
	}
}

// NewClientFromManualCallback creates a new auth client from a manually pasted callback URL
// User must provide the full callback URL after browser redirect.
// The expectedState parameter must match the state used in GetAuthorizationURL for CSRF protection.
func NewClientFromManualCallback(ctx context.Context, cfg OIDCConfig, callbackURL, verifier, expectedState string) (*Client, *LoginResult, error) {
	if cfg.ClientID == "" || cfg.TokenURL == "" {
		return nil, nil, fmt.Errorf("OIDC configuration required: provide ClientID and TokenURL")
	}

	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid callback URL: %w", err)
	}

	query := parsedURL.Query()

	// Check for OAuth errors from the server
	if errorCode := query.Get("error"); errorCode != "" {
		errorDesc := query.Get("error_description")
		if errorDesc != "" {
			return nil, nil, fmt.Errorf("OAuth error: %s - %s", errorCode, errorDesc)
		}
		return nil, nil, fmt.Errorf("OAuth error: %s", errorCode)
	}

	// Validate state parameter for CSRF protection
	returnedState := query.Get("state")
	if returnedState == "" {
		return nil, nil, fmt.Errorf("no state parameter in callback URL")
	}
	if returnedState != expectedState {
		return nil, nil, fmt.Errorf("state mismatch: possible CSRF attack")
	}

	code := query.Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("no authorization code found in callback URL")
	}

	redirectURI := "http://localhost:8000"

	oauth2Config := &oauth2.Config{
		ClientID:    cfg.ClientID,
		Endpoint:    oauth2.Endpoint{TokenURL: cfg.TokenURL, AuthURL: cfg.AuthURL},
		RedirectURL: redirectURI,
		Scopes:      cfg.Scopes,
	}

	token, err := oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	result := &LoginResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	client := &Client{
		oauth2Config: oauth2Config,
		token:        token,
		tokenSource:  oauth2Config.TokenSource(ctx, token),
		httpClient:   oauth2.NewClient(ctx, oauth2Config.TokenSource(ctx, token)),
	}

	return client, result, nil
}

// GenerateVerifier generates a PKCE code verifier.
func GenerateVerifier() string {
	return oauth2.GenerateVerifier()
}

// GenerateState generates a random state string for CSRF protection.
func GenerateState() string {
	return rand.Text()
}
