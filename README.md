<div align="center">

<img src="docs/images/evroc-logo.png" alt="evroc" width="200"/>

# evroc Go SDK

Official Go SDK for the evroc Cloud Platform

Go client for Compute, Networking, IAM, Storage, Quotas, and Think APIs

[![CI](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/ci.yml)
[![Release](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/release.yml/badge.svg)](https://github.com/evroc-oss/evroc-go-sdk/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/evroc-oss/evroc-go-sdk.svg)](https://pkg.go.dev/github.com/evroc-oss/evroc-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/evroc-oss/evroc-go-sdk)](https://goreportcard.com/report/github.com/evroc-oss/evroc-go-sdk)
[![Go Version](https://img.shields.io/github/go-mod-go-version/evroc-oss/evroc-go-sdk)](https://github.com/evroc-oss/evroc-go-sdk/blob/main/go.mod)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Security](https://img.shields.io/badge/Security-Signed%20%26%20Attested-green.svg)](https://github.com/evroc-oss/evroc-go-sdk/releases/latest)
[![SLSA](https://img.shields.io/badge/SLSA-Provenance-blue.svg)](https://slsa.dev/)
[![GitHub release](https://img.shields.io/github/release/evroc-oss/evroc-go-sdk.svg)](https://github.com/evroc-oss/evroc-go-sdk/releases/latest)

</div>

---

Automate your evroc cloud infrastructure with Go - provision VMs, configure networks, manage storage, and control access from your code.

Type-safe API with automatic retries and context support.

## Capabilities

| Service | What you can do |
|---------|----------------|
| **Compute** | Virtual machines, disks, placement groups, hotswap attachments |
| **Networking** | Public IPs, security groups, VPCs (read), subnets (read) |
| **Storage** | S3-compatible buckets and service accounts |
| **IAM** | Projects and permission sets |
| **Quotas** | Organization and project resource quotas (read-only) |
| **Think** | Dedicated GPU instances, AI models, API keys, shared models |

## Installation

```bash
go get github.com/evroc-oss/evroc-go-sdk
```

Requires Go 1.24+

## Authentication

The SDK supports two authentication methods:

### Method 1: User Authentication with evroc CLI (Recommended)

Best for interactive development and automation using your user credentials.

**Step 1:** Install the evroc CLI (see [docs.evroc.com/cli.html](https://docs.evroc.com/cli.html))

**Step 2:** Login using the CLI
```bash
evroc login
```

This creates a config file in your home directory with your credentials and default project/region settings.

**Step 3:** Use the SDK
```go
client, err := evroc.NewFromEnv(ctx)  // Reads CLI config automatically
```

**Alternative: Use login helper**

If you prefer not to install the full CLI, you can use the included login helper:

```bash
go run ./cmd/login

# Copy the tokens and set them as environment variables
export EVROC_TOKEN='...'
export EVROC_REFRESH_TOKEN='...'
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
```

Note: You can use just `EVROC_REFRESH_TOKEN` alone - the SDK will automatically obtain an access token.

### Method 2: Service Account Authentication

Best for production automation, CI/CD pipelines, and infrastructure-as-code.

Service accounts are non-interactive credentials designed for automated systems. They use the same authentication flow as user credentials.

**Step 1:** Contact evroc support (support@evroc.com) to obtain service account credentials

**Step 2:** Set environment variables with the service account credentials
```bash
export EVROC_USERNAME="service-account-id"
export EVROC_PASSWORD="service-account-secret"
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
```

**Step 3:** Use the SDK
```go
client, err := evroc.NewFromEnv(ctx)
```

**Finding your Project ID:** In [cloud.evroc.com](https://cloud.evroc.com), navigate to your project. The project ID is displayed under the project name. Alternatively, extract it from the URL:
```
https://cloud.evroc.com/?rgFullPath=%2F{org-id}%2F{project-id}
                                         ^^^^^^^    ^^^^^^^^^^^
```

## Quick Start

### Create Your First VM

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
    sg, err := client.Networking().SecurityGroups().Create(ctx,
        networking.NewSecurityGroupBuilder("my-sg").
            AllowSSH().      // Allow inbound SSH (port 22) so we can connect
            AllowAllEgress(). // Allow outbound internet access (for updates, etc.)
            Build(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create a boot disk with Ubuntu 24.04
    // This disk contains the operating system the VM will boot from
    disk, err := client.Compute().Disks().Create(ctx,
        compute.NewDiskBuilder("my-disk").
            WithImage(compute.DiskImageUbuntuMinimal2404).  // Ubuntu 24.04 minimal
            WithSizeGB(50).                                  // 50GB disk size
            WithZone("a").                                   // Zone is required for compute resources
            Build(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create the VM
    // FindVMInstanceType finds the right compute profile for our requirements
    profile := compute.FindVMInstanceType(2, 4, 0)  // 2 vCPUs, 4GB RAM, 0 GPUs
    if profile == "" {
		log.Fatal("could not find compute profile with 2 vCPUs and 4GB RAM")
    }

    vm, err := client.Compute().VirtualMachines().Create(ctx,
        compute.NewVirtualMachineBuilder("my-vm").
            WithBootDisk(disk.Ref()).                       // Use .Ref() for resources we just created
            WithVMInstanceType(profile).                    // Compute profile (e.g., a1a.xs)
            WithSecurityGroup(sg.Ref()).                    // Use .Ref() for resources we just created
            WithSSHKey("ssh-rsa AAAAB3NzaC1yc2EA...").     // Your SSH public key for authentication
            WithZone("a").                                  // Zone is required for compute resources
            Build(),
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Created VM: %s", *vm.Metadata.Name)
}
```

## Examples

| Example | Description |
|---------|-------------|
| [authentication](examples/authentication/) | Authentication methods: evroc CLI, environment variables, service accounts |
| [create-vm](examples/create-vm/) | Complete VM creation with public IP and waiters |
| [web-server](examples/web-server/) | Production web server with nginx and cloud-init |
| [k3s-cluster](examples/k3s-cluster/) | Kubernetes cluster across 3 availability zones |
| [vm-backup-to-storage](examples/vm-backup-to-storage/) | S3-compatible storage with file upload/download |
| [hotswap-disk](examples/hotswap-disk/) | Attach disks to running VMs without restart |
| [metrics](examples/metrics/) | Prometheus metrics integration |
| [compute](examples/compute/) | All Compute APIs (VMs, disks, placement groups) |
| [networking](examples/networking/) | All Networking APIs (public IPs, security groups) |
| [storage](examples/storage/) | All Storage APIs (buckets, service accounts, S3) |
| [iam](examples/iam/) | All IAM APIs (projects, permission sets) |
| [think-api-key](examples/think-api-key/) | Think API key management for shared models |
| [think-dedicated-models](examples/think-dedicated-models/) | Dedicated GPU instances for AI model serving |
| [labels](examples/labels/) | Resource labeling and filtering |
| [context-and-retries](examples/context-and-retries/) | Context cancellation, timeouts, and retry configuration |
| [storage-public-urls](examples/storage-public-urls/) | Presigned URLs for S3-compatible storage |

See [examples/README.md](examples/README.md) for full documentation.

## Documentation

**Go Package Documentation:** [pkg.go.dev/github.com/evroc-oss/evroc-go-sdk](https://pkg.go.dev/github.com/evroc-oss/evroc-go-sdk)

| Guide | Description |
|-------|-------------|
| [API Reference](docs/api-reference.md) | Complete API reference for all services |
| [Configuration](docs/configuration.md) | Authentication and configuration methods |
| [SDK Guide](docs/guide.md) | Async operations, waiters, context, and helpers |
| [Security Concepts](docs/vm-security.md) | Security groups, SSH keys, and network access |
| [Metrics](docs/metrics.md) | Prometheus metrics integration |
| [Testing](docs/testing.md) | Unit and E2E testing |

## Security

### Module Integrity Verification

Go modules are automatically verified via the Go checksum database when you install the SDK.

**Verify Go Module Checksums:**

```bash
# Go automatically verifies module checksums during installation
go get github.com/evroc-oss/evroc-go-sdk@v0.1.0

# Manual verification via Go checksum database
curl -s "https://sum.golang.org/lookup/github.com/evroc-oss/evroc-go-sdk@v0.1.0"
```

### Supply Chain Security

- **Pinned Dependencies**: All GitHub Actions are pinned to commit SHAs to prevent supply chain attacks
- **Go Module Checksums**: Module integrity verified via [sum.golang.org](https://sum.golang.org)
- **Automated Testing**: All releases pass comprehensive test suites including unit and E2E tests

### Reporting Security Issues

If you discover a security vulnerability, please contact security@evroc.com. Do not open public issues for security vulnerabilities.

## Contributing

This project does not accept external contributions. If you encounter a bug or
have a feature request, please report it through
[evroc support](mailto:support@evroc.com).

## Support

**Support level:** Best-effort — evroc will address issues as time permits, with
no guaranteed SLA. All issues should be reported through
[evroc support channels](mailto:support@evroc.com).

- **Support:** [support@evroc.com](mailto:support@evroc.com)
- **Documentation:** [github.com/evroc-oss/evroc-go-sdk/tree/main/docs](https://github.com/evroc-oss/evroc-go-sdk/tree/main/docs)
- **Examples:** [github.com/evroc-oss/evroc-go-sdk/tree/main/examples](https://github.com/evroc-oss/evroc-go-sdk/tree/main/examples)

## Versioning and Deprecation

This project follows [Semantic Versioning](https://semver.org/).

- Breaking changes only occur in major version releases.
- Deprecated features are announced at least one minor version in advance.
- Deprecation notices are documented in the [CHANGELOG](CHANGELOG.md) and may
  include runtime warnings.
- Security fixes are provided for the current major version and one prior major
  version.

If this project reaches end-of-life, the README will be updated with archived
status and a final release will be made.

## License

Apache License 2.0 - see [LICENSE](LICENSE) file
