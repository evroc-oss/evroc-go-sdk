// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package evroc provides a Go SDK for the evroc Cloud Platform APIs.
//
// The SDK supports multiple API services:
//   - Compute: Virtual Machines, Disks, Placement Groups
//   - Networking: VPCs, Subnets, Security Groups, Public IPs
//   - IAM: Organizations, Projects, Users, Permissions
//   - Storage: Buckets, Service Accounts
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
// The SDK supports two authentication methods:
//  1. Username/Password (set EVROC_USERNAME, EVROC_PASSWORD)
//  2. Direct bearer token (set EVROC_TOKEN)
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
	"github.com/evroc-oss/evroc-go-sdk/metrics"
	"github.com/evroc-oss/evroc-go-sdk/networking"
	"github.com/evroc-oss/evroc-go-sdk/internal/versions"
	"github.com/evroc-oss/evroc-go-sdk/quotas"
	"github.com/evroc-oss/evroc-go-sdk/storage"
)

// Client provides access to all evroc APIs.
type Client struct {
	config           *config.Config
	authClient       *auth.Client
	customHTTPClient *http.Client         // Optional custom HTTP client
	metrics          *metrics.Manager     // Optional metrics manager

	// API clients
	compute    *compute.Client
	networking *networking.Client
	iam        *iam.Client
	storage    *storage.Client
	quotas     *quotas.Client
}

// Option is a functional option for configuring the Client.
type Option func(*Client) error

// WithHTTPClient sets a custom HTTP client for the REST client.
// This is useful for testing or when you need custom transport behavior.
//
// Note: The custom HTTP client will NOT automatically add authentication headers.
// Authentication is added by the SDK before each request.
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

// NewFromEnv creates a new evroc client using environment variables
//
// Required environment variables:
//   - EVROC_USERNAME: OIDC username/email (or use EVROC_TOKEN for token auth)
//   - EVROC_PASSWORD: OIDC password (or use EVROC_TOKEN for token auth)
//   - EVROC_PROJECT: Project ID
//   - EVROC_REGION: Region
//   - EVROC_ORGANIZATION: Organization ID
func NewFromEnv(ctx context.Context, opts ...Option) (*Client, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}

	return New(ctx, cfg, opts...)
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

	return New(ctx, cfg, opts...)
}

// New creates a new evroc client with the provided configuration.
func New(ctx context.Context, cfg *config.Config, opts ...Option) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Set defaults
	cfg.SetDefaults()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create auth client with either tokens or password
	authClient, err := auth.NewClient(ctx, cfg.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}

	client := &Client{
		config:     cfg,
		authClient: authClient,
	}

	// Apply options (this may set metrics)
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// If metrics were configured, pass them to the auth client
	// We need to recreate the auth client with metrics
	if client.metrics != nil {
		authClient, err = auth.NewClient(ctx, cfg.Auth, auth.WithMetrics(client.metrics))
		if err != nil {
			return nil, fmt.Errorf("failed to recreate auth client with metrics: %w", err)
		}
		client.authClient = authClient
	}

	if err := client.initAPIClients(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize API clients: %w", err)
	}

	return client, nil
}

// NewWithAuthClient creates a new evroc client with a pre-configured auth client
// This is useful when you've performed custom authentication (e.g., interactive login).
func NewWithAuthClient(ctx context.Context, cfg *config.Config, authClient *auth.Client) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if authClient == nil {
		return nil, fmt.Errorf("auth client is required")
	}

	// Set defaults
	cfg.SetDefaults()

	client := &Client{
		config:     cfg,
		authClient: authClient,
	}

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
		httpClient, err = c.authClient.GetHTTPClient(ctx)
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
	c.compute = compute.NewClientWithParent(restClient, c).WithMetrics(c.metrics)
	c.networking = networking.NewClientWithParent(restClient, c).WithMetrics(c.metrics)
	c.iam = iam.NewClientWithParent(restClient, c).WithMetrics(c.metrics)
	c.storage = storage.NewClientWithParent(restClient, c).WithMetrics(c.metrics)
	c.quotas = quotas.NewClientWithParent(restClient, c).WithMetrics(c.metrics)

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

// Config returns the current configuration.
func (c *Client) Config() *config.Config {
	return c.config
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

// Version returns version information for the SDK and all API services.
// This is useful for debugging and understanding which API versions are supported.
func (c *Client) Version() versions.APIVersionInfo {
	return versions.Current()
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
