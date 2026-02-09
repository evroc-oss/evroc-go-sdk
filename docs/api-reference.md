# API Reference

Complete reference for all evroc SDK APIs.

## Compute

### Virtual Machines

```go
// Simple VM with builder
vm := compute.NewVirtualMachineBuilder("my-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("c1a.m").
    WithSSHKey("ssh-rsa AAAA...").
    Build()

createdVM, err := client.Compute().VirtualMachines().Create(ctx, vm)

// VM with full configuration
vm = compute.NewVirtualMachineBuilder("production-vm").
    WithBootDisk("boot-disk").
    WithDataDisk("data-disk-1").
    WithDataDisk("data-disk-2").
    WithVMInstanceType("m1a.xl").
    WithSSHKey(sshPublicKey).
    WithCloudInit(cloudInitScript).
    WithSecurityGroup("web-servers").
    WithPublicIP("my-public-ip").
    WithZone("zone-1").
    Build()
```

### Disks

```go
// Create a 100GB disk with Ubuntu 24.04 image
disk := compute.NewDiskBuilder("my-disk").
    WithImage(string(compute.DiskImageUbuntu2404)).
    WithSizeGB(100).  // Disk capacity in gigabytes
    Build()

createdDisk, err := client.Compute().Disks().Create(ctx, disk)

// Create an empty 500GB data disk (no OS image)
dataDisk := compute.NewDiskBuilder("data-disk").
    WithSizeGB(500).  // Just size, no image
    Build()
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
        WithDisplayName("Development Environment").
        Build(),
)

// Create a permission set
permissions, err := client.IAM().PermissionSets().Create(ctx,
    iam.NewPermissionSetBuilder("developer-permissions", "project-id", "user@example.com").
        WithAdmin(false).
        Build(),
)
```
