// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package config handles SDK configuration loading and validation.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the evroc SDK configuration.
type Config struct {
	// Auth configuration
	Auth AuthConfig `yaml:"auth"`

	// API endpoints
	API APIConfig `yaml:"api"`

	// Project/Organization context
	Context ContextConfig `yaml:"context"`
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	// OAuth2/OIDC token URL
	TokenURL string `yaml:"token_url"`

	// Client ID for OAuth2
	ClientID string `yaml:"client_id"`

	// Username for OIDC password grant
	Username string `yaml:"username"`

	// Password for OIDC password grant
	Password string `yaml:"password"`

	// Direct access token (alternative to username/password)
	Token string `yaml:"token"`

	// Refresh token for automatic token renewal
	RefreshToken string `yaml:"refresh_token"`

	// Optional: Scopes for OAuth2
	Scopes []string `yaml:"scopes"`
}

// APIConfig contains API endpoint configuration.
type APIConfig struct {
	// Base URL for all evroc APIs
	BaseURL string `yaml:"base_url"`
}

// ContextConfig contains project/region context.
type ContextConfig struct {
	// Default project (required for most operations)
	Project string `yaml:"project"`

	// Default region (required for Compute API operations)
	Region string `yaml:"region"`

	// Organization (rarely needed - most APIs use project-scoped paths)
	Organization string `yaml:"organization"`
}

// LoadFromFile loads configuration from a YAML file.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return &cfg, nil
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	cfg.loadFromEnv()

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFromEnv loads values from environment variables, overriding any existing values.
func (c *Config) loadFromEnv() {
	// Auth
	if v := os.Getenv("EVROC_TOKEN_URL"); v != "" {
		c.Auth.TokenURL = v
	}
	if v := os.Getenv("EVROC_CLIENT_ID"); v != "" {
		c.Auth.ClientID = v
	}
	if v := os.Getenv("EVROC_USERNAME"); v != "" {
		c.Auth.Username = v
	}
	if v := os.Getenv("EVROC_PASSWORD"); v != "" {
		c.Auth.Password = v
	}
	if v := os.Getenv("EVROC_TOKEN"); v != "" {
		c.Auth.Token = v
	}
	if v := os.Getenv("EVROC_REFRESH_TOKEN"); v != "" {
		c.Auth.RefreshToken = v
	}

	// API endpoint
	if v := os.Getenv("EVROC_API_URL"); v != "" {
		c.API.BaseURL = v
	}

	// Context
	if v := os.Getenv("EVROC_PROJECT"); v != "" {
		c.Context.Project = v
	}
	if v := os.Getenv("EVROC_REGION"); v != "" {
		c.Context.Region = v
	}
	if v := os.Getenv("EVROC_ORGANIZATION"); v != "" {
		c.Context.Organization = v
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// At least one authentication method must be provided
	hasPasswordAuth := c.Auth.Username != "" && c.Auth.Password != ""
	hasTokenAuth := c.Auth.Token != ""

	if !hasPasswordAuth && !hasTokenAuth {
		return fmt.Errorf("authentication required: provide either EVROC_TOKEN or (EVROC_USERNAME + EVROC_PASSWORD)")
	}

	// If username is set, password must also be set (and vice versa)
	if c.Auth.Username == "" && c.Auth.Password != "" {
		return fmt.Errorf("EVROC_USERNAME is required when EVROC_PASSWORD is set")
	}
	if c.Auth.Password == "" && c.Auth.Username != "" {
		return fmt.Errorf("EVROC_PASSWORD is required when EVROC_USERNAME is set")
	}

	// Project and region are required
	if c.Context.Project == "" {
		return fmt.Errorf("EVROC_PROJECT is required")
	}
	if c.Context.Region == "" {
		return fmt.Errorf("EVROC_REGION is required")
	}

	// Organization is optional (only needed for IAM project creation)

	return nil
}

// SetDefaults sets default values for optional fields.
func (c *Config) SetDefaults() {
	// Set default client_id if not specified
	if c.Auth.ClientID == "" {
		c.Auth.ClientID = defaultClientID
	}

	// Set default token_url if not specified
	if c.Auth.TokenURL == "" {
		c.Auth.TokenURL = defaultTokenURL()
	}

	// Set default scopes if not specified
	if len(c.Auth.Scopes) == 0 {
		c.Auth.Scopes = []string{"openid", "offline_access"}
	}

	// Set default API URL if not specified
	if c.API.BaseURL == "" {
		c.API.BaseURL = defaultBaseURL
	}
}
