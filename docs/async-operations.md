# Understanding Asynchronous Operations

The evroc Cloud Platform API is **asynchronous by nature**. When you create, update, or delete a resource, the API returns immediately with the request accepted, but the actual operation continues in the background.

## Why This Matters

```go
// Create a disk - API returns immediately
disk, err := client.Compute().Disks().Create(ctx, diskReq)
if err != nil {
    log.Fatal(err)
}
// At this point: disk is CREATED but NOT READY
// The disk is still being provisioned in the background

// Delete a VM - API returns immediately
err = client.Compute().VirtualMachines().Delete(ctx, "my-vm")
if err != nil {
    log.Fatal(err)
}
// At this point: VM deletion is INITIATED but NOT COMPLETE
// The VM is still being torn down in the background
```

## Using Waiters for Synchronous Workflows

For most workflows, you'll need to wait for operations to complete before proceeding:

```go
// Create a disk and wait for it to be ready
diskReq := compute.NewDiskBuilder("my-disk").
    WithImage(compute.DiskImageUbuntu2404).
    WithSizeGB(50).
    Build()

disk, err := client.Compute().Disks().Create(ctx, diskReq)
if err != nil {
    log.Fatal(err)
}

// IMPORTANT: Wait for the disk to be ready before using it
err = client.Compute().Disks().WaitForReady(ctx, *disk.Metadata.Name, 5*time.Minute)
if err != nil {
    log.Fatal("Disk did not become ready in time")
}

// Now it's safe to create a VM with this disk
vmReq := compute.NewVirtualMachineBuilder("my-vm").
    WithBootDisk(*disk.Metadata.Name).
    WithVMInstanceType("a1a.xs").
    Build()
vm, err := client.Compute().VirtualMachines().Create(ctx, vmReq)
```

## Common Waiter Operations

The SDK provides waiters for all asynchronous operations:

```go
// Wait for resources to be ready
client.Compute().Disks().WaitForReady(ctx, "my-disk", timeout)
client.Compute().VirtualMachines().WaitForReady(ctx, "my-vm", timeout)
client.Networking().PublicIPs().WaitForReady(ctx, "my-ip", timeout)

// Wait for deletion to complete
client.Compute().Disks().WaitForDeleted(ctx, "my-disk", timeout)
client.Compute().VirtualMachines().WaitForDeleted(ctx, "my-vm", timeout)
client.Networking().SecurityGroups().WaitForDeleted(ctx, "my-sg", timeout)

// Wait for state changes
client.Compute().VirtualMachines().WaitForRunning(ctx, "my-vm", timeout)
client.Compute().VirtualMachines().WaitForStopped(ctx, "my-vm", timeout)
```

## When to Use Waiters

Use waiters when:
- Creating resources that will be used immediately (e.g., disk before creating VM)
- Deleting resources before recreating them with the same name
- Updating resources and need confirmation before proceeding
- Testing or scripts where you need deterministic behavior

You can skip waiters when:
- Fire-and-forget operations (create resources and don't need immediate status)
- Building async workflows (create multiple resources in parallel)
- Using event-driven architectures (rely on status checks or webhooks)
