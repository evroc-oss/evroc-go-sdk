// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package networking

import (
	"testing"

	networking "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

func TestSecurityGroupUpdateBuilder_BuilderMethods(t *testing.T) {
	builder := &SecurityGroupUpdateBuilder{
		name: "test-sg",
	}

	// Test SetRules
	protocol := networking.TCP
	port := int32(80)
	httpName := "http"
	rules := []networking.SecurityGroupSpecRulesItem{
		{
			Name:      &httpName,
			Direction: networking.Ingress,
			Protocol:  &protocol,
			Port:      &port,
		},
	}

	builder.SetRules(rules)
	if builder.rules == nil || len(*builder.rules) != 1 {
		t.Error("SetRules() should set rules")
	}
	if (*builder.rules)[0].Name == nil || *(*builder.rules)[0].Name != "http" {
		t.Error("SetRules() should preserve rule data")
	}

	// Test AddRule
	builder2 := &SecurityGroupUpdateBuilder{
		name: "test-sg",
	}
	https := networking.TCP
	httpsPort := int32(443)
	httpsName := "https"
	builder2.AddRule(networking.SecurityGroupSpecRulesItem{
		Name:      &httpsName,
		Direction: networking.Ingress,
		Protocol:  &https,
		Port:      &httpsPort,
	})
	if builder2.rules == nil || len(*builder2.rules) != 1 {
		t.Error("AddRule() should add to rules list")
	}
}
