// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package main demonstrates comprehensive compute API usage.
// Covers: Disks, VirtualMachines, PlacementGroups, HotswapDiskAttachments
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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Comprehensive Compute API Examples ===")
	fmt.Println()

	// Run all examples
	if err := runDiskExamples(ctx, client); err != nil {
		log.Printf("Disk examples failed: %v", err)
	}

	if err := runPlacementGroupExamples(ctx, client); err != nil {
		log.Printf("Placement group examples failed: %v", err)
	}

	if err := runVMExamples(ctx, client); err != nil {
		log.Printf("VM examples failed: %v", err)
	}

	if err := runHotswapExamples(ctx, client); err != nil {
		log.Printf("Hotswap examples failed: %v", err)
	}

	fmt.Println("\n=== All Compute Examples Complete ===")
}

// runDiskExamples demonstrates all disk operations.
func runDiskExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("--- Disk Examples ---")

	// Example 1: Create a boot disk from an image
	fmt.Println("\n1. Creating boot disk from Ubuntu image...")
	bootDisk := compute.NewDiskBuilder("example-boot-disk").
		WithImage("ubuntu.24-04.1").
		WithSizeGB(50).
		WithZone("a").
		Build()

	createdBootDisk, err := client.Compute().Disks().Create(ctx, bootDisk)
	if err != nil {
		return fmt.Errorf("failed to create boot disk: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (image: %s, size: %dGB)\n",
		*createdBootDisk.Metadata.Name,
		createdBootDisk.Spec.DiskImage.DiskImageRef.Name,
		createdBootDisk.Spec.DiskSize.Amount)

	// Example 2: Create an empty data disk
	fmt.Println("\n2. Creating empty data disk...")
	dataDisk := compute.NewDiskBuilder("example-data-disk").
		WithSizeGB(100).
		WithZone("a").
		Build()

	createdDataDisk, err := client.Compute().Disks().Create(ctx, dataDisk)
	if err != nil {
		return fmt.Errorf("failed to create data disk: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (size: %dGB)\n",
		*createdDataDisk.Metadata.Name,
		createdDataDisk.Spec.DiskSize.Amount)

	// Example 3: Create another boot disk with different image
	fmt.Println("\n3. Creating boot disk with minimal Ubuntu image...")
	bootDisk2 := compute.NewDiskBuilder("example-boot-disk-2").
		WithImage("ubuntu-minimal.24-04.1").
		WithSizeGB(30).
		WithZone("a").
		Build()

	createdBootDisk2, err := client.Compute().Disks().Create(ctx, bootDisk2)
	if err != nil {
		return fmt.Errorf("failed to create boot disk 2: %w", err)
	}
	fmt.Printf("   ✓ Created: %s\n", *createdBootDisk2.Metadata.Name)

	// Example 4: Create another large data disk
	fmt.Println("\n4. Creating 200GB data disk...")
	dataDisk2 := compute.NewDiskBuilder("example-data-disk-2").
		WithSizeGB(200).
		WithZone("a").
		Build()

	createdDataDisk2, err := client.Compute().Disks().Create(ctx, dataDisk2)
	if err != nil {
		return fmt.Errorf("failed to create data disk 2: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (size: %dGB)\n",
		*createdDataDisk2.Metadata.Name,
		createdDataDisk2.Spec.DiskSize.Amount)

	// Wait for disks to be ready
	fmt.Println("\n5. Waiting for disks to be ready...")
	disksToWait := []string{"example-boot-disk", "example-data-disk", "example-boot-disk-2", "example-data-disk-2"}
	for _, diskName := range disksToWait {
		if err := client.Compute().Disks().WaitForReady(ctx, diskName, 2*time.Minute); err != nil {
			return fmt.Errorf("disk %s not ready: %w", diskName, err)
		}
		fmt.Printf("   ✓ %s is ready\n", diskName)
	}

	// Example 6: List all disks
	fmt.Println("\n6. Listing all disks...")
	disks, err := client.Compute().Disks().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list disks: %w", err)
	}
	fmt.Printf("   Found %d disks:\n", len(disks.Items))
	for _, disk := range disks.Items {
		status := "Not Ready"
		if compute.IsDiskReady(&disk) {
			status = "Ready"
		}
		imageInfo := "empty"
		if disk.Spec.DiskImage != nil {
			imageInfo = disk.Spec.DiskImage.DiskImageRef.Name
		}
		fmt.Printf("   - %s: %dGB, image: %s, status: %s\n",
			*disk.Metadata.Name,
			disk.Spec.DiskSize.Amount,
			imageInfo,
			status)
	}

	// Example 7: Get a specific disk
	fmt.Println("\n7. Getting specific disk details...")
	disk, err := client.Compute().Disks().Get(ctx, "example-boot-disk")
	if err != nil {
		return fmt.Errorf("failed to get disk: %w", err)
	}
	fmt.Printf("   ✓ Disk: %s\n", *disk.Metadata.Name)
	fmt.Printf("     Size: %d%s\n", disk.Spec.DiskSize.Amount, disk.Spec.DiskSize.Unit)
	fmt.Printf("     Storage Class: %s\n", disk.Spec.DiskStorageClass.Name)
	if disk.Spec.Placement != nil && disk.Spec.Placement.Zone != nil {
		fmt.Printf("     Zone: %s\n", *disk.Spec.Placement.Zone)
	}

	return nil
}

// runPlacementGroupExamples demonstrates all placement group operations.
func runPlacementGroupExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Placement Group Examples ---")

	// Example 1: Create a spread placement group
	fmt.Println("\n1. Creating spread placement group...")
	pg := compute.NewPlacementGroupBuilder("example-pg-spread", "spread").
		WithZone("a").
		Build()

	createdPG, err := client.Compute().PlacementGroups().Create(ctx, pg)
	if err != nil {
		return fmt.Errorf("failed to create placement group: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (strategy: %s)\n",
		*createdPG.Metadata.Name,
		createdPG.Spec.Strategy.Type)

	// Example 2: Create another placement group
	fmt.Println("\n2. Creating another spread placement group...")
	pg2 := compute.NewPlacementGroupBuilder("example-pg-2", "spread").
		WithZone("a").
		Build()

	_, err = client.Compute().PlacementGroups().Create(ctx, pg2)
	if err != nil {
		return fmt.Errorf("failed to create placement group 2: %w", err)
	}
	fmt.Println("   ✓ Created: example-pg-2")

	// Example 3: List all placement groups
	fmt.Println("\n3. Listing all placement groups...")
	pgs, err := client.Compute().PlacementGroups().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list placement groups: %w", err)
	}
	fmt.Printf("   Found %d placement groups:\n", len(pgs.Items))
	for _, pg := range pgs.Items {
		zone := "unspecified"
		if pg.Spec.Placement != nil && pg.Spec.Placement.Zone != nil {
			zone = *pg.Spec.Placement.Zone
		}
		fmt.Printf("   - %s: strategy=%s, zone=%s\n",
			*pg.Metadata.Name,
			pg.Spec.Strategy.Type,
			zone)
	}

	// Example 4: Get a specific placement group
	fmt.Println("\n4. Getting specific placement group...")
	pgDetails, err := client.Compute().PlacementGroups().Get(ctx, "example-pg-spread")
	if err != nil {
		return fmt.Errorf("failed to get placement group: %w", err)
	}
	fmt.Printf("   ✓ Placement Group: %s\n", *pgDetails.Metadata.Name)
	fmt.Printf("     Strategy: %s\n", pgDetails.Spec.Strategy.Type)

	return nil
}

