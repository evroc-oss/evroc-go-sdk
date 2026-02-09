# Important evroc Security Concepts

## VMs Have Restricted Networking by Default

**All VMs in evroc start with zero network access** - no inbound or outbound traffic is allowed. You must explicitly configure security groups to enable connectivity.

```go
// ❌ This VM will be unreachable - no network access!
vm := compute.NewVirtualMachineBuilder("isolated-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    Build()

// ✅ Proper VM with SSH access
sg := networking.NewSecurityGroupBuilder("ssh-access").
    AllowSSH().      // Allows inbound SSH (port 22)
    AllowEgress().   // Allows all outbound traffic
    Build()

vm := compute.NewVirtualMachineBuilder("accessible-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    WithSecurityGroup("ssh-access").  // Attach security group
    Build()
```

**Critical security group concepts:**

- **Ingress rules** - Control inbound traffic (SSH, HTTP, HTTPS, custom ports)
- **Egress rules** - Control outbound traffic (package updates, API calls, etc.)
- **Both must be explicit** - If you need SSH access AND internet access, you need both ingress and egress rules

Common security group patterns:

```go
// Web server: SSH + HTTP + HTTPS + internet access
sg := networking.NewSecurityGroupBuilder("web-server").
    AllowSSH().       // Inbound SSH from anywhere
    AllowHTTP().      // Inbound HTTP from anywhere
    AllowHTTPS().     // Inbound HTTPS from anywhere
    AllowEgress().    // Outbound to internet (for package updates, etc.)
    Build()

// Private worker: No inbound, only outbound for API calls
sg := networking.NewSecurityGroupBuilder("worker").
    AllowEgress().    // Only outbound traffic
    Build()

// Database: SSH access + custom port from specific subnet
sg := networking.NewSecurityGroupBuilder("database").
    AllowSSH().
    AllowIngressRule("postgres", networking.ProtocolTCP, 5432, 0, "10.0.0.0/24").
    AllowEgress().
    Build()
```

## SSH Keys Are Required for Access

**Without an SSH key, you cannot log into your VM.** Username/password authentication is not supported.

⚠️ **IMPORTANT: SSH keys cannot be changed after VM creation.** You must add all necessary SSH keys during creation. If you lose access, you'll need to destroy and recreate the VM.

```go
// ❌ This VM is inaccessible - no SSH key provided!
vm := compute.NewVirtualMachineBuilder("locked-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    WithSecurityGroup("ssh-access").
    Build()

// ✅ Proper VM with SSH key for authentication
sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
vm := compute.NewVirtualMachineBuilder("accessible-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    WithSecurityGroup("ssh-access").
    WithSSHKey(sshPublicKey).  // Add your public key
    Build()
```

**Best practice:** Add multiple SSH keys during creation for redundancy and team access:

```go
vm := compute.NewVirtualMachineBuilder("team-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    WithSSHKey(devKey1).    // Primary key
    WithSSHKey(devKey2).    // Backup key
    WithSSHKey(devKey3).    // Team member key
    Build()
// SSH keys are immutable - add all you'll need now!
```

## Public IP Addresses and Placement Groups

You can attach public IPs and configure placement groups during VM creation, or change them later when the VM is stopped.

### At Creation Time

```go
// Create the public IP resource
publicIP, err := client.Networking().PublicIPs().Create(ctx,
    networking.NewPublicIPBuilder("my-public-ip").Build(),
)

// Create placement group
placementGroup, err := client.Compute().PlacementGroups().Create(ctx,
    compute.NewPlacementGroupBuilder("my-pg", "spread").
        WithZone("se-sto-1a").
        Build(),
)

// Create VM with public IP and placement group
vm, err := client.Compute().VirtualMachines().Create(ctx,
    compute.NewVirtualMachineBuilder("my-vm").
        WithBootDisk("my-disk").
        WithVMInstanceType("a1a.xs").
        WithSecurityGroup("my-sg").
        WithSSHKey(sshPublicKey).
        WithPublicIP("my-public-ip").
        WithPlacementGroup("my-pg").
        Build(),
)
```

### Changing When VM is Stopped

⚠️ **IMPORTANT: You can only change public IPs and placement groups when the VM is stopped.**

```go
// Stop the VM first
_, err := compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Stop().
    Apply(ctx)

// Wait for VM to be stopped
err = client.Compute().VirtualMachines().WaitForStopped(ctx, "my-vm", 5*time.Minute)

// Change public IP
_, err = compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    SetPublicIP("new-public-ip").  // Or use RemovePublicIP() to remove it
    Apply(ctx)

// Change placement group
_, err = compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    SetPlacementGroup("new-placement-group").  // Or use RemovePlacementGroup() to remove it
    Apply(ctx)

// Start the VM again
_, err = compute.UpdateVM("my-vm", client.Compute().VirtualMachines()).
    Start().
    Apply(ctx)
```

## Cloud-Init Customization

By default, evroc VMs run a standard cloud-init script that configures SSH keys and basic system setup. **If you provide a custom cloud-init script, you are responsible for all VM initialization.**

