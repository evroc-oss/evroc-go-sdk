# Changelog

All notable changes to the evroc Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.1] - 2026-06-17

### Fixed
- `RemovePublicIP()` now correctly clears the public IP reference from a VM — previously `omitempty` dropped the field from the PATCH payload, causing the API to silently ignore the removal

## [0.5.0] - 2026-05-28

### Added
- FileStore (NFS) support in the storage API — CRUD, builder, waiters, and example
- Service account JWT bearer authentication (`EVROC_SERVICE_ACCOUNT_ID` + `EVROC_SERVICE_ACCOUNT_SECRET`)

## [0.4.1] - 2026-05-05

### Fixed
- Fix LICENSE file typo preventing pkg.go.dev documentation display

## [0.4.0] - 2026-05-01
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

[Unreleased]: https://github.com/evroc-oss/evroc-go-sdk/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.4.0
[0.4.1]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.4.1
[0.5.1]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.5.1
