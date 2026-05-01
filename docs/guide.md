# SDK Guide

## Understanding Asynchronous Operations

The evroc Cloud Platform API is asynchronous by nature. When you create, update, or delete a resource, the API returns immediately with the request accepted, but the actual operation continues in the background.

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

## Using Waiters

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
disk, err = client.Compute().Disks().WaitForReady(ctx, *disk.Metadata.Name, 5*time.Minute)
if err != nil {
    log.Fatal("Disk did not become ready in time")
}

// Now it's safe to create a VM with this disk
vmReq := compute.NewVirtualMachineBuilder("my-vm").
    WithBootDisk(disk.Ref()).  // Use disk.Ref() to get type-safe reference
    WithVMInstanceType("a1a.xs").
    Build()
vm, err := client.Compute().VirtualMachines().Create(ctx, vmReq)
```

### Available Waiters

```go
// Wait for resources to be ready
disk, err := client.Compute().Disks().WaitForReady(ctx, "my-disk", timeout)
vm, err := client.Compute().VirtualMachines().WaitForReady(ctx, "my-vm", timeout)
ip, err := client.Networking().PublicIPs().WaitForReady(ctx, "my-ip", timeout)

// Wait for deletion to complete
err := client.Compute().Disks().WaitForDeleted(ctx, "my-disk", timeout)
err := client.Compute().VirtualMachines().WaitForDeleted(ctx, "my-vm", timeout)
err := client.Networking().SecurityGroups().WaitForDeleted(ctx, "my-sg", timeout)
```

### When to Use Waiters

Use waiters when:
- Creating resources that will be used immediately (e.g., disk before creating VM)
- Deleting resources before recreating them with the same name
- Updating resources and need confirmation before proceeding
- Testing or scripts where you need deterministic behavior

You can skip waiters when:
- Fire-and-forget operations (create resources and don't need immediate status)
- Building async workflows (create multiple resources in parallel)
- Using event-driven architectures (rely on status checks or webhooks)

### Customizing Waiter Behavior

Waiters support customizable polling intervals and progress tracking:

```go
// Default behavior: exponential backoff (2s initial, 30s max)
disk, err := client.Compute().Disks().WaitForReady(ctx, "my-disk", 2*time.Minute)

// Fast constant polling for quick operations
disk, err := client.Compute().Disks().WaitForReady(
    ctx, "my-disk", 2*time.Minute,
    compute.WithPollingInterval(1*time.Second),
)

// Custom exponential backoff
vm, err := client.Compute().VirtualMachines().WaitForReady(
    ctx, "my-vm", 5*time.Minute,
    compute.WithExponentialBackoff(1*time.Second, 15*time.Second, 1.5),
)

// Track progress during long waits
vm, err := client.Compute().VirtualMachines().WaitForReady(
    ctx, "my-vm", 5*time.Minute,
    compute.WithProgressCallback(func(attempt int, elapsed time.Duration) {
        log.Printf("Still waiting... attempt %d at %v", attempt, elapsed)
    }),
)
```

## Status Checking

```go
// Check if resources are ready
if compute.IsVMReady(vm) {
    log.Println("VM is ready")
}

if compute.IsDiskReady(disk) {
    log.Println("Disk is ready")
}

if networking.IsPublicIPReady(publicIP) {
    ipAddress := networking.GetPublicIPAddress(publicIP)
    log.Printf("Public IP ready: %s", ipAddress)
}

// Get VM state
state := compute.GetVMState(vm)
log.Printf("VM State: %s", state)
```

## Working with Optional Fields

Many API response fields are pointers to allow for optional/unset values. Handle them correctly:

```go
// Check if field is set before using
if vm.Metadata.Name != nil {
    log.Printf("VM Name: %s", *vm.Metadata.Name)
}

// For fields that are always present after creation, direct dereference is safe
vmName := *vm.Metadata.Name  // Safe - name is always set on created resources

// For optional fields, provide explicit defaults
running := true  // API default
if vm.Spec.Running != nil {
    running = *vm.Spec.Running
}
```

Note: The generic `evroc.Ptr()` helper is only needed if you're using direct construction instead of builders.

## Context Best Practices

The SDK fully respects Go contexts for cancellation and timeouts:

```go
// Use context.WithTimeout for operations that might hang
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

vms, err := client.Compute().VirtualMachines().List(ctx)
// Operation will cancel if it takes more than 30 seconds

// Use signal.NotifyContext for graceful shutdown on Ctrl-C
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

// All SDK operations will stop when user presses Ctrl-C
disk, err := client.Compute().Disks().WaitForReady(ctx, "my-disk", 5*time.Minute)
```

## Automatic Retry Logic

The SDK automatically retries transient errors with exponential backoff:

```go
// SDK handles retries automatically
vm, err := client.Compute().VirtualMachines().Create(ctx, vmReq)
// Will retry on network errors, 5xx server errors, and 429 (rate limit)
// Does NOT retry on 4xx client errors (invalid request, not found, etc.)

// Default retry behavior:
// - Max retries: 3
// - Initial backoff: 1s
// - Max backoff: 30s
// - Backoff multiplier: 2x with jitter
```

Retries are transparent and automatic for all SDK operations. The SDK will:
- Retry network errors (connection refused, timeout, etc.)
- Retry 5xx server errors (500, 502, 503, 504)
- Retry 429 rate limit errors (when implemented)
- NOT retry 4xx client errors (400, 404, 403, etc.)

## Error Handling

The SDK provides typed errors that can be checked using the standard library's `errors.Is()`:

```go
import (
    "errors"
    evroc "github.com/evroc-oss/evroc-go-sdk"
)

// Check for specific HTTP status codes
vm, err := client.Compute().VirtualMachines().Get(ctx, "my-vm")
if errors.Is(err, evroc.ErrNotFound) {
    log.Println("VM not found")
} else if errors.Is(err, evroc.ErrForbidden) {
    log.Println("Access denied")
} else if err != nil {
    log.Printf("API error: %v", err)
}

// Check for conflicts when creating resources
disk, err := client.Compute().Disks().Create(ctx, diskReq)
if errors.Is(err, evroc.ErrConflict) {
    log.Println("Disk already exists")
}
```

Available sentinel errors:
- `evroc.ErrNotFound` (404) - Resource not found
- `evroc.ErrConflict` (409) - Resource already exists
- `evroc.ErrForbidden` (403) - Access denied
- `evroc.ErrBadRequest` (400) - Invalid request
