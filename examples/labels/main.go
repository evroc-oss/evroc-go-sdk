// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package main demonstrates label usage and filtering across all API resources.
// Shows how to use labels for resource organization, filtering, and management.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/iam"
	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/networking"
	"github.com/evroc-oss/evroc-go-sdk/storage"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Label Usage and Filtering Examples ===")
	fmt.Println()

	// Run all examples
	if err := createResourcesWithLabels(ctx, client); err != nil {
		log.Printf("Failed to create resources: %v", err)
		return
	}

	if err := filterResourcesByLabels(ctx, client); err != nil {
		log.Printf("Failed to filter resources: %v", err)
		return
	}

	if err := demonstrateLabelBestPractices(ctx, client); err != nil {
		log.Printf("Failed to demonstrate best practices: %v", err)
	}

	fmt.Println("\n=== All Label Examples Complete ===")
}

// createResourcesWithLabels creates resources with various label schemes.
func createResourcesWithLabels(ctx context.Context, client *evroc.Client) error {
	fmt.Println("--- Creating Resources with Labels ---")

	// Example 1: Create resources with environment labels
	fmt.Println("\n1. Creating resources with environment=production labels...")

	// Create disk with environment label
	prodDisk := compute.NewDiskBuilder("sdk-prod-disk").
		WithImage(string(compute.DiskImageUbuntu2404)).
		WithSizeGB(50).
		WithZone("a").
		WithLabels(map[string]string{
			"environment": "production",
			"managed-by":  "devops",
			"cost-center": "platform",
		}).
		Build()

	createdProdDisk, err := client.Compute().Disks().Create(ctx, prodDisk)
	if err != nil {
		return fmt.Errorf("failed to create prod disk: %w", err)
	}
	fmt.Println("   ✓ Created disk with production labels")

	// Wait for disk
	client.Compute().Disks().WaitForReady(ctx, "sdk-prod-disk", 2*time.Minute)

	// Create VM with environment label
	prodVM := compute.NewVirtualMachineBuilder("sdk-prod-vm").
		WithBootDisk(createdProdDisk.Ref()).
		WithSize(string(compute.VMSizeA1aXS)).
		WithZone("a").
		WithLabels(map[string]string{
			"environment": "production",
			"managed-by":  "devops",
			"app":         "web-server",
			"tier":        "frontend",
		}).
		Build()

	_, err = client.Compute().VirtualMachines().Create(ctx, prodVM)
	if err != nil {
		return fmt.Errorf("failed to create prod VM: %w", err)
	}
	fmt.Println("   ✓ Created VM with production labels")

	// Example 2: Create resources with development labels
	fmt.Println("\n2. Creating resources with environment=development labels...")

	devDisk := compute.NewDiskBuilder("sdk-dev-disk").
		WithImage("ubuntu-minimal.24-04.1").
		WithSizeGB(20).
		WithZone("a").
		WithLabels(map[string]string{
			"environment": "development",
			"managed-by":  "developers",
			"temporary":   "true",
		}).
		Build()

	_, err = client.Compute().Disks().Create(ctx, devDisk)
	if err != nil {
		return fmt.Errorf("failed to create dev disk: %w", err)
	}
	fmt.Println("   ✓ Created disk with development labels")

	// Example 3: Create resources with team labels
	fmt.Println("\n3. Creating resources with team labels...")

	// Create public IP with team label
	teamIP := networking.NewPublicIPBuilder("sdk-team-ip").
		WithLabels(map[string]string{
			"team":        "platform",
			"project":     "infrastructure",
			"environment": "production",
		}).
		Build()

	_, err = client.Networking().PublicIPs().Create(ctx, teamIP)
	if err != nil {
		return fmt.Errorf("failed to create team IP: %w", err)
	}
	fmt.Println("   ✓ Created public IP with team labels")

	// Create security group with team label
	teamSG := networking.NewSecurityGroupBuilder("sdk-team-sg").
		AllowSSH().
		WithLabels(map[string]string{
			"team":    "platform",
			"purpose": "ssh-access",
		}).
		Build()

	_, err = client.Networking().SecurityGroups().Create(ctx, teamSG)
	if err != nil {
		return fmt.Errorf("failed to create team SG: %w", err)
	}
	fmt.Println("   ✓ Created security group with team labels")

	// Example 4: Create storage resources with labels
	fmt.Println("\n4. Creating storage resources with labels...")

	bucket := storage.NewBucketBuilder("sdk-bucket").
		WithLabels(map[string]string{
			"purpose":     "backups",
			"retention":   "30-days",
			"environment": "production",
		}).
		Build()

	_, err = client.Storage().Buckets().Create(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	fmt.Println("   ✓ Created bucket with purpose labels")

	// Example 5: Create IAM resources with labels
	fmt.Println("\n5. Creating IAM resources with labels...")

	orgID := os.Getenv("EVROC_ORGANIZATION")
	if orgID != "" {
		project, err := iam.NewProjectBuilder("sdk-project", orgID).
			WithName("Label Demo Project").
			WithLabels(map[string]string{
				"department":  "engineering",
				"cost-center": "r-and-d",
				"managed-by":  "platform-team",
			}).
			Build()
		if err != nil {
			log.Printf("   Warning: Failed to build project: %v", err)
		} else {

			_, err = client.IAM().Projects().Create(ctx, project)
			if err != nil {
				log.Printf("   Warning: Failed to create project: %v", err)
			} else {
				fmt.Println("   ✓ Created project with organizational labels")
			}
		}
	}

	return nil
}

// filterResourcesByLabels demonstrates filtering resources using labels.
func filterResourcesByLabels(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Filtering Resources by Labels ---")

	// Example 1: Filter disks by environment
	fmt.Println("\n1. Filtering disks by environment=production...")
	prodDisks, err := client.Compute().Disks().List(ctx,
		rest.WithLabelSelector(map[string]string{"environment": "production"}))
	if err != nil {
		return fmt.Errorf("failed to filter prod disks: %w", err)
	}
	fmt.Printf("   Found %d production disks:\n", len(prodDisks.Items))
	for _, disk := range prodDisks.Items {
		fmt.Printf("   - %s\n", disk.Metadata.Id)
	}

	// Example 2: Filter VMs by multiple labels
	fmt.Println("\n2. Filtering VMs by environment=production AND managed-by=devops...")
	prodVMs, err := client.Compute().VirtualMachines().List(ctx,
		rest.WithLabelSelector(map[string]string{
			"environment": "production",
			"managed-by":  "devops",
		}))
	if err != nil {
		return fmt.Errorf("failed to filter prod VMs: %w", err)
	}
	fmt.Printf("   Found %d production VMs managed by devops:\n", len(prodVMs.Items))
	for _, vm := range prodVMs.Items {
		fmt.Printf("   - %s\n", vm.Metadata.Id)
	}

	// Example 3: Filter by team
	fmt.Println("\n3. Filtering public IPs by team=platform...")
	teamIPs, err := client.Networking().PublicIPs().List(ctx,
		rest.WithLabelSelector(map[string]string{"team": "platform"}))
	if err != nil {
		return fmt.Errorf("failed to filter team IPs: %w", err)
	}
	fmt.Printf("   Found %d public IPs for platform team:\n", len(teamIPs.Items))
	for _, ip := range teamIPs.Items {
		fmt.Printf("   - %s\n", ip.Metadata.Id)
	}

	// Example 4: Filter security groups
	fmt.Println("\n4. Filtering security groups by team=platform...")
	teamSGs, err := client.Networking().SecurityGroups().List(ctx,
		rest.WithLabelSelector(map[string]string{"team": "platform"}))
	if err != nil {
		return fmt.Errorf("failed to filter team SGs: %w", err)
	}
	fmt.Printf("   Found %d security groups for platform team:\n", len(teamSGs.Items))
	for _, sg := range teamSGs.Items {
		fmt.Printf("   - %s\n", sg.Metadata.Id)
	}

	// Example 5: Filter buckets by purpose
	fmt.Println("\n5. Filtering buckets by purpose=backups...")
	backupBuckets, err := client.Storage().Buckets().List(ctx,
		rest.WithLabelSelector(map[string]string{"purpose": "backups"}))
	if err != nil {
		return fmt.Errorf("failed to filter backup buckets: %w", err)
	}
	fmt.Printf("   Found %d backup buckets:\n", len(backupBuckets.Items))
	for _, bucket := range backupBuckets.Items {
		fmt.Printf("   - %s\n", bucket.Metadata.Id)
	}

	// Example 6: Filter development resources
	fmt.Println("\n6. Filtering all development disks...")
	devDisks, err := client.Compute().Disks().List(ctx,
		rest.WithLabelSelector(map[string]string{"environment": "development"}))
	if err != nil {
		return fmt.Errorf("failed to filter dev disks: %w", err)
	}
	fmt.Printf("   Found %d development disks:\n", len(devDisks.Items))
	for _, disk := range devDisks.Items {
		fmt.Printf("   - %s\n", disk.Metadata.Id)
	}

	return nil
}

// demonstrateLabelBestPractices shows recommended label usage patterns.
func demonstrateLabelBestPractices(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Label Best Practices ---")

	fmt.Println("\n1. Recommended Label Patterns:")
	fmt.Println("   Common Labels:")
	fmt.Println("   - environment: production, staging, development")
	fmt.Println("   - team: platform, frontend, backend, data")
	fmt.Println("   - cost-center: engineering, marketing, sales")
	fmt.Println("   - managed-by: terraform, manual, sdk")
	fmt.Println("   - project: project-name")
	fmt.Println("   - app: application-name")
	fmt.Println("   - tier: frontend, backend, database")
	fmt.Println("   - temporary: true, false")

	fmt.Println("\n2. Creating resources with comprehensive labels...")

	// Example: Create a well-labeled VM
	wellLabeledVM := compute.NewVirtualMachineBuilder("sdk-best-practice-vm").
		WithSize(string(compute.VMSizeA1aXS)).
		WithZone("a").
		WithLabels(map[string]string{
			// Environment classification
			"environment": "production",

			// Ownership and management
			"team":        "platform",
			"managed-by":  "infrastructure-team",
			"cost-center": "engineering",

			// Application identification
			"app":     "api-server",
			"tier":    "backend",
			"version": "v2",

			// Operational metadata
			"backup":    "daily",
			"monitored": "true",
		}).
		Build()

	fmt.Println("   Example VM labels:")
	if wellLabeledVM.Metadata.UserLabels != nil {
		for k, v := range *wellLabeledVM.Metadata.UserLabels {
			fmt.Printf("     %s: %s\n", k, v)
		}
	}

	fmt.Println("\n3. Label Filtering Tips:")
	fmt.Println("   - Use consistent label keys across all resources")
	fmt.Println("   - Use lowercase values for easier filtering")
	fmt.Println("   - Combine multiple labels for precise filtering")
	fmt.Println("   - Document your labeling schema")
	fmt.Println("   - Use labels for cost allocation and reporting")

	fmt.Println("\n4. Common Label Use Cases:")
	fmt.Println("   - Finding all resources for a specific environment")
	fmt.Println("   - Identifying resources by team or cost center")
	fmt.Println("   - Locating temporary resources for cleanup")
	fmt.Println("   - Filtering resources by application or service")
	fmt.Println("   - Tracking managed vs. manual resources")

	return nil
}
