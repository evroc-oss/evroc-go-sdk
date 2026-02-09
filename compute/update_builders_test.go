// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package compute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

func TestVMUpdateBuilderApply(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"apiVersion":"compute/v1alpha3","kind":"VirtualMachine","metadata":{"id":"test"},"spec":{"size":"c1a.m"},"status":{}}`

		switch r.Method {
		case http.MethodGet, http.MethodPatch, http.MethodPut:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClientWithParent(restClient, &mockContextProvider{})

	ctx := context.Background()

	t.Run("Resize", func(t *testing.T) {
		_, err := NewVirtualMachineUpdateBuilder("test", client.VirtualMachines()).Resize("c1a.l").Apply(ctx)
		if err != nil {
			t.Errorf("VM Resize Apply failed: %v", err)
		}
	})

	t.Run("AddSecurityGroup", func(t *testing.T) {
		_, err := NewVirtualMachineUpdateBuilder("test", client.VirtualMachines()).AddSecurityGroup("sg1").Apply(ctx)
		if err != nil {
			t.Errorf("VM AddSecurityGroup Apply failed: %v", err)
		}
	})

	t.Run("AddLabel", func(t *testing.T) {
		_, err := NewVirtualMachineUpdateBuilder("test", client.VirtualMachines()).AddLabel("env", "prod").Apply(ctx)
		if err != nil {
			t.Errorf("VM AddLabel Apply failed: %v", err)
		}
	})
}

func TestDiskUpdateBuilderApply(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"apiVersion":"compute/v1alpha3","kind":"Disk","metadata":{"id":"test"},"spec":{"sizeGB":50},"status":{}}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClientWithParent(restClient, &mockContextProvider{})
	ctx := context.Background()

	_, err := NewDiskUpdateBuilder("test", client.Disks()).ResizeGB(100).Apply(ctx)
	if err != nil {
		t.Errorf("Disk ResizeGB Apply failed: %v", err)
	}
}

func TestUpdateBuilderMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		diskPath := "/compute/" + apiVersion + "/projects/test-project/regions/test-region/disks/test"
		if r.URL.Path == diskPath {
			w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"Disk","metadata":{"id":"test"},"spec":{},"status":{}}`))
		} else {
			w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"VirtualMachine","metadata":{"id":"test"},"spec":{},"status":{}}`))
		}
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClientWithParent(restClient, &mockContextProvider{})

	// Test VM update builder methods (coverage)
	vmUpdate := NewVirtualMachineUpdateBuilder("test", client.VirtualMachines())
	vmUpdate.Stop().Start().RemoveLabel("old").RemoveSecurityGroup("old-sg")

	// Test Disk update builder methods (coverage)
	diskUpdate := NewDiskUpdateBuilder("test", client.Disks())
	diskUpdate.AddLabel("key", "value").RemoveLabel("old")
}
