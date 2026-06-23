// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	lbtypes "github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

type mockContext struct{}

func (m *mockContext) DefaultProject() string      { return "test-project" }
func (m *mockContext) DefaultRegion() string       { return "test-region" }
func (m *mockContext) DefaultOrganization() string { return "test-org" }

func setupTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClient(restClient, &mockContext{}), server
}

func lbMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		switch {
		case strings.Contains(path, "/backendPools"):
			pool := &lbtypes.Backendpool{
				ApiVersion: builderAPIVersion,
				Kind:       "BackendPool",
				Metadata:   lbtypes.RegionalMetadataResponse{Id: "test-pool"},
				Spec:       lbtypes.BackendpoolSpec{BackendRefs: &[]string{"vm-1"}},
			}
			switch r.Method {
			case http.MethodPost:
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(pool)
			case http.MethodGet:
				if strings.HasSuffix(path, "/backendPools") {
					json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{pool}})
				} else {
					json.NewEncoder(w).Encode(pool)
				}
			case http.MethodPatch:
				json.NewEncoder(w).Encode(pool)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}

		case strings.Contains(path, "/backendServices"):
			svc := &lbtypes.Backendservice{
				ApiVersion: builderAPIVersion,
				Kind:       "BackendService",
				Metadata:   lbtypes.RegionalMetadataResponse{Id: "test-svc"},
				Spec:       lbtypes.BackendserviceSpec{Port: 8080},
			}
			switch r.Method {
			case http.MethodPost:
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(svc)
			case http.MethodGet:
				if strings.HasSuffix(path, "/backendServices") {
					json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{svc}})
				} else {
					json.NewEncoder(w).Encode(svc)
				}
			case http.MethodPatch:
				json.NewEncoder(w).Encode(svc)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}

		case strings.Contains(path, "/l4Routes"):
			route := &lbtypes.L4route{
				ApiVersion: builderAPIVersion,
				Kind:       "L4Route",
				Metadata:   lbtypes.RegionalMetadataResponse{Id: "test-route"},
				Spec:       lbtypes.L4routeSpec{DefaultBackendServiceRef: "/loadbalancer/projects/p/regions/r/backendServices/test-svc"},
			}
			switch r.Method {
			case http.MethodPost:
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(route)
			case http.MethodGet:
				if strings.HasSuffix(path, "/l4Routes") {
					json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{route}})
				} else {
					json.NewEncoder(w).Encode(route)
				}
			case http.MethodPatch:
				json.NewEncoder(w).Encode(route)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}

		case strings.Contains(path, "/loadBalancers"):
			lb := &lbtypes.Loadbalancer{
				ApiVersion: builderAPIVersion,
				Kind:       "LoadBalancer",
				Metadata:   lbtypes.RegionalMetadataResponse{Id: "test-lb"},
			}
			switch r.Method {
			case http.MethodPost:
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(lb)
			case http.MethodGet:
				if strings.HasSuffix(path, "/loadBalancers") {
					json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{lb}})
				} else {
					json.NewEncoder(w).Encode(lb)
				}
			case http.MethodPatch:
				json.NewEncoder(w).Encode(lb)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	return mux
}

