// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package compute_test

import (
	"context"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

func TestE2E_Disk_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	diskName := e2etest.RandomName("disk")

	t.Logf("Creating disk: %s", diskName)

	// Create disk
	disk, err := compute.NewDiskBuilder(diskName).
		WithSizeGB(e2etest.TestDiskSizeGB).
		WithImage(string(e2etest.TestDiskImage)).
		WithZone(e2etest.TestDiskZone).
		Create(ctx, client.Compute().Disks())

	if err != nil {
		t.Fatalf("failed to create disk: %v", err)
	}

	diskID := e2etest.MustGetID(t, disk.Metadata.Id, "disk")
	t.Logf("Created disk with ID: %s", diskID)

	diskDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Compute().Disks().Delete, diskID, "disk", &diskDeleted)

	// Wait for disk to be ready
	t.Logf("Waiting for disk to be ready...")
	if _, err := client.Compute().Disks().WaitForReady(ctx, diskID, e2etest.DiskReadyTimeout); err != nil {
		t.Fatalf("disk never became ready: %v", err)
	}

	// Verify disk was created with correct size
	if disk.Spec.DiskSize == nil {
		t.Error("disk size is nil")
	} else if disk.Spec.DiskSize.Amount != e2etest.TestDiskSizeGB {
		t.Errorf("expected disk size %d, got %d", e2etest.TestDiskSizeGB, disk.Spec.DiskSize.Amount)
	}

	// Read disk
	t.Logf("Reading disk: %s", diskID)
	retrieved, err := client.Compute().Disks().Get(ctx, diskID)
	if err != nil {
		t.Fatalf("failed to get disk: %v", err)
	}

	if retrieved.Metadata.Id != diskID {
		t.Errorf("expected disk ID %s, got %v", diskID, retrieved.Metadata.Id)
	}

	// List disks - should include our disk
	t.Logf("Listing disks")
	disks, err := client.Compute().Disks().List(ctx)
	if err != nil {
		t.Fatalf("failed to list disks: %v", err)
	}
	e2etest.AssertInList(t, disks.Items, diskID, func(d computetypes.Disk) string { return d.Metadata.Id }, "disk")

	// Note: Disk size is immutable in the API, so we skip resize testing

	// Delete disk
	t.Logf("Deleting disk: %s", diskID)
	if err := client.Compute().Disks().Delete(ctx, diskID); err != nil {
		t.Fatalf("failed to delete disk: %v", err)
	}
	diskDeleted = true

	// Verify disk was deleted
	t.Logf("Verifying disk deletion")
	e2etest.AssertDeleted(t, ctx, func(ctx context.Context, id string) (any, error) {
		return client.Compute().Disks().Get(ctx, id)
	}, diskID, "disk")

	t.Logf("Disk lifecycle test completed successfully")
}

