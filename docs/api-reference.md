# API Reference

Complete reference for all evroc SDK APIs.

**Best practice**: When you create a resource, use `.Ref()` to reference it:
```go
disk, _ := client.Compute().Disks().Create(ctx, diskReq)
vm := builder.WithBootDisk(disk.Ref()).Build()  // Use .Ref()
```

When you need to reference an existing resource by name, use client helper methods:
```go
// Use client helpers to construct refs from names
vm := builder.
    WithBootDisk(client.Compute().DiskRef("my-disk")).
    WithPublicIP(client.Networking().PublicIPRef("my-ip")).
    Build()
```

## Compute

**IMPORTANT: All compute resources (VMs, Disks, PlacementGroups) require a zone to be specified. Zones are mandatory.**

Available zones: `a`, `b`, `c`

### Virtual Machines

```go
// Best practice: Create resources first, then reference them
disk, _ := client.Compute().Disks().Create(ctx, diskBuilder.Build())
sg, _ := client.Networking().SecurityGroups().Create(ctx, sgBuilder.Build())

vm := compute.NewVirtualMachineBuilder("my-vm").
    WithBootDisk(disk.Ref()).         // Use .Ref() for created resources
    WithSecurityGroup(sg.Ref()).
    WithVMInstanceType("c1a.m").
    WithSSHKey("ssh-rsa AAAA...").
    WithZone("a").  // REQUIRED: Zone must be specified (a, b, or c)
    Build()

createdVM, _ := client.Compute().VirtualMachines().Create(ctx, vm)

// If referencing existing resources by name using client helpers
vm = compute.NewVirtualMachineBuilder("production-vm").
    WithBootDisk(client.Compute().DiskRef("boot-disk")).
    WithDataDisk(client.Compute().DiskRef("data-disk-1")).
    WithDataDisk(client.Compute().DiskRef("data-disk-2")).
    WithVMInstanceType("m1a.xl").
    WithSSHKey(sshPublicKey).
    WithCloudInit(cloudInitScript).
    WithSecurityGroup(client.Networking().SecurityGroupRef("web-servers")).
    WithPublicIP(client.Networking().PublicIPRef("my-public-ip")).
    WithZone("a").
    Build()
```

### Disks

```go
// Create a 100GB disk with Ubuntu 24.04 image
disk := compute.NewDiskBuilder("my-disk").
    WithImage(string(compute.DiskImageUbuntu2404)).
    WithSizeGB(100).  // Disk capacity in gigabytes
    WithZone("a").  // REQUIRED: Zone must be specified (a, b, or c)
    Build()

createdDisk, err := client.Compute().Disks().Create(ctx, disk)

// Create an empty 500GB data disk (no OS image)
dataDisk := compute.NewDiskBuilder("data-disk").
    WithSizeGB(500).  // Just size, no image
    WithZone("a").  // REQUIRED: Zone must be specified (a, b, or c)
    Build()
```

### Placement Groups

**What are Placement Groups?**

Placement groups control how VMs are distributed across **physical hosts within a single zone**. This provides an additional layer of availability beyond zones.

**Understanding the availability hierarchy:**

1. **Region** (e.g., `se-sto`) - Geographic area with multiple data centers
2. **Zone** (e.g., `a`, `b`, `c`) - Physically separate data centers with independent power, cooling, and networking
3. **Physical hosts** - Individual servers within each data center
4. **Placement Groups** - Control VM distribution across physical hosts within ONE zone

**Why both zones AND placement groups?**

- **Zones** protect against **entire data center failures** (power outage, network failure, etc.)
- **Placement Groups** protect against **individual server failures within a zone**

This two-layer approach maximizes availability: distribute VMs across zones to survive data center failures, and use placement groups within each zone to survive individual host failures.

**Available Strategies:**

| Strategy | Purpose | Constraints | Use Case |
|----------|---------|-------------|----------|
| `spread` | Spreads VMs across different physical hosts within a zone | Max 5 VMs, hard anti-affinity (never on same host) | High availability - if one host fails, other VMs continue running |

