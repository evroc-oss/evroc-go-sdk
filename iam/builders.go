// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package iam provides builder patterns for IAM resources.
package iam

import (
	"context"
	"fmt"
	"time"

	iam "github.com/evroc-oss/evroc-go-sdk/types/iam"
)

// builderAPIVersion is the full API version string for IAM resource requests.
// It references the apiVersion constant from client_generated.go.
const builderAPIVersion = "iam/" + apiVersion

// ============================================================================
// Project Builder
// ============================================================================

// ProjectBuilder provides a fluent interface for creating Project resources.
type ProjectBuilder struct {
	id           string
	organization string
	name         string
	labels       map[string]string
}

// NewProjectBuilder creates a new builder for Project.
func NewProjectBuilder(id string, organization string) *ProjectBuilder {
	return &ProjectBuilder{
		id:           id,
		organization: organization,
	}
}

// WithName sets a human-friendly name for the project.
func (b *ProjectBuilder) WithName(name string) *ProjectBuilder {
	b.name = name
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
		ApiVersion: builderAPIVersion,
		Kind:       "Project",
		Metadata: iam.GlobalMetadataRequest{
			Id: b.id,
		},
		Spec: iam.ProjectSpec{
			Organization: b.organization,
		},
	}

	if b.name != "" {
		req.Spec.Name = &b.name
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
// ServiceAccount Builder
// ============================================================================

// ServiceAccountBuilder provides a fluent interface for creating ServiceAccount resources.
type ServiceAccountBuilder struct {
	id          string
	project     string
	description string
	enabled     bool
	labels      map[string]string
}

// NewServiceAccountBuilder creates a new builder for ServiceAccount.
func NewServiceAccountBuilder(id string, project string) *ServiceAccountBuilder {
	return &ServiceAccountBuilder{
		id:      id,
		project: project,
		enabled: true,
	}
}

// WithDescription sets a human-readable description.
func (b *ServiceAccountBuilder) WithDescription(description string) *ServiceAccountBuilder {
	b.description = description
	return b
}

// WithEnabled sets whether the service account is enabled.
func (b *ServiceAccountBuilder) WithEnabled(enabled bool) *ServiceAccountBuilder {
	b.enabled = enabled
	return b
}

// WithLabels sets user-defined labels.
func (b *ServiceAccountBuilder) WithLabels(labels map[string]string) *ServiceAccountBuilder {
	b.labels = labels
	return b
}

// Build creates the ServiceaccountRequest structure.
func (b *ServiceAccountBuilder) Build() *iam.ServiceaccountRequest {
	req := &iam.ServiceaccountRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "ServiceAccount",
		Metadata: iam.GlobalProjectMetadataRequest{
			Id:      b.id,
			Project: &b.project,
		},
		Spec: iam.ServiceaccountSpec{
			Enabled: &b.enabled,
		},
	}

	if b.description != "" {
		req.Spec.Description = &b.description
	}

	if len(b.labels) > 0 {
		userLabels := iam.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the service account in one call.
func (b *ServiceAccountBuilder) Create(ctx context.Context, client *ServiceAccountsService) (*iam.Serviceaccount, error) {
	req := b.Build()
	return client.Create(ctx, req)
}

// ============================================================================
// ServiceAccountCredential Builder
// ============================================================================

// ServiceAccountCredentialBuilder provides a fluent interface for creating ServiceAccountCredential resources.
type ServiceAccountCredentialBuilder struct {
	id                  string
	project             string
	serviceAccountFQID  string
	expiresAt           time.Time
	description         string
	accessTokenLifetime int
}

// NewServiceAccountCredentialBuilder creates a new builder for ServiceAccountCredential.
// serviceAccountFQID is the full FQID of the parent service account
// (e.g., /iam/projects/<project>/serviceAccounts/<name>).
func NewServiceAccountCredentialBuilder(id string, project string, serviceAccountFQID string, expiresAt time.Time) *ServiceAccountCredentialBuilder {
	return &ServiceAccountCredentialBuilder{
		id:                 id,
		project:            project,
		serviceAccountFQID: serviceAccountFQID,
		expiresAt:          expiresAt,
	}
}

// WithDescription sets a human-readable description.
func (b *ServiceAccountCredentialBuilder) WithDescription(description string) *ServiceAccountCredentialBuilder {
	b.description = description
	return b
}

// WithAccessTokenLifetime sets the access token lifetime in seconds.
func (b *ServiceAccountCredentialBuilder) WithAccessTokenLifetime(seconds int) *ServiceAccountCredentialBuilder {
	b.accessTokenLifetime = seconds
	return b
}

// Build creates the ServiceaccountcredentialRequest structure.
func (b *ServiceAccountCredentialBuilder) Build() *iam.ServiceaccountcredentialRequest {
	req := &iam.ServiceaccountcredentialRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "ServiceAccountCredential",
		Metadata: iam.GlobalProjectMetadataRequest{
			Id:      b.id,
			Project: &b.project,
		},
		Spec: iam.ServiceaccountcredentialSpec{
			AccountRef: iam.ServiceaccountcredentialSpecAccountRef{
				Fqid: b.serviceAccountFQID,
			},
			ExpiresAt: b.expiresAt,
			Type:      iam.Rs256Jwt,
		},
	}

	if b.description != "" {
		req.Spec.Description = &b.description
	}

	if b.accessTokenLifetime > 0 {
		req.Spec.Rs256Jwt = &iam.ServiceaccountcredentialSpecRs256Jwt{
			AccessTokenLifetime: &b.accessTokenLifetime,
		}
	}

	return req
}

// Create is a convenience method that builds and creates the credential in one call.
// The serviceAccountID is the short name (not FQID) needed to route to the correct API path.
func (b *ServiceAccountCredentialBuilder) Create(ctx context.Context, client *ServiceAccountCredentialsService) (*iam.Serviceaccountcredential, error) {
	req := b.Build()
	return client.Create(ctx, req)
}