func TestE2E_VirtualMachine_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	vmName := e2etest.RandomName("vm")
	diskName := e2etest.RandomName("disk")

	t.Logf("Creating boot disk: %s", diskName)

	// Create boot disk first
	disk, err := compute.NewDiskBuilder(diskName).
		WithSizeGB(e2etest.TestDiskSizeGB).
		WithImage(string(e2etest.TestDiskImage)).
		WithZone(e2etest.TestDiskZone).
		Create(ctx, client.Compute().Disks())

	if err != nil {
		t.Fatalf("failed to create boot disk: %v", err)
	}

	diskID := e2etest.MustGetID(t, disk.Metadata.Id, "disk")
	diskDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Compute().Disks().Delete, diskID, "boot disk", &diskDeleted)

	// Wait for boot disk to be ready
	t.Logf("Waiting for boot disk to be ready...")
	if _, err := client.Compute().Disks().WaitForReady(ctx, diskID, e2etest.DiskReadyTimeout); err != nil {
		t.Fatalf("boot disk never became ready: %v", err)
	}

	t.Logf("Creating VM: %s", vmName)

	// Create VM with the boot disk
	vm, err := compute.NewVirtualMachineBuilder(vmName).
		WithVMInstanceType(string(e2etest.TestVMSize)).
		WithBootDisk(disk.Ref()).
		WithSubnet(client.Compute().DefaultSubnetRef(e2etest.TestVMZone)).
		WithZone(e2etest.TestVMZone).
		Create(ctx, client.Compute().VirtualMachines())

	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	vmID := e2etest.MustGetID(t, vm.Metadata.Id, "VM")
	t.Logf("Created VM with ID: %s", vmID)

	vmDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Compute().VirtualMachines().Delete, vmID, "VM", &vmDeleted)

	// Wait for VM to be ready
	t.Logf("Waiting for VM to be ready...")
	if _, err := client.Compute().VirtualMachines().WaitForReady(ctx, vmID, e2etest.VMReadyTimeout); err != nil {
		t.Fatalf("VM never became ready: %v", err)
	}

	// Read VM
	t.Logf("Reading VM: %s", vmID)
	retrieved, err := client.Compute().VirtualMachines().Get(ctx, vmID)
	if err != nil {
		t.Fatalf("failed to get VM: %v", err)
	}

	if retrieved.Metadata.Id != vmID {
		t.Errorf("expected VM ID %s, got %v", vmID, retrieved.Metadata.Id)
	}

	// List VMs - should include our VM
	t.Logf("Listing VMs")
	vms, err := client.Compute().VirtualMachines().List(ctx)
	if err != nil {
		t.Fatalf("failed to list VMs: %v", err)
	}
	e2etest.AssertInList(t, vms.Items, vmID, func(v computetypes.VirtualMachine) string { return v.Metadata.Id }, "VM")

	// Delete VM
	t.Logf("Deleting VM: %s", vmID)
	if err := client.Compute().VirtualMachines().Delete(ctx, vmID); err != nil {
		t.Fatalf("failed to delete VM: %v", err)
	}
	vmDeleted = true

	// Verify VM was deleted
	t.Logf("Verifying VM deletion (this may take 1-2 minutes)...")
	deleted := false
	maxAttempts := int(e2etest.VMDeletionTimeout / e2etest.VMDeletionInterval)
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(e2etest.VMDeletionInterval)
		_, err = client.Compute().VirtualMachines().Get(ctx, vmID)
		if err != nil {
			// VM is gone
			deleted = true
			t.Logf("VM successfully deleted after %v", time.Duration(i+1)*e2etest.VMDeletionInterval)
			break
		}
		// Log every 3 attempts
		if (i+1)%3 == 0 {
			t.Logf("VM still exists, waiting... (%v elapsed)", time.Duration(i+1)*e2etest.VMDeletionInterval)
		}
	}

	if !deleted {
		t.Errorf("VM was not deleted after %v", e2etest.VMDeletionTimeout)
	}

	t.Logf("VM lifecycle test completed successfully")
}

func TestE2E_PlacementGroup_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	pgName := e2etest.RandomName("pg")

	t.Logf("Creating placement group: %s", pgName)

	// Create placement group
	pg, err := compute.NewPlacementGroupBuilder(pgName, e2etest.TestPlacementGroupStrategy).
		WithZone(e2etest.TestPlacementGroupZone).
		Create(ctx, client.Compute().PlacementGroups())

	if err != nil {
		t.Fatalf("failed to create placement group: %v", err)
	}

	pgID := e2etest.MustGetID(t, pg.Metadata.Id, "placement group")
	t.Logf("Created placement group with ID: %s", pgID)

	pgDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Compute().PlacementGroups().Delete, pgID, "placement group", &pgDeleted)

	// Verify placement group was created
	if string(pg.Spec.Strategy.Type) != e2etest.TestPlacementGroupStrategy {
		t.Errorf("expected strategy %s, got %s", e2etest.TestPlacementGroupStrategy, pg.Spec.Strategy.Type)
	}

	// Read placement group
	t.Logf("Reading placement group: %s", pgID)
	retrieved, err := client.Compute().PlacementGroups().Get(ctx, pgID)
	if err != nil {
		t.Fatalf("failed to get placement group: %v", err)
	}

	if retrieved.Metadata.Id != pgID {
		t.Errorf("expected placement group ID %s, got %v", pgID, retrieved.Metadata.Id)
	}

	// List placement groups - should include ours
	t.Logf("Listing placement groups")
	pgs, err := client.Compute().PlacementGroups().List(ctx)
	if err != nil {
		t.Fatalf("failed to list placement groups: %v", err)
	}
	e2etest.AssertInList(t, pgs.Items, pgID, func(p computetypes.PlacementGroup) string { return p.Metadata.Id }, "placement group")

	// Delete placement group
	t.Logf("Deleting placement group: %s", pgID)
	if err := client.Compute().PlacementGroups().Delete(ctx, pgID); err != nil {
		t.Fatalf("failed to delete placement group: %v", err)
	}
	pgDeleted = true

	t.Logf("Placement group lifecycle test completed successfully")
}