**Creating and Using Placement Groups:**

```go
// Create a spread placement group
pg := compute.NewPlacementGroupBuilder("high-availability-pg", "spread").
    WithZone("a").  // REQUIRED: Zone must be specified (a, b, or c)
    Build()

createdPG, err := client.Compute().PlacementGroups().Create(ctx, pg)

// Create disks first
disk1, _ := client.Compute().Disks().Create(ctx, diskBuilder1.Build())
disk2, _ := client.Compute().Disks().Create(ctx, diskBuilder2.Build())

// Create VMs in the placement group - use .Ref() for resources we just created
vm1 := compute.NewVirtualMachineBuilder("web-server-1").
    WithBootDisk(disk1.Ref()).
    WithVMInstanceType("c1a.m").
    WithZone("a").
    WithPlacementGroup(createdPG.Ref()).  // Use the PG we just created
    Build()

vm2 := compute.NewVirtualMachineBuilder("web-server-2").
    WithBootDisk(disk2.Ref()).
    WithVMInstanceType("c1a.m").
    WithZone("a").
    WithPlacementGroup(createdPG.Ref()).  // Same placement group
    Build()

// The spread strategy ensures these VMs run on different physical hosts
```

**Changing Placement Groups:**

You can change a VM's placement group, but the VM must be stopped first:

```go
// Stop the VM
_, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Stop().
    Apply(ctx)

// Change placement group (or use RemovePlacementGroup() to remove)
_, err = compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    SetPlacementGroup("new-placement-group").
    Apply(ctx)

// Start the VM again
_, err = compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Start().
    Apply(ctx)
```

**List and Get Operations:**

```go
// List all placement groups
pgs, err := client.Compute().PlacementGroups().List(ctx)
for _, pg := range pgs.Items {
    fmt.Printf("%s: strategy=%s\n", *pg.Metadata.Name, pg.Spec.Strategy.Type)
}

// Get specific placement group
pg, err := client.Compute().PlacementGroups().Get(ctx, "high-availability-pg")
```

**Important Constraints:**

