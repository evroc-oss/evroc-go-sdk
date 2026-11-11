// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Deploys two nginx VMs behind a load balancer and shows how each listener type
// passes the real client IP to the backends:
//
//   - default: TCP listener + L4Route. The LB adds a PROXY protocol header and
//     nginx reads the client IP from it.
//   - http: HTTP listener + HTTPRoute. The LB terminates HTTP and appends the
//     client IP to the X-Forwarded-For header. nginx trusts that header only
//     from the LB's private range, because clients can send their own.
//
// Usage:
//
//	go run main.go                            # Create all resources (L4, PROXY protocol)
//	go run main.go http                       # Create all resources (L7, X-Forwarded-For)
//	go run main.go destroy                    # Destroy all resources (either mode)
//	SSH_PUBLIC_KEY="ssh-ed25519 AAAA..." go run main.go  # Create with SSH access
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/loadbalancer"
	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
	"github.com/evroc-oss/evroc-go-sdk/networking"
)

const (
	sgName      = "sdk-lb-sg"
	ipName      = "sdk-lb-ip"
	lbName      = "sdk-lb-web"
	poolName    = "sdk-lb-pool"
	svcName     = "sdk-lb-http-svc"
	routeName   = "sdk-lb-http-route"
	httpRoute   = "sdk-lb-httproute"
	lbSubnet    = "10.0.0.0/8" // LB traffic to backends comes from this private range
	diskSize    = 50
	vmType      = "c1a.s" // 2 vCPUs, 4GB RAM
)

var vms = []struct {
	name string
	disk string
	zone string
}{
	{"sdk-lb-vm-a", "sdk-lb-disk-a", "a"},
	{"sdk-lb-vm-b", "sdk-lb-disk-b", "b"},
}

func main() {
	ctx := context.Background()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "destroy":
		destroy(ctx, client)
	case "http":
		create(ctx, client, true)
	case "":
		create(ctx, client, false)
	default:
		log.Fatalf("Unknown mode %q (use http or destroy)", mode)
	}
}