// runVMExamples demonstrates all VM operations.
func runVMExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Virtual Machine Examples ---")

	// Example 1: Create a simple VM with boot disk only
	fmt.Println("\n1. Creating simple VM with boot disk...")
	simpleVM := compute.NewVirtualMachineBuilder("example-vm-simple").
		WithBootDisk("example-boot-disk-2").
		WithSize("a1a.xs").
		WithZone("a").
		WithSSHKey("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFeENOwB0QwUEicJGrFxt44yiShgBWzANhpE/5gNw041 user@example.com").
		Build()

	createdSimpleVM, err := client.Compute().VirtualMachines().Create(ctx, simpleVM)
	if err != nil {
		return fmt.Errorf("failed to create simple VM: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (size: %s)\n",
		*createdSimpleVM.Metadata.Name,
		createdSimpleVM.Spec.VmVirtualResourcesRef.VmVirtualResourcesRefName)

	// Example 2: Create a VM with multiple disks and cloud-init
	fmt.Println("\n2. Creating VM with multiple disks and cloud-init...")
	cloudInit := `#cloud-config
packages:
  - nginx
  - git
runcmd:
  - systemctl start nginx
  - systemctl enable nginx
`
	complexVM := compute.NewVirtualMachineBuilder("example-vm-complex").
		WithBootDisk("example-boot-disk").
		WithDataDisk("example-data-disk").
		WithSize("c1a.m").
		WithZone("a").
		WithPlacementGroup("example-pg-spread").
		WithSSHKey("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFeENOwB0QwUEicJGrFxt44yiShgBWzANhpE/5gNw041 user@example.com").
		WithCloudInit(cloudInit).
		Build()

	createdComplexVM, err := client.Compute().VirtualMachines().Create(ctx, complexVM)
	if err != nil {
		return fmt.Errorf("failed to create complex VM: %w", err)
	}
	fmt.Printf("   ✓ Created: %s (size: %s, disks: %d)\n",
		*createdComplexVM.Metadata.Name,
		createdComplexVM.Spec.VmVirtualResourcesRef.VmVirtualResourcesRefName,
		len(createdComplexVM.Spec.DiskRefs))

	// Example 3: Create VM in stopped state
	fmt.Println("\n3. Creating VM in stopped state...")
	stoppedVM := compute.NewVirtualMachineBuilder("example-vm-stopped").
		WithBootDisk("example-data-disk-2").  // Using data disk as boot for demo
		WithSize("a1a.xs").
		WithZone("a").
		WithRunning(false).  // Create in stopped state
		Build()

	createdStoppedVM, err := client.Compute().VirtualMachines().Create(ctx, stoppedVM)
	if err != nil {
		return fmt.Errorf("failed to create stopped VM: %w", err)
	}
	running := "stopped"
	if createdStoppedVM.Spec.Running != nil && *createdStoppedVM.Spec.Running {
		running = "running"
	}
	fmt.Printf("   ✓ Created: %s (state: %s)\n", *createdStoppedVM.Metadata.Name, running)

	// Wait for VMs to be ready
	fmt.Println("\n4. Waiting for VMs to be ready...")
	vmsToWait := []string{"example-vm-simple", "example-vm-complex"}
	for _, vmName := range vmsToWait {
		if err := client.Compute().VirtualMachines().WaitForReady(ctx, vmName, 5*time.Minute); err != nil {
			log.Printf("   Warning: VM %s not ready: %v", vmName, err)
		} else {
			fmt.Printf("   ✓ %s is ready\n", vmName)
		}
	}

	// Example 5: List all VMs
	fmt.Println("\n5. Listing all VMs...")
	vms, err := client.Compute().VirtualMachines().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list VMs: %w", err)
	}
	fmt.Printf("   Found %d VMs:\n", len(vms.Items))
	for _, vm := range vms.Items {
		state := compute.GetVMState(&vm)
		running := "unknown"
		if vm.Spec.Running != nil {
			if *vm.Spec.Running {
				running = "running"
			} else {
				running = "stopped"
			}
		}
		fmt.Printf("   - %s: size=%s, state=%s, desired=%s\n",
			*vm.Metadata.Name,
			vm.Spec.VmVirtualResourcesRef.VmVirtualResourcesRefName,
			state,
			running)
	}

	// Example 6: Get a specific VM
	fmt.Println("\n6. Getting specific VM details...")
	vmDetails, err := client.Compute().VirtualMachines().Get(ctx, "example-vm-complex")
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	fmt.Printf("   ✓ VM: %s\n", *vmDetails.Metadata.Name)
	fmt.Printf("     Size: %s\n", vmDetails.Spec.VmVirtualResourcesRef.VmVirtualResourcesRefName)
	fmt.Printf("     Disks: %d\n", len(vmDetails.Spec.DiskRefs))
	if vmDetails.Spec.PlacementGroup != nil {
		fmt.Printf("     Placement Group: %s\n", *vmDetails.Spec.PlacementGroup)
	}

	// Example 7: Stop a running VM by updating the running spec
	fmt.Println("\n7. Stopping a running VM...")
	vmToStop, err := client.Compute().VirtualMachines().Get(ctx, "example-vm-simple")
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	stopState := false
	vmToStop.Spec.Running = &stopState
	_, err = client.Compute().VirtualMachines().Update(ctx, "example-vm-simple", vmToStop)
	if err != nil {
		return fmt.Errorf("failed to stop VM: %w", err)
	}
	fmt.Println("   ✓ Stopped example-vm-simple")
	time.Sleep(5 * time.Second)

	// Example 8: Start a stopped VM by updating the running spec
	fmt.Println("\n8. Starting a stopped VM...")
	vmToStart, err := client.Compute().VirtualMachines().Get(ctx, "example-vm-simple")
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	startState := true
	vmToStart.Spec.Running = &startState
	_, err = client.Compute().VirtualMachines().Update(ctx, "example-vm-simple", vmToStart)
	if err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}
	fmt.Println("   ✓ Started example-vm-simple")

	// Example 9: Check if VM is ready
	fmt.Println("\n9. Checking VM readiness...")
	vm, err := client.Compute().VirtualMachines().Get(ctx, "example-vm-complex")
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	if compute.IsVMReady(vm) {
		fmt.Println("   ✓ example-vm-complex is ready")
	} else {
		fmt.Println("   ⚠ example-vm-complex is not ready yet")
	}

	return nil
}

