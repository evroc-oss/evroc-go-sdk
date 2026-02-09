# Common Operations

## Create and Wait for Resources

Resources are created asynchronously. Use built-in waiters instead of manual polling:

```go
// Create a 50GB disk with Ubuntu image (returns immediately, provisioning happens in background)
disk := compute.NewDiskBuilder("my-disk").
    WithImage(compute.DiskImageUbuntu2404).
    WithSizeGB(50).  // Disk size: 50 gigabytes
    Build()

createdDisk, err := client.Compute().Disks().Create(ctx, disk)
if err != nil {
    log.Fatal(err)
}

// IMPORTANT: Use the built-in waiter instead of manual polling
err = client.Compute().Disks().WaitForReady(ctx, *createdDisk.Metadata.Name, 5*time.Minute)
if err != nil {
    log.Fatal("Disk did not become ready")
}

// Now the disk is ready - safe to create VM
vm := compute.NewVirtualMachineBuilder("my-vm").
    WithBootDisk(*createdDisk.Metadata.Name).  // Use the disk we just created
    WithVMInstanceType("a1a.xs").  // 1 vCPU, 4GB RAM
    Build()

createdVM, err := client.Compute().VirtualMachines().Create(ctx, vm)
if err != nil {
    log.Fatal(err)
}

// Wait for VM to be running (if needed)
err = client.Compute().VirtualMachines().WaitForRunning(ctx, "my-vm", 5*time.Minute)
```

## List and Filter Resources

```go
// List all VMs
vms, err := client.Compute().VirtualMachines().List(ctx)
for _, vm := range vms.Items {
    state := compute.GetVMState(&vm)
    log.Printf("VM: %s (State: %s)", *vm.Metadata.Name, state)
}

// List all disks
disks, err := client.Compute().Disks().List(ctx)
for _, disk := range disks.Items {
    if compute.IsDiskReady(&disk) {
        log.Printf("Ready disk: %s", *disk.Metadata.Name)
    }
}
```

## Update Resources with Update Builders

Update builders provide a clean, fluent interface for resource modifications across **all** services (Compute, Networking, Storage, IAM).

**Compute:**

```go
// Stop/Start VM
updatedVM, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Stop().
    AddLabel("maintenance", "true").
    Apply(ctx)

// Update VM configuration (must be stopped first)
// The following operations require the VM to be stopped:
//   - Resize (change compute profile)
//   - Change public IP
//   - Change placement group

// Complete workflow example:
// 1. Stop the VM
_, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Stop().
    Apply(ctx)
err = client.Compute().VirtualMachines().WaitForStopped(ctx, "my-vm", 5*time.Minute)

// 2. Make changes (resize, public IP, placement group)
updatedVM, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Resize("a1a.l").                         // Change compute profile
    SetPublicIP("new-public-ip").            // Change public IP
    SetPlacementGroup("new-placement-group"). // Change placement group
    Apply(ctx)

// 3. Start the VM again
updatedVM, err = compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Start().
    Apply(ctx)

// Individual update examples:

// Resize VM (change compute profile - VM must be stopped)
updatedVM, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Resize("a1a.l").
    Apply(ctx)

// Change public IP (VM must be stopped)
updatedVM, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    SetPublicIP("new-public-ip").
    Apply(ctx)

// Remove public IP (VM must be stopped)
updatedVM, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    RemovePublicIP().
    Apply(ctx)

// Change placement group (VM must be stopped)
updatedVM, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    SetPlacementGroup("new-placement-group").
    Apply(ctx)
```

**Networking:**

```go
// Update security group rules
rules := []networkingtypes.SecurityGroupSpecRulesItem{ /* ... */ }
updatedSG, err := networking.UpdateSecurityGroup("my-sg", client.Networking().SecurityGroups()).
    SetRules(rules).
    Apply(ctx)
```

**Storage:**

```go
// Update bucket retention
updatedBucket, err := storage.UpdateBucket("my-bucket", client.Storage().Buckets()).
    SetRetentionMode(storage.Versioned).
    Apply(ctx)

// Update service account buckets
updatedSA, err := storage.UpdateBucketServiceAccount("my-sa", client.Storage().BucketServiceAccounts()).
    SetBuckets([]string{"bucket1", "bucket2"}).
    Apply(ctx)
```

**Why Update Builders?**

Update builders eliminate the verbose get-modify-update pattern:

```go
// Without update builders (Terraform provider pattern):
sg, err := client.Networking().SecurityGroups().Get(ctx, "my-sg")
if err != nil { return err }
sg.Spec.Rules = &newRules
_, err = client.Networking().SecurityGroups().Update(ctx, "my-sg", sg)

// With update builders (concise):
_, err := networking.UpdateSecurityGroup("my-sg", sgService).
    SetRules(newRules).
    Apply(ctx)
```

Update builders automatically:
- Fetch the current resource state
- Apply your modifications
- Send the update request
- Handle errors gracefully

## Delete Resources

Delete operations are asynchronous - the API returns immediately but deletion continues in the background:

```go
// Delete VM (returns immediately, deletion happens in background)
err := client.Compute().VirtualMachines().Delete(ctx, "my-vm")
if err != nil {
    log.Fatal(err)
}

// If you need to wait for deletion to complete (e.g., recreating with same name):
err = client.Compute().VirtualMachines().WaitForDeleted(ctx, "my-vm", 5*time.Minute)
if err != nil {
    log.Fatal("VM deletion did not complete in time")
}

// Delete disk
err = client.Compute().Disks().Delete(ctx, "my-disk")

// Delete security group
err = client.Networking().SecurityGroups().Delete(ctx, "web-servers")

// Common pattern: Delete and wait before recreating
err = client.Compute().Disks().Delete(ctx, "my-disk")
if err != nil && !errors.Is(err, evroc.ErrNotFound) {
    log.Fatal(err)
}
err = client.Compute().Disks().WaitForDeleted(ctx, "my-disk", 2*time.Minute)
// Now safe to create new disk with same name
```
