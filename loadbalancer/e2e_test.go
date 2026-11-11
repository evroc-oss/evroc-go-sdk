// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package loadbalancer_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"
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
	certTimeout    = 3 * time.Minute
	httpsTimeout   = 5 * time.Minute
)

func newTestClient(t *testing.T) (*evroc.Client, *loadbalancer.Client) {
	t.Helper()
	client := e2etest.NewClient(t)
	return client, client.LoadBalancer()
}

// TestE2E_LoadBalancer_FullStack creates the full LB resource graph
// (IP + disk + VM → pool → service → L4 route + HTTP route + certificate → LB
// with TCP and HTTPS listeners), validates every resource via Get/List, checks
// that HTTPS traffic reaches the VM with the uploaded certificate, renews the
// certificate in place, then tears down in reverse order.
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
	httpRouteName := "httproute-" + suffix
	certName := "cert-" + suffix
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

		httpRouteDeleted bool
		certDeleted      bool
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
		if !certDeleted {
			t.Logf("Cleaning up certificate: %s", certName)
			lb.LoadBalancerCertificates().Delete(ctx, certName)
			lb.LoadBalancerCertificates().WaitForDeleted(ctx, certName, deleteTimeout)
		}
	})
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
		if !httpRouteDeleted {
			t.Logf("Cleaning up HTTP route: %s", httpRouteName)
			lb.HTTPRoutes().Delete(ctx, httpRouteName)
			lb.HTTPRoutes().WaitForDeleted(ctx, httpRouteName, deleteTimeout)
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
		WithIPProtocol(lbtypes.IPv4).
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
	if route.Spec.BackendServiceRef != lb.BackendServiceRef(svcName) {
		t.Errorf("expected backend service ref %s, got %s",
			lb.BackendServiceRef(svcName), route.Spec.BackendServiceRef)
	}

	t.Logf("Listing L4 routes")
	routes, err := lb.L4Routes().List(ctx)
	if err != nil {
		t.Fatalf("failed to list L4 routes: %v", err)
	}
	e2etest.AssertInList(t, routes.Items, routeName,
		func(r lbtypes.L4route) string { return r.Metadata.Id }, "L4 route")

	// --- LoadBalancerCertificate ---
	certKey, certPEM, keyPEM := selfSignedCert(t, nil)
	t.Logf("Creating certificate: %s", certName)
	if _, err := loadbalancer.NewLoadBalancerCertificateBuilder(certName).
		Create(ctx, lb.LoadBalancerCertificates()); err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	t.Logf("Uploading certificate content")
	if err := lb.LoadBalancerCertificates().UploadContent(ctx, certName, certPEM, keyPEM); err != nil {
		t.Fatalf("failed to upload certificate content: %v", err)
	}
	cert, err := lb.LoadBalancerCertificates().WaitForReady(ctx, certName, certTimeout)
	if err != nil {
		t.Fatalf("certificate never became ready: %v", err)
	}
	if got, want := deref(cert.Status.CertificateSha256Fingerprint), pemFingerprint(t, certPEM); !strings.EqualFold(got, want) {
		t.Errorf("certificate fingerprint = %s, want %s", got, want)
	}
	if cert.Status.NotAfter == nil {
		t.Error("expected certificate status to report notAfter")
	}

	t.Logf("Listing certificates")
	certs, err := lb.LoadBalancerCertificates().List(ctx)
	if err != nil {
		t.Fatalf("failed to list certificates: %v", err)
	}
	e2etest.AssertInList(t, certs.Items, certName,
		func(c lbtypes.Loadbalancercertificate) string { return c.Metadata.Id }, "certificate")

	// --- HTTPRoute ---
	t.Logf("Creating HTTP route: %s", httpRouteName)
	httpRoute, err := loadbalancer.NewHTTPRouteBuilder(httpRouteName).
		WithRule(lbtypes.HttprouteSpecRulesItem{Name: "all", BackendServiceRef: lb.BackendServiceRef(svcName)}).
		Create(ctx, lb.HTTPRoutes())
	if err != nil {
		t.Fatalf("failed to create HTTP route: %v", err)
	}
	e2etest.MustGetID(t, httpRoute.Metadata.Id, "HTTP route")

	t.Logf("Listing HTTP routes")
	httpRoutes, err := lb.HTTPRoutes().List(ctx)
	if err != nil {
		t.Fatalf("failed to list HTTP routes: %v", err)
	}
	e2etest.AssertInList(t, httpRoutes.Items, httpRouteName,
		func(r lbtypes.Httproute) string { return r.Metadata.Id }, "HTTP route")

	// --- LoadBalancer ---
	listenerName := "http"
	routeRefs := []string{lb.L4RouteRef(routeName)}
	httpsListenerName := "https"
	httpRouteRefs := []string{lb.HTTPRouteRef(httpRouteName)}

	t.Logf("Creating load balancer: %s", lbName)
	_, err = loadbalancer.NewLoadBalancerBuilder(lbName).
		WithPublicIPRef(string(readyIP.Ref())).
		WithListener(lbtypes.LoadbalancerSpecListenersItem{
			Name:      &listenerName,
			Port:      80,
			Protocol:  lbtypes.TCP,
			RouteRefs: &routeRefs,
		}).
		WithListener(lbtypes.LoadbalancerSpecListenersItem{
			Name:      &httpsListenerName,
			Port:      443,
			Protocol:  lbtypes.HTTPS,
			RouteRefs: &httpRouteRefs,
			Tls: &struct {
				CertificateRef string `json:"certificateRef"`
			}{CertificateRef: lb.LoadBalancerCertificateRef(certName)},
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

	// --- HTTPS traffic through the HTTPS listener and HTTP route ---
	t.Logf("Checking HTTPS traffic to %s", ipAddr)
	waitForServedCert(t, ctx, ipAddr, pemFingerprint(t, certPEM))

	// --- Renew the certificate in place, keeping the stored private key ---
	_, renewedPEM, _ := selfSignedCert(t, certKey)
	t.Logf("Renewing certificate: %s", certName)
	if err := lb.LoadBalancerCertificates().RenewCertificate(ctx, certName, renewedPEM); err != nil {
		t.Fatalf("failed to renew certificate: %v", err)
	}
	waitForServedCert(t, ctx, ipAddr, pemFingerprint(t, renewedPEM))
	renewed, err := lb.LoadBalancerCertificates().Get(ctx, certName)
	if err != nil {
		t.Fatalf("failed to get renewed certificate: %v", err)
	}
	if got := deref(renewed.Status.PublicKeySha256Fingerprint); got != deref(cert.Status.PublicKeySha256Fingerprint) {
		t.Errorf("renewal changed the public key fingerprint: %s -> %s", deref(cert.Status.PublicKeySha256Fingerprint), got)
	}

	// --- Teardown (reverse order: LB → routes → cert → svc → pool → IP → VM → disk) ---
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

	t.Logf("Deleting HTTP route: %s", httpRouteName)
	if err := lb.HTTPRoutes().Delete(ctx, httpRouteName); err != nil {
		t.Fatalf("failed to delete HTTP route: %v", err)
	}
	if err := lb.HTTPRoutes().WaitForDeleted(ctx, httpRouteName, deleteTimeout); err != nil {
		t.Fatalf("HTTP route not deleted in time: %v", err)
	}
	httpRouteDeleted = true

	t.Logf("Deleting certificate: %s", certName)
	if err := lb.LoadBalancerCertificates().Delete(ctx, certName); err != nil {
		t.Fatalf("failed to delete certificate: %v", err)
	}
	if err := lb.LoadBalancerCertificates().WaitForDeleted(ctx, certName, deleteTimeout); err != nil {
		t.Fatalf("certificate not deleted in time: %v", err)
	}
	certDeleted = true

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

// selfSignedCert returns a self-signed certificate for example.com, reusing key
// when given so a renewal keeps the same public key.
func selfSignedCert(t *testing.T, key *ecdsa.PrivateKey) (*ecdsa.PrivateKey, string, string) {
	t.Helper()
	if key == nil {
		var err error
		if key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
			t.Fatalf("generate key: %v", err)
		}
	}
	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "example.com"},
		DNSNames:     []string{"example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	return key,
		string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
}

func pemFingerprint(t *testing.T, certPEM string) string {
	t.Helper()
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		t.Fatal("invalid certificate PEM")
	}
	sum := sha256.Sum256(block.Bytes)
	return hex.EncodeToString(sum[:])
}

// waitForServedCert polls https://ip/ until the backend answers 200 and the LB
// presents the certificate with the given SHA-256 fingerprint.
func waitForServedCert(t *testing.T, ctx context.Context, ip, fingerprint string) {
	t.Helper()
	client := &http.Client{
		Timeout: 10 * time.Second,
		// The certificate is self-signed; the test checks its fingerprint instead.
		// A fresh connection per probe, so a renewed certificate is not hidden by
		// a kept-alive TLS session.
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true, ServerName: "example.com"},
			DisableKeepAlives: true,
		},
	}
	deadline := time.Now().Add(httpsTimeout)
	var last string
	for time.Now().Before(deadline) {
		resp, err := client.Get("https://" + ip + "/")
		if err != nil {
			last = err.Error()
		} else {
			resp.Body.Close()
			sum := sha256.Sum256(resp.TLS.PeerCertificates[0].Raw)
			served := hex.EncodeToString(sum[:])
			if resp.StatusCode == http.StatusOK && served == fingerprint {
				t.Logf("HTTPS OK, serving certificate %s", fingerprint[:16])
				return
			}
			last = fmt.Sprintf("status %d, certificate %s", resp.StatusCode, served[:16])
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Second):
		}
	}
	t.Fatalf("HTTPS never served certificate %s with status 200 (last: %s)", fingerprint[:16], last)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
