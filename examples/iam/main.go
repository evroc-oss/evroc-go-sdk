// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package main demonstrates comprehensive IAM API usage.
// Covers: Projects
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

	fmt.Println("\n=== All IAM Examples Complete ===")
}

// runProjectExamples demonstrates all project operations.
func runProjectExamples(ctx context.Context, client *evroc.Client, orgID string) error {
	fmt.Println("--- Project Examples ---")

	// Example 1: Create a simple project
	fmt.Println("\n1. Creating a simple project...")
	simpleProject, err := iam.NewProjectBuilder("sdk-project-simple", orgID).Build()
	if err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	createdSimple, err := client.IAM().Projects().Create(ctx, simpleProject)
	if err != nil {
		return fmt.Errorf("failed to create simple project: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", createdSimple.Metadata.Id)

	// Example 2: Create a project with name
	fmt.Println("\n2. Creating project with name...")
	namedProject, err := iam.NewProjectBuilder("sdk-project-dev", orgID).WithName("Development Environment").Build()
	if err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	createdNamed, err := client.IAM().Projects().Create(ctx, namedProject)
	if err != nil {
		return fmt.Errorf("failed to create named project: %w", err)
	}
	displayName := "none"
	if createdNamed.Spec.Name != nil {
		displayName = *createdNamed.Spec.Name
	}
	fmt.Printf("   ✓ Created: %s (display name: %s)\n", createdNamed.Metadata.Id, displayName)

	// Example 3: Create a project with labels
	fmt.Println("\n3. Creating project with labels...")
	labeledProject, err := iam.NewProjectBuilder("sdk-project-prod", orgID).
		WithName("Production Environment").
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
	fmt.Printf("   ✓ Created: %s\n", createdLabeled.Metadata.Id)

	// Example 4: Create a staging project
	fmt.Println("\n4. Creating staging project...")
	stagingProject, err := iam.NewProjectBuilder("sdk-project-staging", orgID).
		WithName("Staging Environment").
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
	fmt.Printf("   ✓ Created: %s\n", createdStaging.Metadata.Id)

	// Example 5: List all projects
	fmt.Println("\n5. Listing all projects...")
	projects, err := client.IAM().Projects().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	fmt.Printf("   Found %d projects:\n", len(projects.Items))
	for _, project := range projects.Items {
		displayName := "none"
		if project.Spec.Name != nil {
			displayName = *project.Spec.Name
		}
		fmt.Printf("   - %s: %s (org: %s)\n",
			project.Metadata.Id,
			displayName,
			project.Spec.Organization)
	}

	// Example 6: Get a specific project
	fmt.Println("\n6. Getting specific project...")
	project, err := client.IAM().Projects().Get(ctx, "sdk-project-prod")
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	fmt.Printf("   ✓ Project: %s\n", project.Metadata.Id)
	if project.Spec.Name != nil {
		fmt.Printf("     Display Name: %s\n", *project.Spec.Name)
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
	projectToUpdate, err := client.IAM().Projects().Get(ctx, "sdk-project-simple")
	if err != nil {
		log.Printf("   Warning: Failed to get project for update: %v", err)
	} else {
		newDisplayName := "Simple Demo Project"
		projectToUpdate.Spec.Name = &newDisplayName
		updatedProject, err := client.IAM().Projects().Patch(ctx, "sdk-project-simple", projectToUpdate)
		if err != nil {
			log.Printf("   Warning: Update failed (may not be supported): %v", err)
		} else {
			displayName := "none"
			if updatedProject.Spec.Name != nil {
				displayName = *updatedProject.Spec.Name
			}
			fmt.Printf("   ✓ Updated: %s (new display name: %s)\n",
				updatedProject.Metadata.Id,
				displayName)
		}
	}

	// Example 8: Delete a project
	fmt.Println("\n8. Deleting a project...")
	err = client.IAM().Projects().Delete(ctx, "sdk-project-staging")
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	fmt.Println("   ✓ Deleted sdk-project-staging")

	return nil
}
