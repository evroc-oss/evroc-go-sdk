# Changelog

All notable changes to the evroc Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.2] - 2026-07-16

### Added
- **IAM RoleBindings**: full CRUD for project-scoped role bindings (Create, Get, List, Delete, Patch)
- **IAM RoleBindings**: full CRUD for organization-scoped role bindings
- **IAM RoleBindings**: organization-scoped assign/revoke (`AssignOrgRole`, `RevokeOrgRole`)
- **IAM Access Evaluation**: `TestPermissions` (check caller's own permissions) and `CheckAccess` (check another principal's permissions)
- **IAM Caller Bindings**: `ListMyRoleBindings` returns all role bindings for the authenticated caller
- **IAM Roles**: `RoleKubernetesCSIAgent` and `RoleKubernetesCCMAgent` role constants
- **IAM ServiceAccountCredentials**: `HmacSigv4` credential type for S3 SigV4 authentication
- **REST Client**: `OrgScopedCollectionPath` and `OrgScopedResourcePath` path builders

### Breaking Changes
- **IAM Types**: regenerated from OpenAPI spec — `RoleInfo.ID` is now `RoleInfo.Id`, `RoleInfo.Description` is now `*string`, `RoleInfo.Scope` is now `RoleInfoScope` (typed string)
- **IAM Types**: `RoleList` is now an alias for `RoleListResponse` with `Items *[]RoleInfo` (pointer to slice)
- **IAM Types**: `AssignRoleRequest.Resources` is now `*[]string` (pointer to slice)
- **IAM Types**: `AssignProjectRole` and `RevokeProjectRole` now return `*Rolebinding` instead of `*RoleBindingResponse`
- **IAM Types**: `RoleBindingResponse` and `RoleBindingEntry` removed — replaced by generated `Rolebinding`, `RolebindingSpec`, `RoleEntry`

## [0.7.1] - 2026-07-10

### Fixed
- **LoadBalancer BackendService**: `status.backends` is now correctly typed as `[]BackendserviceStatusBackendsItem` (was `int`)

## [0.7.0] - 2026-07-08

### Added
- **IAM ServiceAccounts**: create, get, list, patch, delete service accounts
- **IAM ServiceAccountCredentials**: create, get, list, delete credentials (private key returned only at creation)
- **LoadBalancer BackendService**: `ipProtocolSelection` field (IPv4/IPv6)
- **Compute Snapshots**: Patch operation
- **Storage BucketServiceAccountSecrets**: migrated from handwritten to codegen
- New example: `service-account-lifecycle` — full SA + credential creation flow
- E2E test for IAM ServiceAccount/Credential lifecycle

### Breaking Changes
- `ServiceAccountCredentials()` now requires a `serviceAccountID` parameter
- Removed `BackendserviceSpecHealthCheck.Https` and related types
- Removed `VirtualMachineSpec.LoadBalancerMemberships`

## [0.6.0] - 2026-06-18

### Added
- **LoadBalancer API** (v1alpha1): L4 load balancers, backend pools/services, L4 routes — builders, waiters, helpers
- **Snapshots**: create disks from snapshots, full CRUD, waiters
- **VPC and Subnet CRUD**: create, delete, patch — previously read-only
- VM dual-stack networking: `WithDualStack()`, `WithSubnet()`
- Auto-default `vpcRef` on SecurityGroup and `subnetRef` on VM create
- Label filtering helpers (`filter` package)
- Update builders for IAM, Think, HotswapDiskAttachment

### Changed
- Compute and Networking APIs upgraded from v1beta1 to v1beta2

### Fixed
- `RemovePublicIP()` now correctly clears the public IP reference from a VM

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
[0.7.0]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.7.0
[0.7.2]: https://github.com/evroc-oss/evroc-go-sdk/releases/tag/v0.7.2
