// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates comprehensive networking API usage.
// Covers: PublicIPs, SecurityGroups, VPCs (list only), Subnets (list only)
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
	"github.com/evroc-oss/evroc-go-sdk/networking"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Comprehensive Networking API Examples ===")
	fmt.Println()

	// Cleanup at start to ensure clean slate
	fmt.Println("0. Cleaning up any existing resources from previous runs...")
	cleanupResources(ctx, client)
	fmt.Println()

	// Run all examples
	if err := runPublicIPExamples(ctx, client); err != nil {
		log.Printf("Public IP examples failed: %v", err)
	}

	if err := runSecurityGroupExamples(ctx, client); err != nil {
		log.Printf("Security group examples failed: %v", err)
	}

	if err := runVPCAndSubnetExamples(ctx, client); err != nil {
		log.Printf("VPC/Subnet examples failed: %v", err)
	}

	// Cleanup at end
	fmt.Println("\n=== Cleaning Up Resources ===")
	cleanupResources(ctx, client)

	fmt.Println("\n=== All Networking Examples Complete ===")
}

// cleanupResources deletes all resources created by this example
func cleanupResources(ctx context.Context, client *evroc.Client) {
	// Delete public IPs
	publicIPsToDelete := []string{"sdk-public-ip-1", "sdk-public-ip-2", "sdk-public-ip-3"}
	for _, ipName := range publicIPsToDelete {
		err := client.Networking().PublicIPs().Delete(ctx, ipName)
		if err == nil {
			fmt.Printf("   ✓ Deleted %s\n", ipName)
		}
	}

	// Delete security groups
	securityGroupsToDelete := []string{"sdk-sg-ssh", "sdk-sg-web", "sdk-sg-custom", "sdk-sg-database"}
	for _, sgName := range securityGroupsToDelete {
		err := client.Networking().SecurityGroups().Delete(ctx, sgName)
		if err == nil {
			fmt.Printf("   ✓ Deleted %s\n", sgName)
		}
	}
}

