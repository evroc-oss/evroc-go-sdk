// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

// Role FQIDs for the evroc IAM role catalog.
// Use these constants when assigning roles via RoleBindings.
const (
	// Project roles
	RoleProjectOwner               = "/iam/roles/projectOwner"
	RoleProjectContributor         = "/iam/roles/projectContributor"
	RoleProjectAccessAdministrator = "/iam/roles/projectAccessAdministrator"

	// Compute
	RoleComputeViewer   = "/iam/roles/computeViewer"
	RoleComputeOperator = "/iam/roles/computeOperator"

	// Networking
	RoleNetworkingViewer   = "/iam/roles/networkingViewer"
	RoleNetworkingOperator = "/iam/roles/networkingOperator"
	RoleNetworkingUser     = "/iam/roles/networkingUser"

	// Load Balancer
	RoleLoadBalancerViewer   = "/iam/roles/loadBalancerViewer"
	RoleLoadBalancerOperator = "/iam/roles/loadBalancerOperator"

	// Storage
	RoleStorageOperator = "/iam/roles/storageOperator"

	// Think (AI inference)
	RoleThinkViewer   = "/iam/roles/thinkViewer"
	RoleThinkOperator = "/iam/roles/thinkOperator"

	// Cross-cutting
	RoleQuotaViewer               = "/iam/roles/quotaViewer"
	RoleCICDInfrastructureManager = "/iam/roles/cicdInfrastructureManager"
	RoleSecurityIncidentResponder = "/iam/roles/securityIncidentResponder"

	// Organization roles
	RoleOrganizationOwner               = "/iam/roles/organizationOwner"
	RoleOrganizationViewer              = "/iam/roles/organizationViewer"
	RoleOrganizationAccessAdministrator = "/iam/roles/organizationAccessAdministrator"
	RoleOrganizationProjectCreator      = "/iam/roles/organizationProjectCreator"
)
