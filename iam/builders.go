// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package iam provides builder patterns for IAM resources.
package iam

import (
	"context"
	"fmt"

	iam "github.com/evroc-oss/evroc-go-sdk/types/iam"
)

// ============================================================================
// Project Builder
// ============================================================================

// ProjectBuilder provides a fluent interface for creating Project resources.
type ProjectBuilder struct {
	name         string
	organization string
	displayName  string
	labels       map[string]string
}

// NewProjectBuilder creates a new builder for Project.
func NewProjectBuilder(name string, organization string) *ProjectBuilder {
	return &ProjectBuilder{
		name:         name,
		organization: organization,
	}
}

// WithDisplayName sets a human-friendly display name for the project.
func (b *ProjectBuilder) WithDisplayName(displayName string) *ProjectBuilder {
	b.displayName = displayName
	return b
}

// WithLabels sets user-defined labels for the project.
func (b *ProjectBuilder) WithLabels(labels map[string]string) *ProjectBuilder {
	b.labels = labels
	return b
}

// Build creates the ProjectRequest structure.
func (b *ProjectBuilder) Build() (*iam.ProjectRequest, error) {
	if b.organization == "" {
		return nil, fmt.Errorf("organization is required for creating projects")
	}

	req := &iam.ProjectRequest{
		ApiVersion: "iam/v1alpha4",
		Kind:       "Project",
		Metadata: iam.GlobalMetadataRequest{
			Name: &b.name,
		},
		Spec: iam.ProjectSpec{
			Organization: b.organization,
		},
	}

	if b.displayName != "" {
		req.Spec.DisplayName = &b.displayName
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := iam.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req, nil
}

// Create is a convenience method that builds and creates the project in one call.
func (b *ProjectBuilder) Create(ctx context.Context, client *ProjectsService) (*iam.Project, error) {
	req, err := b.Build()
	if err != nil {
		return nil, err
	}
	return client.Create(ctx, req)
}

// ============================================================================
// PermissionSet Builder
// ============================================================================

// PermissionSetBuilder provides a fluent interface for creating PermissionSet resources.
type PermissionSetBuilder struct {
	name    string
	project string
	email   string
	admin   bool
	labels  map[string]string
}

// NewPermissionSetBuilder creates a new builder for PermissionSet
// email is the user email to grant permissions to.
func NewPermissionSetBuilder(name string, project string, email string) *PermissionSetBuilder {
	return &PermissionSetBuilder{
		name:    name,
		project: project,
		email:   email,
		admin:   false,
	}
}

// WithAdmin sets whether this permission set grants admin privileges.
func (b *PermissionSetBuilder) WithAdmin(admin bool) *PermissionSetBuilder {
	b.admin = admin
	return b
}

// WithLabels sets user-defined labels for the permission set.
func (b *PermissionSetBuilder) WithLabels(labels map[string]string) *PermissionSetBuilder {
	b.labels = labels
	return b
}

// Build creates the PermissionSetRequest structure.
func (b *PermissionSetBuilder) Build() *iam.PermissionSetRequest {
	req := &iam.PermissionSetRequest{
		ApiVersion: "iam/v1alpha4",
		Kind:       "PermissionSet",
		Metadata: iam.GlobalProjectMetadataRequest{
			Name:    &b.name,
			Project: &b.project,
		},
		Spec: iam.PermissionSetSpec{
			Admin: b.admin,
			Subject: iam.PermissionSetSpecSubject{
				Type: iam.User,
				User: struct {
					Email string `json:"email"`
				}{
					Email: b.email,
				},
			},
		},
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := iam.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the permission set in one call.
func (b *PermissionSetBuilder) Create(ctx context.Context, client *PermissionSetsService) (*iam.PermissionSet, error) {
	req := b.Build()
	return client.Create(ctx, req)
}