func create(ctx context.Context, client *evroc.Client, l7 bool) {
	if l7 {
		fmt.Println("=== Creating Load-Balanced Web Servers with X-Forwarded-For ===")
	} else {
		fmt.Println("=== Creating Load-Balanced Web Servers with PROXY Protocol ===")
	}
	fmt.Println()

	// Step 1: Security group
	fmt.Println("Step 1: Creating security group...")
	sg, err := client.Networking().SecurityGroups().Create(ctx,
		networking.NewSecurityGroupBuilder(sgName).
			AllowSSH().
			AllowIngressRule("app", "TCP", 8080, 0, lbSubnet).
			AllowAllEgress().
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create security group: %v", err)
	}
	fmt.Printf("  Created: %s\n\n", sg.Metadata.Id)

	// Step 2: Public IP for the load balancer
	fmt.Println("Step 2: Creating public IP for load balancer...")
	_, err = client.Networking().PublicIPs().Create(ctx,
		networking.NewPublicIPBuilder(ipName).Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create public IP: %v", err)
	}

	fmt.Println("  Waiting for IP allocation...")
	readyIP, err := client.Networking().PublicIPs().WaitForReady(ctx, ipName, 2*time.Minute)
	if err != nil {
		log.Fatalf("Public IP never became ready: %v", err)
	}
	ipAddress := networking.GetPublicIPAddress(readyIP)
	fmt.Printf("  Allocated: %s\n\n", ipAddress)

	// Step 3: Create disks
	fmt.Println("Step 3: Creating boot disks...")
	for _, vm := range vms {
		_, err := client.Compute().Disks().Create(ctx,
			compute.NewDiskBuilder(vm.disk).
				WithImage(string(compute.DiskImageUbuntuMinimal2404)).
				WithSizeGB(diskSize).
				WithZone(vm.zone).
				Build(),
		)
		if err != nil {
			log.Fatalf("Failed to create disk %s: %v", vm.disk, err)
		}
		fmt.Printf("  Created: %s (zone %s)\n", vm.disk, vm.zone)
	}

	fmt.Println("  Waiting for disks...")
	for _, vm := range vms {
		if _, err := client.Compute().Disks().WaitForReady(ctx, vm.disk, 5*time.Minute); err != nil {
			log.Fatalf("Disk %s never became ready: %v", vm.disk, err)
		}
	}
	fmt.Println("  All disks ready")
	fmt.Println()

	// Step 4: Create VMs with nginx configured to read the client IP
	fmt.Println("Step 4: Creating VMs with nginx...")
	vmRefs := make([]compute.VMRef, len(vms))
	for i, vm := range vms {
		cloudInit := makeCloudInit(vm.name, lbName, os.Getenv("SSH_PUBLIC_KEY"), l7)
		diskRef := client.Compute().DiskRef(vm.disk)

		created, err := client.Compute().VirtualMachines().Create(ctx,
			compute.NewVirtualMachineBuilder(vm.name).
				WithBootDisk(diskRef).
				WithVMInstanceType(vmType).
				WithSecurityGroup(sg.Ref()).
				WithCloudInit(cloudInit).
				WithSubnet(client.Compute().DefaultSubnetRef(vm.zone)).
				WithZone(vm.zone).
				Build())
		if err != nil {
			log.Fatalf("Failed to create VM %s: %v", vm.name, err)
		}
		vmRefs[i] = created.Ref()
		fmt.Printf("  Created: %s (zone %s)\n", vm.name, vm.zone)
	}

	fmt.Println("  Waiting for VMs...")
	for _, vm := range vms {
		if _, err := client.Compute().VirtualMachines().WaitForReady(ctx, vm.name, 5*time.Minute); err != nil {
			log.Fatalf("VM %s never became ready: %v", vm.name, err)
		}
		fmt.Printf("  Ready: %s\n", vm.name)
	}
	fmt.Println()

	// Step 5: Create load balancer resources
	fmt.Println("Step 5: Creating load balancer...")
	lbClient := client.LoadBalancer()

	vmRefStrings := make([]string, len(vmRefs))
	for i, ref := range vmRefs {
		vmRefStrings[i] = string(ref)
	}

	lb, err := createLBStack(ctx, lbClient, string(readyIP.Ref()), vmRefStrings, l7)
	if err != nil {
		log.Fatalf("Failed to create load balancer stack: %v", err)
	}
	fmt.Printf("  Ready: %s\n", lb.Metadata.Id)
	if l7 {
		fmt.Printf("  Listener: HTTP port 80 -> port 8080 (X-Forwarded-For)\n")
	} else {
		fmt.Printf("  Listener: TCP port 80 -> port 8080 (PROXY protocol)\n")
	}
	fmt.Printf("  Backends: %s, %s\n\n", vms[0].name, vms[1].name)

	fmt.Printf("\n=== Ready: http://%s ===\n", ipAddress)
	fmt.Println("Note: cloud-init takes ~2 min after VMs are ready.")
	fmt.Println("To destroy: go run main.go destroy")
}