- **Maximum 5 VMs** per placement group
- **Hard anti-affinity** - VMs will NEVER be placed on the same physical host
- **Strict enforcement** - If there aren't enough distinct physical hosts available, some VMs won't be scheduled (they'll remain pending)
- **Zonal scope** - All VMs in a placement group must be in the same zone

**Best Practices:**

1. **High Availability** - Use `spread` strategy for production workloads that need to survive host failures
2. **Multi-zone distribution** - For even higher availability, create separate placement groups in each zone (e.g., `web-tier-zone-a`, `web-tier-zone-b`, `web-tier-zone-c`)
3. **Zone Matching** - Placement group zone must match the VM zone
4. **Plan Ahead** - Create placement groups before VMs, or you'll need to stop/start VMs to add them
5. **Naming** - Use descriptive names that include zone, like "web-tier-zone-a-spread" or "db-cluster-zone-b-spread"

**Example: Multi-zone high availability setup**

```go
// Create placement groups in three zones
pgZoneA := compute.NewPlacementGroupBuilder("web-tier-zone-a", "spread").
    WithZone("a").
    Build()
pgZoneB := compute.NewPlacementGroupBuilder("web-tier-zone-b", "spread").
    WithZone("b").
    Build()
pgZoneC := compute.NewPlacementGroupBuilder("web-tier-zone-c", "spread").
    WithZone("c").
    Build()

// Create VMs distributed across zones and physical hosts
// Zone A VMs (each on different host within zone A)
vm1 := compute.NewVirtualMachineBuilder("web-1").
    WithZone("a").
    WithPlacementGroup(client.Compute().PlacementGroupRef("web-tier-zone-a")).  // Protected from host failures in zone A
    Build()
vm2 := compute.NewVirtualMachineBuilder("web-2").
    WithZone("a").
    WithPlacementGroup(client.Compute().PlacementGroupRef("web-tier-zone-a")).
    Build()

// Zone B VMs (each on different host within zone B)
vm3 := compute.NewVirtualMachineBuilder("web-3").
    WithZone("b").
    WithPlacementGroup(client.Compute().PlacementGroupRef("web-tier-zone-b")).  // Protected from zone A failures
    Build()
vm4 := compute.NewVirtualMachineBuilder("web-4").
    WithZone("b").
    WithPlacementGroup(client.Compute().PlacementGroupRef("web-tier-zone-b")).
    Build()

// Zone C VMs (each on different host within zone C)
vm5 := compute.NewVirtualMachineBuilder("web-5").
    WithZone("c").
    WithPlacementGroup(client.Compute().PlacementGroupRef("web-tier-zone-c")).  // Protected from zone A and B failures
    Build()

// Result: 5 VMs protected from both data center failures (zones) AND server failures (placement groups)
```

## Networking

### Security Groups

```go
// Web server security group
sg := networking.NewSecurityGroupBuilder("web-servers").
    AllowSSH().         // Port 22 from anywhere
    AllowHTTP().        // Port 80 from anywhere
    AllowHTTPS().       // Port 443 from anywhere
    AllowIngressRule("custom-range", networking.TCP, 8000, 8999, "10.0.0.0/8").
    Build()

createdSG, err := client.Networking().SecurityGroups().Create(ctx, sg)
```

### VPCs and Subnets

VPCs and Subnets are read-only (you cannot create them via SDK yet):

```go
// List all VPCs
vpcs, err := client.Networking().VirtualPrivateClouds().List(ctx)

// Get specific VPC
vpc, err := client.Networking().VirtualPrivateClouds().Get(ctx, "vpc-name")

// List all subnets
subnets, err := client.Networking().Subnets().List(ctx)

// Get specific subnet
subnet, err := client.Networking().Subnets().Get(ctx, "subnet-name")
```

## Storage

### Buckets

```go
// Simple bucket
bucket := storage.NewBucketBuilder("my-bucket").Build()
created, err := client.Storage().Buckets().Create(ctx, bucket)

// Bucket with versioning
versionedBucket := storage.NewBucketBuilder("versioned-bucket").
    WithObjectRetentionMode("Versioned").
    Build()

// Bucket with object locking (GOVERNANCE mode, 30 days)
lockingBucket := storage.NewBucketBuilder("locked-bucket").
    WithObjectRetentionMode("Locking").
    WithDefaultObjectLocking("GOVERNANCE", 30).
    Build()

// Bucket with compliance locking (COMPLIANCE mode, 90 days)
complianceBucket := storage.NewBucketBuilder("compliance-bucket").
    WithObjectRetentionMode("Locking").
    WithDefaultObjectLocking("COMPLIANCE", 90).
    Build()
```

### Bucket Service Accounts

```go
// Create service account for S3 access
serviceAccount := storage.NewBucketServiceAccountBuilder(
    "my-service-account",
    "my-bucket",
).Build()

sa, err := client.Storage().BucketServiceAccounts().Create(ctx, serviceAccount)

// Get S3 credentials
credentials, err := client.Storage().GetS3Credentials(ctx, "my-bucket", "my-service-account")
fmt.Printf("Access Key: %s\n", credentials.AccessKeyID)
fmt.Printf("Secret Key: %s\n", credentials.SecretAccessKey)
fmt.Printf("Endpoint: %s\n", credentials.Endpoint)

// Use with MinIO SDK (see examples/vm-backup-to-storage for complete example)
```

## IAM

### Projects and Permission Sets

```go
// Create a project
project, err := client.IAM().Projects().Create(ctx,
    iam.NewProjectBuilder("dev-project", "organization-id").
        WithName("Development Environment").
        Build(),
)

// Create a permission set
permissions, err := client.IAM().PermissionSets().Create(ctx,
    iam.NewPermissionSetBuilder("developer-permissions", "project-id", "user@example.com").
        WithAdmin(false).
        Build(),
)
```
