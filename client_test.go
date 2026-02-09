// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package evroc

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/config"
	"github.com/evroc-oss/evroc-go-sdk/metrics"
)

func newTestConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			Token:    "test-token",
			TokenURL: "https://auth.test.com/token",
		},
		API: config.APIConfig{
			BaseURL: "https://api.test.com",
		},
		Context: config.ContextConfig{
			Project:      "test-project",
			Region:       "test-region",
			Organization: "test-org",
		},
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	t.Run("nil config", func(t *testing.T) {
		_, err := New(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})

	t.Run("valid config with token", func(t *testing.T) {
		cfg := newTestConfig()
		client, err := New(ctx, cfg)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		if client == nil {
			t.Fatal("client should not be nil")
		}

		if client.Config() != cfg {
			t.Error("config mismatch")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &config.Config{
			// Missing required fields
		}

		_, err := New(ctx, cfg)
		if err == nil {
			t.Fatal("expected error for invalid config")
		}
	})
}

func TestNewWithAuthClient(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig()

	t.Run("nil config", func(t *testing.T) {
		_, err := NewWithAuthClient(ctx, nil, nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})

	t.Run("nil auth client", func(t *testing.T) {
		_, err := NewWithAuthClient(ctx, cfg, nil)
		if err == nil {
			t.Fatal("expected error for nil auth client")
		}
	})
}

func TestWithHTTPClient(t *testing.T) {
	t.Run("valid http client", func(t *testing.T) {
		client := &Client{}
		httpClient := &http.Client{}

		opt := WithHTTPClient(httpClient)
		err := opt(client)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.customHTTPClient != httpClient {
			t.Error("http client not set correctly")
		}
	})

	t.Run("nil http client", func(t *testing.T) {
		client := &Client{}

		opt := WithHTTPClient(nil)
		err := opt(client)

		if err == nil {
			t.Fatal("expected error for nil http client")
		}
	})
}

func TestWithMetrics(t *testing.T) {
	t.Run("valid metrics manager", func(t *testing.T) {
		client := &Client{}
		manager := metrics.NewManager()

		opt := WithMetrics(manager)
		err := opt(client)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.metrics != manager {
			t.Error("metrics manager not set correctly")
		}
	})

	t.Run("nil metrics manager", func(t *testing.T) {
		client := &Client{}

		opt := WithMetrics(nil)
		err := opt(client)

		if err == nil {
			t.Fatal("expected error for nil metrics manager")
		}
	})
}

func TestServiceGetters(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig()
	client, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("Compute", func(t *testing.T) {
		compute := client.Compute()
		if compute == nil {
			t.Error("compute client should not be nil")
		}
	})

	t.Run("Networking", func(t *testing.T) {
		networking := client.Networking()
		if networking == nil {
			t.Error("networking client should not be nil")
		}
	})

	t.Run("IAM", func(t *testing.T) {
		iam := client.IAM()
		if iam == nil {
			t.Error("iam client should not be nil")
		}
	})

	t.Run("Storage", func(t *testing.T) {
		storage := client.Storage()
		if storage == nil {
			t.Error("storage client should not be nil")
		}
	})

	t.Run("Quotas", func(t *testing.T) {
		quotas := client.Quotas()
		if quotas == nil {
			t.Error("quotas client should not be nil")
		}
	})
}

func TestContextGetters(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig()
	client, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("DefaultProject", func(t *testing.T) {
		project := client.DefaultProject()
		if project != "test-project" {
			t.Errorf("expected test-project, got %s", project)
		}
	})

	t.Run("DefaultRegion", func(t *testing.T) {
		region := client.DefaultRegion()
		if region != "test-region" {
			t.Errorf("expected test-region, got %s", region)
		}
	})

	t.Run("DefaultOrganization", func(t *testing.T) {
		org := client.DefaultOrganization()
		if org != "test-org" {
			t.Errorf("expected test-org, got %s", org)
		}
	})

	t.Run("Config", func(t *testing.T) {
		config := client.Config()
		if config != cfg {
			t.Error("config mismatch")
		}
	})
}

func TestVersionGetters(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig()
	client, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("Version", func(t *testing.T) {
		version := client.Version()
		if version.SDKVersion == "" {
			t.Error("SDK version should not be empty")
		}
	})

	t.Run("SDKVersion", func(t *testing.T) {
		version := client.SDKVersion()
		if version == "" {
			t.Error("SDK version should not be empty")
		}
	})
}

func TestPtr(t *testing.T) {
	t.Run("string pointer", func(t *testing.T) {
		val := "test"
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("pointer should not be nil")
		}
		if *ptr != val {
			t.Errorf("expected %s, got %s", val, *ptr)
		}
	})

	t.Run("int pointer", func(t *testing.T) {
		val := 42
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("pointer should not be nil")
		}
		if *ptr != val {
			t.Errorf("expected %d, got %d", val, *ptr)
		}
	})

	t.Run("bool pointer", func(t *testing.T) {
		val := true
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("pointer should not be nil")
		}
		if *ptr != val {
			t.Errorf("expected %v, got %v", val, *ptr)
		}
	})
}

func TestNewFromEnv(t *testing.T) {
	t.Run("missing env vars", func(t *testing.T) {
		// Clear all environment variables
		os.Clearenv()

		ctx := context.Background()
		_, err := NewFromEnv(ctx)
		if err == nil {
			t.Fatal("expected error when env vars missing")
		}
	})

	t.Run("valid env vars", func(t *testing.T) {
		// Set up required environment variables
		os.Setenv("EVROC_TOKEN", "test-token")
		os.Setenv("EVROC_PROJECT", "test-project")
		os.Setenv("EVROC_REGION", "test-region")
		os.Setenv("EVROC_ORGANIZATION", "test-org")
		defer os.Clearenv()

		ctx := context.Background()
		client, err := NewFromEnv(ctx)
		if err != nil {
			t.Fatalf("failed to create client from env: %v", err)
		}

		if client == nil {
			t.Fatal("client should not be nil")
		}
	})
}
