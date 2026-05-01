// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"context"
	"fmt"

	networking "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

// Update Builders
//
// Update builders provide a fluent interface for modifying existing networking resources.
//
// Example - Updating security group rules:
//
//	rules := []networking.SecurityGroupSpecRulesItem{
//		// ... your rules ...
//	}
//	updated, err := UpdateSecurityGroup("my-sg", sgService).
//		SetRules(rules).
//		Apply(ctx)
//

// SecurityGroupUpdateBuilder provides a fluent interface for updating SecurityGroup resources.
//
// This builder simplifies updating security group rules, avoiding the verbose
// get-modify-update pattern commonly used in Terraform providers.
type SecurityGroupUpdateBuilder struct {
	name                   string
	service                *SecurityGroupsService
	rules                  *[]networking.SecurityGroupSpecRulesItem
	overwriteExistingRules bool
}

// NewSecurityGroupUpdateBuilder creates a new builder for updating a security group.
func NewSecurityGroupUpdateBuilder(name string, service *SecurityGroupsService) *SecurityGroupUpdateBuilder {
	return &SecurityGroupUpdateBuilder{
		name:    name,
		service: service,
	}
}

// SetRules replaces all rules in the security group.
func (b *SecurityGroupUpdateBuilder) SetRules(rules []networking.SecurityGroupSpecRulesItem) *SecurityGroupUpdateBuilder {
	b.rules = &rules
	b.overwriteExistingRules = true
	return b
}

// AddRule adds a single rule to the security group.
// This method fetches current rules, appends the new rule, and sets the result.
func (b *SecurityGroupUpdateBuilder) AddRule(rule networking.SecurityGroupSpecRulesItem) *SecurityGroupUpdateBuilder {
	// Note: This will be applied in Apply() by fetching current state first
	if b.rules == nil {
		b.rules = &[]networking.SecurityGroupSpecRulesItem{}
	}
	*b.rules = append(*b.rules, rule)
	return b
}

// Apply applies all pending updates to the security group.
func (b *SecurityGroupUpdateBuilder) Apply(ctx context.Context) (*networking.SecurityGroup, error) {
	if b.rules == nil {
		return nil, fmt.Errorf("no updates to apply")
	}

	// Fetch current security group state
	sg, err := b.service.Get(ctx, b.name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch security group: %w", err)
	}

	// Merge existing rules with new rules when using AddRule
	// If b.rules is set, it contains the new rules to add
	var finalRules []networking.SecurityGroupSpecRulesItem

	// Start with existing rules if any
	if sg.Spec.Rules != nil && !b.overwriteExistingRules {
		finalRules = *sg.Spec.Rules
	}

	// Append new rules from builder
	if b.rules != nil {
		finalRules = append(finalRules, *b.rules...)
	}

	sg.Spec.Rules = &finalRules

	// Send the update
	updated, err := b.service.Patch(ctx, b.name, sg)
	if err != nil {
		return nil, fmt.Errorf("failed to update security group: %w", err)
	}

	return updated, nil
}

// Convenience functions for common update patterns

// UpdateSecurityGroup creates an update builder for a security group.
func UpdateSecurityGroup(name string, service *SecurityGroupsService) *SecurityGroupUpdateBuilder {
	return NewSecurityGroupUpdateBuilder(name, service)
}
