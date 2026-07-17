// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

// Convenience aliases for generated role binding types.
type (
	RoleBinding          = Rolebinding
	RoleBindingList      = RolebindingList
	RoleBindingRequest   = RolebindingRequest
	RoleBindingSpec      = RolebindingSpec
	RoleBindingStatus    = RolebindingStatus

	OrgRoleBinding          = OrgResponse
	OrgRoleBindingList      = OrgListResponse
	OrgRoleBindingRequest   = OrgRequest
	OrgRoleBindingPatch     = OrgPatchRequest

	RoleList = RoleListResponse
)
