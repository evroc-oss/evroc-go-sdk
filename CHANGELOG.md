# Changelog

All notable changes to the evroc Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.14] - 2026-04-27
Public release

### Added

#### Core Features
- OAuth2/OIDC authentication with automatic token refresh
- Configuration via environment variables, YAML file, or evroc CLI config
- Context support for cancellation and timeouts
- Automatic retry logic with exponential backoff
- Prometheus metrics integration
- Type-safe API clients generated from OpenAPI specifications
- Go module integrity verification via checksum database
- Automated pkg.go.dev indexing

#### Compute API
- VirtualMachines - Create, read, update, delete VMs
- Disks - Manage persistent storage volumes
- PlacementGroups - Control VM placement for high availability
- HotswapDiskAttachments - Attach/detach disks without VM restart

#### Networking API
- VirtualPrivateClouds - Read VPC configurations
- Subnets - Read subnet configurations
- SecurityGroups - Manage firewall rules with builder pattern
- PublicIPs - Allocate and manage public IP addresses

#### IAM API
- PermissionSets - Manage access control policies
- Projects - Manage project resources

#### Storage API
- Buckets - S3-compatible object storage with versioning and locking
- BucketServiceAccounts - Service account credentials for S3 access
- Presigned URL support for browser uploads/downloads
- Waiter utilities for credential availability

#### Developer Experience
- Builder pattern for resource creation with fluent APIs
- Waiter utilities for async operations (WaitForReady, WaitForDeleted)
- Comprehensive examples with inline documentation
- End-to-end test suite for all major resources
- Supply chain security documentation

[Unreleased]: https://github.com/evroc-oss/evroc-go-sdk/compare/v0.2.10...HEAD
[0.2.10]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.2.10
[0.2.11]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.2.11
[0.2.12]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.2.12
[0.2.13]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.2.13
[0.2.14]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.2.14
