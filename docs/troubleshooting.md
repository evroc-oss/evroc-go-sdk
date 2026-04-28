# Troubleshooting Guide

## Enable Debug Logging

Set `EVROC_SDK_DEBUG=1` to see detailed HTTP requests/responses:

```bash
EVROC_SDK_DEBUG=1 go run main.go
```

## Common Issues

### Authentication Failures

**401 Unauthorized** - Check credentials in config.yaml or environment variables:
- Verify username/password or refresh_token is correct
- Use `cmd/login` to obtain a fresh refresh token

### Resource Not Found

**404 Not Found** - Verify:
- Project ID and region are correct in config.yaml
- Resource exists: `client.Compute().VirtualMachines().Get(ctx, "vm-id")`
- Resource references use correct format (short ID or FQID)

### SSH Connection Issues

**Cannot connect to VM** - Check:
- VM has a public IP attached
- Security group allows SSH (port 22) ingress
- Security group allows egress traffic
- SSH key was added during VM creation (cannot be changed later!)
- Using correct username (default: `evroc-user`, or `ubuntu` for custom cloud-init)

### Timeout Errors

**Context deadline exceeded** - Increase timeout or check resource state:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()
```

Use `WaitForReady()` for resources that need provisioning:
```go
disk, err := client.Compute().Disks().WaitForReady(ctx, "disk-id", 5*time.Minute)
```

### Validation Errors

**400 Bad Request** - Common mistakes:
- Zone not specified (required for most resources)
- Invalid resource references (use `resource.Ref()` or `client.Service().ResourceRef()`)
- Missing required fields (boot disk, VM instance type, etc.)
- Custom cloud-init without SSH keys configured

### Cloud-Init Issues

**VM created but cannot SSH** - If using custom cloud-init:
- Ensure SSH keys are in your cloud-init script under `ssh_authorized_keys`
- `WithSSHKey()` is ignored when `WithCloudInit()` is set
- `evroc-user` is NOT created (use distro default like `ubuntu` or create your own user)
