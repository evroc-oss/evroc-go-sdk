// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates creating a complete web server setup.
// This example shows the complete workflow:
// 1. Create security group with SSH, HTTP, HTTPS, and egress
// 2. Create public IP for internet access
// 3. Create boot disk with Ubuntu image
// 4. Create VM with all components attached
// 5. Wait for VM to be ready
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/networking"
)

//go:embed index.html
var indexHTML string

const (
	securityGroupName = "sdk-web-sg"
	publicIPName      = "sdk-web-ip"
	diskName          = "sdk-web-disk"
	vmName            = "sdk-web-vm"
	diskSizeGB        = 50
	vmInstanceType    = "c1a.m" // 4 vCPUs, 8GB RAM
)

func main() {
	ctx := context.Background()

	// Create client from environment variables
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Creating Web Server Infrastructure ===")

	// Step 1: Create security group
	fmt.Println("Step 1: Creating security group with SSH, HTTP, HTTPS, and egress rules...")
	sg, err := client.Networking().SecurityGroups().Create(ctx,
		networking.NewSecurityGroupBuilder(securityGroupName).
			AllowSSH().       // Port 22 for management
			AllowHTTP().      // Port 80 for web traffic
			AllowHTTPS().     // Port 443 for secure web traffic
			AllowAllEgress(). // Allow all outbound traffic for package updates
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create security group: %v", err)
	}
	fmt.Printf("✓ Security group created: %s\n\n", *sg.Metadata.Id)

	// Step 2: Create public IP
	fmt.Println("Step 2: Creating public IP for internet access...")
	publicIP, err := client.Networking().PublicIPs().Create(ctx,
		networking.NewPublicIPBuilder(publicIPName).Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create public IP: %v", err)
	}
	fmt.Printf("✓ Public IP created: %s\n\n", *publicIP.Metadata.Id)

	// Public IPs are usually ready immediately, but check status
	var ipAddress string
	if networking.IsPublicIPReady(publicIP) {
		ipAddress = networking.GetPublicIPAddress(publicIP)
		fmt.Printf("✓ Public IP ready: %s\n\n", ipAddress)
	} else {
		fmt.Println("   Public IP is being allocated...")
		// In production, you'd implement polling here
		time.Sleep(5 * time.Second)
		publicIP, err = client.Networking().PublicIPs().Get(ctx, publicIPName)
		if err != nil {
			log.Fatalf("Failed to get public IP: %v", err)
		}
		ipAddress = networking.GetPublicIPAddress(publicIP)
		fmt.Printf("✓ Public IP ready: %s\n\n", ipAddress)
	}

	// Step 3: Create boot disk
	fmt.Printf("Step 3: Creating %dGB boot disk with Ubuntu 24.04...\n", diskSizeGB)
	disk, err := client.Compute().Disks().Create(ctx,
		compute.NewDiskBuilder(diskName).
			WithImage(string(compute.DiskImageUbuntuMinimal2404)).
			WithSizeGB(diskSizeGB).
			WithZone("a"). // Zone is required
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create disk: %v", err)
	}
	fmt.Printf("✓ Disk created: %s\n\n", *disk.Metadata.Id)

	// Wait for disk to be ready
	fmt.Println("   Waiting for disk to be ready...")
	err = client.Compute().Disks().WaitForReady(ctx, diskName, 5*time.Minute)
	if err != nil {
		log.Fatalf("Disk did not become ready: %v", err)
	}
	fmt.Println("✓ Disk is ready")

	// Step 4: Create VM with nginx cloud-init
	fmt.Printf("Step 4: Creating VM (%s instance) with nginx setup...\n", vmInstanceType)

	// IMPORTANT: SSH key must be provided or VM will be inaccessible
	sshPublicKey := os.Getenv("SSH_PUBLIC_KEY")
	if sshPublicKey == "" {
		// Use default SSH key if not provided
		sshPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFeENOwB0QwUEicJGrFxt44yiShgBWzANhpE/5gNw041"
		fmt.Println("   Using default SSH key (set SSH_PUBLIC_KEY to override)")
	}

	// Prepare HTML content with instance details
	htmlContent := strings.ReplaceAll(indexHTML, "{{.VMName}}", vmName)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.InstanceType}}", vmInstanceType)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.IPAddress}}", ipAddress)

	// Indent HTML for cloud-init YAML format (6 spaces)
	indentedHTML := ""
	for _, line := range strings.Split(htmlContent, "\n") {
		indentedHTML += "      " + line + "\n"
	}

	// Cloud-init script to set up nginx web server
	cloudInit := fmt.Sprintf(`#cloud-config
package_update: true
package_upgrade: true

packages:
  - nginx
  - curl

write_files:
  - path: /var/www/html/index.html
    owner: www-data:www-data
    permissions: '0644'
    content: |
%s
runcmd:
  - systemctl enable nginx
  - systemctl start nginx
  - systemctl status nginx
`, indentedHTML)

	vm, err := client.Compute().VirtualMachines().Create(ctx,
		compute.NewVirtualMachineBuilder(vmName).
			WithBootDisk(diskName).
			WithVMInstanceType(vmInstanceType).
			WithSecurityGroup(securityGroupName). // Enable network traffic
			WithPublicIP(publicIPName).           // Attach public IP (immutable!)
			WithSSHKey(sshPublicKey).             // SSH authentication (immutable!)
			WithCloudInit(cloudInit).             // Install and configure nginx
			WithZone("a").                        // Zone is required
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}
	fmt.Printf("✓ VM created: %s\n\n", *vm.Metadata.Id)

	// Step 5: Wait for VM to be ready
	fmt.Println("Step 5: Waiting for VM to be ready (this may take 1-2 minutes)...")
	err = client.Compute().VirtualMachines().WaitForReady(ctx, vmName, 5*time.Minute)
	if err != nil {
		log.Fatalf("VM did not become ready: %v", err)
	}
	fmt.Println("✓ VM is ready")

	// Print summary
	fmt.Println("=== Web Server Successfully Created ===")
	fmt.Printf("\nNginx web server is now running!\n")
	fmt.Printf("  View in browser: http://%s\n", ipAddress)
	fmt.Printf("  SSH access:      ssh ubuntu@%s\n\n", ipAddress)
	fmt.Println("Resources created:")
	fmt.Printf("  - Security Group: %s (SSH, HTTP, HTTPS, egress)\n", securityGroupName)
	fmt.Printf("  - Public IP: %s (%s)\n", publicIPName, ipAddress)
	fmt.Printf("  - Disk: %s (%dGB, Ubuntu 24.04)\n", diskName, diskSizeGB)
	fmt.Printf("  - VM: %s (%s, with nginx)\n\n", vmName, vmInstanceType)
	fmt.Println("Note: Cloud-init is configuring nginx. It may take 1-2 minutes for the web page to be available.")

	/*
		// Cleanup - delete resources in reverse order
		fmt.Println("=== Cleaning Up Resources ===\n")

		// Step 6: Delete VM
		fmt.Printf("Step 6: Deleting VM '%s'...\n", vmName)
		err = client.Compute().VirtualMachines().Delete(ctx, vmName)
		if err != nil {
			log.Printf("Failed to delete VM: %v", err)
		} else {
			fmt.Println("✓ VM deleted")
		}

		// Wait for VM to be fully deleted before deleting disk
		fmt.Println("   Waiting for VM to be fully deleted...")
		err = client.Compute().VirtualMachines().WaitForDeleted(ctx, vmName, 5*time.Minute)
		if err != nil {
			log.Printf("Warning: VM deletion wait failed: %v", err)
		} else {
			fmt.Println("✓ VM fully deleted\n")
		}

		// Step 7: Delete disk
		fmt.Printf("Step 7: Deleting disk '%s'...\n", diskName)
		err = client.Compute().Disks().Delete(ctx, diskName)
		if err != nil {
			log.Printf("Failed to delete disk: %v", err)
		} else {
			fmt.Println("✓ Disk deleted\n")
		}

		// Step 8: Delete public IP
		fmt.Printf("Step 8: Deleting public IP '%s'...\n", publicIPName)
		err = client.Networking().PublicIPs().Delete(ctx, publicIPName)
		if err != nil {
			log.Printf("Failed to delete public IP: %v", err)
		} else {
			fmt.Println("✓ Public IP deleted\n")
		}

		// Step 9: Delete security group
		fmt.Printf("Step 9: Deleting security group '%s'...\n", securityGroupName)
		err = client.Networking().SecurityGroups().Delete(ctx, securityGroupName)
		if err != nil {
			log.Printf("Failed to delete security group: %v", err)
		} else {
			fmt.Println("✓ Security group deleted\n")
		}

		fmt.Println("=== Cleanup Complete ===")
	*/
}
