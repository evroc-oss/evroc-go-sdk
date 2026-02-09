// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package quotas provides access to the evroc Quotas API.
//
// The Quotas API enables you to view and manage resource quotas at both
// organization and project levels in the evroc Cloud Platform.
//
// # Resources
//
// The quotas package provides access to the following resources:
//
//   - Organization Quotas: Quotas applied at the organization level
//   - Project Quotas: Quotas applied at the project level
//
// # Getting Started
//
// Create a client and list organization quotas:
//
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	orgQuotas, err := client.Quotas().OrganizationQuotas().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Organization Quotas
//
// View quotas for an organization:
//
//	orgQuota, err := client.Quotas().OrganizationQuotas().Get(
//	    ctx,
//	    client.DefaultOrganization(),
//	)
//
// Organization quotas define resource limits at the organization level,
// such as maximum number of VMs, storage capacity, and network resources.
//
// # Project Quotas
//
// View quotas for a project:
//
//	projectQuota, err := client.Quotas().ProjectQuotas().Get(
//	    ctx,
//	    client.DefaultProject(),
//	)
//
// Project quotas define resource limits for individual projects within
// an organization.
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	quota, err := client.Quotas().ProjectQuotas().Get(ctx, "my-project")
package quotas
