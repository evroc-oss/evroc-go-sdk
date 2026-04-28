// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package networking provides builder patterns for networking resources.
package networking

import (
	"context"

	networking "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// builderAPIVersion is the full API version string for networking resource requests.
// It references the apiVersion constant from client_generated.go.
const builderAPIVersion = "networking/" + apiVersion

// ============================================================================
// PublicIP Builder
// ============================================================================

// PublicIPBuilder provides a fluent interface for creating PublicIP resources.
type PublicIPBuilder struct {
	id     string
	labels map[string]string
}

// NewPublicIPBuilder creates a new builder for PublicIP.
func NewPublicIPBuilder(id string) *PublicIPBuilder {
	return &PublicIPBuilder{
		id: id,
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
		ApiVersion: builderAPIVersion,
		Kind:       "PublicIP",
		Metadata: networking.RegionalMetadataRequest{
			Id: b.id,
		},
		// TODO: Once the API spec is updated to include omitempty tag for spec field,
		// we can set this to nil and it will be omitted from JSON entirely.
		// For now, v1beta1 API requires spec to be present, so we send an empty object {}.
		Spec: make(networking.PublicIPSpec), // or: map[string]interface{}{}
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
	id     string
	rules  []networking.SecurityGroupSpecRulesItem
	labels map[string]string
}

// NewSecurityGroupBuilder creates a new builder for SecurityGroup.
func NewSecurityGroupBuilder(id string) *SecurityGroupBuilder {
	return &SecurityGroupBuilder{
		id:    id,
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
		Name:      &ruleName,
		Direction: networking.SecurityGroupSpecRulesItemDirectionIngress,
		Protocol:  &proto,
	}

	if port > 0 {
		rule.Port = &port
	}
	if endPort > 0 {
		rule.EndPort = &endPort
	}
	if source != "" {
		rule.Remote = struct {
			Address          *networking.SecurityGroupSpecRulesItemAddress `json:"address,omitempty"`
			SecurityGroupRef *string                                       `json:"securityGroupRef,omitempty"`
			SubnetRef        *string                                       `json:"subnetRef,omitempty"`
		}{
			Address: &networking.SecurityGroupSpecRulesItemAddress{
				IpAddressOrCIDR: source,
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
		Name:      &ruleName,
		Direction: networking.SecurityGroupSpecRulesItemDirectionEgress,
		Protocol:  &proto,
	}

	if port > 0 {
		rule.Port = &port
	}
	if endPort > 0 {
		rule.EndPort = &endPort
	}
	if destination != "" {
		rule.Remote = struct {
			Address          *networking.SecurityGroupSpecRulesItemAddress `json:"address,omitempty"`
			SecurityGroupRef *string                                       `json:"securityGroupRef,omitempty"`
			SubnetRef        *string                                       `json:"subnetRef,omitempty"`
		}{
			Address: &networking.SecurityGroupSpecRulesItemAddress{
				IpAddressOrCIDR: destination,
			},
		}
	}

	b.rules = append(b.rules, rule)
	return b
}

// AllowIngressFromSecurityGroup adds an ingress rule allowing traffic from another security group.
// The sgRef must be a fully qualified resource ID (e.g. /networking/projects/{project}/regions/{region}/securityGroups/{name}).
func (b *SecurityGroupBuilder) AllowIngressFromSecurityGroup(ruleName string, protocol string, port int32, endPort int32, sgRef string) *SecurityGroupBuilder {
	proto := networking.SecurityGroupSpecRulesItemProtocol(protocol)
	rule := networking.SecurityGroupSpecRulesItem{
		Name:      &ruleName,
		Direction: networking.SecurityGroupSpecRulesItemDirectionIngress,
		Protocol:  &proto,
	}

	if port > 0 {
		rule.Port = &port
	}
	if endPort > 0 {
		rule.EndPort = &endPort
	}

	// For now, expecting fully qualified ID or will be resolved by service
	// Format: /networking/projects/{project}/regions/{region}/securityGroups/{name}
	rule.Remote = struct {
		Address          *networking.SecurityGroupSpecRulesItemAddress `json:"address,omitempty"`
		SecurityGroupRef *string                                       `json:"securityGroupRef,omitempty"`
		SubnetRef        *string                                       `json:"subnetRef,omitempty"`
	}{
		SecurityGroupRef: &sgRef,
	}

	b.rules = append(b.rules, rule)
	return b
}

// AllowIngressFromSubnet adds an ingress rule allowing traffic from a subnet.
// The subnetRef must be a fully qualified resource ID (e.g. /networking/projects/{project}/regions/{region}/subnets/{name}).
func (b *SecurityGroupBuilder) AllowIngressFromSubnet(ruleName string, protocol string, port int32, endPort int32, subnetRef string) *SecurityGroupBuilder {
	proto := networking.SecurityGroupSpecRulesItemProtocol(protocol)
	rule := networking.SecurityGroupSpecRulesItem{
		Name:      &ruleName,
		Direction: networking.SecurityGroupSpecRulesItemDirectionIngress,
		Protocol:  &proto,
	}

	if port > 0 {
		rule.Port = &port
	}
	if endPort > 0 {
		rule.EndPort = &endPort
	}

	// For now, expecting fully qualified ID or will be resolved by service
	// Format: /networking/projects/{project}/regions/{region}/subnets/{name}
	rule.Remote = struct {
		Address          *networking.SecurityGroupSpecRulesItemAddress `json:"address,omitempty"`
		SecurityGroupRef *string                                       `json:"securityGroupRef,omitempty"`
		SubnetRef        *string                                       `json:"subnetRef,omitempty"`
	}{
		SubnetRef: &subnetRef,
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
	return b.AllowEgressRule("allow-all-egress", string(ProtocolAll), 0, 0, "0.0.0.0/0")
}

// WithLabels sets user-defined labels for the security group.
func (b *SecurityGroupBuilder) WithLabels(labels map[string]string) *SecurityGroupBuilder {
	b.labels = labels
	return b
}

// Build creates the SecurityGroupRequest structure.
func (b *SecurityGroupBuilder) Build() *networking.SecurityGroupRequest {
	req := &networking.SecurityGroupRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "SecurityGroup",
		Metadata: networking.RegionalMetadataRequest{
			Id: b.id,
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
