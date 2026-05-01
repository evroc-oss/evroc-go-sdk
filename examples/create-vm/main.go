// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package main demonstrates creating a VM with disk, public IP, and SSH access.
package main

import (
	"context"
	"errors"
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
	// Resource names.
	vmName       = "sdk-vm"
	diskName     = "sdk-disk"
	publicIPName = "sdk-public-ip"
	sgName       = "default-allow-ssh"

	// SSH key.
	sshPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFeENOwB0QwUEicJGrFxt44yiShgBWzANhpE/5gNw041 user@example.com"

	// Configuration.
	diskImage  = string(compute.DiskImageUbuntu2404)
	diskSizeGB = 50
	vmSize     = string(compute.VMSizeA1aXS)
	zone       = "a"

	// Timeouts.
	diskReadyTimeout = 2 * time.Minute
	vmReadyTimeout   = 5 * time.Minute
	deleteTimeout    = 3 * time.Minute
)

func main() {
	// Create context that cancels on SIGINT/SIGTERM (Ctrl-C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create client - config has project/region
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Creating VM with New Disk Example ===")

	// Step 0: Clean up any existing resources
	fmt.Println("0. Cleaning up any existing resources...")
	cleanupResources(ctx, client)

	// Example 1: Create a public IP
	fmt.Println("\n1. Creating public IP...")
	pubIPReq := networking.NewPublicIPBuilder(publicIPName).Build()

	pubIP, err := client.Networking().PublicIPs().Create(ctx, pubIPReq)
	if err != nil {
		log.Fatalf("Failed to create public IP: %v", err)
	}
	fmt.Printf("✓ Created public IP: %s\n", pubIP.Metadata.Id)

	// Example 2: Create a disk using the builder pattern
	fmt.Println("\n2. Creating disk with builder pattern...")
	diskReq := compute.NewDiskBuilder(diskName).
		WithImage(diskImage).
		WithSizeGB(diskSizeGB).
		WithZone(zone).
		Build()

	disk, err := client.Compute().Disks().Create(ctx, diskReq)
	if err != nil {
		log.Fatalf("Failed to create disk: %v", err)
	}
	fmt.Printf("✓ Created disk: %s\n", disk.Metadata.Id)

	// Wait for disk to be ready
	fmt.Println("  Waiting for disk to be ready...")
	if _, err := client.Compute().Disks().WaitForReady(ctx, diskName, diskReadyTimeout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✓ Disk is ready")

	// Example 3: Create a VM with the disk using builder pattern
	// We use .Ref() to get type-safe references to created resources.
	fmt.Println("\n3. Creating VM with public IP and SSH access...")
	vmReq := compute.NewVirtualMachineBuilder(vmName).
		WithSize(vmSize).
		WithBootDisk(disk.Ref()).                                        // Use the disk we just created
		WithPublicIP(pubIP.Ref()).                                       // Use the public IP we just created
		WithSecurityGroup(client.Networking().SecurityGroupRef(sgName)). // Pre-existing SG - construct ref from name
		WithSSHKey(sshPublicKey).
		WithZone(zone).
		Build()

	vm, err := client.Compute().VirtualMachines().Create(ctx, vmReq)
	if err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}
	fmt.Printf("✓ Created VM: %s\n", vm.Metadata.Id)

	// Wait for VM to be ready and running
	fmt.Println("  Waiting for VM to be ready...")
	if _, err := client.Compute().VirtualMachines().WaitForReady(ctx, vmName, vmReadyTimeout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✓ VM is ready and running")

	// Get the public IP address
	pubIP, err = client.Networking().PublicIPs().Get(ctx, publicIPName)
	if err == nil && pubIP.Status.PublicIPv4Address != nil {
		fmt.Printf("  Public IP: %s\n", *pubIP.Status.PublicIPv4Address)
		fmt.Printf("  SSH access: ssh ubuntu@%s\n", *pubIP.Status.PublicIPv4Address)
	}

	// Example 4: List all VMs
	fmt.Println("\n4. Listing all VMs...")
	vms, err := client.Compute().VirtualMachines().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d VMs:\n", len(vms.Items))
	for _, vm := range vms.Items {
		fmt.Printf("  - %s (size: %s)\n",
			vm.Metadata.Id,
			vm.Spec.ComputeProfileRef)
	}

	// Example 5: Get specific VM
	fmt.Println("\n5. Getting specific VM...")
	vm, err = client.Compute().VirtualMachines().Get(ctx, vmName)
	if err != nil {
		log.Printf("Failed to get VM: %v", err)
	} else {
		fmt.Printf("✓ Got VM: %s\n", vm.Metadata.Id)
		fmt.Printf("  Status: %+v\n", vm.Status)
	}

	fmt.Println("\n✓ All resources created and ready!")
}

// cleanupResources checks for existing resources and deletes them if found.
func cleanupResources(ctx context.Context, client *evroc.Client) {
	// Check and delete VM first (it depends on disk and public IP)
	fmt.Println("  Checking for existing VM...")
	_, err := client.Compute().VirtualMachines().Get(ctx, vmName)
	if err != nil {
		if errors.Is(err, evroc.ErrNotFound) {
			fmt.Println("  No existing VM found")
		} else {
			log.Printf("  Warning: Failed to check for VM: %v", err)
		}
	} else {
		fmt.Println("  Found existing VM, deleting...")
		err = client.Compute().VirtualMachines().Delete(ctx, vmName)
		if err != nil {
			log.Printf("  Warning: Failed to delete VM: %v", err)
		} else {
			fmt.Println("  Waiting for VM to be deleted...")
			if err := client.Compute().VirtualMachines().WaitForDeleted(ctx, vmName, deleteTimeout); err != nil {
				log.Printf("  Warning: VM deletion timeout: %v", err)
			} else {
				fmt.Println("  ✓ VM deleted")
			}
		}
	}

	// Check and delete disk
	fmt.Println("  Checking for existing disk...")
	_, err = client.Compute().Disks().Get(ctx, diskName)
	if err != nil {
		if errors.Is(err, evroc.ErrNotFound) {
			fmt.Println("  No existing disk found")
		} else {
			log.Printf("  Warning: Failed to check for disk: %v", err)
		}
	} else {
		fmt.Println("  Found existing disk, deleting...")
		err = client.Compute().Disks().Delete(ctx, diskName)
		if err != nil {
			log.Printf("  Warning: Failed to delete disk: %v", err)
		} else {
			fmt.Println("  Waiting for disk to be deleted...")
			if err := client.Compute().Disks().WaitForDeleted(ctx, diskName, deleteTimeout); err != nil {
				log.Printf("  Warning: Disk deletion timeout: %v", err)
			} else {
				fmt.Println("  ✓ Disk deleted")
			}
		}
	}

	// Check and delete public IP (VM must be deleted first)
	fmt.Println("  Checking for existing public IP...")
	_, err = client.Networking().PublicIPs().Get(ctx, publicIPName)
	if err != nil {
		if errors.Is(err, evroc.ErrNotFound) {
			fmt.Println("  No existing public IP found")
		} else {
			log.Printf("  Warning: Failed to check for public IP: %v", err)
		}
	} else {
		fmt.Println("  Found existing public IP, deleting...")
		err = client.Networking().PublicIPs().Delete(ctx, publicIPName)
		if err != nil {
			log.Printf("  Warning: Failed to delete public IP: %v", err)
		} else {
			fmt.Println("  ✓ Public IP deleted")
		}
	}

	fmt.Println()
}
