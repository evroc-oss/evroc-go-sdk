// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"

// ============================================================================
// Security Group Protocols
// ============================================================================

// Re-export protocol constants from types for convenience
const (
	ProtocolTCP  = networkingtypes.TCP
	ProtocolUDP  = networkingtypes.UDP
	ProtocolICMP = networkingtypes.ICMP
	ProtocolAll  = networkingtypes.All
)

// ============================================================================
// Security Group Directions
// ============================================================================

// Re-export direction constants from types for convenience
const (
	DirectionIngress = networkingtypes.Ingress
	DirectionEgress  = networkingtypes.Egress
)
