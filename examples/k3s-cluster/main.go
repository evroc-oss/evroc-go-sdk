// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates deploying a k3s Kubernetes cluster across 3 availability zones.
// This example shows:
// 1. Creating infrastructure across multiple zones for high availability
// 2. Deploying k3s server (control plane) in zone A
// 3. Deploying k3s agents (workers) in zones B and C
// 4. Using cloud-init for automated cluster setup
// 5. Retrieving kubeconfig for cluster access
//
// The cluster is configured without local-storage, ready for the Evroc CSI Driver:
// https://github.com/evroc-oss/evroc-csi-driver
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
	"github.com/evroc-oss/evroc-go-sdk/networking"
)

const (
	// Cluster configuration
	clusterName = "sdk-k3s"

	// Node configuration
	serverInstanceType = "c1a.m" // 4 vCPUs, 8GB RAM for control plane
	agentInstanceType  = "c1a.s" // 2 vCPUs, 4GB RAM for workers
	diskSizeGB         = 50

	// Network configuration
	sgName       = "sdk-k3s-sg"
	serverIPName = "sdk-k3s-server-ip"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "destroy":
			fmt.Println("=== Destroying k3s Cluster ===")
			fmt.Println()
			destroyCluster(ctx, client)
			fmt.Println("\n=== Cluster Destroyed ===")
			return
		case "create":
			// Continue to creation below
		default:
			fmt.Println("Usage: go run main.go [create|destroy]")
			fmt.Println("  create  - Deploy k3s cluster (default)")
			fmt.Println("  destroy - Delete k3s cluster and all resources")
			os.Exit(1)
		}
	}

	// SSH key is required for creation
	sshPublicKey := os.Getenv("SSH_PUBLIC_KEY")
	if sshPublicKey == "" {
		log.Fatal("SSH_PUBLIC_KEY environment variable is required")
	}

	fmt.Println("=== Deploying k3s Kubernetes Cluster Across 3 Zones ===")
	fmt.Println()

	// Step 1: Create security group
	fmt.Println("Step 1: Creating security group...")
	sg, err := client.Networking().SecurityGroups().Create(ctx,
		networking.NewSecurityGroupBuilder(sgName).
			AllowSSH().                                                     // SSH access for management
			AllowIngressRule("k3s-api", "TCP", 6443, 0, "0.0.0.0/0").       // Kubernetes API
			AllowIngressRule("k3s-flannel", "UDP", 8472, 0, "10.0.0.0/8").  // Flannel VXLAN
			AllowIngressRule("k3s-metrics", "TCP", 10250, 0, "10.0.0.0/8"). // Kubelet metrics
			AllowAllEgress().                                               // Allow all outbound traffic
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create security group: %v", err)
	}
	fmt.Printf("✓ Security group created: %s\n\n", *sg.Metadata.Id)

	// Step 2: Create public IP for server
	fmt.Println("Step 2: Creating public IP for k3s server...")
	publicIP, err := client.Networking().PublicIPs().Create(ctx,
		networking.NewPublicIPBuilder(serverIPName).Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create public IP: %v", err)
	}

	// Get the IP address (usually ready immediately)
	time.Sleep(2 * time.Second) // Brief wait for IP allocation
	publicIP, err = client.Networking().PublicIPs().Get(ctx, serverIPName)
	if err != nil {
		log.Fatalf("Failed to get public IP: %v", err)
	}

	serverIP := networking.GetPublicIPAddress(publicIP)
	fmt.Printf("✓ Public IP ready: %s\n\n", serverIP)

	// Step 3: Create disks for all nodes
	fmt.Println("Step 3: Creating disks for all nodes (3 zones)...")
	zones := []string{"a", "b", "c"}
	nodeNames := []string{
		fmt.Sprintf("%s-server", clusterName),
		fmt.Sprintf("%s-agent-1", clusterName),
		fmt.Sprintf("%s-agent-2", clusterName),
	}

	for i, zone := range zones {
		diskName := fmt.Sprintf("%s-disk", nodeNames[i])
		_, err := client.Compute().Disks().Create(ctx,
			compute.NewDiskBuilder(diskName).
				WithImage(string(compute.DiskImageUbuntuMinimal2404)).
				WithSizeGB(diskSizeGB).
				WithZone(zone).
				Build(),
		)
		if err != nil {
			log.Fatalf("Failed to create disk %s: %v", diskName, err)
		}
		fmt.Printf("   ✓ Created disk in zone %s: %s\n", zone, diskName)
	}

	// Wait for all disks to be ready
	fmt.Println("\n   Waiting for disks to be ready...")
	for _, nodeName := range nodeNames {
		diskName := fmt.Sprintf("%s-disk", nodeName)
		err = client.Compute().Disks().WaitForReady(ctx, diskName, 5*time.Minute)
		if err != nil {
			log.Fatalf("Disk %s did not become ready: %v", diskName, err)
		}
	}
	fmt.Println("✓ All disks ready\n")

	// Step 4: Generate k3s token (shared secret for agents to join)
	k3sToken := fmt.Sprintf("K%d", time.Now().Unix())

	// Step 5: Create k3s server (control plane) in zone A
	fmt.Println("Step 4: Creating k3s server (control plane) in zone A...")
	serverCloudInit := generateServerCloudInit(serverIP, k3sToken, "a", sshPublicKey)

	_, err = client.Compute().VirtualMachines().Create(ctx,
		compute.NewVirtualMachineBuilder(nodeNames[0]).
			WithBootDisk(fmt.Sprintf("%s-disk", nodeNames[0])).
			WithVMInstanceType(serverInstanceType).
			WithSecurityGroup(sgName).
			WithPublicIP(serverIPName).
			WithCloudInit(serverCloudInit).
			WithZone("a").
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create server VM: %v", err)
	}
	fmt.Printf("✓ Server VM created: %s (zone: a)\n\n", nodeNames[0])

	// Wait for server to be ready
	fmt.Println("   Waiting for server to be ready...")
	err = client.Compute().VirtualMachines().WaitForReady(ctx, nodeNames[0], 5*time.Minute)
	if err != nil {
		log.Fatalf("Server VM did not become ready: %v", err)
	}
	fmt.Println("✓ Server is ready\n")

	// Step 6: Create k3s agents (workers) in zones B and C
	fmt.Println("Step 5: Creating k3s agents (workers) in zones B and C...")

	// Get server's private IP for agents to connect
	// Note: Agents will use the server's public IP initially
	// In production, you'd use internal networking or VPC peering
	fmt.Printf("   Agents will connect to server using public IP: %s\n", serverIP)

	// Create agents in zones B and C
	for i, zone := range []string{"b", "c"} {
		agentIndex := i + 1
		agentName := nodeNames[agentIndex]
		agentCloudInit := generateAgentCloudInit(serverIP, k3sToken, zone, sshPublicKey)

		_, err = client.Compute().VirtualMachines().Create(ctx,
			compute.NewVirtualMachineBuilder(agentName).
				WithBootDisk(fmt.Sprintf("%s-disk", agentName)).
				WithVMInstanceType(agentInstanceType).
				WithSecurityGroup(sgName).
				WithCloudInit(agentCloudInit).
				WithZone(zone).
				Build(),
		)
		if err != nil {
			log.Fatalf("Failed to create agent VM %s: %v", agentName, err)
		}
		fmt.Printf("   ✓ Agent VM created: %s (zone: %s)\n", agentName, zone)
	}

	// Wait for agents to be ready
	fmt.Println("\n   Waiting for agents to be ready...")
	for i := 1; i <= 2; i++ {
		agentName := nodeNames[i]
		err = client.Compute().VirtualMachines().WaitForReady(ctx, agentName, 5*time.Minute)
		if err != nil {
			log.Fatalf("Agent VM %s did not become ready: %v", agentName, err)
		}
	}
	fmt.Println("✓ All agents ready\n")

	// Step 7: Display cluster information
	fmt.Println("=== k3s Cluster Infrastructure Deployed ===")
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: k3s is now installing via cloud-init")
	fmt.Println("    This takes 2-3 minutes. The kubeconfig will not be available immediately.")
	fmt.Println()
	fmt.Println("Cluster Configuration:")
	fmt.Printf("  Name: %s\n", clusterName)
	fmt.Printf("  Server: %s (zone: a, %s vCPUs)\n", nodeNames[0], serverInstanceType)
	fmt.Printf("  Agent 1: %s (zone: b, %s vCPUs)\n", nodeNames[1], agentInstanceType)
	fmt.Printf("  Agent 2: %s (zone: c, %s vCPUs)\n", nodeNames[2], agentInstanceType)
	fmt.Println()
	fmt.Println("Access Information:")
	fmt.Printf("  Server Public IP: %s\n", serverIP)
	fmt.Printf("  SSH: ssh ubuntu@%s\n", serverIP)
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("NEXT STEPS:")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("1. Wait for k3s installation to complete (2-3 minutes)")
	fmt.Println()
	fmt.Println("   Check if k3s is ready:")
	fmt.Printf("   ssh ubuntu@%s 'until sudo test -f /etc/rancher/k3s/k3s.yaml; do echo \"Waiting for k3s...\"; sleep 5; done; echo \"k3s is ready!\"'\n", serverIP)
	fmt.Println()
	fmt.Println("2. Retrieve and configure kubeconfig:")
	fmt.Printf("   ssh ubuntu@%s sudo cat /etc/rancher/k3s/k3s.yaml | \\\n", serverIP)
	fmt.Printf("     sed 's/127.0.0.1/%s/g' | \\\n", serverIP)
	fmt.Println("     sed '/certificate-authority-data:/d' | \\")
	fmt.Println("     sed 's/server: https/insecure-skip-tls-verify: true\\n    server: https/' \\")
	fmt.Printf("     > ~/.kube/%s-config\n", clusterName)
	fmt.Println()
	fmt.Println("3. Set permissions:")
	fmt.Printf("   chmod 600 ~/.kube/%s-config\n", clusterName)
	fmt.Println()
	fmt.Println("4. Verify cluster:")
	fmt.Printf("   export KUBECONFIG=$HOME/.kube/%s-config\n", clusterName)
	fmt.Println("   kubectl get nodes -o wide")
	fmt.Println("   kubectl get nodes --show-labels | grep topology.kubernetes.io/zone")
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
}

