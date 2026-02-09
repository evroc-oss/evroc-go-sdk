# VM Backup to Object Storage Example

This example demonstrates using the Evroc SDK to:
1. Create an S3-compatible bucket
2. Create a service account with S3 credentials
3. Use the credentials to upload/download files via S3 API (using MinIO Go SDK)
4. Demonstrate object versioning
5. Clean up all resources

## Running

This example uses a separate Go module to keep MinIO SDK out of the main SDK dependencies.

```bash
# From repo root:
go run examples/vm-backup-to-storage/main.go

# Or from the example directory:
cd examples/vm-backup-to-storage
go run main.go
```

## What it does

- Creates a bucket and service account via Evroc SDK
- Retrieves S3 credentials programmatically (shows full credentials in output)
- Uploads a 20MB file and measures upload speed (MB/s)
- Downloads the file and measures download speed (MB/s)
- Verifies data integrity
- Uploads a second version of the same file (different content)
- Lists all object versions to demonstrate versioning
- Deletes all objects from bucket
- Cleans up service account and bucket

## Dependencies

This example has its own `go.mod` to isolate the MinIO SDK dependency from the main SDK.
The root `go.work` file manages both modules together.