// runPublicIPExamples demonstrates all public IP operations.
func runPublicIPExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("--- Public IP Examples ---")

	// Example 1: Create a public IP using builder
	fmt.Println("\n1. Creating public IP using builder...")
	pubIP1 := networking.NewPublicIPBuilder("sdk-public-ip-1").Build()

	createdPubIP1, err := client.Networking().PublicIPs().Create(ctx, pubIP1)
	if err != nil {
		return fmt.Errorf("failed to create public IP 1: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdPubIP1.Metadata.Id)

	// Example 2: Create another public IP
	fmt.Println("\n2. Creating another public IP...")
	pubIP2 := networking.NewPublicIPBuilder("sdk-public-ip-2").Build()

	createdPubIP2, err := client.Networking().PublicIPs().Create(ctx, pubIP2)
	if err != nil {
		return fmt.Errorf("failed to create public IP 2: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdPubIP2.Metadata.Id)

	// Example 3: Create public IP using convenience method
	fmt.Println("\n3. Creating public IP using Create method directly...")
	pubIP3, err := client.Networking().PublicIPs().Create(ctx,
		networking.NewPublicIPBuilder("sdk-public-ip-3").Build())
	if err != nil {
		return fmt.Errorf("failed to create public IP 3: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *pubIP3.Metadata.Id)

	// Wait for public IPs to be ready
	fmt.Println("\n4. Waiting for public IPs to be ready...")
	time.Sleep(5 * time.Second)  // Give some time for IPs to be allocated

	// Example 5: List all public IPs
	fmt.Println("\n5. Listing all public IPs...")
	publicIPs, err := client.Networking().PublicIPs().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list public IPs: %w", err)
	}
	fmt.Printf("   Found %d public IPs:\n", len(publicIPs.Items))
	for _, ip := range publicIPs.Items {
		status := "Pending"
		ipAddr := "Not allocated"
		if networking.IsPublicIPReady(&ip) {
			status = "Ready"
			ipAddr = networking.GetPublicIPAddress(&ip)
		}
		fmt.Printf("   - %s: %s [%s]\n", *ip.Metadata.Id, ipAddr, status)
	}

	// Example 6: Get a specific public IP
	fmt.Println("\n6. Getting specific public IP...")
	pubIP, err := client.Networking().PublicIPs().Get(ctx, "sdk-public-ip-1")
	if err != nil {
		return fmt.Errorf("failed to get public IP: %w", err)
	}
	fmt.Printf("   ✓ Public IP: %s\n", *pubIP.Metadata.Id)
	if pubIP.Status.PublicIPv4Address != nil {
		fmt.Printf("     IPv4 Address: %s\n", *pubIP.Status.PublicIPv4Address)
	}
	if pubIP.Status.Conditions != nil {
		fmt.Printf("     Conditions: %d\n", len(*pubIP.Status.Conditions))
	}

	// Example 7: Check if public IP is ready
	fmt.Println("\n7. Checking public IP readiness...")
	if networking.IsPublicIPReady(pubIP) {
		fmt.Printf("   ✓ %s is ready\n", *pubIP.Metadata.Id)
		fmt.Printf("   Address: %s\n", networking.GetPublicIPAddress(pubIP))
	} else {
		fmt.Printf("   ⚠ %s is not ready yet\n", *pubIP.Metadata.Id)
	}

	return nil
}

// runSecurityGroupExamples demonstrates all security group operations.
func runSecurityGroupExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Security Group Examples ---")

	// Example 1: Create a basic security group with SSH access
	fmt.Println("\n1. Creating security group with SSH access...")
	sshSG := networking.NewSecurityGroupBuilder("sdk-sg-ssh").
		AllowSSH().
		Build()

	createdSSHSG, err := client.Networking().SecurityGroups().Create(ctx, sshSG)
	if err != nil {
		return fmt.Errorf("failed to create SSH security group: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdSSHSG.Metadata.Id)

	// Example 2: Create a web server security group
	fmt.Println("\n2. Creating web server security group...")
	webSG := networking.NewSecurityGroupBuilder("sdk-sg-web").
		AllowSSH().
		AllowHTTP().
		AllowHTTPS().
		AllowAllEgress().
		Build()

	createdWebSG, err := client.Networking().SecurityGroups().Create(ctx, webSG)
	if err != nil {
		return fmt.Errorf("failed to create web security group: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (rules: %d)\n",
		*createdWebSG.Metadata.Id,
		len(*createdWebSG.Spec.Rules))

	// Example 3: Create a custom security group with specific rules
	fmt.Println("\n3. Creating custom security group with port ranges...")
	customSG := networking.NewSecurityGroupBuilder("sdk-sg-custom").
		AllowIngressRule("allow-postgres", "TCP", 5432, 0, "10.0.0.0/16").
		AllowIngressRule("allow-app-range", "TCP", 8000, 8999, "0.0.0.0/0").
		AllowEgressRule("allow-outbound-dns", "UDP", 53, 0, "0.0.0.0/0").
		AllowAllEgress().
		Build()

	createdCustomSG, err := client.Networking().SecurityGroups().Create(ctx, customSG)
	if err != nil {
		return fmt.Errorf("failed to create custom security group: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (rules: %d)\n",
		*createdCustomSG.Metadata.Id,
		len(*createdCustomSG.Spec.Rules))

	// Example 4: Create a database security group with restricted access
	fmt.Println("\n4. Creating database security group...")
	dbSG := networking.NewSecurityGroupBuilder("sdk-sg-database").
		AllowIngressRule("allow-mysql", "TCP", 3306, 0, "10.0.1.0/24").
		AllowIngressRule("allow-redis", "TCP", 6379, 0, "10.0.1.0/24").
		AllowEgressRule("allow-outbound", "all", 0, 0, "0.0.0.0/0").
		Build()

	createdDBSG, err := client.Networking().SecurityGroups().Create(ctx, dbSG)
	if err != nil {
		return fmt.Errorf("failed to create database security group: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdDBSG.Metadata.Id)

	// Example 5: List all security groups
	fmt.Println("\n5. Listing all security groups...")
	securityGroups, err := client.Networking().SecurityGroups().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list security groups: %w", err)
	}
	fmt.Printf("   Found %d security groups:\n", len(securityGroups.Items))
	for _, sg := range securityGroups.Items {
		ruleCount := 0
		if sg.Spec.Rules != nil {
			ruleCount = len(*sg.Spec.Rules)
		}
		fmt.Printf("   - %s: %d rules\n", *sg.Metadata.Id, ruleCount)
	}

	// Example 6: Get a specific security group
	fmt.Println("\n6. Getting specific security group...")
	sg, err := client.Networking().SecurityGroups().Get(ctx, "sdk-sg-web")
	if err != nil {
		return fmt.Errorf("failed to get security group: %w", err)
	}
	fmt.Printf("   ✓ Security Group: %s\n", *sg.Metadata.Id)
	if sg.Spec.Rules != nil {
		fmt.Printf("     Rules:\n")
		for _, rule := range *sg.Spec.Rules {
			protocol := "all"
			if rule.Protocol != nil {
				protocol = string(*rule.Protocol)
			}
			port := "all"
			if rule.Port != nil {
				port = fmt.Sprintf("%d", *rule.Port)
				if rule.EndPort != nil {
					port = fmt.Sprintf("%d-%d", *rule.Port, *rule.EndPort)
				}
			}
			fmt.Printf("       - %s: %s/%s (%s)\n", *rule.Name, protocol, port, rule.Direction)
		}
	}

	// Example 7: Delete a security group (demonstrates delete operation)
	fmt.Println("\n7. Deleting a security group...")
	err = client.Networking().SecurityGroups().Delete(ctx, "sdk-sg-custom")
	if err != nil {
		return fmt.Errorf("failed to delete security group: %w", err)
	}
	fmt.Println("   ✓ Deleted sdk-sg-custom")
	fmt.Println("   Note: Remaining resources will be cleaned up at the end")

	return nil
}

// runVPCAndSubnetExamples demonstrates VPC and Subnet operations (read-only).
func runVPCAndSubnetExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- VPC and Subnet Examples (Read-Only) ---")

	// Example 1: List all VPCs
	fmt.Println("\n1. Listing all VPCs...")
	vpcs, err := client.Networking().VirtualPrivateClouds().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list VPCs: %w", err)
	}
	fmt.Printf("   Found %d VPCs:\n", len(vpcs.Items))
	for _, vpc := range vpcs.Items {
		fmt.Printf("   - %s\n", *vpc.Metadata.Id)
	}

	// Example 2: Get a specific VPC (if any exist)
	if len(vpcs.Items) > 0 {
		fmt.Println("\n2. Getting specific VPC...")
		vpcName := *vpcs.Items[0].Metadata.Id
		vpc, err := client.Networking().VirtualPrivateClouds().Get(ctx, vpcName)
		if err != nil {
			return fmt.Errorf("failed to get VPC: %w", err)
		}
		fmt.Printf("   ✓ VPC: %s\n", *vpc.Metadata.Id)
		fmt.Printf("     Spec: %+v\n", vpc.Spec)
	}

	// Example 3: List all subnets
	fmt.Println("\n3. Listing all subnets...")
	subnets, err := client.Networking().Subnets().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list subnets: %w", err)
	}
	fmt.Printf("   Found %d subnets:\n", len(subnets.Items))
	for _, subnet := range subnets.Items {
		fmt.Printf("   - %s\n", *subnet.Metadata.Id)
	}

	// Example 4: Get a specific subnet (if any exist)
	if len(subnets.Items) > 0 {
		fmt.Println("\n4. Getting specific subnet...")
		subnetName := *subnets.Items[0].Metadata.Id
		subnet, err := client.Networking().Subnets().Get(ctx, subnetName)
		if err != nil {
			return fmt.Errorf("failed to get subnet: %w", err)
		}
		fmt.Printf("   ✓ Subnet: %s\n", *subnet.Metadata.Id)
		fmt.Printf("     Spec: %+v\n", subnet.Spec)
	}

	return nil
}
