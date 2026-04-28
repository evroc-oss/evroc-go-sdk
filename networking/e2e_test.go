// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package networking_test

import (
	"context"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	"github.com/evroc-oss/evroc-go-sdk/networking"
	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

func TestE2E_PublicIP_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	ipName := e2etest.RandomName("ip")

	t.Logf("Creating public IP: %s", ipName)

	// Create public IP
	ip, err := networking.NewPublicIPBuilder(ipName).
		Create(ctx, client.Networking().PublicIPs())

	if err != nil {
		t.Fatalf("failed to create public IP: %v", err)
	}

	ipID := e2etest.MustGetID(t, ip.Metadata.Id, "public IP")
	t.Logf("Created public IP with ID: %s", ipID)

	ipDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Networking().PublicIPs().Delete, ipID, "public IP", &ipDeleted)

	// Wait for public IP to be ready
	t.Logf("Waiting for public IP to be ready...")
	if _, err := client.Networking().PublicIPs().WaitForReady(ctx, ipID, 2*time.Minute); err != nil {
		t.Fatalf("public IP never became ready: %v", err)
	}

	// Re-fetch to get the assigned address
	ip, err = client.Networking().PublicIPs().Get(ctx, ipID)
	if err != nil {
		t.Fatalf("failed to re-fetch public IP: %v", err)
	}

	// Verify public IP was created with an address
	if ip.Status.PublicIPv4Address == nil || *ip.Status.PublicIPv4Address == "" {
		t.Error("expected public IP address to be assigned after ready")
	} else {
		t.Logf("Public IP assigned: %s", *ip.Status.PublicIPv4Address)
	}

	// Read public IP
	t.Logf("Reading public IP: %s", ipID)
	retrieved, err := client.Networking().PublicIPs().Get(ctx, ipID)
	if err != nil {
		t.Fatalf("failed to get public IP: %v", err)
	}

	if retrieved.Metadata.Id != ipID {
		t.Errorf("expected public IP ID %s, got %v", ipID, retrieved.Metadata.Id)
	}

	// List public IPs - should include ours
	t.Logf("Listing public IPs")
	ips, err := client.Networking().PublicIPs().List(ctx)
	if err != nil {
		t.Fatalf("failed to list public IPs: %v", err)
	}
	e2etest.AssertInList(t, ips.Items, ipID, func(p networkingtypes.PublicIP) string { return p.Metadata.Id }, "public IP")

	// Delete public IP
	t.Logf("Deleting public IP: %s", ipID)
	if err := client.Networking().PublicIPs().Delete(ctx, ipID); err != nil {
		t.Fatalf("failed to delete public IP: %v", err)
	}
	ipDeleted = true

	// Verify public IP was deleted
	t.Logf("Verifying public IP deletion")
	e2etest.AssertDeleted(t, ctx, func(ctx context.Context, id string) (any, error) {
		return client.Networking().PublicIPs().Get(ctx, id)
	}, ipID, "public IP")

	t.Logf("Public IP lifecycle test completed successfully")
}

func TestE2E_SecurityGroup_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	sgName := e2etest.RandomName("sg")

	t.Logf("Creating security group: %s", sgName)

	// Create security group with a simple rule
	sg, err := networking.NewSecurityGroupBuilder(sgName).
		AllowIngressRule("ssh", "TCP", 22, 0, "0.0.0.0/0").
		Create(ctx, client.Networking().SecurityGroups())

	if err != nil {
		t.Fatalf("failed to create security group: %v", err)
	}

	sgID := e2etest.MustGetID(t, sg.Metadata.Id, "security group")
	t.Logf("Created security group with ID: %s", sgID)

	sgDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Networking().SecurityGroups().Delete, sgID, "security group", &sgDeleted)

	// Wait for security group to be ready
	t.Logf("Waiting for security group to be ready...")
	if _, err := client.Networking().SecurityGroups().WaitForReady(ctx, sgID, 2*time.Minute); err != nil {
		t.Fatalf("security group never became ready: %v", err)
	}

	// Verify security group was created with the rule
	if sg.Spec.Rules == nil || len(*sg.Spec.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(*sg.Spec.Rules))
	}

	// Read security group
	t.Logf("Reading security group: %s", sgID)
	retrieved, err := client.Networking().SecurityGroups().Get(ctx, sgID)
	if err != nil {
		t.Fatalf("failed to get security group: %v", err)
	}

	if retrieved.Metadata.Id != sgID {
		t.Errorf("expected security group ID %s, got %v", sgID, retrieved.Metadata.Id)
	}

	// List security groups - should include ours
	t.Logf("Listing security groups")
	sgs, err := client.Networking().SecurityGroups().List(ctx)
	if err != nil {
		t.Fatalf("failed to list security groups: %v", err)
	}
	e2etest.AssertInList(t, sgs.Items, sgID, func(s networkingtypes.SecurityGroup) string { return s.Metadata.Id }, "security group")

	// Update security group (add another rule)
	t.Logf("Updating security group: %s", sgID)
	httpRuleName := "http"
	httpPort := int32(80)
	httpProto := networkingtypes.SecurityGroupSpecRulesItemProtocol("TCP")
	httpAddress := networkingtypes.SecurityGroupSpecRulesItemAddress{
		IpAddressOrCIDR: "0.0.0.0/0",
	}

	// Create HTTP rule (port 80, no endPort means single port)
	httpRule := networkingtypes.SecurityGroupSpecRulesItem{
		Name:      &httpRuleName,
		Direction: networking.DirectionIngress,
		Protocol:  &httpProto,
		Port:      &httpPort,
	}
	httpRule.Remote.Address = &httpAddress

	updated, err := networking.UpdateSecurityGroup(sgID, client.Networking().SecurityGroups()).
		AddRule(httpRule).
		Apply(ctx)

	if err != nil {
		t.Fatalf("failed to update security group: %v", err)
	}

	if updated.Spec.Rules == nil || len(*updated.Spec.Rules) != 2 {
		t.Errorf("expected 2 rules after update, got %d", len(*updated.Spec.Rules))
	}

	// Delete security group
	t.Logf("Deleting security group: %s", sgID)
	if err := client.Networking().SecurityGroups().Delete(ctx, sgID); err != nil {
		t.Fatalf("failed to delete security group: %v", err)
	}
	sgDeleted = true

	// Verify security group was deleted
	t.Logf("Verifying security group deletion")
	e2etest.AssertDeleted(t, ctx, func(ctx context.Context, id string) (any, error) {
		return client.Networking().SecurityGroups().Get(ctx, id)
	}, sgID, "security group")

	t.Logf("Security group lifecycle test completed successfully")
}