// createLBStack creates the load balancer stack and waits for it to be ready.
// With l7 it uses an HTTP listener and HTTPRoute instead of TCP and an L4Route.
func createLBStack(ctx context.Context, client *loadbalancer.Client, publicIPRef string, backendRefs []string, l7 bool) (*lbtypes.Loadbalancer, error) {
	if _, err := loadbalancer.NewBackendPoolBuilder(poolName).
		WithBackendRefs(backendRefs).
		Create(ctx, client.BackendPools()); err != nil {
		return nil, fmt.Errorf("backend pool: %w", err)
	}

	if _, err := loadbalancer.NewBackendServiceBuilder(svcName).
		WithPort(8080).
		WithIPProtocol(lbtypes.IPv4).
		WithBackendPoolRef(client.BackendPoolRef(poolName)).
		// L7 carries the client IP in X-Forwarded-For, so PROXY protocol stays off.
		WithProxyProtocol(!l7).
		WithHTTPHealthCheck("/").
		Create(ctx, client.BackendServices()); err != nil {
		return nil, fmt.Errorf("backend service: %w", err)
	}

	listener := lbtypes.LoadbalancerSpecListenersItem{Port: 80}
	if l7 {
		if _, err := loadbalancer.NewHTTPRouteBuilder(httpRoute).
			WithRule(lbtypes.HttprouteSpecRulesItem{Name: "all", BackendServiceRef: client.BackendServiceRef(svcName)}).
			Create(ctx, client.HTTPRoutes()); err != nil {
			return nil, fmt.Errorf("HTTP route: %w", err)
		}
		listener.Protocol = lbtypes.HTTP
		listener.RouteRefs = &[]string{client.HTTPRouteRef(httpRoute)}
	} else {
		if _, err := loadbalancer.NewL4RouteBuilder(routeName).
			WithBackendServiceRef(client.BackendServiceRef(svcName)).
			Create(ctx, client.L4Routes()); err != nil {
			return nil, fmt.Errorf("L4 route: %w", err)
		}
		listener.Protocol = lbtypes.TCP
		listener.RouteRefs = &[]string{client.L4RouteRef(routeName)}
	}

	listenerName := "http"
	listener.Name = &listenerName
	if _, err := loadbalancer.NewLoadBalancerBuilder(lbName).
		WithPublicIPRef(publicIPRef).
		WithListener(listener).
		Create(ctx, client.LoadBalancers()); err != nil {
		return nil, fmt.Errorf("load balancer: %w", err)
	}

	return client.LoadBalancers().WaitForReady(ctx, lbName, 5*time.Minute)
}

// deleteLBStack tears down a load balancer stack and waits for each deletion.
func deleteLBStack(ctx context.Context, client *loadbalancer.Client) error {
	timeout := 2 * time.Minute

	client.LoadBalancers().Delete(ctx, lbName)
	client.LoadBalancers().WaitForDeleted(ctx, lbName, timeout)

	// Delete both route kinds so destroy works for either mode.
	client.L4Routes().Delete(ctx, routeName)
	client.L4Routes().WaitForDeleted(ctx, routeName, timeout)
	client.HTTPRoutes().Delete(ctx, httpRoute)
	client.HTTPRoutes().WaitForDeleted(ctx, httpRoute, timeout)

	client.BackendServices().Delete(ctx, svcName)
	client.BackendServices().WaitForDeleted(ctx, svcName, timeout)

	client.BackendPools().Delete(ctx, poolName)
	client.BackendPools().WaitForDeleted(ctx, poolName, timeout)

	return nil
}

// nginxProxyProtocol reads the client IP from the PROXY protocol header the LB
// adds on TCP listeners. The header can only come from the LB, since clients
// cannot reach port 8080 directly.
const nginxProxyProtocol = `# Accept PROXY protocol from the load balancer on port 8080.
      # $proxy_protocol_addr contains the real client IP.

      set_real_ip_from 0.0.0.0/0;
      real_ip_header proxy_protocol;

      server {
          listen 8080 proxy_protocol;

          location / {
              root /var/www/html;
              index index.html;
              sub_filter '{{CLIENT_IP}}' $proxy_protocol_addr;
              sub_filter '{{XFF}}' 'n/a';`

