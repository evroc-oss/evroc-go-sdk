// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"regexp"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

// RoleBindingsService provides operations for role bindings (assign/revoke roles).
type RoleBindingsService struct {
	client *Client
}

var (
	userPrincipalRe           = regexp.MustCompile(`^/iam/users/([^/]+)$`)
	serviceAccountPrincipalRe = regexp.MustCompile(`^/iam/projects/([^/]+)/serviceAccounts/([^/]+)$`)
)

// DeriveRoleBindingName derives the conventional role binding name from a
// principal FQID: u-{user uuid} for users, sa-{project}.{service account name}
// for service accounts. IAM enforces this convention server-side, so
// CreateProjectRoleBinding and CreateOrgRoleBinding always set it themselves.
// Exported so callers (e.g. the Terraform provider) can recompute the name a
// binding must have from a principal alone, without an extra API call.
func DeriveRoleBindingName(principal string) (string, error) {
	if m := userPrincipalRe.FindStringSubmatch(principal); m != nil {
		return "u-" + m[1], nil
	}
	if m := serviceAccountPrincipalRe.FindStringSubmatch(principal); m != nil {
		return fmt.Sprintf("sa-%s.%s", m[1], m[2]), nil
	}
	return "", fmt.Errorf("cannot derive role binding name from principal %q: expected /iam/users/<uuid> or /iam/projects/<project>/serviceAccounts/<name>", principal)
}

// --- Project-scoped CRUD ---

// CreateProjectRoleBinding creates a project-scoped role binding. The name is
// always derived from request.Spec.Principal; any Metadata.Id the caller sets
// is overwritten.
func (s *RoleBindingsService) CreateProjectRoleBinding(ctx context.Context, request *iam.RolebindingRequest) (*iam.Rolebinding, error) {
	name, err := DeriveRoleBindingName(request.Spec.Principal)
	if err != nil {
		return nil, err
	}
	request.Metadata.Id = name

	path := s.client.path.ProjectCollectionPath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings)
	return rest.CreateResource[*iam.Rolebinding](ctx, s.client.rest, path, request)
}

// GetProjectRoleBinding retrieves a project-scoped role binding by name.
func (s *RoleBindingsService) GetProjectRoleBinding(ctx context.Context, name string) (*iam.Rolebinding, error) {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings, name)
	return rest.GetResource[*iam.Rolebinding](ctx, s.client.rest, path)
}

// ListProjectRoleBindings retrieves all project-scoped role bindings with optional filters.
func (s *RoleBindingsService) ListProjectRoleBindings(ctx context.Context, filters ...rest.ListFilter) (*iam.RolebindingList, error) {
	path := s.client.path.ProjectCollectionPath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings)
	return rest.ListWithFilters[*iam.RolebindingList](ctx, s.client.rest, path, filters...)
}

// DeleteProjectRoleBinding deletes a project-scoped role binding.
func (s *RoleBindingsService) DeleteProjectRoleBinding(ctx context.Context, name string) error {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings, name)
	return rest.DeleteResource(ctx, s.client.rest, path)
}

// PatchProjectRoleBinding partially updates a project-scoped role binding.
func (s *RoleBindingsService) PatchProjectRoleBinding(ctx context.Context, name string, patch *iam.RolebindingPatchRequest) (*iam.Rolebinding, error) {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings, name)
	return rest.PatchResource[*iam.Rolebinding](ctx, s.client.rest, path, patch)
}

// --- Project-scoped assign / revoke ---

// AssignProjectRole grants a role to a principal at project scope.
func (s *RoleBindingsService) AssignProjectRole(ctx context.Context, req *iam.AssignRoleRequest) (*iam.Rolebinding, error) {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings, "assign")
	return rest.CreateResource[*iam.Rolebinding](ctx, s.client.rest, path, req)
}

// RevokeProjectRole removes a role from a principal at project scope.
func (s *RoleBindingsService) RevokeProjectRole(ctx context.Context, req *iam.RevokeRoleRequest) (*iam.Rolebinding, error) {
	path := s.client.path.ProjectResourcePath(
		s.client.parent.DefaultProject(),
		resourceRoleBindings, "revoke")
	return rest.CreateResource[*iam.Rolebinding](ctx, s.client.rest, path, req)
}

// --- Organization-scoped CRUD ---

