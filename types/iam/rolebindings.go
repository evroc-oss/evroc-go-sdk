// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

// AssignRoleRequest is the request body for assigning a role to a principal.
type AssignRoleRequest struct {
	// Principal is the fully-qualified principal FQID, e.g.
	// "/iam/users/{uuid}" or "/iam/projects/{project}/serviceAccounts/{saId}".
	Principal string `json:"principal"`

	// Role is the role FQID, e.g. "/iam/roles/computeOperator".
	Role string `json:"role"`

	// Resources optionally scopes the role to specific resource FQIDs.
	Resources []string `json:"resources,omitempty"`
}

// RevokeRoleRequest is the request body for revoking a role from a principal.
type RevokeRoleRequest struct {
	// Principal is the fully-qualified principal FQID.
	Principal string `json:"principal"`

	// Role is the role FQID to remove.
	Role string `json:"role"`
}

// RoleBindingResponse is returned by assign and revoke operations.
type RoleBindingResponse struct {
	PrincipalType     string              `json:"principalType"`
	PrincipalID       string              `json:"principalID"`
	DisplayName       string              `json:"displayName,omitempty"`
	Roles             []RoleBindingEntry  `json:"roles"`
	UID               string              `json:"uid"`
	ResourceVersion   string              `json:"resourceVersion"`
	CreationTimestamp string              `json:"creationTimestamp"`
}

// RoleBindingEntry represents a single role assignment within a binding.
type RoleBindingEntry struct {
	Name      string   `json:"name"`
	Resources []string `json:"resources,omitempty"`
}

// RoleInfo describes a role from the IAM role catalog.
type RoleInfo struct {
	// ID is the role FQID, e.g. "/iam/roles/computeOperator".
	ID string `json:"id"`

	// Description is a human-readable description of the role.
	Description string `json:"description"`

	// Scope is "organization" or "project".
	Scope string `json:"scope"`

	// Permissions lists the permissions this role grants.
	Permissions []string `json:"permissions"`
}

// RoleList is the response from the role catalog endpoint.
type RoleList struct {
	Items []RoleInfo `json:"items"`
}
