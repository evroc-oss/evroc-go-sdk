# evroc SDK Examples

Examples demonstrating the evroc Go SDK, from quick start to complete API coverage.

## 📁 Examples

### Quick Start

- **`simple/`** - Minimal example showing basic SDK usage (list VMs)
- **`create-vm/`** - Step-by-step VM creation with disk, public IP, and SSH

### API Examples

- **`compute/`** - Complete Compute API coverage
  - Disks (boot, data, images, resizing)
  - Virtual Machines (creation, configuration, lifecycle)
  - Placement Groups (spread strategy, high availability)
  - Hotswap Disk Attachments (dynamic disk management)

- **`networking/`** - Complete Networking API coverage
  - Public IPs (allocation, assignment)
  - Security Groups (rules, ingress/egress, convenience methods)
  - VPCs and Subnets (read-only operations)

- **`storage/`** - Complete Storage API coverage
  - Buckets (retention modes, object locking, versioning)
  - Bucket Service Accounts (access management)

- **`iam/`** - Complete IAM API coverage
  - Projects (creation, labels, organization management)
  - Permission Sets (user access, admin privileges)

- **`labels/`** - Label filtering across all APIs
  - Creating labeled resources
  - Single and multi-label filtering
  - Label best practices

## 🚀 Running the Examples

### Prerequisites

**Required environment variables:**

```bash
export EVROC_PROJECT="your-project-uuid"
export EVROC_REGION="se-sto"
export EVROC_ORGANIZATION="your-org-uuid"  # Required for IAM examples
```

**Authentication (choose one):**

Option 1: Username and password
```bash
export EVROC_USERNAME="your-username@example.com"
export EVROC_PASSWORD="your-password"
```

Option 2: OAuth tokens (get these from `go run cmd/evroc-login/main.go`)
```bash
export EVROC_TOKEN='eyJhbGci...'
export EVROC_REFRESH_TOKEN='eyJhbGci...'
```

### Running Individual Examples

Each example is a standalone Go program:

```bash
# Compute examples
cd compute
go run main.go

# Networking examples
cd networking
go run main.go

# Storage examples
cd storage
go run main.go

# IAM examples (requires EVROC_ORGANIZATION)
cd iam
go run main.go

# Label filtering examples
cd labels
go run main.go
```

## 📚 What Each Example Demonstrates

### Compute Examples (`compute/main.go`)

**Disk Operations:**
- Creating boot disks from images
- Creating empty data disks
- Using convenience methods (CreateBootDisk, CreateDataDisk)
- Waiting for disks to be ready
- Listing and filtering disks
- Getting disk details
- Resizing disks

**Placement Groups:**
- Creating spread placement groups
- Listing placement groups
- Understanding placement strategies

**Virtual Machines:**
- Simple VMs with boot disk only
- Complex VMs with multiple disks
- VMs with cloud-init configuration
- VMs with placement groups
- Creating VMs in stopped state
- Starting and stopping VMs
- Checking VM readiness
- VM lifecycle management

**Hotswap Disk Attachments:**
- Attaching disks to running VMs
- Listing attachments
- Detaching disks

### Networking Examples (`networking/main.go`)

**Public IPs:**
- Creating public IPs
- Listing public IPs
- Checking IP readiness
- Extracting IP addresses

**Security Groups:**
- Basic security groups with SSH
- Web server security groups (HTTP, HTTPS)
- Custom port ranges
- Database security groups with restricted access
- Listing rules
- Deleting security groups

**VPCs and Subnets:**
- Listing VPCs
- Viewing VPC configuration
- Listing subnets
- Understanding subnet placement

### Storage Examples (`storage/main.go`)

**Buckets:**
- Simple buckets
- Buckets with versioning
- Buckets with object locking (GOVERNANCE, COMPLIANCE)
- Different retention modes
- Updating buckets
- Listing and filtering buckets

**Bucket Service Accounts:**
- Single bucket access
- Multiple bucket access
- Different ways to add buckets
- Updating service accounts

### IAM Examples (`iam/main.go`)

**Projects:**
- Simple projects
- Projects with display names
- Projects with labels
- Listing all projects
- Updating projects
- Organizational structure

**Permission Sets:**
- Standard user permissions
- Admin permissions
- Permission sets with labels
- Creating permissions for multiple users
- Listing permission sets
- Updating permissions

### Label Examples (`labels/main.go`)

**Creating Labeled Resources:**
- Environment-based labels (production, development)
- Team-based labels
- Purpose-based labels
- Organizational labels

**Filtering by Labels:**
- Single label filtering
- Multi-label filtering
- Filtering across different resource types
- Finding resources by team, environment, purpose

**Best Practices:**
- Recommended label patterns
- Common label keys and values
- Label schema documentation
- Use cases for labels

## 💡 Key Concepts Demonstrated

### Builder Pattern
All examples use the builder pattern for clean, readable code:

```go
disk := compute.NewDiskBuilder("my-disk").
    WithImage("ubuntu.24-04.1").
    WithSizeGB(100).
    WithZone("a").
    WithLabels(map[string]string{"env": "prod"}).
    Build()
```

### Resource Lifecycle
- Creating resources
- Waiting for readiness
- Listing and filtering
- Getting specific resources
- Updating resources
- Deleting resources

### Labels and Organization
- Consistent labeling across resources
- Filtering by labels
- Multi-label queries
- Best practices for resource organization

### Error Handling
- Proper error checking
- Graceful degradation
- Informative error messages

### Waiters
- Using built-in waiters for resource readiness
- Timeouts and polling
- Checking resource status

## 🎯 Best Practices Shown

1. **Resource Naming**: Consistent, descriptive names
2. **Labels**: Standardized label keys and values
3. **Error Handling**: Always check and handle errors
4. **Waiters**: Use waiters for asynchronous operations
5. **Security**: Proper security group configuration
6. **High Availability**: Placement groups for resilience
7. **Storage**: Appropriate retention and locking policies
8. **Organization**: Clear resource grouping and management

## 📖 Additional Resources

- [Main SDK README](../README.md) - SDK overview and quick start
- [API Documentation](https://docs.evroc.com) - Complete API reference

## 🤝 Contributing

Found an issue or want to add more examples? Please open an issue or pull request!

## 📄 License

MIT License - See [LICENSE](../LICENSE) for details
