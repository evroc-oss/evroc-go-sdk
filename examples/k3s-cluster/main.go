// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// k3s Kubernetes Cluster Across 3 Availability Zones
//
// This example deploys a production-ready k3s Kubernetes cluster across 3 availability zones for high availability.
//
// # What This Example Does
//
// 1. Creates Infrastructure:
//   - Security group with k3s networking rules (SSH, API 6443, Flannel 8472, Kubelet 10250)
//   - Public IP for the control plane
//   - 3 boot disks across zones A, B, and C
//
// 2. Deploys k3s Cluster:
//   - 1 server (control plane) in zone A with public IP (c1a.m: 4vCPU/8GB)
//   - 2 agents (workers) in zones B and C (c1a.s: 2vCPU/4GB each)
//   - Automatic cluster join using k3s token
//   - Zone topology labels on all nodes
//
// 3. Configuration:
//   - Disables default components (Traefik, ServiceLB, local-storage)
//   - Enables kubeconfig access with proper TLS SANs
//   - Sets up kubelet and Flannel VXLAN networking
//
// # Prerequisites
//
//	export EVROC_PROJECT="your-project-uuid"
//	export EVROC_REGION="se-sto"
//	export SSH_PUBLIC_KEY="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
//
// # Running
//
//	# Create the cluster
//	go run main.go
//	go run main.go create
//
//	# Destroy the cluster
//	go run main.go destroy
//
// # Timeline
//
//   - VM creation: ~2 minutes (the example waits for this)
//   - Cloud-init execution: 2-3 minutes (k3s installation - happens in background)
//   - Cluster ready: Total ~5 minutes from start
//
// The example completes after VMs are ready, but k3s installation continues via cloud-init.
// You must wait 2-3 additional minutes before accessing the cluster.
//
// # Accessing the Cluster
//
// After the example completes, wait for k3s installation:
//
//	# Wait for k3s to be ready (replace IP with your server IP)
//	ssh ubuntu@<server-public-ip> 'until sudo test -f /etc/rancher/k3s/k3s.yaml; do echo "Waiting for k3s..."; sleep 5; done; echo "k3s is ready!"'
//
//	# Download kubeconfig
//	ssh ubuntu@<server-public-ip> sudo cat /etc/rancher/k3s/k3s.yaml | \
//	  sed 's/127.0.0.1/<server-public-ip>/g' | \
//	  sed '/certificate-authority-data:/d' | \
//	  sed 's/server: https/insecure-skip-tls-verify: true\n    server: https/' \
//	  > ~/.kube/sdk-k3s-config
//
//	chmod 600 ~/.kube/sdk-k3s-config
//	export KUBECONFIG=$HOME/.kube/sdk-k3s-config
//
//	# Verify cluster
//	kubectl get nodes -o wide
//	kubectl get nodes --show-labels | grep topology.kubernetes.io/zone
//
// Note: We use insecure-skip-tls-verify because the k3s certificate is only valid for 127.0.0.1.
// For production, configure proper TLS certificates.
//
// # Architecture
//
//	┌─────────────────────────────────────────────────────────┐
//	│                    Public Internet                      │
//	└────────────────────────┬────────────────────────────────┘
//	                         │
//	                         ▼
//	              ┌──────────────────────┐
//	              │   Security Group     │
//	              │  (SSH, API, Flannel) │
//	              └──────────────────────┘
//	                         │
//	        ┌────────────────┼────────────────┐
//	        │                │                │
//	┌───────▼────────┐ ┌────▼──────┐ ┌───────▼────────┐
//	│  Zone A        │ │  Zone B   │ │  Zone C        │
//	│                │ │           │ │                │
//	│ sdk-k3s-server │ │sdk-k3s-   │ │ sdk-k3s-       │
//	│ (Control Plane)│ │agent-1    │ │ agent-2        │
//	│                │ │(Worker)   │ │ (Worker)       │
//	│ c1a.m (4v/8GB) │ │c1a.s      │ │ c1a.s          │
//	│ Public IP      │ │(2v/4GB)   │ │ (2v/4GB)       │
//	│                │ │Private IP │ │ Private IP     │
//	└────────────────┘ └───────────┘ └────────────────┘
//
// # Next Steps
//
//   - Install CSI Driver: Deploy the Evroc CSI Driver for persistent volumes
//     https://github.com/evroc-oss/evroc-csi-driver
//   - Setup Ingress: Install nginx-ingress or Traefik
//   - Add Monitoring: Deploy Prometheus and Grafana
//   - Scale Cluster: Add more agents in each zone
//
// # Troubleshooting
//
// If kubeconfig retrieval fails, k3s hasn't finished installing. Check progress:
//
//	ssh ubuntu@<server-ip> 'tail -f /var/log/k3s-install.log'
//	ssh ubuntu@<server-ip> 'systemctl status k3s'
//	ssh ubuntu@<server-ip> 'tail -f /var/log/cloud-init-output.log'
//
// # References
//
//   - k3s Documentation: https://docs.k3s.io/
//   - Kubernetes Topology Awareness: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
//   - Evroc CSI Driver: https://github.com/evroc-oss/evroc-csi-driver
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
	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
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
	fmt.Printf("✓ Security group created: %s\n\n", sg.Metadata.Id)

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

	// Create and save disk references
	disks := make([]*computetypes.Disk, len(nodeNames))
	for i, zone := range zones {
		diskName := fmt.Sprintf("%s-disk", nodeNames[i])
		disk, err := client.Compute().Disks().Create(ctx,
			compute.NewDiskBuilder(diskName).
				WithImage(string(compute.DiskImageUbuntuMinimal2404)).
				WithSizeGB(diskSizeGB).
				WithZone(zone).
				Build(),
		)
		if err != nil {
			log.Fatalf("Failed to create disk %s: %v", diskName, err)
		}
		disks[i] = disk
		fmt.Printf("   ✓ Created disk in zone %s: %s\n", zone, diskName)
	}

	// Wait for all disks to be ready
	fmt.Println("\n   Waiting for disks to be ready...")
	for _, nodeName := range nodeNames {
		diskName := fmt.Sprintf("%s-disk", nodeName)
		_, err = client.Compute().Disks().WaitForReady(ctx, diskName, 5*time.Minute)
		if err != nil {
			log.Fatalf("Disk %s did not become ready: %v", diskName, err)
		}
	}
	fmt.Println("✓ All disks ready")

	// Step 4: Generate k3s token (shared secret for agents to join)
	k3sToken := fmt.Sprintf("K%d", time.Now().Unix())

	// Step 5: Create k3s server (control plane) in zone A
	fmt.Println("Step 4: Creating k3s server (control plane) in zone A...")
	// Note: We'll use the public IP in TLS SAN, but agents will connect via private IP
	serverCloudInit := generateServerCloudInit(serverIP, k3sToken, "a", sshPublicKey)

	_, err = client.Compute().VirtualMachines().Create(ctx,
		compute.NewVirtualMachineBuilder(nodeNames[0]).
			WithBootDisk(disks[0].Ref()).
			WithVMInstanceType(serverInstanceType).
			WithSecurityGroup(sg.Ref()).
			WithPublicIP(publicIP.Ref()).
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
	_, err = client.Compute().VirtualMachines().WaitForReady(ctx, nodeNames[0], 5*time.Minute)
	if err != nil {
		log.Fatalf("Server VM did not become ready: %v", err)
	}
	fmt.Println("✓ Server is ready")

	// Step 6: Get server's private IP for agent connections
	fmt.Println("Step 6: Getting server's private IP for agent connections...")
	serverVM, err := client.Compute().VirtualMachines().Get(ctx, nodeNames[0])
	if err != nil {
		log.Fatalf("Failed to get server VM: %v", err)
	}

	var serverPrivateIP string
	if serverVM.Status.Networking != nil && serverVM.Status.Networking.PrivateIPv4Address != nil {
		serverPrivateIP = *serverVM.Status.Networking.PrivateIPv4Address
	} else {
		log.Fatal("Server VM does not have a private IP address")
	}
	fmt.Printf("✓ Server private IP: %s\n\n", serverPrivateIP)

	// Step 7: Create k3s agents (workers) in zones B and C
	fmt.Println("Step 7: Creating k3s agents (workers) in zones B and C...")
	fmt.Printf("   Agents will connect to server using private IP: %s\n", serverPrivateIP)

	// Create agents in zones B and C
	for i, zone := range []string{"b", "c"} {
		agentIndex := i + 1
		agentName := nodeNames[agentIndex]
		agentCloudInit := generateAgentCloudInit(serverPrivateIP, k3sToken, zone, sshPublicKey)

		_, err = client.Compute().VirtualMachines().Create(ctx,
			compute.NewVirtualMachineBuilder(agentName).
				WithBootDisk(disks[agentIndex].Ref()).
				WithVMInstanceType(agentInstanceType).
				WithSecurityGroup(sg.Ref()).
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
		_, err = client.Compute().VirtualMachines().WaitForReady(ctx, agentName, 5*time.Minute)
		if err != nil {
			log.Fatalf("Agent VM %s did not become ready: %v", agentName, err)
		}
	}
	fmt.Println("✓ All agents ready")

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
  - curl -sfL https://get.k3s.io | K3S_URL=https://%s:6443 K3S_TOKEN=%s sh -s - agent --node-label topology.kubernetes.io/zone=%s --node-label node-role.kubernetes.io/worker=true
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