func TestLoadBalancers_CRUD(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	lb, err := client.LoadBalancers().Create(ctx, &lbtypes.LoadbalancerRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "LoadBalancer",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: "test-lb"},
		Spec:       lbtypes.LoadbalancerSpec{},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if lb.Metadata.Id != "test-lb" {
		t.Errorf("expected ID 'test-lb', got %s", lb.Metadata.Id)
	}

	if _, err := client.LoadBalancers().Get(ctx, "test-lb"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	list, err := client.LoadBalancers().List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(list.Items))
	}

	if err := client.LoadBalancers().Delete(ctx, "test-lb"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestBackendPools_CRUD(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	pool, err := client.BackendPools().Create(ctx, &lbtypes.BackendpoolRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "BackendPool",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: "test-pool"},
		Spec:       lbtypes.BackendpoolSpec{BackendRefs: &[]string{"vm-1"}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pool.Kind != "BackendPool" {
		t.Errorf("expected Kind 'BackendPool', got %s", pool.Kind)
	}
}

func TestIsReady(t *testing.T) {
	if IsReady(nil) {
		t.Error("nil should not be ready")
	}

	lb := &lbtypes.Loadbalancer{}
	if IsReady(lb) {
		t.Error("no conditions should not be ready")
	}

	conditions := []lbtypes.LoadbalancerStatusConditionsItem{
		{Type: "Ready", Status: "True"},
	}
	lb.Status.Conditions = &conditions
	if !IsReady(lb) {
		t.Error("Ready=True should be ready")
	}
}

func TestRefHelpers(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()

	if got := client.LoadBalancerRef("my-lb"); got != "/loadbalancer/projects/test-project/regions/test-region/loadBalancers/my-lb" {
		t.Errorf("LoadBalancerRef = %q", got)
	}
	if got := client.BackendPoolRef("my-pool"); got != "/loadbalancer/projects/test-project/regions/test-region/backendPools/my-pool" {
		t.Errorf("BackendPoolRef = %q", got)
	}
	if got := client.BackendServiceRef("my-svc"); got != "/loadbalancer/projects/test-project/regions/test-region/backendServices/my-svc" {
		t.Errorf("BackendServiceRef = %q", got)
	}
	if got := client.L4RouteRef("my-route"); got != "/loadbalancer/projects/test-project/regions/test-region/l4Routes/my-route" {
		t.Errorf("L4RouteRef = %q", got)
	}
}

func TestLoadBalancerBuilder(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()

	name := "http"
	req := NewLoadBalancerBuilder("web-lb").
		WithPublicIPRef("/networking/projects/p/regions/r/publicIPs/ip").
		WithListener(lbtypes.LoadbalancerSpecListenersItem{
			Name:     &name,
			Port:     80,
			Protocol: lbtypes.TCP,
		}).
		WithLabels(map[string]string{"app": "web"}).
		Build()

	if req.Metadata.Id != "web-lb" {
		t.Errorf("expected ID 'web-lb', got %s", req.Metadata.Id)
	}
	if req.ApiVersion != builderAPIVersion {
		t.Errorf("expected ApiVersion %s, got %s", builderAPIVersion, req.ApiVersion)
	}
	if req.Spec.PublicIPRef != "/networking/projects/p/regions/r/publicIPs/ip" {
		t.Error("PublicIPRef not set")
	}
	if req.Spec.Listeners == nil || len(*req.Spec.Listeners) != 1 {
		t.Fatal("expected 1 listener")
	}
	if req.Metadata.UserLabels == nil {
		t.Fatal("expected labels")
	}
	if (*req.Metadata.UserLabels)["app"] != "web" {
		t.Error("label 'app' not set")
	}

	lb, err := NewLoadBalancerBuilder("test-lb").
		Create(context.Background(), client.LoadBalancers())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if lb.Metadata.Id != "test-lb" {
		t.Errorf("expected ID 'test-lb', got %s", lb.Metadata.Id)
	}
}

func TestBackendPoolBuilder(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()

	req := NewBackendPoolBuilder("my-pool").
		WithBackendRefs([]string{"vm-1", "vm-2"}).
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	if req.Metadata.Id != "my-pool" {
		t.Errorf("expected ID 'my-pool', got %s", req.Metadata.Id)
	}
	if req.Spec.BackendRefs == nil || len(*req.Spec.BackendRefs) != 2 {
		t.Fatal("expected 2 backend refs")
	}

	pool, err := NewBackendPoolBuilder("test-pool").
		WithBackendRefs([]string{"vm-1"}).
		Create(context.Background(), client.BackendPools())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pool.Metadata.Id != "test-pool" {
		t.Errorf("expected ID 'test-pool', got %s", pool.Metadata.Id)
	}
}

func TestBackendServiceBuilder(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()

	req := NewBackendServiceBuilder("my-svc").
		WithPort(8080).
		WithBackendPoolRef("/loadbalancer/projects/p/regions/r/backendPools/my-pool").
		WithProxyProtocol(true).
		Build()

	if req.Metadata.Id != "my-svc" {
		t.Errorf("expected ID 'my-svc', got %s", req.Metadata.Id)
	}
	if req.Spec.Port != 8080 {
		t.Errorf("expected port 8080, got %d", req.Spec.Port)
	}
	if req.Spec.BackendPoolRef == nil || *req.Spec.BackendPoolRef != "/loadbalancer/projects/p/regions/r/backendPools/my-pool" {
		t.Error("BackendPoolRef not set")
	}
	if req.Spec.ProxyProtocol == nil || !*req.Spec.ProxyProtocol {
		t.Error("ProxyProtocol not set")
	}

	svc, err := NewBackendServiceBuilder("test-svc").
		WithPort(8080).
		Create(context.Background(), client.BackendServices())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if svc.Metadata.Id != "test-svc" {
		t.Errorf("expected ID 'test-svc', got %s", svc.Metadata.Id)
	}
}

func TestL4RouteBuilder(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()

	req := NewL4RouteBuilder("my-route").
		WithBackendServiceRef("/loadbalancer/projects/p/regions/r/backendServices/my-svc").
		WithLabels(map[string]string{"env": "prod"}).
		Build()

	if req.Metadata.Id != "my-route" {
		t.Errorf("expected ID 'my-route', got %s", req.Metadata.Id)
	}
	if req.Spec.DefaultBackendServiceRef != "/loadbalancer/projects/p/regions/r/backendServices/my-svc" {
		t.Error("DefaultBackendServiceRef not set")
	}
	if req.Metadata.UserLabels == nil || (*req.Metadata.UserLabels)["env"] != "prod" {
		t.Error("labels not set on L4Route")
	}

	route, err := NewL4RouteBuilder("test-route").
		WithBackendServiceRef("/loadbalancer/projects/p/regions/r/backendServices/test-svc").
		Create(context.Background(), client.L4Routes())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if route.Metadata.Id != "test-route" {
		t.Errorf("expected ID 'test-route', got %s", route.Metadata.Id)
	}
}

func TestBackendServiceBuilder_WithLabels(t *testing.T) {
	req := NewBackendServiceBuilder("svc").
		WithPort(80).
		WithLabels(map[string]string{"team": "infra"}).
		Build()

	if req.Metadata.UserLabels == nil || (*req.Metadata.UserLabels)["team"] != "infra" {
		t.Error("labels not set on BackendService")
	}
}

func TestBackendServices_CRUD(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	svc, err := client.BackendServices().Create(ctx, &lbtypes.BackendserviceRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "BackendService",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: "test-svc"},
		Spec:       lbtypes.BackendserviceSpec{Port: 8080},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if svc.Metadata.Id != "test-svc" {
		t.Errorf("expected ID 'test-svc', got %s", svc.Metadata.Id)
	}

	if _, err := client.BackendServices().Get(ctx, "test-svc"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if err := client.BackendServices().Delete(ctx, "test-svc"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestL4Routes_CRUD(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	route, err := client.L4Routes().Create(ctx, &lbtypes.L4routeRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "L4Route",
		Metadata:   lbtypes.RegionalMetadataRequest{Id: "test-route"},
		Spec:       lbtypes.L4routeSpec{DefaultBackendServiceRef: "ref"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if route.Metadata.Id != "test-route" {
		t.Errorf("expected ID 'test-route', got %s", route.Metadata.Id)
	}

	if _, err := client.L4Routes().Get(ctx, "test-route"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if err := client.L4Routes().Delete(ctx, "test-route"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestBackendPools_GetDeleteListPatch(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	if _, err := client.BackendPools().Get(ctx, "test-pool"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if err := client.BackendPools().Delete(ctx, "test-pool"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	list, err := client.BackendPools().List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if list == nil {
		t.Fatal("expected non-nil list")
	}
	if _, err := client.BackendPools().Patch(ctx, "test-pool", map[string]interface{}{
		"spec": lbtypes.BackendpoolSpec{BackendRefs: &[]string{"vm-1", "vm-2"}},
	}); err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestBackendServices_ListPatch(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	list, err := client.BackendServices().List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if list == nil {
		t.Fatal("expected non-nil list")
	}
	if _, err := client.BackendServices().Patch(ctx, "test-svc", map[string]interface{}{
		"spec": lbtypes.BackendserviceSpec{Port: 9090},
	}); err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestL4Routes_ListPatch(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	list, err := client.L4Routes().List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if list == nil {
		t.Fatal("expected non-nil list")
	}
	if _, err := client.L4Routes().Patch(ctx, "test-route", map[string]interface{}{
		"spec": lbtypes.L4routeSpec{DefaultBackendServiceRef: "new-ref"},
	}); err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestLoadBalancers_Patch(t *testing.T) {
	client, server := setupTestClient(t, lbMux())
	defer server.Close()
	ctx := context.Background()

	if _, err := client.LoadBalancers().Patch(ctx, "test-lb", map[string]interface{}{
		"spec": lbtypes.LoadbalancerSpec{PublicIPRef: "new-ref"},
	}); err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestWaiterOptions(t *testing.T) {
	opt := WithExponentialBackoff(1*time.Second, 10*time.Second, 2.0)
	if opt == nil {
		t.Fatal("expected non-nil option")
	}

	opt = WithProgressCallback(func(attempt int, elapsed time.Duration) {})
	if opt == nil {
		t.Fatal("expected non-nil option")
	}
}

func TestWaitForReady(t *testing.T) {
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		lb := &lbtypes.Loadbalancer{
			ApiVersion: builderAPIVersion,
			Kind:       "LoadBalancer",
			Metadata:   lbtypes.RegionalMetadataResponse{Id: "test-lb"},
		}
		if calls >= 2 {
			conditions := []lbtypes.LoadbalancerStatusConditionsItem{
				{Type: "Ready", Status: "True"},
			}
			lb.Status.Conditions = &conditions
		}
		json.NewEncoder(w).Encode(lb)
	})

	client, server := setupTestClient(t, mux)
	defer server.Close()

	lb, err := client.LoadBalancers().WaitForReady(context.Background(), "test-lb", 10*time.Second, WithPollingInterval(100*time.Millisecond))
	if err != nil {
		t.Fatalf("WaitForReady: %v", err)
	}
	if !IsReady(lb) {
		t.Error("expected LB to be ready")
	}
	if calls < 2 {
		t.Errorf("expected at least 2 polls, got %d", calls)
	}
}

func TestWaitForDeleted(t *testing.T) {
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		if calls >= 2 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
			return
		}
		lb := &lbtypes.Loadbalancer{
			ApiVersion: builderAPIVersion,
			Kind:       "LoadBalancer",
			Metadata:   lbtypes.RegionalMetadataResponse{Id: "test-lb"},
		}
		json.NewEncoder(w).Encode(lb)
	})

	client, server := setupTestClient(t, mux)
	defer server.Close()

	err := client.LoadBalancers().WaitForDeleted(context.Background(), "test-lb", 10*time.Second, WithPollingInterval(100*time.Millisecond))
	if err != nil {
		t.Fatalf("WaitForDeleted: %v", err)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 polls, got %d", calls)
	}
}
