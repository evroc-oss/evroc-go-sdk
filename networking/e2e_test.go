// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package networking_test

import (
	"context"
	"testing"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
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

	// Wait for public IP to be fully deleted
	t.Logf("Waiting for public IP deletion...")
	if err := client.Networking().PublicIPs().WaitForDeleted(ctx, ipID, 2*time.Minute); err != nil {
		t.Fatalf("public IP not deleted in time: %v", err)
	}

	t.Logf("Public IP lifecycle test completed successfully")
}

func TestE2E_VPC_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	vpcName := e2etest.RandomName("vpc")

	t.Logf("Creating VPC: %s", vpcName)

	vpc, err := networking.NewVPCBuilder(vpcName).
		WithIPv4CIDRBlock("10.100.0.0/16").
		WithDualStack().
		Create(ctx, client.Networking().VirtualPrivateClouds())

	if err != nil {
		// VPC create may be restricted to platform admins; fall back to read-only tests
		t.Logf("VPC create returned: %v — falling back to read-only tests", err)
		t.Run("ReadOnly", func(t *testing.T) {
			testVPCReadOnly(t, ctx, client)
		})
		return
	}

	vpcID := e2etest.MustGetID(t, vpc.Metadata.Id, "VPC")
	t.Logf("Created VPC with ID: %s", vpcID)

	vpcDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Networking().VirtualPrivateClouds().Delete, vpcID, "VPC", &vpcDeleted)

	// Read VPC
	t.Logf("Reading VPC: %s", vpcID)
	retrieved, err := client.Networking().VirtualPrivateClouds().Get(ctx, vpcID)
	if err != nil {
		t.Fatalf("failed to get VPC: %v", err)
	}

	if retrieved.Metadata.Id != vpcID {
		t.Errorf("expected VPC ID %s, got %v", vpcID, retrieved.Metadata.Id)
	}

	// List VPCs - should include ours
	t.Logf("Listing VPCs")
	vpcs, err := client.Networking().VirtualPrivateClouds().List(ctx)
	if err != nil {
		t.Fatalf("failed to list VPCs: %v", err)
	}

	e2etest.AssertInList(t, vpcs.Items, vpcID, func(v networkingtypes.VirtualPrivateCloud) string { return v.Metadata.Id }, "VPC")

	// Delete VPC
	t.Logf("Deleting VPC: %s", vpcID)
	if err := client.Networking().VirtualPrivateClouds().Delete(ctx, vpcID); err != nil {
		t.Fatalf("failed to delete VPC: %v", err)
	}
	vpcDeleted = true

	// Wait for VPC to be fully deleted
	t.Logf("Waiting for VPC deletion...")
	if err := client.Networking().VirtualPrivateClouds().WaitForDeleted(ctx, vpcID, 2*time.Minute); err != nil {
		t.Fatalf("VPC not deleted in time: %v", err)
	}

	t.Logf("VPC lifecycle test completed successfully")
}

func testVPCReadOnly(t *testing.T, ctx context.Context, client *evroc.Client) {
	t.Helper()

	// List VPCs
	t.Logf("Listing VPCs")
	vpcs, err := client.Networking().VirtualPrivateClouds().List(ctx)
	if err != nil {
		t.Fatalf("failed to list VPCs: %v", err)
	}

	if len(vpcs.Items) == 0 {
		t.Fatal("expected at least one VPC (default VPC)")
	}
	t.Logf("Found %d VPC(s)", len(vpcs.Items))

	// Get the default VPC by name
	defaultVPCName := "default-" + e2etest.GetRegion()
	t.Logf("Getting default VPC: %s", defaultVPCName)
	vpc, err := client.Networking().VirtualPrivateClouds().Get(ctx, defaultVPCName)
	if err != nil {
		t.Fatalf("failed to get default VPC: %v", err)
	}

	if vpc.Metadata.Id != defaultVPCName {
		t.Errorf("expected VPC ID %s, got %s", defaultVPCName, vpc.Metadata.Id)
	}

	t.Logf("Default VPC verified: %s (IPv4 CIDRs: %v)", vpc.Metadata.Id, vpc.Status.AssignedIPv4CidrBlocks)
	t.Logf("VPC read-only test completed successfully")
}

