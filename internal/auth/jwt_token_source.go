// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
)

type jwtClaims struct {
	JTI string `json:"jti"`
	ISS string `json:"iss"`
	SUB string `json:"sub"`
	AUD string `json:"aud"`
	IAT int64  `json:"iat"`
	EXP int64  `json:"exp"`
	NBF int64  `json:"nbf"`
}

// jwtTokenSource implements oauth2.TokenSource using a signed JWT client assertion
// exchanged via the client_credentials grant.
type jwtTokenSource struct {
	ctx        context.Context
	clientID   string
	tokenURL   string
	jwk        *jose.JSONWebKey
	httpClient *http.Client

	mu    sync.Mutex
	token *oauth2.Token
}

func loadJWK(input string) (*jose.JSONWebKey, error) {
	var data []byte

	trimmed := strings.TrimSpace(input)

	// If the file exists on disk, read it; otherwise decode as base64
	if _, err := os.Stat(trimmed); err == nil {
		fileData, err := os.ReadFile(trimmed)
		if err != nil {
			return nil, fmt.Errorf("read secret file: %w", err)
		}
		data = fileData
	} else {
		decoded, err := decodeBase64(trimmed)
		if err != nil {
			return nil, fmt.Errorf("secret is neither a readable file nor valid base64: %w", err)
		}
		data = decoded
	}

	var jwk jose.JSONWebKey
	if err := json.Unmarshal(data, &jwk); err != nil {
		return nil, fmt.Errorf("parse JWK: %w", err)
	}

	if !jwk.Valid() {
		return nil, fmt.Errorf("invalid JWK")
	}

	return &jwk, nil
}

func decodeBase64(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		if decoded, err := enc.DecodeString(s); err == nil {
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("not valid base64 (tried standard, URL-safe, with and without padding)")
}

func newJWTTokenSource(ctx context.Context, clientID, tokenURL, jwkInput string, httpClient *http.Client) (*jwtTokenSource, error) {
	jwk, err := loadJWK(jwkInput)
	if err != nil {
		return nil, fmt.Errorf("load service account JWK: %w", err)
	}

	return &jwtTokenSource{
		ctx:        ctx,
		clientID:   clientID,
		tokenURL:   tokenURL,
		jwk:        jwk,
		httpClient: httpClient,
	}, nil
}

func (s *jwtTokenSource) signAssertion() (string, error) {
	opts := &jose.SignerOptions{}
	if s.jwk.KeyID != "" {
		opts.WithHeader("kid", s.jwk.KeyID)
	}
	opts.WithType("JWT")

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: s.jwk.Key},
		opts,
	)
	if err != nil {
		return "", fmt.Errorf("create signer: %w", err)
	}

	now := time.Now().Unix()
	jti := make([]byte, 16)
	rand.Read(jti)
	claims := jwtClaims{
		JTI: hex.EncodeToString(jti),
		ISS: s.clientID,
		SUB: s.clientID,
		AUD: s.tokenURL,
		IAT: now,
		EXP: now + 300,
		NBF: now,
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	jws, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	compact, err := jws.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("serialize JWT: %w", err)
	}

	return compact, nil
}

// Token implements oauth2.TokenSource. It returns a cached token if still valid,
// otherwise signs a fresh JWT assertion and exchanges it for a new access token.
func (s *jwtTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != nil && time.Until(s.token.Expiry) > TokenRefreshBuffer {
		return s.token, nil
	}

	token, err := s.fetchToken()
	if err != nil {
		return nil, err
	}
	s.token = token
	return token, nil
}

func (s *jwtTokenSource) fetchToken() (*oauth2.Token, error) {
	assertion, err := s.signAssertion()
	if err != nil {
		return nil, err
	}

	form := url.Values{
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {assertion},
		"client_id":             {s.clientID},
	}

	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned empty access_token")
	}

	return &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}
