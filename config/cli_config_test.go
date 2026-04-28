// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCLIConfigPath(t *testing.T) {
	path, err := DefaultCLIConfigPath()
	if err != nil {
		t.Fatalf("DefaultCLIConfigPath() failed: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got: %s", path)
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("Expected config.yaml, got: %s", filepath.Base(path))
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".evroc", "config.yaml")
	if path != expected {
		t.Errorf("Got %s, want %s", path, expected)
	}
}

func TestLoadFromCLIConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	wantProject := "proj-456"
	wantRegion := "se-sto"
	wantOrg := "org-123"
	wantAPIURL := "https://api.evroc.com"
	wantIssuerURL := "https://authn.iam.evroc.com/realms/evroc-customer"
	wantRefreshToken := "test-token"

	content := "formatVersion: v1\n" +
		"profiles:\n" +
		"  default:\n" +
		"    apiURL: " + wantAPIURL + "\n" +
		"    issuerURL: " + wantIssuerURL + "\n" +
		"    organization: " + wantOrg + "\n" +
		"    project: " + wantProject + "\n" +
		"    region: " + wantRegion + "\n" +
		"    user:\n" +
		"      refreshToken: " + wantRefreshToken + "\n" +
		"currentProfile: default\n"

	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := LoadFromCLIConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromCLIConfig() failed: %v", err)
	}

	if cfg.Context.Project != wantProject {
		t.Errorf("Project = %v, want %v", cfg.Context.Project, wantProject)
	}
	if cfg.Context.Region != wantRegion {
		t.Errorf("Region = %v, want %v", cfg.Context.Region, wantRegion)
	}
	if cfg.Context.Organization != wantOrg {
		t.Errorf("Organization = %v, want %v", cfg.Context.Organization, wantOrg)
	}
	if cfg.API.BaseURL != wantAPIURL {
		t.Errorf("BaseURL = %v, want %v", cfg.API.BaseURL, wantAPIURL)
	}
	if cfg.Auth.RefreshToken != wantRefreshToken {
		t.Errorf("RefreshToken = %v, want %v", cfg.Auth.RefreshToken, wantRefreshToken)
	}
	if cfg.Auth.TokenURL != wantIssuerURL+"/protocol/openid-connect/token" {
		t.Errorf("TokenURL = %v, want %v/protocol/openid-connect/token", cfg.Auth.TokenURL, wantIssuerURL)
	}
	if cfg.Auth.ClientID != "evroc-cli" {
		t.Errorf("ClientID = %v, want evroc-cli", cfg.Auth.ClientID)
	}
}