// generateServerCloudInit creates cloud-init for k3s server node
// Based on evroc-csi-driver/test/e2e/ansible/playbooks/setup-k3s.yaml
func generateServerCloudInit(publicIP, token, zone, sshKey string) string {
	return fmt.Sprintf(`#cloud-config
package_update: true
package_upgrade: true

users:
  - name: ubuntu
    ssh_authorized_keys:
      - %s
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash

packages:
  - curl
  - util-linux
  - e2fsprogs
  - jq

runcmd:
  # Install k3s server
  - curl -sfL https://get.k3s.io | sh -s - server --disable traefik --disable servicelb --disable local-storage --write-kubeconfig-mode 644 --tls-san %s --node-label topology.kubernetes.io/zone=%s --token %s
`, sshKey, publicIP, zone, token)
}

// generateAgentCloudInit creates cloud-init for k3s agent nodes
func generateAgentCloudInit(serverIP, token, zone, sshKey string) string {
	return fmt.Sprintf(`#cloud-config
package_update: true
package_upgrade: true

users:
  - name: ubuntu
    ssh_authorized_keys:
      - %s
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash

packages:
  - curl
  - util-linux
  - e2fsprogs

runcmd:
  # Install k3s agent
  - curl -sfL https://get.k3s.io | K3S_URL=https://%s:6443 K3S_TOKEN=%s sh -s - agent --node-label topology.kubernetes.io/zone=%s
`, sshKey, serverIP, token, zone)
}

