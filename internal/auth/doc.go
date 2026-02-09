// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package auth provides authentication for the evroc SDK.
//
// The auth package handles OAuth2/OIDC authentication with the evroc platform
// and manages access tokens automatically.
//
// # Authentication Methods
//
// The SDK supports two authentication methods:
//
//  1. Username/Password (OAuth2 Resource Owner Password Credentials)
//  2. Direct Bearer Token
//
// # Username/Password Authentication
//
// Recommended for most use cases. The SDK handles token acquisition and refresh:
//
//	cfg := config.Config{
//	    Auth: config.AuthConfig{
//	        Username:  "user@example.com",
//	        Password:  "secret",
//	    },
//	    Region:  "se-sto",
//	    Project: "my-project",
//	}
//
//	client, err := evroc.New(ctx, &cfg)
//
// # Environment Variables
//
// Configure authentication via environment variables:
//
//	export EVROC_USERNAME="user@example.com"
//	export EVROC_PASSWORD="secret"
//	export EVROC_REGION="se-sto"
//	export EVROC_PROJECT="my-project"
//
//	client, err := evroc.NewFromEnv(ctx)
//
// # Bearer Token Authentication
//
// Use when you've obtained a token through other means:
//
//	cfg := config.Config{
//	    Auth: config.AuthConfig{
//	        Token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
//	    },
//	    Region:  "se-sto",
//	    Project: "my-project",
//	}
//
//	client, err := evroc.New(ctx, &cfg)
//
// Or via environment variable:
//
//	export EVROC_TOKEN="eyJhbGci..."
//	export EVROC_REGION="se-sto"
//	export EVROC_PROJECT="my-project"
//	client, err := evroc.NewFromEnv(ctx)
//
// # Token Management
//
// The SDK automatically:
//
//   - Acquires access tokens on first request
//   - Includes tokens in Authorization headers
//   - Refreshes tokens when they expire
//
// # Custom HTTP Client
//
// Provide your own HTTP client (e.g., for testing):
//
//	client, err := evroc.NewFromEnv(ctx,
//	    evroc.WithHTTPClient(customHTTPClient),
//	)
//
// Note: Custom HTTP clients do not automatically add authentication headers.
// The SDK adds authentication before each request.
//
// # Security Best Practices
//
//   - Never commit credentials to source control
//   - Use environment variables or secure credential storage
//   - Rotate credentials regularly
//   - Use short-lived tokens when possible
package auth