// nginxXForwardedFor reads the client IP from X-Forwarded-For, which the LB
// appends on HTTP/HTTPS listeners. Clients can send their own X-Forwarded-For,
// so only the LB's range is trusted and real_ip_recursive skips trusted
// addresses from the right: the result is the address the LB itself saw.
// The LB appends to an existing header, so behind another proxy (e.g. a CDN)
// add that proxy's ranges as further set_real_ip_from lines to get the
// original client instead of the proxy's address.
const nginxXForwardedFor = `# Trust X-Forwarded-For only when it comes from the load balancer.
      # $remote_addr becomes the right-most address not in the trusted range.

      set_real_ip_from ` + lbSubnet + `;
      real_ip_header X-Forwarded-For;
      real_ip_recursive on;

      server {
          listen 8080;

          location / {
              root /var/www/html;
              index index.html;
              sub_filter '{{CLIENT_IP}}' $remote_addr;
              sub_filter '{{XFF}}' $http_x_forwarded_for;`

// makeCloudInit generates a cloud-init that installs nginx reading the client IP
// from PROXY protocol, or from X-Forwarded-For when l7 is set.
// If sshPubKey is non-empty, it is added to the ubuntu user's authorized_keys.
func makeCloudInit(hostname, loadBalancerName, sshPubKey string, l7 bool) string {
	sshBlock := ""
	if sshPubKey != "" {
		sshBlock = fmt.Sprintf("\nssh_authorized_keys:\n  - %s\n", sshPubKey)
	}
	nginxConf := nginxProxyProtocol
	if l7 {
		nginxConf = nginxXForwardedFor
	}

	return fmt.Sprintf(`#cloud-config
package_update: true
%s
packages:
  - nginx

write_files:
  - path: /etc/nginx/sites-available/default
    owner: root:root
    permissions: '0644'
    content: |
      %s
              sub_filter '{{HOSTNAME}}' '%s';
              sub_filter '{{LB_NAME}}' '%s';
              sub_filter_once off;
              sub_filter_types text/html;
          }
      }

  - path: /var/www/html/index.html
    owner: www-data:www-data
    permissions: '0644'
    content: |
      <pre>client_ip={{CLIENT_IP}} xff={{XFF}} hostname={{HOSTNAME}} lb={{LB_NAME}}</pre>

runcmd:
  - systemctl enable nginx
  - systemctl restart nginx
`, sshBlock, nginxConf, hostname, loadBalancerName)
}

func destroy(ctx context.Context, client *evroc.Client) {
	fmt.Println("=== Destroying Load-Balanced Web Server Infrastructure ===")
	fmt.Println()

	fmt.Println("Deleting load balancer stack...")
	if err := deleteLBStack(ctx, client.LoadBalancer()); err != nil {
		log.Printf("  Warning: %v", err)
	} else {
		fmt.Println("  Deleted")
	}

	for _, vm := range vms {
		fmt.Printf("Deleting VM %s...\n", vm.name)
		if err := client.Compute().VirtualMachines().Delete(ctx, vm.name); err != nil {
			log.Printf("  Failed (may not exist): %v", err)
		} else {
			fmt.Println("  Deleted")
		}
	}

	fmt.Println("Waiting for VM deletion...")
	for _, vm := range vms {
		if err := client.Compute().VirtualMachines().WaitForDeleted(ctx, vm.name, 2*time.Minute); err != nil {
			log.Printf("  Warning: %v", err)
		}
	}

	for _, vm := range vms {
		fmt.Printf("Deleting disk %s...\n", vm.disk)
		if err := client.Compute().Disks().Delete(ctx, vm.disk); err != nil {
			log.Printf("  Failed (may not exist): %v", err)
		} else {
			fmt.Println("  Deleted")
		}
	}

	fmt.Printf("Deleting public IP %s...\n", ipName)
	if err := client.Networking().PublicIPs().Delete(ctx, ipName); err != nil {
		log.Printf("  Failed (may not exist): %v", err)
	} else {
		fmt.Println("  Deleted")
	}

	fmt.Printf("Deleting security group %s...\n", sgName)
	if err := client.Networking().SecurityGroups().Delete(ctx, sgName); err != nil {
		log.Printf("  Failed (may not exist): %v", err)
	} else {
		fmt.Println("  Deleted")
	}

	fmt.Println("\n=== Cleanup Complete ===")
}
