<div align="center">

<img src="docs/images/evroc-logo.png" alt="evroc" width="200"/>

# evroc Go SDK

Official Go SDK for the evroc Cloud Platform

Go client for Compute, Networking, IAM, and Storage APIs

[![CI](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/ci.yml)
[![Release](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/release.yml/badge.svg)](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/evroc-oss/evroc-go-sdk.svg)](https://pkg.go.dev/github.com/evroc-oss/evroc-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/evroc-oss/evroc-go-sdk.svg)](https://github.com/evroc-oss/evroc-go-sdk/releases/latest)

</div>

---

Automate your evroc cloud infrastructure with Go - provision VMs, configure networks, manage storage, and control access from your code.

## Capabilities

| Service | What you can do |
|---------|----------------|
| **Compute** | Virtual machines, disks, placement groups, hotswap attachments |
| **Networking** | Public IPs, security groups, VPCs (read), subnets (read) |
| **Storage** | S3-compatible buckets and service accounts |
| **IAM** | Projects and permission sets |

## Installation

```bash
go get github.com/evroc-oss/evroc-go-sdk
```

Requires Go 1.24+

## Quick Start

### 1. Authenticate

```bash
# Run evroc-login (included in this SDK)
go run ./cmd/evroc-login

# Copy the output:
# export EVROC_TOKEN='...'
# export EVROC_REFRESH_TOKEN='...'
```

### 2. Set Environment Variables

```bash
# Paste the tokens from evroc-login
export EVROC_TOKEN='your-access-token'
export EVROC_REFRESH_TOKEN='your-refresh-token'

# Add your project and region
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
```

**Finding your Project ID:** Go to [console.evroc.com](https://console.evroc.com), select your project, and extract the UUID from the URL:
```
https://console.evroc.com/?rgFullPath=%2F{org-id}%2F{project-id}
                                          ^^^^^^^         ^^^^^^^^^^^
```

### 3. Create Your First VM

```go
package main

import (
    "context"
    "log"

    evroc "github.com/evroc-oss/evroc-go-sdk"
    "github.com/evroc-oss/evroc-go-sdk/compute"
    "github.com/evroc-oss/evroc-go-sdk/networking"
)

func main() {
    ctx := context.Background()

    // Initialize client from environment variables
    client, err := evroc.NewFromEnv(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Create security group to allow network access
    // Without this, the VM has no inbound or outbound connectivity
    sg, _ := client.Networking().SecurityGroups().Create(ctx,
        networking.NewSecurityGroupBuilder("my-sg").
            AllowSSH().      // Allow inbound SSH (port 22) so we can connect
            AllowEgress().   // Allow outbound internet access (for updates, etc.)
            Build(),
    )

    // Create a boot disk with Ubuntu 24.04
    // This disk contains the operating system the VM will boot from
    disk, _ := client.Compute().Disks().Create(ctx,
        compute.NewDiskBuilder("my-disk").
            WithImage(compute.DiskImageUbuntuMinimal2404).  // Ubuntu 24.04 minimal
            WithSizeGB(50).                                  // 50GB disk size
            Build(),
    )

    // Create the VM
    // FindVMInstanceType finds the right compute profile for our requirements
    profile := compute.FindVMInstanceType(2, 4, 0)  // 2 vCPUs, 4GB RAM, 0 GPUs
    vm, _ := client.Compute().VirtualMachines().Create(ctx,
        compute.NewVirtualMachineBuilder("my-vm").
            WithBootDisk("my-disk").                        // Boot from our Ubuntu disk
            WithVMInstanceType(profile).                    // Compute profile (e.g., a1a.xs)
            WithSecurityGroup("my-sg").                     // Network access rules
            WithSSHKey("ssh-rsa AAAAB3NzaC1yc2EA...").     // Your SSH public key for authentication
            Build(),
    )

    log.Printf("Created VM: %s", *vm.Metadata.Name)
}
```

## Examples

Complete, runnable examples in `examples/`:

- **[create-vm](examples/create-vm/)** - Complete VM creation (based on the example above, with public IP and waiters)
- **[web-server](examples/web-server/)** - Production web server with nginx and cloud-init
- **[k3s-cluster](examples/k3s-cluster/)** - Kubernetes cluster across 3 availability zones
- **[vm-backup-to-storage](examples/vm-backup-to-storage/)** - S3-compatible storage with file upload/download
- **[hotswap-disk](examples/hotswap-disk/)** - Attach disks to running VMs without restart
- **[compute](examples/compute/)** - All Compute APIs (VMs, disks, placement groups)
- **[networking](examples/networking/)** - All Networking APIs (public IPs, security groups)
- **[storage](examples/storage/)** - All Storage APIs (buckets, service accounts, S3)
- **[iam](examples/iam/)** - All IAM APIs (projects, permission sets)

See [examples/README.md](examples/README.md) for full documentation.

## Documentation

- **[API Reference](docs/api-reference.md)** - Detailed API usage for all services
- **[Configuration](docs/configuration.md)** - Environment variables, config files, programmatic setup
- **[Asynchronous Operations](docs/async-operations.md)** - Waiters, polling, and timeouts
- **[Security Concepts](docs/security.md)** - Security groups, SSH keys, and network access
- **[Helper Functions](docs/helpers.md)** - Convenience functions for common tasks
- **[Common Operations](docs/common-operations.md)** - Frequently used patterns

**Go Package Documentation:** [pkg.go.dev/github.com/evroc-oss/evroc-go-sdk](https://pkg.go.dev/github.com/evroc-oss/evroc-go-sdk)

## Features

- **Builder Pattern** - Fluent API for creating resources
- **Async Waiters** - Wait for operations to complete with timeouts
- **Auto Retry** - Automatic retry with exponential backoff
- **Context Support** - Full context support for timeouts and cancellation
- **Type Safety** - Strongly typed API with compile-time checks

## License

MIT License - see [LICENSE](LICENSE) file

## Support

- **Issues:** [github.com/evroc-oss/evroc-go-sdk/issues](https://github.com/evroc-oss/evroc-go-sdk/issues)
- **Documentation:** [docs/](docs/)
- **Examples:** [examples/](examples/)
