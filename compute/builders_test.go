// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"testing"
)

func TestDiskBuilder(t *testing.T) {
	req := NewDiskBuilder("test-disk").
		WithImage("ubuntu.24-04.1").
		WithSizeGB(100).
		WithSize(50, DiskSizeUnitGB).
		WithZone("a").
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	// Validate basic fields
	if req.Kind != "Disk" {
		t.Errorf("Expected Kind 'Disk', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-disk" {
		t.Errorf("Expected Id 'test-disk', got %s", req.Metadata.Id)
	}

	// Validate disk source
	if req.Spec.Source == nil {
		t.Error("Source should not be nil")
	} else {
		if req.Spec.Source.Type != "image" {
			t.Errorf("Expected source type 'image', got %s", req.Spec.Source.Type)
		}
		if req.Spec.Source.DiskImageRef == nil || *req.Spec.Source.DiskImageRef != "/compute/global/diskImages/evroc/ubuntu.24-04.1" {
			t.Errorf("Expected image ref '/compute/global/diskImages/evroc/ubuntu.24-04.1', got %v", req.Spec.Source.DiskImageRef)
		}
	}

	// Validate disk size (WithSize should override WithSizeGB)
	if req.Spec.DiskSize == nil {
		t.Error("DiskSize should not be nil")
	} else {
		if req.Spec.DiskSize.Amount != 50 {
			t.Errorf("Expected size 50, got %d", req.Spec.DiskSize.Amount)
		}
		if req.Spec.DiskSize.Unit != "GB" {
			t.Errorf("Expected unit 'GB', got %s", req.Spec.DiskSize.Unit)
		}
	}

	// Validate zone (Placement is now non-pointer)
	if req.Spec.Placement.Zone == nil {
		t.Error("Zone should not be nil")
	} else if *req.Spec.Placement.Zone != "a" {
		t.Errorf("Expected zone 'a', got %s", *req.Spec.Placement.Zone)
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else if (*req.Metadata.UserLabels)["env"] != "prod" {
		t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
	}
}

func TestVirtualMachineBuilder(t *testing.T) {
	req := NewVirtualMachineBuilder("test-vm").
		WithBootDisk(DiskRef("/compute/projects/p/regions/r/disks/boot-disk")).
		WithDataDisk(DiskRef("/compute/projects/p/regions/r/disks/data-disk")).
		WithSize("c1a.m").
		WithVMInstanceType("c1a.l").
		WithSSHKey("ssh-ed25519 KEY1").
		WithSSHKey("ssh-rsa KEY2").
		WithSecurityGroup(SecurityGroupRef("/networking/projects/p/regions/r/securityGroups/sg-1")).
		WithSecurityGroup(SecurityGroupRef("/networking/projects/p/regions/r/securityGroups/sg-2")).
		WithPublicIP(PublicIPRef("/networking/projects/p/regions/r/publicIPs/public-ip")).
		WithCloudInit("#cloud-config\npackages:\n  - nginx").
		WithZone("a").
		WithPlacementGroup(PlacementGroupRef("/compute/projects/p/regions/r/placementGroups/pg")).
		WithRunning(true).
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	// Validate basic fields
	if req.Kind != "VirtualMachine" {
		t.Errorf("Expected Kind 'VirtualMachine', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-vm" {
		t.Errorf("Expected Id 'test-vm', got %s", req.Metadata.Id)
	}

	// Validate VM instance type (WithVMInstanceType should override WithSize)
	if req.Spec.ComputeProfileRef != "c1a.l" {
		t.Errorf("Expected instance type 'c1a.l', got %s", req.Spec.ComputeProfileRef)
	}

	// Validate running state
	if req.Spec.Running == nil {
		t.Error("Running should not be nil")
	} else if *req.Spec.Running != true {
		t.Error("Expected running=true")
	}

	// Validate disks
	if req.Spec.Disks == nil || len(*req.Spec.Disks) != 2 {
		t.Errorf("Expected 2 disks, got %v", req.Spec.Disks)
	} else {
		disks := *req.Spec.Disks
		// First disk should be boot disk with FQID
		if disks[0].DiskRef != "/compute/projects/p/regions/r/disks/boot-disk" {
			t.Errorf("Expected boot disk FQID, got %s", disks[0].DiskRef)
		}
		if disks[0].BootFrom == nil || *disks[0].BootFrom != true {
			t.Error("First disk should have BootFrom=true")
		}
		// Second disk should be data disk with FQID
		if disks[1].DiskRef != "/compute/projects/p/regions/r/disks/data-disk" {
			t.Errorf("Expected data disk FQID, got %s", disks[1].DiskRef)
		}
		if disks[1].BootFrom != nil && *disks[1].BootFrom != false {
			t.Error("Data disk should not have BootFrom=true")
		}
	}

	// Validate networking (value type in v1beta2, always present)
	{
		// Validate public IP
		if req.Spec.Networking.PublicIPv4Address == nil {
			t.Error("PublicIPv4Address should not be nil")
		} else if req.Spec.Networking.PublicIPv4Address.Static == nil {
			t.Error("PublicIPv4Address.Static should not be nil")
		} else if req.Spec.Networking.PublicIPv4Address.Static.PublicIPRef == nil {
			t.Error("PublicIPRef should not be nil")
		} else if *req.Spec.Networking.PublicIPv4Address.Static.PublicIPRef != "/networking/projects/p/regions/r/publicIPs/public-ip" {
			t.Errorf("Expected public IP FQID, got %s", *req.Spec.Networking.PublicIPv4Address.Static.PublicIPRef)
		}

		// Validate security groups
		if req.Spec.Networking.SecurityGroupSettings == nil {
			t.Error("SecurityGroupSettings should not be nil")
		} else if req.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs == nil {
			t.Error("SecurityGroupMemberRefs should not be nil")
		} else {
			sgRefs := *req.Spec.Networking.SecurityGroupSettings.SecurityGroupMemberRefs
			if len(sgRefs) != 2 {
				t.Errorf("Expected 2 security groups, got %d", len(sgRefs))
			} else {
				if sgRefs[0] != "/networking/projects/p/regions/r/securityGroups/sg-1" {
					t.Errorf("Expected security group FQID, got %s", sgRefs[0])
				}
				if sgRefs[1] != "/networking/projects/p/regions/r/securityGroups/sg-2" {
					t.Errorf("Expected security group FQID, got %s", sgRefs[1])
				}
			}
		}
	}

	// Validate OS settings
	if req.Spec.OsSettings == nil {
		t.Error("OsSettings should not be nil")
	} else {
		// Validate cloud-init
		if req.Spec.OsSettings.CloudInitUserData == nil {
			t.Error("CloudInitUserData should not be nil")
		} else if *req.Spec.OsSettings.CloudInitUserData != "#cloud-config\npackages:\n  - nginx" {
			t.Errorf("CloudInitUserData mismatch, got: %s", *req.Spec.OsSettings.CloudInitUserData)
		}

		// Validate SSH keys
		if req.Spec.OsSettings.Ssh == nil {
			t.Error("Ssh should not be nil")
		} else if req.Spec.OsSettings.Ssh.AuthorizedKeys == nil {
			t.Error("AuthorizedKeys should not be nil")
		} else {
			keys := *req.Spec.OsSettings.Ssh.AuthorizedKeys
			if len(keys) != 2 {
				t.Errorf("Expected 2 SSH keys, got %d", len(keys))
			} else {
				if keys[0].Value != "ssh-ed25519 KEY1" {
					t.Errorf("Expected first key 'ssh-ed25519 KEY1', got %s", keys[0].Value)
				}
				if keys[1].Value != "ssh-rsa KEY2" {
					t.Errorf("Expected second key 'ssh-rsa KEY2', got %s", keys[1].Value)
				}
			}
		}
	}

	// Validate placement zone (Placement is now non-pointer)
	if req.Spec.Placement.Zone == nil {
		t.Error("Zone should not be nil")
	} else if *req.Spec.Placement.Zone != "a" {
		t.Errorf("Expected zone 'a', got %s", *req.Spec.Placement.Zone)
	}

	// Validate placement group
	if req.Spec.Placement.PlacementGroupRef == nil {
		t.Error("PlacementGroupRef should not be nil")
	} else if *req.Spec.Placement.PlacementGroupRef != "/compute/projects/p/regions/r/placementGroups/pg" {
		t.Errorf("Expected placement group FQID, got %s", *req.Spec.Placement.PlacementGroupRef)
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else if (*req.Metadata.UserLabels)["env"] != "prod" {
		t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
	}
}

// Example demonstrates how to create a disk using the builder pattern.
func Example_diskBuilder() {
	// Create a 100GB disk from Ubuntu image in zone 'a'
	diskRequest := NewDiskBuilder("my-boot-disk").
		WithImage("ubuntu.24-04.1").
		WithSizeGB(100).
		WithZone("a").
		WithLabels(map[string]string{
			"environment": "production",
			"team":        "platform",
		}).
		Build()

	// diskRequest can now be passed to client.Compute().Disks().Create()
	_ = diskRequest
}

// Example demonstrates how to create a virtual machine using the builder pattern.
func Example_virtualMachineBuilder() {
	// Create a VM with boot disk, public IP, and SSH access
	vmRequest := NewVirtualMachineBuilder("my-web-server").
		WithVMInstanceType("c1a.m"). // 2 vCPU, 4GB RAM
		WithBootDisk(DiskRef("/compute/projects/p/regions/r/disks/my-boot-disk")).
		WithDataDisk(DiskRef("/compute/projects/p/regions/r/disks/my-data-disk")).
		WithPublicIP(PublicIPRef("/networking/projects/p/regions/r/publicIPs/my-public-ip")).
		WithSecurityGroup(SecurityGroupRef("/networking/projects/p/regions/r/securityGroups/web-sg")).
		WithSecurityGroup(SecurityGroupRef("/networking/projects/p/regions/r/securityGroups/ssh-sg")).
		WithSSHKey("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ...").
		WithZone("a").
		WithLabels(map[string]string{
			"environment": "production",
			"service":     "web",
		}).
		Build()

	// vmRequest can now be passed to client.Compute().VirtualMachines().Create()
	_ = vmRequest
}

// Example demonstrates how to hot-attach a disk to a running VM.
func Example_hotswapDiskAttachmentBuilder() {
	// Create a hotswap disk attachment to attach a data disk to a running VM
	// Parameters: attachment name, VM reference, disk reference
	attachmentRequest := NewHotswapDiskAttachmentBuilder(
		"my-attachment",
		VMRef("/compute/projects/p/regions/r/virtualMachines/my-running-vm"),
		DiskRef("/compute/projects/p/regions/r/disks/my-data-disk"),
	).WithLabels(map[string]string{
		"purpose": "additional-storage",
	}).Build()

	// attachmentRequest can now be passed to client.Compute().HotswapDiskAttachments().Create()
	// This allows attaching storage to a VM without stopping it
	_ = attachmentRequest
}

func TestSnapshotBuilder(t *testing.T) {
	req := NewSnapshotBuilder("test-snap").
		WithDiskRef("/compute/projects/p/regions/r/disks/my-disk").
		Build()

	if req.Kind != "Snapshot" {
		t.Errorf("Expected Kind 'Snapshot', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-snap" {
		t.Errorf("Expected Id 'test-snap', got %s", req.Metadata.Id)
	}
	if req.Spec.DiskRef == nil || *req.Spec.DiskRef != "/compute/projects/p/regions/r/disks/my-disk" {
		t.Errorf("Expected disk ref, got %v", req.Spec.DiskRef)
	}
}

func TestDiskBuilderWithSnapshot(t *testing.T) {
	req := NewDiskBuilder("restore-disk").
		WithSnapshot("/compute/projects/p/regions/r/snapshots/my-snap").
		WithSizeGB(100).
		WithZone("a").
		Build()

	if req.Spec.Source == nil {
		t.Fatal("Source should not be nil")
	}
	if req.Spec.Source.Type != "snapshot" {
		t.Errorf("Expected source type 'snapshot', got %s", req.Spec.Source.Type)
	}
	if req.Spec.Source.SnapshotRef == nil || *req.Spec.Source.SnapshotRef != "/compute/projects/p/regions/r/snapshots/my-snap" {
		t.Errorf("Expected snapshot ref, got %v", req.Spec.Source.SnapshotRef)
	}
	if req.Spec.Source.DiskImageRef != nil {
		t.Error("DiskImageRef should be nil when using snapshot source")
	}
}

func TestIsSnapshotReady(t *testing.T) {
	if IsSnapshotReady(nil) {
		t.Error("nil snapshot should not be ready")
	}
}
