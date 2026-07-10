// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"path"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

// ServiceAccountCredentialsService provides operations for ServiceAccountCredentials.
// Credentials are a sub-resource of ServiceAccounts, so all operations require a
// parent service account ID.
type ServiceAccountCredentialsService struct {
	client           *Client
	serviceAccountID string
}

func (s *ServiceAccountCredentialsService) collectionPath() string {
	return path.Join("/", "iam", apiVersion, "projects",
		s.client.parent.DefaultProject(),
		resourceServiceAccounts, s.serviceAccountID,
		resourceServiceAccountCredentials)
}

func (s *ServiceAccountCredentialsService) resourcePath(name string) string {
	return path.Join(s.collectionPath(), name)
}

// Create creates a Serviceaccountcredential under the parent service account
func (s *ServiceAccountCredentialsService) Create(ctx context.Context, request *iam.ServiceaccountcredentialRequest) (*iam.Serviceaccountcredential, error) {
	return rest.CreateResource[*iam.Serviceaccountcredential](ctx, s.client.rest, s.collectionPath(), request)
}

// Delete deletes a Serviceaccountcredential by name
func (s *ServiceAccountCredentialsService) Delete(ctx context.Context, name string) error {
	return rest.DeleteResource(ctx, s.client.rest, s.resourcePath(name))
}

// Get retrieves a Serviceaccountcredential by name
func (s *ServiceAccountCredentialsService) Get(ctx context.Context, name string) (*iam.Serviceaccountcredential, error) {
	return rest.GetResource[*iam.Serviceaccountcredential](ctx, s.client.rest, s.resourcePath(name))
}

// List retrieves all Serviceaccountcredentials for the parent service account
func (s *ServiceAccountCredentialsService) List(ctx context.Context, filters ...rest.ListFilter) (*ServiceaccountcredentialList, error) {
	return rest.ListWithFilters[*ServiceaccountcredentialList](ctx, s.client.rest, s.collectionPath(), filters...)
}
