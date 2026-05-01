// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Package networking provides access to the evroc Networking API.
//
// The Networking API enables you to manage virtual private clouds (VPCs),
// subnets, security groups, and public IP addresses in the evroc Cloud Platform.
//
// # Resources
//
// The networking package provides access to the following resources:
//
//   - VPCs: Virtual Private Clouds for network isolation (read-only)
//   - Subnets: Network segments within VPCs (read-only)
//   - Security Groups: Firewall rules for controlling traffic
//   - Public IPs: Static public IP addresses
//
// # Getting Started
//
// Create a client and list security groups:
//
//	client, err := evroc.NewFromEnv(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	sgs, err := client.Networking().SecurityGroups().List(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # VPCs and Subnets
//
// VPCs and Subnets are read-only resources managed by the platform:
//
//	// List available VPCs
//	vpcs, err := client.Networking().VirtualPrivateClouds().List(ctx)
//
//	// Get VPC details
//	vpc, err := client.Networking().VirtualPrivateClouds().Get(ctx, "vpc-name")
//
//	// List subnets in a VPC
//	subnets, err := client.Networking().Subnets().List(ctx)
//
// # Security Groups
//
// Create security groups with firewall rules:
//
//	sg, err := client.Networking().SecurityGroups().Create(ctx,
//	    networking.NewSecurityGroupBuilder("web-sg").
//	        AllowHTTP().    // Convenience method for port 80
//	        AllowHTTPS().   // Convenience method for port 443
//	        Build(),
//	)
//
// Or use custom rules with protocol constants:
//
//	sg, err := client.Networking().SecurityGroups().Create(ctx,
//	    networking.NewSecurityGroupBuilder("web-sg").
//	        AllowIngressRule("http", string(networking.ProtocolTCP), 80, 0, "0.0.0.0/0").
//	        AllowIngressRule("custom-udp", string(networking.ProtocolUDP), 9000, 0, "0.0.0.0/0").
//	        Build(),
//	)
//
// Available protocols: ProtocolTCP, ProtocolUDP, ProtocolICMP, ProtocolAll
//
// Update security group rules:
//
//	rules := []networkingtypes.SecurityGroupSpecRulesItem{ /* ... */ }
//	sg, err := networking.UpdateSecurityGroup("web-sg", client.Networking().SecurityGroups()).
//	    SetRules(rules).
//	    Apply(ctx)
//
// # Public IP Addresses
//
// Reserve and manage public IP addresses:
//
//	ip, err := client.Networking().PublicIPs().Create(ctx,
//	    networking.NewPublicIPBuilder("my-ip").
//	        WithDescription("Load balancer IP").
//	        Build(),
//	)
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	sg, err := client.Networking().SecurityGroups().Get(ctx, "web-sg")
package networking
