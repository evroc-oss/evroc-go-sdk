// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// DiskRef is a type-safe reference to a disk resource.
// DiskRef must always be a Fully Qualified ID (FQID).
// Use client.Compute().DiskRef(name) to construct from a name, or disk.Ref() from a resource.
type DiskRef = computetypes.DiskRef

// VMRef is a type-safe reference to a virtual machine resource.
// VMRef must always be a Fully Qualified ID (FQID).
// Use client.Compute().VMRef(name) to construct from a name, or vm.Ref() from a resource.
type VMRef = computetypes.VMRef

// PublicIPRef is a type-safe reference to a public IP resource.
// PublicIPRef must always be a Fully Qualified ID (FQID).
// Use client.Networking().PublicIPRef(name) to construct from a name, or publicIP.Ref() from a resource.
type PublicIPRef = networkingtypes.PublicIPRef

// SecurityGroupRef is a type-safe reference to a security group resource.
// SecurityGroupRef must always be a Fully Qualified ID (FQID).
// Use client.Networking().SecurityGroupRef(name) to construct from a name, or sg.Ref() from a resource.
type SecurityGroupRef = networkingtypes.SecurityGroupRef

// PlacementGroupRef is a type-safe reference to a placement group resource.
// PlacementGroupRef must always be a Fully Qualified ID (FQID).
// Use client.Compute().PlacementGroupRef(name) to construct from a name, or pg.Ref() from a resource.
type PlacementGroupRef = computetypes.PlacementGroupRef
