// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package compute provides access to the evroc Compute API.
//
// The Compute API enables you to manage virtual machines, disks, placement groups,
// and disk attachments in the evroc Cloud Platform.
//
// # Resources
//
// The compute package provides access to the following resources:
//
//   - Virtual Machines: Create and manage VM instances
//   - Disks: Create and manage persistent block storage
//   - Placement Groups: Control VM placement for high availability
//   - Hotswap Disk Attachments: Attach/detach disks without VM restart
//
// # Getting Started
//
// Create a client and list virtual machines:
//
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	vms, err := client.Compute().VirtualMachines().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Creating Resources
//
// Use builders for a fluent API to create resources:
//
//	vm, err := client.Compute().VirtualMachines().Create(ctx,
//	    compute.NewVirtualMachineBuilder("my-vm").
//	        WithBootDisk("my-disk").
//	        WithSize("c1a.m").
//	        WithSSHKey("ssh-rsa AAAA...").
//	        Build(),
//	)
//
// Or use the convenience function to create a VM with a new disk:
//
//	disk, vm, err := compute.CreateVMWithNewDisk(
//	    ctx,
//	    client.Compute(),
//	    "my-vm",
//	    "my-disk",
//	    "ubuntu-minimal.24-04.1",
//	    "a1a.xs",
//	)
//
// # Updating Resources
//
// Use update builders to modify existing resources:
//
//	vm, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
//	    Stop().
//	    Apply(ctx)
//
// # Waiters
//
// Wait for resources to reach desired states:
//
//	err := client.Compute().VirtualMachines().WaitForReady(
//	    ctx,
//	    "my-vm",
//	    5*time.Minute,
//	)
//
// Customize waiter behavior:
//
//	err := client.Compute().Disks().WaitForReady(
//	    ctx,
//	    "my-disk",
//	    10*time.Minute,
//	    compute.WithProgressCallback(func(attempt int, elapsed time.Duration) {
//	        log.Printf("Waiting... attempt %d, elapsed %s", attempt, elapsed)
//	    }),
//	)
//
// # Error Handling
//
// The SDK automatically retries transient errors with exponential backoff.
// Customize retry behavior:
//
//	client, err := evroc.NewFromEnv(ctx,
//	    evroc.WithHTTPClient(customClient),
//	)
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	vm, err := client.Compute().VirtualMachines().Get(ctx, "my-vm")
package compute