// runHotswapExamples demonstrates hotswap disk attachment operations.
func runHotswapExamples(ctx context.Context, client *evroc.Client) error {
	fmt.Println("\n--- Hotswap Disk Attachment Examples ---")

	// Create an additional disk for hotswap
	fmt.Println("\n1. Creating disk for hotswap attachment...")
	hotswapDisk := compute.NewDiskBuilder("example-hotswap-disk").
		WithSizeGB(50).
		WithZone("a").
		Build()

	_, err := client.Compute().Disks().Create(ctx, hotswapDisk)
	if err != nil {
		return fmt.Errorf("failed to create hotswap disk: %w", err)
	}
	fmt.Println("   ✓ Created example-hotswap-disk")

	// Wait for disk
	if err := client.Compute().Disks().WaitForReady(ctx, "example-hotswap-disk", 2*time.Minute); err != nil {
		return fmt.Errorf("hotswap disk not ready: %w", err)
	}
	fmt.Println("   ✓ Disk is ready")

	// Example 2: Attach disk to VM using hotswap
	fmt.Println("\n2. Creating hotswap attachment to VM...")
	attachment := compute.NewHotswapDiskAttachmentBuilder(
		"example-hotswap-1",
		"example-vm-complex",
		"example-hotswap-disk",
	).Build()

	createdAttachment, err := client.Compute().HotswapDiskAttachments().Create(ctx, attachment)
	if err != nil {
		return fmt.Errorf("failed to create hotswap attachment: %w", err)
	}
	fmt.Printf("   ✓ Attached disk %s to VM %s\n",
		createdAttachment.Spec.DiskRef,
		createdAttachment.Spec.VmRef)

	// Example 3: List all hotswap attachments
	fmt.Println("\n3. Listing all hotswap attachments...")
	attachments, err := client.Compute().HotswapDiskAttachments().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list attachments: %w", err)
	}
	fmt.Printf("   Found %d hotswap attachments:\n", len(attachments.Items))
	for _, att := range attachments.Items {
		fmt.Printf("   - %s: VM=%s, Disk=%s\n",
			*att.Metadata.Name,
			att.Spec.VmRef,
			att.Spec.DiskRef)
	}

	// Example 4: Get specific attachment details
	fmt.Println("\n4. Getting specific hotswap attachment...")
	att, err := client.Compute().HotswapDiskAttachments().Get(ctx, "example-hotswap-1")
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}
	fmt.Printf("   ✓ Attachment: %s\n", *att.Metadata.Name)
	fmt.Printf("     VM: %s\n", att.Spec.VmRef)
	fmt.Printf("     Disk: %s\n", att.Spec.DiskRef)

	// Example 5: Delete hotswap attachment (detach disk)
	fmt.Println("\n5. Deleting hotswap attachment (detaching disk)...")
	err = client.Compute().HotswapDiskAttachments().Delete(ctx, "example-hotswap-1")
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	fmt.Println("   ✓ Detached disk from VM")

	return nil
}
