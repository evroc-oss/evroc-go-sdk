// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package auth provides OAuth2 authentication for the evroc API.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/config"
	"golang.org/x/oauth2"
)

const (
	// TokenRefreshBuffer is the duration before token expiry when we should refresh.
	// Tokens are refreshed 30 seconds before expiry to avoid race conditions
	// where a token expires between validation and use.
	TokenRefreshBuffer = 30 * time.Second
)

// AuthMetricsRecorder defines the interface for recording auth metrics.
type AuthMetricsRecorder interface {
	RecordTokenRefresh(duration float64)
	RecordTokenRefreshError(duration float64, errorType string)
	RecordInitialAuth(duration float64)
	RecordInitialAuthError(duration float64, errorType string)
}

// Client provides authentication for API requests.
type Client struct {
	tokenSource  oauth2.TokenSource
	httpClient   *http.Client
	oauth2Config *oauth2.Config
	token        *oauth2.Token
	username     string
	password     string
	tokenMu      sync.RWMutex
	metrics      AuthMetricsRecorder
}

// Option is a functional option for configuring the auth Client.
type Option func(*Client) error

// WithHTTPClient sets a custom HTTP client for the auth client.
// This is useful for testing or when you need custom transport behavior.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("httpClient cannot be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithMetrics sets the metrics recorder for the auth client.
func WithMetrics(metrics AuthMetricsRecorder) Option {
	return func(c *Client) error {
		c.metrics = metrics
		return nil
	}
}

// NewClient creates a new authentication client.
func NewClient(ctx context.Context, cfg config.AuthConfig, opts ...Option) (*Client, error) {
	client := &Client{
		username: cfg.Username,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if err := client.buildTokenSource(ctx, cfg); err != nil {
		return nil, err
	}

	return client.initTokenSource()
}

// buildTokenSource creates an oauth2.TokenSource from OIDC config (token or password)
// and sets it on c.tokenSource.
func (c *Client) buildTokenSource(ctx context.Context, cfg config.AuthConfig) error {
	if cfg.ClientID == "" || cfg.TokenURL == "" {
		return fmt.Errorf("OIDC configuration required: provide ClientID and TokenURL")
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "offline_access"}
	} else {
		hasOfflineAccess := false
		for _, s := range scopes {
			if s == "offline_access" {
				hasOfflineAccess = true
				break
			}
		}
		if !hasOfflineAccess {
			scopes = append(scopes, "offline_access")
		}
	}

	c.oauth2Config = &oauth2.Config{
		ClientID: cfg.ClientID,
		Endpoint: oauth2.Endpoint{TokenURL: cfg.TokenURL},
		Scopes:   scopes,
	}

	if c.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)
	}

	// Token/RefreshToken path
	if cfg.Token != "" || cfg.RefreshToken != "" {
		var expiry time.Time
		accessToken := cfg.Token

		if cfg.Token != "" {
			expiry = extractTokenExpiry(cfg.Token)
		} else {
			expiry = time.Now().Add(-time.Hour)
			accessToken = ""
		}

		token := &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: cfg.RefreshToken,
			TokenType:    "Bearer",
			Expiry:       expiry,
		}

		c.tokenMu.Lock()
		c.token = token
		c.tokenMu.Unlock()

		c.tokenSource = c.oauth2Config.TokenSource(ctx, token)
		return nil
	}

	// Password path
	if cfg.Username == "" || cfg.Password == "" {
		return fmt.Errorf("authentication required: provide either (Token/RefreshToken) or (Username + Password)")
	}

	start := time.Now()
	token, err := c.oauth2Config.PasswordCredentialsToken(ctx, c.username, c.password)
	duration := time.Since(start).Seconds()
	if err != nil {
		if c.metrics != nil {
			c.metrics.RecordInitialAuthError(duration, classifyAuthError(err))
		}
		return fmt.Errorf("failed to obtain token from %s using client_id=%s and username=%s: %w",
			c.oauth2Config.Endpoint.TokenURL, c.oauth2Config.ClientID, c.username, err)
	}

	c.tokenMu.Lock()
	c.token = token
	c.tokenMu.Unlock()

	c.tokenSource = c.oauth2Config.TokenSource(ctx, token)
	return nil
}

// initTokenSource fetches the initial token and wraps with metrics if configured.
// This is the single convergence point for all auth methods.
func (c *Client) initTokenSource() (*Client, error) {
	ts := c.tokenSource

	// If we don't have a token yet, fetch one now
	c.tokenMu.RLock()
	needsToken := c.token == nil
	c.tokenMu.RUnlock()

	if needsToken {
		start := time.Now()
		token, err := ts.Token()
		duration := time.Since(start).Seconds()
		if err != nil {
			if c.metrics != nil {
				c.metrics.RecordInitialAuthError(duration, classifyAuthError(err))
			}
			return nil, fmt.Errorf("token source: initial token fetch failed: %w", err)
		}
		if c.metrics != nil {
			c.metrics.RecordInitialAuth(duration)
		}
		c.tokenMu.Lock()
		c.token = token
		c.tokenMu.Unlock()
	} else if c.metrics != nil {
		// Password/token paths already fetched a token during buildTokenSource
		c.metrics.RecordInitialAuth(0)
	}

	// Wrap with metrics
	if c.metrics != nil {
		c.tokenSource = &metricsTokenSource{source: ts, metrics: c.metrics}
	}

	return c, nil
}

