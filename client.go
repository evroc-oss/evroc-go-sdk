// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package evroc provides a Go SDK for the evroc Cloud Platform APIs.
//
// The SDK supports multiple API services:
//   - Compute: Virtual Machines, Disks, Placement Groups
//   - Networking: VPCs, Subnets, Security Groups, Public IPs
//   - IAM: Organizations, Projects, Users, Permissions
//   - Storage: Buckets, Service Accounts, File Stores
//   - Think: Models, Instances, API Keys
//
// Quick Start:
//
//	// Create client from environment variables
//	ctx := context.Background()
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Or with custom HTTP client for testing
//	client, err := evroc.NewFromEnv(ctx, evroc.WithHTTPClient(customClient))
//
//	// Use the API services
//	vms, err := client.Compute().VirtualMachines().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Authentication:
//
// The SDK supports multiple authentication methods:
//   - Service Account (recommended): EVROC_SERVICE_ACCOUNT_ID + EVROC_SERVICE_ACCOUNT_SECRET
//   - Bearer token: EVROC_TOKEN, EVROC_REFRESH_TOKEN
//   - Username/Password (deprecated): EVROC_USERNAME, EVROC_PASSWORD
//
// Configuration:
//
// Configuration can be provided via:
//  1. Environment variables (EVROC_*)
//  2. YAML configuration file
//  3. Programmatically via config.Config struct
package evroc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/config"
	"github.com/evroc-oss/evroc-go-sdk/iam"
	"github.com/evroc-oss/evroc-go-sdk/internal/auth"
	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/internal/versions"
	"github.com/evroc-oss/evroc-go-sdk/metrics"
	"github.com/evroc-oss/evroc-go-sdk/networking"
	"github.com/evroc-oss/evroc-go-sdk/quotas"
	"github.com/evroc-oss/evroc-go-sdk/storage"
	"github.com/evroc-oss/evroc-go-sdk/think"
)

// Client provides access to all evroc APIs.
type Client struct {
	config           *config.Config
	authClient       *auth.Client
	customHTTPClient *http.Client
	metrics          *metrics.Manager

	// API clients
	compute    *compute.Client
	networking *networking.Client
	iam        *iam.Client
	storage    *storage.Client
	quotas     *quotas.Client
	think      *think.Client
}

// Option is a functional option for configuring the Client.
type Option func(*Client) error

// WithHTTPClient sets a custom HTTP client for the REST client.
// This is useful for testing or when you need custom transport behavior.
//
// Note: The custom HTTP client will NOT automatically add authentication headers.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("httpClient cannot be nil")
		}
		// Store for use during initAPIClients
		c.customHTTPClient = httpClient
		return nil
	}
}

// WithMetrics enables Prometheus metrics collection for SDK operations.
// Pass a metrics.Manager instance to collect metrics for:
//   - API call durations and error rates
//   - Retry attempts and backoff times
//   - Waiter operation durations and polling attempts
//
// Example:
//
//	manager := metrics.NewManager()
//	client, err := evroc.NewFromEnv(ctx, evroc.WithMetrics(manager))
//
//	// Expose metrics endpoint
//	http.Handle("/metrics", promhttp.HandlerFor(manager.Registry(), promhttp.HandlerOpts{}))
func WithMetrics(manager *metrics.Manager) Option {
	return func(c *Client) error {
		if manager == nil {
			return fmt.Errorf("metrics manager cannot be nil")
		}
		c.metrics = manager
		return nil
	}
}

// NewFromEnv creates a new evroc client using environment variables.
//
// Service account authentication (recommended):
//   - EVROC_SERVICE_ACCOUNT_ID: Service account name
//   - EVROC_SERVICE_ACCOUNT_SECRET: Private key (file path or base64-encoded JWK)
//   - EVROC_PROJECT: Project name
//   - EVROC_REGION: Region
//
// The client_id is derived automatically as <EVROC_SERVICE_ACCOUNT_ID>_<EVROC_PROJECT>.
//
// Deprecated authentication methods (username/password, bearer token) are still
// supported — see [config.AuthConfig] for details.
func NewFromEnv(ctx context.Context, opts ...Option) (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}

	return New(ctx, *cfg, opts...)
}

// NewFromFile creates a new evroc client from a YAML configuration file
//
// The configuration file should have the following structure:
//
//	auth:
//	  username: "your-username@evroc.com"
//	  password: "your-password"
//	  client_id: "evroc-cli"  # optional, has default
//	  token_url: "https://authn.iam.evroc.com/..."  # optional, has default
//	api:
//	  base_url: "https://api.evroc.com"  # optional, has default
//	context:
//	  project: "project-uuid"
//	  region: "se-sto"
//	  organization: "org-uuid"
//
// Environment variables override file configuration.
func NewFromFile(ctx context.Context, path string, opts ...Option) (*Client, error) {
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from file: %w", err)
	}

	return New(ctx, *cfg, opts...)
}

// NewFromCLIConfig creates a new evroc client from the evroc CLI configuration.
// Reads from ~/.evroc/config.yaml by default, or from configPath if provided.
//
// Parameters:
//   - ctx: Context for the client initialization
//   - configPath: Path to CLI config file (empty string uses default ~/.evroc/config.yaml)
//   - opts: Optional client configuration options
func NewFromCLIConfig(ctx context.Context, configPath string, opts ...Option) (*Client, error) {
	if configPath == "" {
		var err error
		configPath, err = config.DefaultCLIConfigPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get default CLI config path: %w", err)
		}
	}

	cfg, err := config.LoadFromCLIConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CLI config: %w", err)
	}

	return New(ctx, *cfg, opts...)
}