// destroyCluster deletes all k3s cluster resources
func destroyCluster(ctx context.Context, client *evroc.Client) {
	nodeNames := []string{
		fmt.Sprintf("%s-server", clusterName),
		fmt.Sprintf("%s-agent-1", clusterName),
		fmt.Sprintf("%s-agent-2", clusterName),
	}

	// Step 1: Delete VMs
	fmt.Println("Step 1: Deleting VMs...")
	for _, nodeName := range nodeNames {
		err := client.Compute().VirtualMachines().Delete(ctx, nodeName)
		if err != nil {
			log.Printf("   Warning: Failed to delete VM %s: %v", nodeName, err)
		} else {
			fmt.Printf("   ✓ Deleted VM %s\n", nodeName)
		}
	}

	// Step 2: Wait for VMs to be deleted
	fmt.Println("\nStep 2: Waiting for VMs to be deleted...")
	for _, nodeName := range nodeNames {
		err := client.Compute().VirtualMachines().WaitForDeleted(ctx, nodeName, 5*time.Minute)
		if err != nil {
			log.Printf("   Warning: Timeout waiting for %s deletion: %v", nodeName, err)
		} else {
			fmt.Printf("   ✓ VM %s deleted\n", nodeName)
		}
	}

	// Step 3: Delete disks
	fmt.Println("\nStep 3: Deleting disks...")
	for _, nodeName := range nodeNames {
		diskName := fmt.Sprintf("%s-disk", nodeName)
		err := client.Compute().Disks().Delete(ctx, diskName)
		if err != nil {
			log.Printf("   Warning: Failed to delete disk %s: %v", diskName, err)
		} else {
			fmt.Printf("   ✓ Deleted disk %s\n", diskName)
		}
	}

	// Step 4: Delete public IP
	fmt.Println("\nStep 4: Deleting public IP...")
	err := client.Networking().PublicIPs().Delete(ctx, serverIPName)
	if err != nil {
		log.Printf("   Warning: Failed to delete public IP: %v", err)
	} else {
		fmt.Printf("   ✓ Deleted public IP %s\n", serverIPName)
	}

	// Step 5: Delete security group
	fmt.Println("\nStep 5: Deleting security group...")
	err = client.Networking().SecurityGroups().Delete(ctx, sgName)
	if err != nil {
		log.Printf("   Warning: Failed to delete security group: %v", err)
	} else {
		fmt.Printf("   ✓ Deleted security group %s\n", sgName)
	}
}
