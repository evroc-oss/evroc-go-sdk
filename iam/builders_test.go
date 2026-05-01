// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import "testing"

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

func TestPermissionSetBuilder(t *testing.T) {
	req := NewPermissionSetBuilder("test-ps", "project-123", "user@example.com").
		WithAdmin(true).
		WithLabels(map[string]string{"env": "prod", "role": "admin"}).
		Build()

	// Validate basic fields
	if req.Kind != "PermissionSet" {
		t.Errorf("Expected Kind 'PermissionSet', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-ps" {
		t.Errorf("Expected Id 'test-ps', got %s", req.Metadata.Id)
	}

	// Validate project
	if req.Metadata.Project == nil {
		t.Error("Project should not be nil")
	} else if *req.Metadata.Project != "project-123" {
		t.Errorf("Expected project 'project-123', got %s", *req.Metadata.Project)
	}

	// Validate admin flag
	if req.Spec.Admin != true {
		t.Error("Expected admin=true")
	}

	// Validate subject
	if req.Spec.Subject.Type != "user" {
		t.Errorf("Expected subject type 'user', got %s", req.Spec.Subject.Type)
	}
	if req.Spec.Subject.User.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got %s", req.Spec.Subject.User.Email)
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else {
		if (*req.Metadata.UserLabels)["env"] != "prod" {
			t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
		}
		if (*req.Metadata.UserLabels)["role"] != "admin" {
			t.Errorf("Expected label role='admin', got %s", (*req.Metadata.UserLabels)["role"])
		}
	}
}
