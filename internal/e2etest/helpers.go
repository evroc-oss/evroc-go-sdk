// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

// Package e2etest provides helpers for end-to-end testing against the real evroc API.
// These tests require valid credentials and will create/destroy real resources.
package e2etest

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/config"
)

// SkipIfNotE2E skips the test if E2E environment variable is not set.
// This prevents e2e tests from running during normal unit test execution.
func SkipIfNotE2E(t *testing.T) {
	if os.Getenv("E2E") == "" {
		t.Skip("Skipping e2e test. Set E2E=1 to run e2e tests")
	}
}

// PreCheck validates that required configuration is available for e2e testing.
// It uses config.Load() which automatically checks environment variables and falls back to CLI config.
func PreCheck(t *testing.T) {
	t.Helper()

	SkipIfNotE2E(t)

	// Let the SDK's config loader do all the autodetection work
	// It checks env vars first, then falls back to ~/.evroc/config.yaml
	if _, err := config.Load(); err != nil {
		t.Fatalf("Configuration required for e2e tests: %v\n\nEither:\n  - Set EVROC_TOKEN (or EVROC_REFRESH_TOKEN) and EVROC_PROJECT\n  - Or run 'evroc login'", err)
	}
}

// NewClient creates a new SDK client configured from environment variables.
// This uses the same authentication mechanism as the SDK itself.
func NewClient(t *testing.T) *evroc.Client {
	t.Helper()

	ctx := context.Background()
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return client
}

// RandomName generates a unique test resource name with the given prefix.
// All test resources use the "e2e-test-" prefix for easy identification.
func RandomName(prefix string) string {
	return fmt.Sprintf("e2e-test-%s-%d", prefix, time.Now().Unix())
}

// GetProject returns the project ID from environment variables.
func GetProject(t *testing.T) string {
	t.Helper()
	project := os.Getenv("EVROC_PROJECT")
	if project == "" {
		t.Fatal("EVROC_PROJECT must be set")
	}
	return project
}

// GetRegion returns the region from environment variables, or defaults to "se-sto".
func GetRegion() string {
	region := os.Getenv("EVROC_REGION")
	if region == "" {
		return "se-sto"
	}
	return region
}

// GetOrganization returns the organization ID from environment variables if set.
func GetOrganization() string {
	return os.Getenv("EVROC_ORGANIZATION")
}

// MustGetID validates that the ID is not empty and returns it.
func MustGetID(t *testing.T, id string, resourceType string) string {
	t.Helper()
	if id == "" {
		t.Fatalf("%s ID is empty", resourceType)
	}
	return id
}

// AssertInList verifies that a resource with the expected ID appears in a list.
// The getID function extracts the ID from each item in the list.
func AssertInList[T any](t *testing.T, items []T, expectedID string, getID func(T) string, resourceType string) {
	t.Helper()
	for _, item := range items {
		if id := getID(item); id == expectedID {
			return
		}
	}
	t.Errorf("%s %s not found in list", resourceType, expectedID)
}

// AssertDeleted verifies that a resource has been deleted by attempting to get it
// and expecting an error. It waits for API propagation before checking.
func AssertDeleted(t *testing.T, ctx context.Context, getFunc func(context.Context, string) (any, error), id, resourceType string) {
	t.Helper()
	time.Sleep(APIPropagationDelay)
	_, err := getFunc(ctx, id)
	if err == nil {
		t.Errorf("expected error when getting deleted %s, got nil", resourceType)
	}
}

// DeferCleanup sets up a deferred cleanup function that deletes a resource.
// It uses the deleted pointer to track whether the resource was already deleted in the test.
func DeferCleanup(t *testing.T, ctx context.Context, deleteFunc func(context.Context, string) error, id, resourceType string, deleted *bool) {
	t.Helper()
	t.Cleanup(func() {
		if !*deleted {
			t.Logf("Cleaning up %s: %s", resourceType, id)
			if err := deleteFunc(ctx, id); err != nil {
				t.Errorf("failed to delete %s during cleanup: %v", resourceType, err)
			}
		}
	})
}