// HTTPClient returns an HTTP client that automatically handles OAuth2 authentication.
// The client will automatically add Authorization headers and refresh tokens when needed.
func (c *Client) HTTPClient(ctx context.Context) (*http.Client, error) {
	c.tokenMu.RLock()
	tokenSource := c.tokenSource
	c.tokenMu.RUnlock()

	if tokenSource == nil {
		return nil, fmt.Errorf("token source not initialized")
	}

	if c.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

// Token returns the current access token, refreshing if necessary.
// TokenSource automatically handles token refresh, so this method just
// delegates to it and updates the cached token.
func (c *Client) Token() (string, error) {
	c.tokenMu.RLock()
	tokenSource := c.tokenSource
	c.tokenMu.RUnlock()

	if tokenSource == nil {
		return "", fmt.Errorf("token source not initialized")
	}

	// TokenSource.Token() automatically refreshes if expired
	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get valid token: %w", err)
	}

	// Update cached token
	c.tokenMu.Lock()
	c.token = token
	c.tokenMu.Unlock()

	return token.AccessToken, nil
}

// IsTokenValid checks if the current token is valid and not expired
// Returns true if token is valid with at least 30 seconds remaining.
func (c *Client) IsTokenValid() bool {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()

	if c.token == nil {
		return false
	}

	// Check if token is expired (with buffer to avoid race conditions)
	return c.token.Valid() && time.Until(c.token.Expiry) > TokenRefreshBuffer
}

// ForceRefresh forces a token refresh even if the current token is still valid.
// This can be useful for testing or recovering from authentication issues.
func (c *Client) ForceRefresh(ctx context.Context) error {
	c.tokenMu.RLock()
	ts := c.tokenSource
	c.tokenMu.RUnlock()

	token, err := ts.Token()
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	c.tokenMu.Lock()
	c.token = token
	c.tokenMu.Unlock()
	return nil
}

// extractTokenExpiry extracts the expiry time from a JWT token
// Returns the expiry time, or current time if extraction fails (forcing immediate refresh).
func extractTokenExpiry(tokenString string) time.Time {
	// JWT tokens have 3 parts separated by dots: header.payload.signature
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		// Not a valid JWT, assume expired to force refresh
		return time.Now()
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Can't decode, assume expired to force refresh
		return time.Now()
	}

	// Parse the JSON payload
	var claims struct {
		Exp int64 `json:"exp"` // Expiry time as Unix timestamp
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		// Can't parse, assume expired to force refresh
		return time.Now()
	}

	if claims.Exp == 0 {
		// No expiry in token, assume expired to force refresh
		return time.Now()
	}

	// Convert Unix timestamp to time.Time
	return time.Unix(claims.Exp, 0)
}

// metricsTokenSource wraps an oauth2.TokenSource to capture token refresh metrics
type metricsTokenSource struct {
	source  oauth2.TokenSource
	metrics AuthMetricsRecorder
}

// Token implements oauth2.TokenSource, capturing metrics for token refresh operations
func (m *metricsTokenSource) Token() (*oauth2.Token, error) {
	start := time.Now()
	token, err := m.source.Token()
	duration := time.Since(start).Seconds()

	if err != nil {
		// Record refresh error
		if m.metrics != nil {
			errorType := classifyAuthError(err)
			m.metrics.RecordTokenRefreshError(duration, errorType)
		}
		return nil, err
	}

	// Record successful refresh (only if token was actually refreshed, not just returned from cache)
	// The oauth2 library's TokenSource caches tokens and only refreshes when expired
	if m.metrics != nil && duration > 0.01 { // Only count if operation took > 10ms (actual refresh)
		m.metrics.RecordTokenRefresh(duration)
	}

	return token, nil
}

// classifyAuthError classifies authentication errors for metrics
func classifyAuthError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "invalid_grant"):
		return "invalid_credentials"
	case strings.Contains(errStr, "invalid_client"):
		return "invalid_client"
	case strings.Contains(errStr, "unauthorized_client"):
		return "unauthorized_client"
	case strings.Contains(errStr, "unsupported_grant_type"):
		return "unsupported_grant"
	case strings.Contains(errStr, "invalid_scope"):
		return "invalid_scope"
	case strings.Contains(errStr, "context deadline exceeded") || strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host"):
		return "network_error"
	default:
		return "unknown_error"
	}
}
