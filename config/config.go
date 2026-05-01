// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package config handles SDK configuration loading and validation.
package config

import (
	"fmt"
	"os"
	"path/filepath"

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

// Load loads configuration using a credential chain.
// It first attempts to load from environment variables, and if that fails,
// it falls back to the evroc CLI config (typically at ~/.evroc/config.yaml on Unix
// or %USERPROFILE%\.evroc\config.yaml on Windows).
func Load() (*Config, error) {
	cfg := &Config{}
	cfg.loadFromEnv()

	// Try to validate with just env vars
	envErr := cfg.Validate()
	if envErr == nil {
		return cfg, nil
	}

	// If env vars aren't sufficient, try loading from CLI config as fallback
	cliPath, pathErr := DefaultCLIConfigPath()
	if pathErr != nil {
		// Can't get CLI config path, return original env validation error
		return nil, envErr
	}

	// Check if CLI config exists
	if _, statErr := os.Stat(cliPath); statErr != nil {
		// CLI config doesn't exist, return original env validation error
		return nil, envErr
	}

	// Try to load from CLI config
	cliCfg, cliErr := LoadFromCLIConfig(cliPath)
	if cliErr != nil {
		// CLI config failed to load, return original env validation error
		return nil, fmt.Errorf("environment variables incomplete and CLI config failed to load: %w (original error: %w)", cliErr, envErr)
	}

	// Successfully loaded from CLI config
	return cliCfg, nil
}

type cliConfig struct {
	CurrentProfile string `yaml:"currentProfile"`
	Profiles       map[string]struct {
		APIURL       string `yaml:"apiURL"`
		IssuerURL    string `yaml:"issuerURL"`
		Organization string `yaml:"organization"`
		Project      string `yaml:"project"`
		Region       string `yaml:"region"`
		User         struct {
			RefreshToken string `yaml:"refreshToken"`
		} `yaml:"user"`
	} `yaml:"profiles"`
}

// LoadFromCLIConfig loads configuration from the evroc CLI config file.
func LoadFromCLIConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read CLI config file: %w", err)
	}

	var cli cliConfig
	if err := yaml.Unmarshal(data, &cli); err != nil {
		return nil, fmt.Errorf("failed to parse CLI config file: %w", err)
	}

	prof, ok := cli.Profiles[cli.CurrentProfile]
	if !ok {
		return nil, fmt.Errorf("current profile %q not found", cli.CurrentProfile)
	}

	cfg := &Config{
		Auth: AuthConfig{
			RefreshToken: prof.User.RefreshToken,
			TokenURL:     prof.IssuerURL + "/protocol/openid-connect/token",
			ClientID:     "evroc-cli",
		},
		API: APIConfig{
			BaseURL: prof.APIURL,
		},
		Context: ContextConfig{
			Project:      prof.Project,
			Region:       prof.Region,
			Organization: prof.Organization,
		},
	}

	// Override with environment variables
	cfg.loadFromEnv()

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid CLI config: %w", err)
	}

	return cfg, nil
}

// DefaultCLIConfigPath returns the default path to the evroc CLI config.
// On Unix systems: ~/.evroc/config.yaml
// On Windows: %USERPROFILE%\.evroc\config.yaml
func DefaultCLIConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".evroc", "config.yaml"), nil
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
	hasTokenAuth := c.Auth.Token != "" || c.Auth.RefreshToken != ""

	if !hasPasswordAuth && !hasTokenAuth {
		return fmt.Errorf("authentication required: provide either EVROC_TOKEN/EVROC_REFRESH_TOKEN or (EVROC_USERNAME + EVROC_PASSWORD)")
	}

	// Username and password must be together
	if (c.Auth.Username != "") != (c.Auth.Password != "") {
		return fmt.Errorf("EVROC_USERNAME and EVROC_PASSWORD must be provided together")
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