// New creates a new evroc client with the provided configuration.
func New(ctx context.Context, cfg config.Config, opts ...Option) (*Client, error) {
	// Set defaults
	cfg.SetDefaults()

	client := &Client{
		config: &cfg,
	}

	// Apply options first — these may set metrics, HTTP client, or token source
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create auth client
	var metricsOpt auth.Option
	if client.metrics != nil {
		metricsOpt = auth.WithMetrics(client.metrics)
	}
	authClient, err := auth.NewClient(ctx, cfg.Auth, metricsOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}
	client.authClient = authClient

	if err := client.initAPIClients(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize API clients: %w", err)
	}

	return client, nil
}

// initAPIClients initializes all API-specific clients.
func (c *Client) initAPIClients(ctx context.Context) error {
	// All APIs share the same base URL
	baseURL := c.config.API.BaseURL

	// Get HTTP client with automatic authentication
	// Use custom HTTP client if provided, otherwise get auto-auth client
	var httpClient *http.Client
	var err error

	if c.customHTTPClient != nil {
		httpClient = c.customHTTPClient
	} else {
		httpClient, err = c.authClient.HTTPClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to get authenticated HTTP client: %w", err)
		}
	}

	// Create shared REST client
	restConfig := rest.Config{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		Metrics:    c.metrics, // Pass metrics to REST client
	}

	restClient, err := rest.NewClient(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Initialize all API clients with REST client and parent context provider
	c.compute = compute.NewClient(restClient, c).WithMetrics(c.metrics)
	c.networking = networking.NewClient(restClient, c).WithMetrics(c.metrics)
	c.iam = iam.NewClient(restClient, c).WithMetrics(c.metrics)
	c.storage = storage.NewClient(restClient, c).WithMetrics(c.metrics)
	c.quotas = quotas.NewClient(restClient, c).WithMetrics(c.metrics)
	c.think = think.NewClient(restClient, c).WithMetrics(c.metrics)

	return nil
}

// Compute returns the Compute API client
//
// Provides access to:
//   - Virtual Machines
//   - Disks
//   - Placement Groups
//   - Hotswap Disk Attachments
func (c *Client) Compute() *compute.Client {
	return c.compute
}

// Networking returns the Networking API client
//
// Provides access to:
//   - Virtual Private Clouds (VPCs)
//   - Subnets
//   - Security Groups
//   - Public IPs
func (c *Client) Networking() *networking.Client {
	return c.networking
}

// IAM returns the IAM API client
//
// Provides access to:
//   - Organizations
//   - Projects
//   - Users
//   - Permission Sets
//   - Quotas
func (c *Client) IAM() *iam.Client {
	return c.iam
}

// Storage returns the Storage API client
//
// Provides access to:
//   - Buckets
//   - Bucket Service Accounts
//   - Bucket Service Account Secrets
//   - File Stores
func (c *Client) Storage() *storage.Client {
	return c.storage
}

// Quotas returns the Quotas API client
//
// Provides access to:
//   - Organization Quotas
//   - Project Quotas
func (c *Client) Quotas() *quotas.Client {
	return c.quotas
}

// Think returns the Think API client
//
// Provides access to:
//   - API Keys
//   - Models
//   - Shared Models
//   - Sizes
//   - Dedicated Instances
func (c *Client) Think() *think.Client {
	return c.think
}

// Config returns the current configuration.
func (c *Client) Config() config.Config {
	return *c.config
}

// DefaultProject returns the default project from config.
func (c *Client) DefaultProject() string {
	return c.config.Context.Project
}

// DefaultRegion returns the default region from config.
func (c *Client) DefaultRegion() string {
	return c.config.Context.Region
}

// DefaultOrganization returns the default organization from config (rarely needed).
func (c *Client) DefaultOrganization() string {
	return c.config.Context.Organization
}

// APIVersionInfo contains version information for the SDK and all API services.
type APIVersionInfo struct {
	SDKVersion string
	Compute    string
	Networking string
	IAM        string
	Storage    string
	Quotas     string
	Think      string
}

// Version returns version information for the SDK and all API services.
// This is useful for debugging and understanding which API versions are supported.
func (c *Client) Version() APIVersionInfo {
	v := versions.Current()
	return APIVersionInfo{
		SDKVersion: v.SDKVersion,
		Compute:    v.Compute,
		Networking: v.Networking,
		IAM:        v.IAM,
		Storage:    v.Storage,
		Quotas:     v.Quotas,
		Think:      v.Think,
	}
}

// SDKVersion returns the semantic version of this SDK.
func (c *Client) SDKVersion() string {
	return versions.SDKVersion
}

// Ptr returns a pointer to any value.
// This is useful when you need to pass a pointer to a literal value
// when using direct construction instead of builders.
//
// Example:
//
//	name := evroc.Ptr("my-vm")
//	enabled := evroc.Ptr(true)
//	size := evroc.Ptr(int32(100))
func Ptr[T any](v T) *T {
	return &v
}
