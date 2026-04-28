// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"

// ============================================================================
// Security Group Protocols
// ============================================================================

// Re-export protocol constants from types for convenience
const (
	ProtocolTCP  = networkingtypes.SecurityGroupSpecRulesItemProtocolTCP
	ProtocolUDP  = networkingtypes.SecurityGroupSpecRulesItemProtocolUDP
	ProtocolICMP = networkingtypes.SecurityGroupSpecRulesItemProtocolICMP
	ProtocolAll  = networkingtypes.SecurityGroupSpecRulesItemProtocolAll
)

// ============================================================================
// Security Group Directions
// ============================================================================

// Re-export direction constants from types for convenience
const (
	DirectionIngress = networkingtypes.SecurityGroupSpecRulesItemDirectionIngress
	DirectionEgress  = networkingtypes.SecurityGroupSpecRulesItemDirectionEgress
)