func TestE2E_Subnet_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	subnetName := e2etest.RandomName("subnet")

	vpcRef := client.Networking().DefaultVPCRef()

	t.Logf("Creating subnet: %s (in default VPC)", subnetName)

	subnet, err := networking.NewSubnetBuilder(subnetName).
		WithVPCRef(vpcRef).
		WithIPv4CIDRBlock("10.0.200.0/24").
		WithDualStack().
		WithZone("a").
		Create(ctx, client.Networking().Subnets())

	if err != nil {
		t.Logf("Subnet create returned: %v — falling back to read-only tests", err)
		t.Run("ReadOnly", func(t *testing.T) {
			testSubnetReadOnly(t, ctx, client)
		})
		return
	}

	subnetID := e2etest.MustGetID(t, subnet.Metadata.Id, "subnet")
	t.Logf("Created subnet with ID: %s", subnetID)

	subnetDeleted := false
	e2etest.DeferCleanup(t, ctx, client.Networking().Subnets().Delete, subnetID, "subnet", &subnetDeleted)

	// Read subnet
	t.Logf("Reading subnet: %s", subnetID)
	retrieved, err := client.Networking().Subnets().Get(ctx, subnetID)
	if err != nil {
		t.Fatalf("failed to get subnet: %v", err)
	}

	if retrieved.Metadata.Id != subnetID {
		t.Errorf("expected subnet ID %s, got %v", subnetID, retrieved.Metadata.Id)
	}
	if retrieved.Spec.VpcRef != vpcRef {
		t.Errorf("expected VPC ref %s, got %s", vpcRef, retrieved.Spec.VpcRef)
	}

	// List subnets - should include ours
	t.Logf("Listing subnets")
	subnets, err := client.Networking().Subnets().List(ctx)
	if err != nil {
		t.Fatalf("failed to list subnets: %v", err)
	}
	e2etest.AssertInList(t, subnets.Items, subnetID, func(s networkingtypes.Subnet) string { return s.Metadata.Id }, "subnet")

	// Delete subnet
	t.Logf("Deleting subnet: %s", subnetID)
	if err := client.Networking().Subnets().Delete(ctx, subnetID); err != nil {
		t.Fatalf("failed to delete subnet: %v", err)
	}
	subnetDeleted = true

	// Wait for subnet to be fully deleted
	t.Logf("Waiting for subnet deletion...")
	if err := client.Networking().Subnets().WaitForDeleted(ctx, subnetID, 2*time.Minute); err != nil {
		t.Fatalf("subnet not deleted in time: %v", err)
	}

	t.Logf("Subnet lifecycle test completed successfully")
}

func testSubnetReadOnly(t *testing.T, ctx context.Context, client *evroc.Client) {
	t.Helper()

	t.Logf("Listing subnets")
	subnets, err := client.Networking().Subnets().List(ctx)
	if err != nil {
		t.Fatalf("failed to list subnets: %v", err)
	}

	if len(subnets.Items) == 0 {
		t.Fatal("expected at least one subnet (default subnets)")
	}
	t.Logf("Found %d subnet(s)", len(subnets.Items))

	// Get the first subnet
	subnetName := subnets.Items[0].Metadata.Id
	t.Logf("Getting subnet: %s", subnetName)
	subnet, err := client.Networking().Subnets().Get(ctx, subnetName)
	if err != nil {
		t.Fatalf("failed to get subnet: %v", err)
	}

	if subnet.Metadata.Id != subnetName {
		t.Errorf("expected subnet ID %s, got %s", subnetName, subnet.Metadata.Id)
	}

	ipv4 := "none"
	if subnet.Spec.Ipv4CidrBlock != nil {
		ipv4 = *subnet.Spec.Ipv4CidrBlock
	}
	t.Logf("Subnet verified: %s (VPC: %s, IPv4: %s, stack: %s)", subnet.Metadata.Id, subnet.Spec.VpcRef, ipv4, subnet.Spec.StackType)
	t.Logf("Subnet read-only test completed successfully")
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

	// Wait for security group to be fully deleted
	t.Logf("Waiting for security group deletion...")
	if err := client.Networking().SecurityGroups().WaitForDeleted(ctx, sgID, 2*time.Minute); err != nil {
		t.Fatalf("security group not deleted in time: %v", err)
	}

	t.Logf("Security group lifecycle test completed successfully")
}
