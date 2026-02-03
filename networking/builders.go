// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package networking provides builder patterns for networking resources.
package networking

import (
	"context"

	networking "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// ============================================================================
// PublicIP Builder
// ============================================================================

// PublicIPBuilder provides a fluent interface for creating PublicIP resources.
type PublicIPBuilder struct {
	name   string
	labels map[string]string
}

// NewPublicIPBuilder creates a new builder for PublicIP.
func NewPublicIPBuilder(name string) *PublicIPBuilder {
	return &PublicIPBuilder{
		name: name,
	}
}

// WithLabels sets user-defined labels for the public IP.
func (b *PublicIPBuilder) WithLabels(labels map[string]string) *PublicIPBuilder {
	b.labels = labels
	return b
}

// Build creates the PublicIPRequest structure.
func (b *PublicIPBuilder) Build() *networking.PublicIPRequest {
	req := &networking.PublicIPRequest{
		ApiVersion: "networking/v1alpha2",
		Kind:       "PublicIP",
		Metadata: networking.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: networking.PublicIPSpec{},
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := networking.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the public IP in one call.
func (b *PublicIPBuilder) Create(ctx context.Context, client *PublicIPsService) (*networking.PublicIP, error) {
	req := b.Build()
	return client.Create(ctx, req)
}

// ============================================================================
// SecurityGroup Builder
// ============================================================================

// SecurityGroupBuilder provides a fluent interface for creating SecurityGroup resources.
type SecurityGroupBuilder struct {
	name   string
	rules  []networking.SecurityGroupSpecRulesItem
	labels map[string]string
}

// NewSecurityGroupBuilder creates a new builder for SecurityGroup.
func NewSecurityGroupBuilder(name string) *SecurityGroupBuilder {
	return &SecurityGroupBuilder{
		name:  name,
		rules: []networking.SecurityGroupSpecRulesItem{},
	}
}

// AllowIngressRule adds an ingress (inbound) rule to the security group
// direction: "Ingress"
// protocol: "TCP", "UDP", "ICMP", or "all"
// port: 0 means all ports (or use specific port like 22, 80, 443)
// endPort: optional, for port ranges (e.g., port=8000, endPort=9000)
// source: CIDR block like "0.0.0.0/0" or specific IP
func (b *SecurityGroupBuilder) AllowIngressRule(ruleName string, protocol string, port int32, endPort int32, source string) *SecurityGroupBuilder {
	proto := networking.SecurityGroupSpecRulesItemProtocol(protocol)
	rule := networking.SecurityGroupSpecRulesItem{
		Name:      ruleName,
		Direction: networking.Ingress,
		Protocol:  &proto,
	}

	if port > 0 {
		rule.Port = &port
	}
	if endPort > 0 {
		rule.EndPort = &endPort
	}
	if source != "" {
		rule.Remote = &struct {
			Address          *networking.SecurityGroupSpecRulesItemAddress          `json:"address,omitempty"`
			SecurityGroupRef *networking.SecurityGroupSpecRulesItemSecurityGroupRef `json:"securityGroupRef,omitempty"`
			SubnetRef        *networking.SecurityGroupSpecRulesItemSubnetRef        `json:"subnetRef,omitempty"`
		}{
			Address: &networking.SecurityGroupSpecRulesItemAddress{
				IPAddressOrCIDR: &source,
			},
		}
	}

	b.rules = append(b.rules, rule)
	return b
}

// AllowEgressRule adds an egress (outbound) rule to the security group.
func (b *SecurityGroupBuilder) AllowEgressRule(ruleName string, protocol string, port int32, endPort int32, destination string) *SecurityGroupBuilder {
	proto := networking.SecurityGroupSpecRulesItemProtocol(protocol)
	rule := networking.SecurityGroupSpecRulesItem{
		Name:      ruleName,
		Direction: networking.Egress,
		Protocol:  &proto,
	}

	if port > 0 {
		rule.Port = &port
	}
	if endPort > 0 {
		rule.EndPort = &endPort
	}
	if destination != "" {
		rule.Remote = &struct {
			Address          *networking.SecurityGroupSpecRulesItemAddress          `json:"address,omitempty"`
			SecurityGroupRef *networking.SecurityGroupSpecRulesItemSecurityGroupRef `json:"securityGroupRef,omitempty"`
			SubnetRef        *networking.SecurityGroupSpecRulesItemSubnetRef        `json:"subnetRef,omitempty"`
		}{
			Address: &networking.SecurityGroupSpecRulesItemAddress{
				IPAddressOrCIDR: &destination,
			},
		}
	}

	b.rules = append(b.rules, rule)
	return b
}

// AllowSSH is a convenience method to allow SSH (port 22) from anywhere.
func (b *SecurityGroupBuilder) AllowSSH() *SecurityGroupBuilder {
	return b.AllowIngressRule("allow-ssh", "TCP", 22, 0, "0.0.0.0/0")
}

// AllowHTTP is a convenience method to allow HTTP (port 80) from anywhere.
func (b *SecurityGroupBuilder) AllowHTTP() *SecurityGroupBuilder {
	return b.AllowIngressRule("allow-http", "TCP", 80, 0, "0.0.0.0/0")
}

// AllowHTTPS is a convenience method to allow HTTPS (port 443) from anywhere.
func (b *SecurityGroupBuilder) AllowHTTPS() *SecurityGroupBuilder {
	return b.AllowIngressRule("allow-https", "TCP", 443, 0, "0.0.0.0/0")
}

// AllowAllEgress is a convenience method to allow all outbound traffic.
func (b *SecurityGroupBuilder) AllowAllEgress() *SecurityGroupBuilder {
	return b.AllowEgressRule("allow-all-egress", "all", 0, 0, "0.0.0.0/0")
}

// WithLabels sets user-defined labels for the security group.
func (b *SecurityGroupBuilder) WithLabels(labels map[string]string) *SecurityGroupBuilder {
	b.labels = labels
	return b
}

// Build creates the SecurityGroupRequest structure.
func (b *SecurityGroupBuilder) Build() *networking.SecurityGroupRequest {
	req := &networking.SecurityGroupRequest{
		ApiVersion: "networking/v1alpha2",
		Kind:       "SecurityGroup",
		Metadata: networking.RegionalMetadataRequest{
			Name: &b.name,
		},
		Spec: networking.SecurityGroupSpec{},
	}

	if len(b.rules) > 0 {
		req.Spec.Rules = &b.rules
	}

	// Add labels if specified
	if len(b.labels) > 0 {
		userLabels := networking.UserLabels(b.labels)
		req.Metadata.UserLabels = &userLabels
	}

	return req
}

// Create is a convenience method that builds and creates the security group in one call.
func (b *SecurityGroupBuilder) Create(ctx context.Context, client *SecurityGroupsService) (*networking.SecurityGroup, error) {
	req := b.Build()
	return client.Create(ctx, req)
}
