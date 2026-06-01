// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// Re-export commonly used constants from types/networking for easier access.
// This allows users to write networking.Ingress instead of networkingtypes.Ingress.

// Security Group Rule Direction constants
const (
	// Ingress represents inbound traffic rules.
	Ingress = string(networkingtypes.Ingress)

	// Egress represents outbound traffic rules.
	Egress = string(networkingtypes.Egress)
)

// Security Group Rule Protocol constants
const (
	// All allows all protocols.
	All = string(networkingtypes.All)

	// TCP allows TCP protocol only.
	TCP = string(networkingtypes.TCP)

	// UDP allows UDP protocol only.
	UDP = string(networkingtypes.UDP)

	// ICMP allows ICMP protocol only (ping, etc.).
	ICMP = string(networkingtypes.ICMP)
)

// Security Group Member Status constants
const (
	// Active indicates the member is actively part of the security group.
	Active = string(networkingtypes.Active)

	// Pending indicates the member is being added to the security group.
	Pending = string(networkingtypes.Pending)
)
