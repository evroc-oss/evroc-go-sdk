# Examples

Complete examples demonstrating the evroc Go SDK. Examples are self-contained `main.go` programs with inline usage notes.

## Available Examples

| Example | Description |
|---------|-------------|
| [authentication](authentication/) | Authentication methods |
| [create-vm](create-vm/) | VM creation with disk, public IP, and SSH |
| [web-server](web-server/) | Web server with nginx and cloud-init |
| [k3s-cluster](k3s-cluster/) | Kubernetes cluster across 3 availability zones |
| [vm-backup-to-storage](vm-backup-to-storage/) | S3-compatible storage integration |
| [hotswap-disk](hotswap-disk/) | Dynamic disk attachment |
| [metrics](metrics/) | Prometheus metrics integration |
| [compute](compute/) | Complete Compute API coverage |
| [networking](networking/) | Complete Networking API coverage |
| [storage](storage/) | Complete Storage API coverage |
| [iam](iam/) | Complete IAM API coverage |
| [labels](labels/) | Label filtering across APIs |
| [context-and-retries](context-and-retries/) | Context usage and retry configuration |

## Running Examples

Each example can be run directly:

```bash
cd <example-directory>
go run main.go
```

See each example's source comments for specific requirements and configuration.
