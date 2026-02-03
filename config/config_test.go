// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package config

import (
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Set up test environment (t.Setenv automatically cleans up)
	t.Setenv("EVROC_USERNAME", "test@example.com")
	t.Setenv("EVROC_PASSWORD", "testpass")
	t.Setenv("EVROC_PROJECT", "test-project")
	t.Setenv("EVROC_REGION", "test-region")
	t.Setenv("EVROC_ORGANIZATION", "test-org")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() failed: %v", err)
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
	if cfg.API.BaseURL != "https://api.cloud.evroc.com" {
		t.Errorf("Default BaseURL = %v, want https://api.cloud.evroc.com", cfg.API.BaseURL)
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
			name: "username without password",
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
					// Organization omitted - should be optional
				},
			},
			wantErr: false,
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