// CreateOrgRoleBinding creates an organization-scoped role binding. The name
// is always derived from req.Spec.Principal; any Metadata.Id the caller sets
// is overwritten.
func (s *RoleBindingsService) CreateOrgRoleBinding(ctx context.Context, req *iam.OrgRequest) (*iam.OrgResponse, error) {
	name, err := DeriveRoleBindingName(req.Spec.Principal)
	if err != nil {
		return nil, err
	}
	req.Metadata.Id = name

	path := s.client.path.OrgScopedCollectionPath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings)
	return rest.CreateResource[*iam.OrgResponse](ctx, s.client.rest, path, req)
}

// GetOrgRoleBinding retrieves an organization-scoped role binding by name.
func (s *RoleBindingsService) GetOrgRoleBinding(ctx context.Context, name string) (*iam.OrgResponse, error) {
	path := s.client.path.OrgScopedResourcePath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings, name)
	return rest.GetResource[*iam.OrgResponse](ctx, s.client.rest, path)
}

// ListOrgRoleBindings lists the role bindings in the default organization.
func (s *RoleBindingsService) ListOrgRoleBindings(ctx context.Context, filters ...rest.ListFilter) (*iam.OrgListResponse, error) {
	path := s.client.path.OrgScopedCollectionPath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings)
	return rest.ListWithFilters[*iam.OrgListResponse](ctx, s.client.rest, path, filters...)
}

// DeleteOrgRoleBinding deletes an organization-scoped role binding.
func (s *RoleBindingsService) DeleteOrgRoleBinding(ctx context.Context, name string) error {
	path := s.client.path.OrgScopedResourcePath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings, name)
	return rest.DeleteResource(ctx, s.client.rest, path)
}

// PatchOrgRoleBinding partially updates an organization-scoped role binding.
func (s *RoleBindingsService) PatchOrgRoleBinding(ctx context.Context, name string, patch *iam.OrgPatchRequest) (*iam.OrgResponse, error) {
	path := s.client.path.OrgScopedResourcePath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings, name)
	return rest.PatchResource[*iam.OrgResponse](ctx, s.client.rest, path, patch)
}

// --- Organization-scoped assign / revoke ---

// AssignOrgRole grants a role to a principal at organization scope.
func (s *RoleBindingsService) AssignOrgRole(ctx context.Context, req *iam.AssignRoleRequest) (*iam.OrgResponse, error) {
	path := s.client.path.OrgScopedResourcePath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings, "assign")
	return rest.CreateResource[*iam.OrgResponse](ctx, s.client.rest, path, req)
}

// RevokeOrgRole removes a role from a principal at organization scope.
func (s *RoleBindingsService) RevokeOrgRole(ctx context.Context, req *iam.RevokeRoleRequest) (*iam.OrgResponse, error) {
	path := s.client.path.OrgScopedResourcePath(
		s.client.parent.DefaultOrganization(),
		resourceRoleBindings, "revoke")
	return rest.CreateResource[*iam.OrgResponse](ctx, s.client.rest, path, req)
}

// --- Role catalog ---

// ListRoles returns the IAM role catalog.
func (s *RoleBindingsService) ListRoles(ctx context.Context) (*iam.RoleListResponse, error) {
	path := s.client.path.GlobalCollectionPath("roles")
	return rest.GetResource[*iam.RoleListResponse](ctx, s.client.rest, path)
}

// --- Caller's own role bindings ---

// ListMyRoleBindings returns every role binding granted to the authenticated caller.
func (s *RoleBindingsService) ListMyRoleBindings(ctx context.Context) (*iam.RolebindingList, error) {
	path := s.client.path.GlobalResourcePath("me", "roleBindings")
	return rest.GetResource[*iam.RolebindingList](ctx, s.client.rest, path)
}

// --- Access evaluation ---

// TestPermissions evaluates which of the listed permissions the calling identity holds.
func (s *RoleBindingsService) TestPermissions(ctx context.Context, req *iam.TestPermissionsRequest) (*iam.TestPermissionsResponse, error) {
	path := s.client.path.GlobalCollectionPath("testPermissions")
	return rest.CreateResource[*iam.TestPermissionsResponse](ctx, s.client.rest, path, req)
}

// CheckAccess evaluates which of the listed permissions a given principal holds.
func (s *RoleBindingsService) CheckAccess(ctx context.Context, req *iam.CheckAccessRequest) (*iam.CheckAccessResponse, error) {
	path := s.client.path.GlobalCollectionPath("checkAccess")
	return rest.CreateResource[*iam.CheckAccessResponse](ctx, s.client.rest, path, req)
}
