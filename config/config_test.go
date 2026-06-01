// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	// Set up test environment (t.Setenv automatically cleans up)
	t.Setenv("EVROC_USERNAME", "test@example.com")
	t.Setenv("EVROC_PASSWORD", "testpass")
	t.Setenv("EVROC_PROJECT", "test-project")
	t.Setenv("EVROC_REGION", "test-region")
	t.Setenv("EVROC_ORGANIZATION", "test-org")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Auth.Username != "test@example.com" {
		t.Errorf("Username = %v, want test@example.com", cfg.Auth.Username)
	}
	if cfg.Context.Project != "test-project" {
		t.Errorf("Project = %v, want test-project", cfg.Context.Project)
	}
	if cfg.Context.Region != "test-region" {
		t.Errorf("Region = %v, want test-region", cfg.Context.Region)
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefaults()

	if cfg.Auth.ClientID != "evroc-cli" {
		t.Errorf("Default ClientID = %v, want evroc-cli", cfg.Auth.ClientID)
	}
	if cfg.Auth.TokenURL == "" {
		t.Error("Default TokenURL should be set")
	}
	if cfg.API.BaseURL != "https://api.evroc.com" {
		t.Errorf("Default BaseURL = %v, want https://api.evroc.com", cfg.API.BaseURL)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Auth: AuthConfig{
					Username: "user@example.com",
					Password: "password",
				},
				Context: ContextConfig{
					Project:      "project",
					Region:       "region",
					Organization: "org",
				},
			},
			wantErr: false,
		},
		{
			name: "refresh token only (valid)",
			cfg: &Config{
				Auth: AuthConfig{
					RefreshToken: "refresh-token",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: false,
		},
		{
			name: "access token only (valid)",
			cfg: &Config{
				Auth: AuthConfig{
					Token: "access-token",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: false,
		},
		{
			name: "both tokens (valid)",
			cfg: &Config{
				Auth: AuthConfig{
					Token:        "access-token",
					RefreshToken: "refresh-token",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: false,
		},
		{
			name: "missing project",
			cfg: &Config{
				Auth: AuthConfig{
					Username: "user@example.com",
					Password: "password",
				},
				Context: ContextConfig{
					Region:       "region",
					Organization: "org",
				},
			},
			wantErr: true,
		},
		{
			name: "missing region",
			cfg: &Config{
				Auth: AuthConfig{
					Username: "user@example.com",
					Password: "password",
				},
				Context: ContextConfig{
					Project:      "project",
					Organization: "org",
				},
			},
			wantErr: true,
		},
		{
			name: "username without password or refresh token",
			cfg: &Config{
				Auth: AuthConfig{
					Username: "user@example.com",
				},
				Context: ContextConfig{
					Project:      "project",
					Region:       "region",
					Organization: "org",
				},
			},
			wantErr: true,
		},
		{
			name: "missing organization (should be valid)",
			cfg: &Config{
				Auth: AuthConfig{
					Username: "user@example.com",
					Password: "password",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: false,
		},
		{
			name: "service account with ID and secret (valid)",
			cfg: &Config{
				Auth: AuthConfig{
					ServiceAccountID:     "sa-test",
					ServiceAccountSecret: "secret-key",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: false,
		},
		{
			name: "service account with client_id override (valid)",
			cfg: &Config{
				Auth: AuthConfig{
					ClientID:             "custom-client-id",
					ServiceAccountSecret: "secret-key",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: false,
		},
		{
			name: "service account secret without ID",
			cfg: &Config{
				Auth: AuthConfig{
					ServiceAccountSecret: "secret-key",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: true,
		},
		{
			name: "service account ID without secret",
			cfg: &Config{
				Auth: AuthConfig{
					ServiceAccountID: "sa-test",
				},
				Context: ContextConfig{
					Project: "project",
					Region:  "region",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
