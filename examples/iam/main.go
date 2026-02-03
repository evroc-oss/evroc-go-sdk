// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates comprehensive IAM API usage.
// Covers: Projects, PermissionSets
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/iam"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get organization from environment or config
	orgID := os.Getenv("EVROC_ORGANIZATION")
	if orgID == "" {
		log.Fatal("EVROC_ORGANIZATION environment variable must be set for IAM examples")
	}

	fmt.Println("=== Comprehensive IAM API Examples ===")
	fmt.Println()

	// Run all examples
	if err := runProjectExamples(ctx, client, orgID); err != nil {
		log.Printf("Project examples failed: %v", err)
	}

	if err := runPermissionSetExamples(ctx, client); err != nil {
		log.Printf("Permission set examples failed: %v", err)
	}

	fmt.Println("\n=== All IAM Examples Complete ===")
}

// runProjectExamples demonstrates all project operations.
func runProjectExamples(ctx context.Context, client *evroc.Client, orgID string) error {
	fmt.Println("--- Project Examples ---")

	// Example 1: Create a simple project
	fmt.Println("\n1. Creating a simple project...")
	simpleProject, err := iam.NewProjectBuilder("example-project-simple", orgID).Build()
	if err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	createdSimple, err := client.IAM().Projects().Create(ctx, simpleProject)
	if err != nil {
		return fmt.Errorf("failed to create simple project: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdSimple.Metadata.Name)

	// Example 2: Create a project with display name
	fmt.Println("\n2. Creating project with display name...")
	namedProject, err := iam.NewProjectBuilder("example-project-dev", orgID).WithDisplayName("Development Environment").Build()
	if err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	createdNamed, err := client.IAM().Projects().Create(ctx, namedProject)
	if err != nil {
		return fmt.Errorf("failed to create named project: %w", err)
	}
	displayName := "none"
	if createdNamed.Spec.DisplayName != nil {
		displayName = *createdNamed.Spec.DisplayName
	}
	fmt.Printf("   ✓ Created: %s (display name: %s)\n", *createdNamed.Metadata.Name, displayName)

	// Example 3: Create a project with labels
	fmt.Println("\n3. Creating project with labels...")
	labeledProject, err := iam.NewProjectBuilder("example-project-prod", orgID).
		WithDisplayName("Production Environment").
		WithLabels(map[string]string{
			"environment": "production",
			"team":        "platform",
			"cost-center": "engineering",
		}).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	createdLabeled, err := client.IAM().Projects().Create(ctx, labeledProject)
	if err != nil {
		return fmt.Errorf("failed to create labeled project: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdLabeled.Metadata.Name)

	// Example 4: Create a staging project
	fmt.Println("\n4. Creating staging project...")
	stagingProject, err := iam.NewProjectBuilder("example-project-staging", orgID).
		WithDisplayName("Staging Environment").
		WithLabels(map[string]string{
			"environment": "staging",
			"team":        "qa",
		}).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	createdStaging, err := client.IAM().Projects().Create(ctx, stagingProject)
	if err != nil {
		return fmt.Errorf("failed to create staging project: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdStaging.Metadata.Name)

	// Example 5: List all projects
	fmt.Println("\n5. Listing all projects...")
	projects, err := client.IAM().Projects().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	fmt.Printf("   Found %d projects:\n", len(projects.Items))
	for _, project := range projects.Items {
		displayName := "none"
		if project.Spec.DisplayName != nil {
			displayName = *project.Spec.DisplayName
		}
		fmt.Printf("   - %s: %s (org: %s)\n",
			*project.Metadata.Name,
			displayName,
			project.Spec.Organization)
	}

	// Example 6: Get a specific project
	fmt.Println("\n6. Getting specific project...")
	project, err := client.IAM().Projects().Get(ctx, "example-project-prod")
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	fmt.Printf("   ✓ Project: %s\n", *project.Metadata.Name)
	if project.Spec.DisplayName != nil {
		fmt.Printf("     Display Name: %s\n", *project.Spec.DisplayName)
	}
	fmt.Printf("     Organization: %s\n", project.Spec.Organization)
	if project.Metadata.UserLabels != nil {
		fmt.Println("     Labels:")
		for k, v := range *project.Metadata.UserLabels {
			fmt.Printf("       %s: %s\n", k, v)
		}
	}

	// Example 7: Update project display name
	fmt.Println("\n7. Updating project display name...")
	projectToUpdate, err := client.IAM().Projects().Get(ctx, "example-project-simple")
	if err != nil {
		log.Printf("   Warning: Failed to get project for update: %v", err)
	} else {
		newDisplayName := "Simple Demo Project"
		projectToUpdate.Spec.DisplayName = &newDisplayName
		updatedProject, err := client.IAM().Projects().Update(ctx, "example-project-simple", projectToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			displayName := "none"
			if updatedProject.Spec.DisplayName != nil {
				displayName = *updatedProject.Spec.DisplayName
			}
			fmt.Printf("   ✓ Updated: %s (new display name: %s)\n",
				*updatedProject.Metadata.Name,
				displayName)
		}
	}

	// Example 8: Delete a project
	fmt.Println("\n8. Deleting a project...")
	err = client.IAM().Projects().Delete(ctx, "example-project-staging")
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	fmt.Println("   ✓ Deleted example-project-staging")

	return nil
}

// runPermissionSetExamples demonstrates all permission set operations.
func runPermissionSetExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Permission Set Examples ---")

	// Note: Permission sets require a valid project ID
	// We'll use one of the projects created above
	projectID := "example-project-dev"

	// Example 1: Create a standard user permission set
	fmt.Println("\n1. Creating standard user permission set...")
	userPS := iam.NewPermissionSetBuilder("example-ps-user", projectID, "user@example.com").
		WithAdmin(false).
		Build()

	createdUser, err := client.IAM().PermissionSets().Create(ctx, userPS)
	if err != nil {
		return fmt.Errorf("failed to create user permission set: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (admin: %t, email: %s)\n",
		*createdUser.Metadata.Name,
		createdUser.Spec.Admin,
		createdUser.Spec.Subject.User.Email)

	// Example 2: Create an admin permission set
	fmt.Println("\n2. Creating admin permission set...")
	adminPS := iam.NewPermissionSetBuilder("example-ps-admin", projectID, "admin@example.com").
		WithAdmin(true).
		Build()

	createdAdmin, err := client.IAM().PermissionSets().Create(ctx, adminPS)
	if err != nil {
		return fmt.Errorf("failed to create admin permission set: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (admin: %t)\n",
		*createdAdmin.Metadata.Name,
		createdAdmin.Spec.Admin)

	// Example 3: Create a permission set with labels
	fmt.Println("\n3. Creating permission set with labels...")
	labeledPS := iam.NewPermissionSetBuilder("example-ps-developer", projectID, "dev@example.com").
		WithAdmin(false).
		WithLabels(map[string]string{
			"role":       "developer",
			"department": "engineering",
			"access":     "limited",
		}).
		Build()

	createdLabeled, err := client.IAM().PermissionSets().Create(ctx, labeledPS)
	if err != nil {
		return fmt.Errorf("failed to create labeled permission set: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdLabeled.Metadata.Name)

	// Example 4: Create permission sets for different users
	fmt.Println("\n4. Creating permission sets for team members...")
	teamMembers := []struct {
		name  string
		email string
		admin bool
	}{
		{"example-ps-alice", "alice@example.com", false},
		{"example-ps-bob", "bob@example.com", false},
		{"example-ps-charlie", "charlie@example.com", true},
	}

	for _, member := range teamMembers {
		ps := iam.NewPermissionSetBuilder(member.name, projectID, member.email).
			WithAdmin(member.admin).
			Build()

		createdPS, err := client.IAM().PermissionSets().Create(ctx, ps)
		if err != nil {
			log.Printf("   Warning: Failed to create permission set for %s: %v", member.email, err)
			continue
		}
		role := "user"
		if member.admin {
			role = "admin"
		}
		fmt.Printf("   ✓ Created: %s (email: %s, role: %s)\n",
			*createdPS.Metadata.Name,
			member.email,
			role)
	}

	// Example 5: List all permission sets
	fmt.Println("\n5. Listing all permission sets...")
	permissionSets, err := client.IAM().PermissionSets().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list permission sets: %w", err)
	}
	fmt.Printf("   Found %d permission sets:\n", len(permissionSets.Items))
	for _, ps := range permissionSets.Items {
		role := "user"
		if ps.Spec.Admin {
			role = "admin"
		}
		fmt.Printf("   - %s: %s [%s]\n",
			*ps.Metadata.Name,
			ps.Spec.Subject.User.Email,
			role)
	}

	// Example 6: Get a specific permission set
	fmt.Println("\n6. Getting specific permission set...")
	ps, err := client.IAM().PermissionSets().Get(ctx, "example-ps-developer")
	if err != nil {
		return fmt.Errorf("failed to get permission set: %w", err)
	}
	fmt.Printf("   ✓ Permission Set: %s\n", *ps.Metadata.Name)
	fmt.Printf("     Subject: %s (%s)\n", ps.Spec.Subject.User.Email, ps.Spec.Subject.Type)
	fmt.Printf("     Admin: %t\n", ps.Spec.Admin)
	if ps.Metadata.UserLabels != nil {
		fmt.Println("     Labels:")
		for k, v := range *ps.Metadata.UserLabels {
			fmt.Printf("       %s: %s\n", k, v)
		}
	}

	// Example 7: Update permission set (promote user to admin)
	fmt.Println("\n7. Updating permission set (promote to admin)...")
	psToUpdate, err := client.IAM().PermissionSets().Get(ctx, "example-ps-user")
	if err != nil {
		log.Printf("   Warning: Failed to get permission set for update: %v", err)
	} else {
		psToUpdate.Spec.Admin = true
		updatedPS, err := client.IAM().PermissionSets().Update(ctx, "example-ps-user", psToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			fmt.Printf("   ✓ Updated: %s (admin: %t)\n",
				*updatedPS.Metadata.Name,
				updatedPS.Spec.Admin)
		}
	}

	// Example 8: Delete a permission set
	fmt.Println("\n8. Deleting a permission set...")
	err = client.IAM().PermissionSets().Delete(ctx, "example-ps-bob")
	if err != nil {
		return fmt.Errorf("failed to delete permission set: %w", err)
	}
	fmt.Println("   ✓ Deleted example-ps-bob")

	return nil
}
