// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import "testing"

func TestPublicIPBuilder(t *testing.T) {
	req := NewPublicIPBuilder("test-ip").
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	// Validate basic fields
	if req.Kind != "PublicIP" {
		t.Errorf("Expected Kind 'PublicIP', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-ip" {
		t.Errorf("Expected Id 'test-ip', got %s", req.Metadata.Id)
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else if (*req.Metadata.UserLabels)["env"] != "prod" {
		t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
	}
}

func TestSecurityGroupBuilder(t *testing.T) {
	req := NewSecurityGroupBuilder("test-sg").
		AllowIngressRule("custom-ssh", "TCP", 22, 0, "10.0.0.0/8").
		AllowEgressRule("custom-all", "All", 0, 0, "0.0.0.0/0").
		AllowSSH().
		AllowHTTP().
		AllowHTTPS().
		AllowAllEgress().
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	// Validate basic fields
	if req.Kind != "SecurityGroup" {
		t.Errorf("Expected Kind 'SecurityGroup', got %s", req.Kind)
	}
	if req.Metadata.Id != "test-sg" {
		t.Errorf("Expected Id 'test-sg', got %s", req.Metadata.Id)
	}

	// Validate rules
	if req.Spec.Rules == nil {
		t.Fatal("Rules should not be nil")
	}
	rules := *req.Spec.Rules
	if len(rules) != 6 {
		t.Fatalf("Expected 6 rules, got %d", len(rules))
	}

	// Rule 0: Custom SSH rule
	if rules[0].Name == nil || *rules[0].Name != "custom-ssh" {
		t.Errorf("Rule 0: Expected name 'custom-ssh', got %v", rules[0].Name)
	}
	if rules[0].Direction != "Ingress" {
		t.Errorf("Rule 0: Expected direction 'Ingress', got %s", rules[0].Direction)
	}
	if rules[0].Protocol == nil || *rules[0].Protocol != "TCP" {
		t.Errorf("Rule 0: Expected protocol 'TCP'")
	}
	if rules[0].Port == nil || *rules[0].Port != 22 {
		t.Errorf("Rule 0: Expected port 22")
	}

	// Rule 1: Custom egress rule
	if rules[1].Name == nil || *rules[1].Name != "custom-all" {
		t.Errorf("Rule 1: Expected name 'custom-all', got %v", rules[1].Name)
	}
	if rules[1].Direction != "Egress" {
		t.Errorf("Rule 1: Expected direction 'Egress', got %s", rules[1].Direction)
	}

	// Rule 2: AllowSSH
	if rules[2].Name == nil || *rules[2].Name != "allow-ssh" {
		t.Errorf("Rule 2: Expected name 'allow-ssh', got %v", rules[2].Name)
	}

	// Rule 3: AllowHTTP
	if rules[3].Name == nil || *rules[3].Name != "allow-http" {
		t.Errorf("Rule 3: Expected name 'allow-http', got %v", rules[3].Name)
	}
	if rules[3].Port == nil || *rules[3].Port != 80 {
		t.Errorf("Rule 3: Expected port 80")
	}

	// Rule 4: AllowHTTPS
	if rules[4].Name == nil || *rules[4].Name != "allow-https" {
		t.Errorf("Rule 4: Expected name 'allow-https', got %v", rules[4].Name)
	}
	if rules[4].Port == nil || *rules[4].Port != 443 {
		t.Errorf("Rule 4: Expected port 443")
	}

	// Rule 5: AllowAllEgress
	if rules[5].Name == nil || *rules[5].Name != "allow-all-egress" {
		t.Errorf("Rule 5: Expected name 'allow-all-egress', got %v", rules[5].Name)
	}
	if rules[5].Direction != "Egress" {
		t.Errorf("Rule 5: Expected direction 'Egress', got %s", rules[5].Direction)
	}

	// Validate labels
	if req.Metadata.UserLabels == nil {
		t.Error("UserLabels should not be nil")
	} else if (*req.Metadata.UserLabels)["env"] != "prod" {
		t.Errorf("Expected label env='prod', got %s", (*req.Metadata.UserLabels)["env"])
	}
}

func TestProtocolConstants(t *testing.T) {
	if ProtocolTCP == "" || ProtocolUDP == "" || ProtocolICMP == "" || ProtocolAll == "" {
		t.Error("Protocol constants should not be empty")
	}
}
