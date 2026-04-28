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
	Ingress = string(networkingtypes.SecurityGroupSpecRulesItemDirectionIngress)

	// Egress represents outbound traffic rules.
	Egress = string(networkingtypes.SecurityGroupSpecRulesItemDirectionEgress)
)

// Security Group Rule Protocol constants
const (
	// All allows all protocols.
	All = string(networkingtypes.SecurityGroupSpecRulesItemProtocolAll)

	// TCP allows TCP protocol only.
	TCP = string(networkingtypes.SecurityGroupSpecRulesItemProtocolTCP)

	// UDP allows UDP protocol only.
	UDP = string(networkingtypes.SecurityGroupSpecRulesItemProtocolUDP)

	// ICMP allows ICMP protocol only (ping, etc.).
	ICMP = string(networkingtypes.SecurityGroupSpecRulesItemProtocolICMP)
)

// Security Group Member Status constants
const (
	// Active indicates the member is actively part of the security group.
	Active = string(networkingtypes.Active)

	// Pending indicates the member is being added to the security group.
	Pending = string(networkingtypes.Pending)
)
