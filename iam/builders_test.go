// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

func TestProjectBuilder(t *testing.T) {
	req, err := NewProjectBuilder("test-project", "org-123").
		WithName("Test Project").
		WithLabels(map[string]string{"env": "prod", "tier": "production"}).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Validate basic fields
	if req.Kind != "Project" {
		t.Errorf("Expected Kind 'Project', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-project" {
		t.Errorf("Expected Id 'test-project', got %s", req.Metadata.Id)
	}

	// Validate organization
	if req.Spec.Organization != "org-123" {
		t.Errorf("Expected organization 'org-123', got %s", req.Spec.Organization)
	}

	// Validate display name
	if req.Spec.Name == nil {
		t.Error("Display name should not be nil")
	} else if *req.Spec.Name != "Test Project" {
		t.Errorf("Expected display name 'Test Project', got %s", *req.Spec.Name)
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else {
		if (*req.Metadata.UserLabels)["env"] != "prod" {
			t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
		}
		if (*req.Metadata.UserLabels)["tier"] != "production" {
			t.Errorf("Expected label tier='production', got %s", (*req.Metadata.UserLabels)["tier"])
		}
	}
}

func TestServiceAccountCredentialBuilderHMACSigV4(t *testing.T) {
	req := NewServiceAccountCredentialBuilder("key-1", "proj", "/iam/projects/proj/serviceAccounts/sa", time.Now()).
		WithHMACSigV4().
		WithAccessTokenLifetime(600).
		Build()

	if req.Spec.Type != iam.HmacSigv4 {
		t.Errorf("expected type %q, got %q", iam.HmacSigv4, req.Spec.Type)
	}
	if req.Spec.Rs256Jwt != nil {
		t.Errorf("expected no rs256Jwt block for HMAC, got %+v", req.Spec.Rs256Jwt)
	}
}

func TestServiceAccountCredentialBuilderDefaultsToRS256(t *testing.T) {
	req := NewServiceAccountCredentialBuilder("key-1", "proj", "/iam/projects/proj/serviceAccounts/sa", time.Now()).Build()

	if req.Spec.Type != iam.Rs256Jwt {
		t.Errorf("expected type %q, got %q", iam.Rs256Jwt, req.Spec.Type)
	}
}
