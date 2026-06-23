// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package loadbalancer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/compute"
	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	"github.com/evroc-oss/evroc-go-sdk/loadbalancer"
	"github.com/evroc-oss/evroc-go-sdk/networking"
	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

const (
	lbReadyTimeout = 5 * time.Minute
	deleteTimeout  = 2 * time.Minute
	ipReadyTimeout = 2 * time.Minute
	backendPort    = 8080
)

func newTestClient(t *testing.T) (*evroc.Client, *loadbalancer.Client) {
	t.Helper()
	client := e2etest.NewClient(t)
	return client, client.LoadBalancer()
}

// TestE2E_LoadBalancer_FullStack creates the full LB resource graph
// (IP + disk + VM → pool → service → route → LB), validates every
// resource via Get/List, then tears down in reverse order.
func TestE2E_LoadBalancer_FullStack(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client, lb := newTestClient(t)

	suffix := e2etest.RandomName("lb")
	sgName := "sg-" + suffix
	diskName := "disk-" + suffix
	vmName := "vm-" + suffix
	ipName := "ip-" + suffix
	poolName := "pool-" + suffix
	svcName := "svc-" + suffix
	routeName := "route-" + suffix
	lbName := "lb-" + suffix

	// Track deletions so cleanups can be skipped when the test deletes successfully.
	var (
		sgDeleted    bool
		diskDeleted  bool
		vmDeleted    bool
		ipDeleted    bool
		poolDeleted  bool
		svcDeleted   bool
		routeDeleted bool
		lbDeleted    bool
	)

	// --- Security Group (allow port 80 for LB health checks) ---
	t.Logf("Creating security group: %s", sgName)
	sg, err := client.Networking().SecurityGroups().Create(ctx,
		networking.NewSecurityGroupBuilder(sgName).
			AllowIngressRule("allow-http", "TCP", 80, 0, "0.0.0.0/0").
			AllowIngressRule("allow-backend", "TCP", int32(backendPort), 0, "0.0.0.0/0").
			AllowAllEgress().
			Build(),
	)
	if err != nil {
		t.Fatalf("failed to create security group: %v", err)
	}
	t.Cleanup(func() {
		if !sgDeleted {
			t.Logf("Cleaning up security group: %s", sgName)
			client.Networking().SecurityGroups().Delete(ctx, sgName)
			client.Networking().SecurityGroups().WaitForDeleted(ctx, sgName, deleteTimeout)
		}
	})

	// --- Disk ---
	t.Logf("Creating disk: %s", diskName)
	disk, err := compute.NewDiskBuilder(diskName).
		WithSizeGB(e2etest.TestDiskSizeGB).
		WithImage(string(e2etest.TestDiskImage)).
		WithZone(e2etest.TestDiskZone).
		Create(ctx, client.Compute().Disks())
	if err != nil {
		t.Fatalf("failed to create disk: %v", err)
	}
	e2etest.DeferCleanup(t, ctx, client.Compute().Disks().Delete, disk.Metadata.Id, "disk", &diskDeleted)

	t.Logf("Waiting for disk to be ready...")
	if _, err := client.Compute().Disks().WaitForReady(ctx, diskName, e2etest.DiskReadyTimeout); err != nil {
		t.Fatalf("disk never became ready: %v", err)
	}

	// --- VM ---
	// cloud-init: start a minimal HTTP listener on the backend port so the LB health check passes.
	cloudInit := fmt.Sprintf(`#cloud-config
runcmd:
  - nohup python3 -m http.server %d &
`, backendPort)

	t.Logf("Creating VM: %s", vmName)
	vm, err := compute.NewVirtualMachineBuilder(vmName).
		WithVMInstanceType(string(e2etest.TestVMSize)).
		WithBootDisk(disk.Ref()).
		WithSecurityGroup(sg.Ref()).
		WithCloudInit(cloudInit).
		WithZone(e2etest.TestVMZone).
		Create(ctx, client.Compute().VirtualMachines())
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}
	t.Cleanup(func() {
		if !vmDeleted {
			t.Logf("Cleaning up VM: %s", vmName)
			client.Compute().VirtualMachines().Delete(ctx, vmName)
			client.Compute().VirtualMachines().WaitForDeleted(ctx, vmName, e2etest.VMDeletionTimeout)
		}
	})

	t.Logf("Waiting for VM to be ready...")
	if _, err := client.Compute().VirtualMachines().WaitForReady(ctx, vmName, e2etest.VMReadyTimeout); err != nil {
		t.Fatalf("VM never became ready: %v", err)
	}

	// --- Public IP ---
	t.Logf("Creating public IP: %s", ipName)
	_, err = client.Networking().PublicIPs().Create(ctx,
		networking.NewPublicIPBuilder(ipName).Build(),
	)
	if err != nil {
		t.Fatalf("failed to create public IP: %v", err)
	}
	t.Cleanup(func() {
		if !ipDeleted {
			t.Logf("Cleaning up public IP: %s", ipName)
			client.Networking().PublicIPs().Delete(ctx, ipName)
			client.Networking().PublicIPs().WaitForDeleted(ctx, ipName, deleteTimeout)
		}
	})

	t.Logf("Waiting for public IP to be ready...")
	readyIP, err := client.Networking().PublicIPs().WaitForReady(ctx, ipName, ipReadyTimeout)
	if err != nil {
		t.Fatalf("public IP never became ready: %v", err)
	}
	ipAddr := networking.GetPublicIPAddress(readyIP)
	t.Logf("Public IP allocated: %s", ipAddr)

	// --- LB cleanup registrations (LIFO: first registered = last to run) ---
	t.Cleanup(func() {
		if !poolDeleted {
			t.Logf("Cleaning up backend pool: %s", poolName)
			lb.BackendPools().Delete(ctx, poolName)
			lb.BackendPools().WaitForDeleted(ctx, poolName, deleteTimeout)
		}
	})
	t.Cleanup(func() {
		if !svcDeleted {
			t.Logf("Cleaning up backend service: %s", svcName)
			lb.BackendServices().Delete(ctx, svcName)
			lb.BackendServices().WaitForDeleted(ctx, svcName, deleteTimeout)
		}
	})
	t.Cleanup(func() {
		if !routeDeleted {
			t.Logf("Cleaning up L4 route: %s", routeName)
			lb.L4Routes().Delete(ctx, routeName)
			lb.L4Routes().WaitForDeleted(ctx, routeName, deleteTimeout)
		}
	})
	t.Cleanup(func() {
		if !lbDeleted {
			t.Logf("Cleaning up load balancer: %s", lbName)
			lb.LoadBalancers().Delete(ctx, lbName)
			lb.LoadBalancers().WaitForDeleted(ctx, lbName, deleteTimeout)
		}
	})

	// --- BackendPool (with the VM as backend) ---
	vmRef := string(vm.Ref())
	t.Logf("Creating backend pool: %s (backend: %s)", poolName, vmName)
	pool, err := loadbalancer.NewBackendPoolBuilder(poolName).
		WithBackendRefs([]string{vmRef}).
		Create(ctx, lb.BackendPools())
	if err != nil {
		t.Fatalf("failed to create backend pool: %v", err)
	}
	e2etest.MustGetID(t, pool.Metadata.Id, "backend pool")

	t.Logf("Getting backend pool: %s", poolName)
	pool, err = lb.BackendPools().Get(ctx, poolName)
	if err != nil {
		t.Fatalf("failed to get backend pool: %v", err)
	}
	if pool.Metadata.Id != poolName {
		t.Errorf("expected pool ID %s, got %s", poolName, pool.Metadata.Id)
	}

	t.Logf("Listing backend pools")
	pools, err := lb.BackendPools().List(ctx)
	if err != nil {
		t.Fatalf("failed to list backend pools: %v", err)
	}
	e2etest.AssertInList(t, pools.Items, poolName,
		func(p lbtypes.Backendpool) string { return p.Metadata.Id }, "backend pool")

	// --- BackendService ---
	t.Logf("Creating backend service: %s", svcName)
	svc, err := loadbalancer.NewBackendServiceBuilder(svcName).
		WithPort(backendPort).
		WithBackendPoolRef(lb.BackendPoolRef(poolName)).
		WithHTTPHealthCheck("/").
		Create(ctx, lb.BackendServices())
	if err != nil {
		t.Fatalf("failed to create backend service: %v", err)
	}
	e2etest.MustGetID(t, svc.Metadata.Id, "backend service")

	t.Logf("Getting backend service: %s", svcName)
	svc, err = lb.BackendServices().Get(ctx, svcName)
	if err != nil {
		t.Fatalf("failed to get backend service: %v", err)
	}
	if svc.Spec.Port != backendPort {
		t.Errorf("expected port %d, got %d", backendPort, svc.Spec.Port)
	}

	t.Logf("Listing backend services")
	svcs, err := lb.BackendServices().List(ctx)
	if err != nil {
		t.Fatalf("failed to list backend services: %v", err)
	}
	e2etest.AssertInList(t, svcs.Items, svcName,
		func(s lbtypes.Backendservice) string { return s.Metadata.Id }, "backend service")

	// --- L4Route ---
	t.Logf("Creating L4 route: %s", routeName)
	route, err := loadbalancer.NewL4RouteBuilder(routeName).
		WithBackendServiceRef(lb.BackendServiceRef(svcName)).
		Create(ctx, lb.L4Routes())
	if err != nil {
		t.Fatalf("failed to create L4 route: %v", err)
	}
	e2etest.MustGetID(t, route.Metadata.Id, "L4 route")

	t.Logf("Getting L4 route: %s", routeName)
	route, err = lb.L4Routes().Get(ctx, routeName)
	if err != nil {
		t.Fatalf("failed to get L4 route: %v", err)
	}
	if route.Spec.DefaultBackendServiceRef != lb.BackendServiceRef(svcName) {
		t.Errorf("expected backend service ref %s, got %s",
			lb.BackendServiceRef(svcName), route.Spec.DefaultBackendServiceRef)
	}

	t.Logf("Listing L4 routes")
	routes, err := lb.L4Routes().List(ctx)
	if err != nil {
		t.Fatalf("failed to list L4 routes: %v", err)
	}
	e2etest.AssertInList(t, routes.Items, routeName,
		func(r lbtypes.L4route) string { return r.Metadata.Id }, "L4 route")

	// --- LoadBalancer ---
	listenerName := "http"
	routeRefs := []string{lb.L4RouteRef(routeName)}

	t.Logf("Creating load balancer: %s", lbName)
	_, err = loadbalancer.NewLoadBalancerBuilder(lbName).
		WithPublicIPRef(string(readyIP.Ref())).
		WithListener(lbtypes.LoadbalancerSpecListenersItem{
			Name:      &listenerName,
			Port:      80,
			Protocol:  lbtypes.TCP,
			RouteRefs: &routeRefs,
		}).
		Create(ctx, lb.LoadBalancers())
	if err != nil {
		t.Fatalf("failed to create load balancer: %v", err)
	}

	t.Logf("Waiting for load balancer to be ready...")
	readyLB, err := lb.LoadBalancers().WaitForReady(ctx, lbName, lbReadyTimeout)
	if err != nil {
		t.Fatalf("load balancer never became ready: %v", err)
	}
	if !loadbalancer.IsReady(readyLB) {
		t.Error("expected load balancer to report ready")
	}

	t.Logf("Getting load balancer: %s", lbName)
	gotLB, err := lb.LoadBalancers().Get(ctx, lbName)
	if err != nil {
		t.Fatalf("failed to get load balancer: %v", err)
	}
	if gotLB.Metadata.Id != lbName {
		t.Errorf("expected LB ID %s, got %s", lbName, gotLB.Metadata.Id)
	}

	t.Logf("Listing load balancers")
	lbs, err := lb.LoadBalancers().List(ctx)
	if err != nil {
		t.Fatalf("failed to list load balancers: %v", err)
	}
	e2etest.AssertInList(t, lbs.Items, lbName,
		func(l lbtypes.Loadbalancer) string { return l.Metadata.Id }, "load balancer")

	// --- Teardown (reverse order: LB → route → svc → pool → IP → VM → disk) ---
	t.Logf("Deleting load balancer: %s", lbName)
	if err := lb.LoadBalancers().Delete(ctx, lbName); err != nil {
		t.Fatalf("failed to delete load balancer: %v", err)
	}
	if err := lb.LoadBalancers().WaitForDeleted(ctx, lbName, deleteTimeout); err != nil {
		t.Fatalf("load balancer not deleted in time: %v", err)
	}
	lbDeleted = true

	t.Logf("Deleting L4 route: %s", routeName)
	if err := lb.L4Routes().Delete(ctx, routeName); err != nil {
		t.Fatalf("failed to delete L4 route: %v", err)
	}
	if err := lb.L4Routes().WaitForDeleted(ctx, routeName, deleteTimeout); err != nil {
		t.Fatalf("L4 route not deleted in time: %v", err)
	}
	routeDeleted = true

	t.Logf("Deleting backend service: %s", svcName)
	if err := lb.BackendServices().Delete(ctx, svcName); err != nil {
		t.Fatalf("failed to delete backend service: %v", err)
	}
	if err := lb.BackendServices().WaitForDeleted(ctx, svcName, deleteTimeout); err != nil {
		t.Fatalf("backend service not deleted in time: %v", err)
	}
	svcDeleted = true

	t.Logf("Deleting backend pool: %s", poolName)
	if err := lb.BackendPools().Delete(ctx, poolName); err != nil {
		t.Fatalf("failed to delete backend pool: %v", err)
	}
	if err := lb.BackendPools().WaitForDeleted(ctx, poolName, deleteTimeout); err != nil {
		t.Fatalf("backend pool not deleted in time: %v", err)
	}
	poolDeleted = true

	t.Logf("Deleting public IP: %s", ipName)
	if err := client.Networking().PublicIPs().Delete(ctx, ipName); err != nil {
		t.Fatalf("failed to delete public IP: %v", err)
	}
	if err := client.Networking().PublicIPs().WaitForDeleted(ctx, ipName, deleteTimeout); err != nil {
		t.Fatalf("public IP not deleted in time: %v", err)
	}
	ipDeleted = true

	t.Logf("Deleting VM: %s", vmName)
	if err := client.Compute().VirtualMachines().Delete(ctx, vmName); err != nil {
		t.Fatalf("failed to delete VM: %v", err)
	}
	if err := client.Compute().VirtualMachines().WaitForDeleted(ctx, vmName, e2etest.VMDeletionTimeout); err != nil {
		t.Fatalf("VM not deleted in time: %v", err)
	}
	vmDeleted = true

	t.Logf("Deleting disk: %s", diskName)
	if err := client.Compute().Disks().Delete(ctx, diskName); err != nil {
		t.Fatalf("failed to delete disk: %v", err)
	}
	diskDeleted = true

	t.Logf("Deleting security group: %s", sgName)
	if err := client.Networking().SecurityGroups().Delete(ctx, sgName); err != nil {
		t.Fatalf("failed to delete security group: %v", err)
	}
	sgDeleted = true

	t.Log("LoadBalancer full stack E2E test completed successfully")
}
