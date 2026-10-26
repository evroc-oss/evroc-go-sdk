// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2etest

import (
	"time"

	"github.com/evroc-oss/evroc-go-sdk/compute"
)

// Test resource configuration constants
const (
	// Disk constants
	TestDiskSizeGB   = 100
	TestDiskImage    = compute.DiskImageUbuntu2404
	TestDiskZone     = "a"
	DiskReadyTimeout = 6 * time.Minute

	// VM constants
	TestVMSize         = compute.VMSizeC1aM
	TestVMZone         = "a"
	VMReadyTimeout     = 6 * time.Minute
	VMDeletionTimeout  = 3 * time.Minute
	VMDeletionInterval = 10 * time.Second

	// IAM constants
	CredentialRevokeTimeout  = 2 * time.Minute
	CredentialRevokeInterval = 5 * time.Second

	// KMS constants
	KMSStateTimeout  = 3 * time.Minute
	KMSStateInterval = 5 * time.Second

	// Placement group constants
	TestPlacementGroupStrategy = "spread"
	TestPlacementGroupZone     = "a"

	// General timeouts
	APIPropagationDelay = 2 * time.Second
)