```go
// Default behavior - evroc handles SSH key setup
vm := compute.NewVirtualMachineBuilder("standard-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    WithSSHKey(sshPublicKey).  // evroc's cloud-init will configure this
    Build()

// Custom cloud-init - you must handle SSH keys yourself!
customCloudInit := `#cloud-config
users:
  - name: admin
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2E...
    sudo: ALL=(ALL) NOPASSWD:ALL
packages:
  - nginx
  - postgresql
runcmd:
  - systemctl start nginx
`

vm := compute.NewVirtualMachineBuilder("custom-vm").
    WithBootDisk("my-disk").
    WithVMInstanceType("a1a.xs").
    WithCloudInit(customCloudInit).  // You're now responsible for everything
    Build()
```

**When using custom cloud-init:**
- You must configure SSH keys manually in your script
- The `WithSSHKey()` builder method is ignored
- You're responsible for user creation, package installation, and all setup
- Errors in your cloud-init script can make the VM unusable

## Complete Secure VM Example

Here's a complete example showing all security concepts:

```go
// 1. Create security group with proper rules
sg, err := client.Networking().SecurityGroups().Create(ctx,
    networking.NewSecurityGroupBuilder("web-app-sg").
        AllowSSH().       // SSH access for management
        AllowHTTP().      // Web traffic
        AllowHTTPS().     // Secure web traffic
        AllowEgress().    // Outbound for updates and API calls
        Build(),
)

// 2. Create disk
disk, err := client.Compute().Disks().Create(ctx,
    compute.NewDiskBuilder("web-app-disk").
        WithImage(compute.DiskImageUbuntuMinimal2404).
        WithSizeGB(50).
        Build(),
)

// 3. Create public IP (must be done before VM creation!)
publicIP, err := client.Networking().PublicIPs().Create(ctx,
    networking.NewPublicIPBuilder("web-app-ip").Build(),
)

// 4. Wait for disk to be ready
err = client.Compute().Disks().WaitForReady(ctx, "web-app-disk", 5*time.Minute)

// 5. Create VM with security group, SSH key, and public IP
sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
vm, err := client.Compute().VirtualMachines().Create(ctx,
    compute.NewVirtualMachineBuilder("web-app-vm").
        WithBootDisk("web-app-disk").
        WithVMInstanceType("c1a.m").
        WithSecurityGroup("web-app-sg").  // Network access
        WithSSHKey(sshPublicKey).          // Authentication (immutable!)
        WithPublicIP("web-app-ip").        // Public IP (must attach at creation!)
        Build(),
)

// Now you can SSH to the VM: ssh ubuntu@<public-ip-address>
```

## Example with Custom Cloud-Init

Here's an example showing custom cloud-init with nginx installation and SSH key configuration:

```go
// 1. Create a 50GB disk with Ubuntu 24.04
disk, err := client.Compute().Disks().Create(ctx,
    compute.NewDiskBuilder("custom-vm-disk").
        WithImage(compute.DiskImageUbuntuMinimal2404).
        WithSizeGB(50).  // 50 gigabytes
        Build(),
)
if err != nil {
    log.Fatal(err)
}

// 2. Wait for disk to be ready
err = client.Compute().Disks().WaitForReady(ctx, "custom-vm-disk", 5*time.Minute)
if err != nil {
    log.Fatal(err)
}

// 3. Create security group
sg, err := client.Networking().SecurityGroups().Create(ctx,
    networking.NewSecurityGroupBuilder("web-sg").
        AllowSSH().
        AllowHTTP().
        AllowHTTPS().
        AllowEgress().
        Build(),
)
if err != nil {
    log.Fatal(err)
}

// 4. Custom cloud-init script - YOU are responsible for SSH key setup!
customCloudInit := `#cloud-config
users:
  - name: ubuntu
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... your-key-here
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD... backup-key-here
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash

packages:
  - nginx
  - curl
  - git

write_files:
  - path: /var/www/html/index.html
    content: |
      <h1>Welcome to my evroc VM</h1>
      <p>Deployed with custom cloud-init</p>

runcmd:
  - systemctl enable nginx
  - systemctl start nginx
  - ufw allow 'Nginx Full'
`

// 5. Create VM with custom cloud-init
vm, err := client.Compute().VirtualMachines().Create(ctx,
    compute.NewVirtualMachineBuilder("custom-vm").
        WithBootDisk("custom-vm-disk").       // Using the 50GB disk we created
        WithVMInstanceType("a1a.m").          // 4 vCPUs, 16GB RAM
        WithSecurityGroup("web-sg").
        WithCloudInit(customCloudInit).       // Using custom cloud-init
        // Note: WithSSHKey() is ignored when custom cloud-init is provided
        // You MUST configure SSH keys in your cloud-init script!
        Build(),
)
if err != nil {
    log.Fatal(err)
}
```

**Important notes about custom cloud-init:**
- You MUST include SSH keys in your cloud-init script under `ssh_authorized_keys`
- The `WithSSHKey()` builder method is completely ignored
- If your cloud-init script has errors, the VM may be unusable
- Always test your cloud-init scripts thoroughly
- Include multiple SSH keys for redundancy (you can't change them later!)
