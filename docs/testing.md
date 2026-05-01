# Testing

## Unit Tests

Run without credentials using mock servers:

```bash
go test -v ./... -short
```

## E2E Tests

Test against the real evroc API. Creates actual resources.

### Prerequisites

Set credentials using one of:

```bash
# Option 1: API Token (recommended)
export EVROC_TOKEN="your-token"
export EVROC_PROJECT="project-id"
export EVROC_REGION="se-sto"

# Option 2: Refresh Token
export EVROC_REFRESH_TOKEN="your-token"
export EVROC_PROJECT="project-id"

# Option 3: Username/Password
export EVROC_USERNAME="user"
export EVROC_PASSWORD="pass"
export EVROC_PROJECT="project-id"
```

### Run Tests

```bash
# By service
E2E=1 go test -v -tags=e2e ./compute -timeout 60m
E2E=1 go test -v -tags=e2e ./storage -timeout 60m
E2E=1 go test -v -tags=e2e ./networking -timeout 60m

# Single test
E2E=1 go test -v -tags=e2e ./compute -run TestE2E_Disk_Lifecycle -timeout 60m
```

### Important Notes

- E2E tests create real resources that may incur costs
- Always use a dedicated test project, never production
- Tests attempt cleanup but failures may leave orphaned resources
- All test resources are prefixed with `e2e-test-`
- Default timeout is 120 minutes

### Writing E2E Tests

```go
//go:build e2e

package yourpackage_test

import (
	"context"
	"testing"
	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
)

func TestE2E_Resource_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	name := e2etest.RandomName("resource")

	// Create
	resource, err := client.Service().Create(ctx, ...)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Cleanup
	defer func() {
		if err := client.Service().Delete(ctx, resource.Metadata.Id); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	}()

	// Test assertions...
}
```

## Coverage

- Compute: Disks, VMs, Placement Groups
- Storage: Buckets, Service Accounts
- Networking: Public IPs, Security Groups
