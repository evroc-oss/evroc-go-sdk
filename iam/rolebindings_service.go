// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

// RoleBindingsService provides operations for role bindings (assign/revoke roles).
type RoleBindingsService struct {
	client *Client
}

// AssignProjectRole grants a role to a principal at project scope.
func (s *RoleBindingsService) AssignProjectRole(ctx context.Context, req *iam.AssignRoleRequest) (*iam.RoleBindingResponse, error) {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		"roleBindings", "assign")
	return rest.CreateResource[*iam.RoleBindingResponse](ctx, s.client.rest, path, req)
}

// RevokeProjectRole removes a role from a principal at project scope.
func (s *RoleBindingsService) RevokeProjectRole(ctx context.Context, req *iam.RevokeRoleRequest) (*iam.RoleBindingResponse, error) {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		"roleBindings", "revoke")
	return rest.CreateResource[*iam.RoleBindingResponse](ctx, s.client.rest, path, req)
}

// ListRoles returns the IAM role catalog.
func (s *RoleBindingsService) ListRoles(ctx context.Context) (*iam.RoleList, error) {
	path := s.client.path.GlobalCollectionPath("roles")
	return rest.GetResource[*iam.RoleList](ctx, s.client.rest, path)
}
