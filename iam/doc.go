// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package iam provides access to the evroc Identity and Access Management (IAM) API.
//
// The IAM API enables you to manage organizations, projects, users, and permissions
// in the evroc Cloud Platform.
//
// # Resources
//
// The iam package provides access to the following resources:
//
//   - Organizations: Top-level containers for resources
//   - Projects: Isolated environments within organizations
//   - Users: User accounts and identities
//   - Permission Sets: Role-based access control policies
//
// # Getting Started
//
// Create a client and list projects:
//
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	projects, err := client.IAM().Projects().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Organizations
//
// List and view organizations:
//
//	orgs, err := client.IAM().Organizations().List(ctx)
//
//	org, err := client.IAM().Organizations().Get(ctx, "my-org")
//
// # Projects
//
// Create projects with builders:
//
//	orgID := client.DefaultOrganization()
//	project, err := client.IAM().Projects().Create(ctx,
//	    iam.NewProjectBuilder("dev-project", orgID).
//	        WithName("Development environment").
//	        WithLabels(map[string]string{"env": "dev"}).
//	        Build(),
//	)
//
// # Users
//
// Manage user accounts:
//
//	users, err := client.IAM().Users().List(ctx)
//
//	user, err := client.IAM().Users().Get(ctx, "user@example.com")
//
// # Permission Sets
//
// Create and manage permission sets for access control:
//
//	permSet, err := client.IAM().PermissionSets().Create(ctx,
//	    iam.NewPermissionSetBuilder("developer-access").
//	        WithDescription("Developer permissions").
//	        WithProject("dev-project").
//	        Build(),
//	)
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	project, err := client.IAM().Projects().Get(ctx, "my-project")
package iam
