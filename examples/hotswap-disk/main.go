// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package main demonstrates hot-attaching a disk to a running VM without stopping it.
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
)

const (
	// Resource names
	vmName         = "sdk-hotswap-vm"
	bootDiskName   = "hotswap-boot-disk"
	dataDiskName   = "hotswap-data-disk"
	attachmentName = "hotswap-attachment"

	// Configuration
	diskImage  = string(compute.DiskImageUbuntu2404)
	bootSizeGB = 50
	dataSizeGB = 100
	vmSize     = string(compute.VMSizeA1aXS)
	zone       = "a"

	// Timeouts
	diskReadyTimeout       = 2 * time.Minute
	vmReadyTimeout         = 5 * time.Minute
	attachmentReadyTimeout = 2 * time.Minute
	deleteTimeout          = 3 * time.Minute
)

func main() {
	// Create context that cancels on SIGINT/SIGTERM (Ctrl-C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create client
	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Hot-Attach Disk to Running VM Example ===")
	fmt.Println()
	fmt.Println("This example demonstrates:")
	fmt.Println("  1. Creating a VM with a boot disk")
	fmt.Println("  2. Creating a separate data disk")
	fmt.Println("  3. Hot-attaching the data disk to the running VM (no downtime!)")
	fmt.Println("  4. Detaching the disk")
	fmt.Println()

	// Cleanup any existing resources
	fmt.Println("0. Cleaning up any existing resources...")
	cleanupResources(ctx, client)

	// Step 1: Create boot disk
	fmt.Println("\n1. Creating boot disk...")
	bootDiskReq := compute.NewDiskBuilder(bootDiskName).
		WithImage(diskImage).
		WithSizeGB(bootSizeGB).
		WithZone(zone).
		Build()

	bootDisk, err := client.Compute().Disks().Create(ctx, bootDiskReq)
	if err != nil {
		log.Fatalf("Failed to create boot disk: %v", err)
	}
	fmt.Printf("✓ Created boot disk: %s (%dGB)\n", bootDisk.Metadata.Id, bootSizeGB)

	// Wait for boot disk to be ready
	fmt.Println("  Waiting for boot disk to be ready...")
	if _, err := client.Compute().Disks().WaitForReady(ctx, bootDiskName, diskReadyTimeout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✓ Boot disk is ready")

	// Step 2: Create VM with boot disk
	fmt.Println("\n2. Creating VM...")
	vmReq := compute.NewVirtualMachineBuilder(vmName).
		WithSize(vmSize).
		WithBootDisk(bootDisk.Ref()). // Use the disk we just created
		WithZone(zone).
		Build()

	vm, err := client.Compute().VirtualMachines().Create(ctx, vmReq)
	if err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}
	fmt.Printf("✓ Created VM: %s\n", vm.Metadata.Id)

	// Wait for VM to be ready and running
	fmt.Println("  Waiting for VM to be ready and running...")
	if _, err := client.Compute().VirtualMachines().WaitForReady(ctx, vmName, vmReadyTimeout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✓ VM is ready and running")

	// Step 3: Create data disk (while VM is running)
	fmt.Println("\n3. Creating data disk (while VM is running)...")
	dataDiskReq := compute.NewDiskBuilder(dataDiskName).
		WithSizeGB(dataSizeGB).
		WithZone(zone).
		Build()

	dataDisk, err := client.Compute().Disks().Create(ctx, dataDiskReq)
	if err != nil {
		log.Fatalf("Failed to create data disk: %v", err)
	}
	fmt.Printf("✓ Created data disk: %s (%dGB)\n", dataDisk.Metadata.Id, dataSizeGB)

	// Wait for data disk to be ready
	fmt.Println("  Waiting for data disk to be ready...")
	if _, err := client.Compute().Disks().WaitForReady(ctx, dataDiskName, diskReadyTimeout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✓ Data disk is ready")

	// Step 4: Hot-attach the data disk to the running VM
	fmt.Println("\n4. Hot-attaching data disk to running VM (no downtime!)...")
	attachmentReq := compute.NewHotswapDiskAttachmentBuilder(
		attachmentName,
		vm.Ref(),       // Use the VM we just created
		dataDisk.Ref(), // Use the disk we just created
	).Build()

	attachment, err := client.Compute().HotswapDiskAttachments().Create(ctx, attachmentReq)
	if err != nil {
		log.Fatalf("Failed to create hotswap attachment: %v", err)
	}
	fmt.Printf("✓ Created hotswap attachment: %s\n", attachment.Metadata.Id)

	// Wait for attachment to be ready
	fmt.Println("  Waiting for attachment to be ready...")
	if _, err := client.Compute().HotswapDiskAttachments().WaitForReady(ctx, attachmentName, attachmentReadyTimeout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("  ✓ Disk successfully attached to running VM!")

	// Step 5: Verify the attachment
	fmt.Println("\n5. Verifying attachment...")
	attachment, err = client.Compute().HotswapDiskAttachments().Get(ctx, attachmentName)
	if err != nil {
		log.Printf("Failed to get attachment: %v", err)
	} else {
		fmt.Printf("✓ Attachment verified:\n")
		fmt.Printf("  VM: %s\n", attachment.Spec.VirtualMachineRef)
		fmt.Printf("  Disk: %s\n", attachment.Spec.DiskRef)
		fmt.Printf("  Status: %+v\n", attachment.Status)
	}

	// Step 6: List all attachments
	fmt.Println("\n6. Listing all hotswap disk attachments...")
	attachments, err := client.Compute().HotswapDiskAttachments().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d hotswap disk attachment(s):\n", len(attachments.Items))
	for _, a := range attachments.Items {
		fmt.Printf("  - %s: %s → %s\n",
			a.Metadata.Id,
			a.Spec.DiskRef,
			a.Spec.VirtualMachineRef)
	}

	// Step 7: Detach the disk (also hot operation)
	fmt.Println("\n7. Hot-detaching disk from running VM...")
	err = client.Compute().HotswapDiskAttachments().Delete(ctx, attachmentName)
	if err != nil {
		log.Printf("Failed to delete attachment: %v", err)
	} else {
		fmt.Println("✓ Disk detached from running VM")
	}

	fmt.Println("\n✓ Hot-attach/detach demonstration complete!")
	fmt.Println("The VM remained running throughout the entire process - no downtime required.")
}

// cleanupResources checks for existing resources and deletes them if found.
func cleanupResources(ctx context.Context, client *evroc.Client) {
	// Delete attachment first
	fmt.Println("  Checking for existing attachment...")
	_, err := client.Compute().HotswapDiskAttachments().Get(ctx, attachmentName)
	if err == nil {
		fmt.Println("  Found existing attachment, deleting...")
		err = client.Compute().HotswapDiskAttachments().Delete(ctx, attachmentName)
		if err != nil {
			log.Printf("  Warning: Failed to delete attachment: %v", err)
		} else {
			time.Sleep(2 * time.Second) // Give it time to detach
			fmt.Println("  ✓ Attachment deleted")
		}
	} else {
		fmt.Println("  No existing attachment found")
	}

	// Delete VM
	fmt.Println("  Checking for existing VM...")
	_, err = client.Compute().VirtualMachines().Get(ctx, vmName)
	if err == nil {
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
	} else {
		fmt.Println("  No existing VM found")
	}

	// Delete boot disk
	fmt.Println("  Checking for existing boot disk...")
	_, err = client.Compute().Disks().Get(ctx, bootDiskName)
	if err == nil {
		fmt.Println("  Found existing boot disk, deleting...")
		err = client.Compute().Disks().Delete(ctx, bootDiskName)
		if err != nil {
			log.Printf("  Warning: Failed to delete boot disk: %v", err)
		} else {
			fmt.Println("  Waiting for boot disk to be deleted...")
			if err := client.Compute().Disks().WaitForDeleted(ctx, bootDiskName, deleteTimeout); err != nil {
				log.Printf("  Warning: Boot disk deletion timeout: %v", err)
			} else {
				fmt.Println("  ✓ Boot disk deleted")
			}
		}
	} else {
		fmt.Println("  No existing boot disk found")
	}

	// Delete data disk
	fmt.Println("  Checking for existing data disk...")
	_, err = client.Compute().Disks().Get(ctx, dataDiskName)
	if err == nil {
		fmt.Println("  Found existing data disk, deleting...")
		err = client.Compute().Disks().Delete(ctx, dataDiskName)
		if err != nil {
			log.Printf("  Warning: Failed to delete data disk: %v", err)
		} else {
			fmt.Println("  Waiting for data disk to be deleted...")
			if err := client.Compute().Disks().WaitForDeleted(ctx, dataDiskName, deleteTimeout); err != nil {
				log.Printf("  Warning: Data disk deletion timeout: %v", err)
			} else {
				fmt.Println("  ✓ Data disk deleted")
			}
		}
	} else {
		fmt.Println("  No existing data disk found")
	}

	fmt.Println()
}
