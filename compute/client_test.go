// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package compute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/compute"
)

type mockContextProvider struct{}

func (m *mockContextProvider) DefaultProject() string      { return "test-project" }
func (m *mockContextProvider) DefaultRegion() string       { return "test-region" }
func (m *mockContextProvider) DefaultOrganization() string { return "test-org" }

func setupClient(t *testing.T) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"VirtualMachine","metadata":{"id":"test"},"spec":{},"status":{}}`))
	}))

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	return NewClientWithParent(restClient, &mockContextProvider{}), server
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	client, server := setupClient(t)
	defer server.Close()

	if client.VirtualMachines() == nil || client.Disks() == nil || client.HotswapDiskAttachments() == nil || client.PlacementGroups() == nil {
		t.Fatal("service getters failed")
	}

	vmReq := NewVirtualMachineBuilder("test").Build()
	if _, err := client.VirtualMachines().Create(ctx, vmReq); err != nil {
		t.Errorf("VM Create: %v", err)
	}
	if _, err := client.VirtualMachines().Get(ctx, "test"); err != nil {
		t.Errorf("VM Get: %v", err)
	}
	if _, err := client.VirtualMachines().List(ctx); err != nil {
		t.Errorf("VM List: %v", err)
	}
	if err := client.VirtualMachines().Delete(ctx, "test"); err != nil {
		t.Errorf("VM Delete: %v", err)
	}
	if _, err := client.VirtualMachines().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("VM Patch: %v", err)
	}
	if _, err := client.VirtualMachines().Update(ctx, "test", &compute.VirtualMachine{ApiVersion: builderAPIVersion, Kind: "VirtualMachine"}); err != nil {
		t.Errorf("VM Update: %v", err)
	}

	diskReq := NewDiskBuilder("test").Build()
	if _, err := client.Disks().Create(ctx, diskReq); err != nil {
		t.Errorf("Disk Create: %v", err)
	}
	if _, err := client.Disks().Get(ctx, "test"); err != nil {
		t.Errorf("Disk Get: %v", err)
	}
	if _, err := client.Disks().List(ctx); err != nil {
		t.Errorf("Disk List: %v", err)
	}
	if err := client.Disks().Delete(ctx, "test"); err != nil {
		t.Errorf("Disk Delete: %v", err)
	}

	// Test builder Create methods and additional builder options
	if _, err := NewDiskBuilder("test2").Create(ctx, client.Disks()); err != nil {
		t.Errorf("Disk builder Create: %v", err)
	}
	if _, err := NewVirtualMachineBuilder("test2").WithDataDisk("data-disk").WithPlacementGroup("pg").WithRunning(true).Create(ctx, client.VirtualMachines()); err != nil {
		t.Errorf("VM builder Create: %v", err)
	}

	// Test Placement Groups
	pgReq := NewPlacementGroupBuilder("test", "spread").WithZone("a").Build()
	if _, err := client.PlacementGroups().Create(ctx, pgReq); err != nil {
		t.Errorf("PG Create: %v", err)
	}
	if _, err := client.PlacementGroups().Get(ctx, "test"); err != nil {
		t.Errorf("PG Get: %v", err)
	}
	if _, err := client.PlacementGroups().List(ctx); err != nil {
		t.Errorf("PG List: %v", err)
	}
	if err := client.PlacementGroups().Delete(ctx, "test"); err != nil {
		t.Errorf("PG Delete: %v", err)
	}

	// Test HotswapDiskAttachments
	hotswapReq := NewHotswapDiskAttachmentBuilder("test", "vm", "disk").Build()
	if _, err := client.HotswapDiskAttachments().Create(ctx, hotswapReq); err != nil {
		t.Errorf("Hotswap Create: %v", err)
	}
	if _, err := client.HotswapDiskAttachments().Get(ctx, "test"); err != nil {
		t.Errorf("Hotswap Get: %v", err)
	}
	if _, err := client.HotswapDiskAttachments().List(ctx); err != nil {
		t.Errorf("Hotswap List: %v", err)
	}
	if err := client.HotswapDiskAttachments().Delete(ctx, "test"); err != nil {
		t.Errorf("Hotswap Delete: %v", err)
	}

	//Test Patch and Update
	if _, err := client.Disks().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("Disk Patch: %v", err)
	}
	if _, err := client.Disks().Update(ctx, "test", &compute.Disk{ApiVersion: builderAPIVersion, Kind: "Disk"}); err != nil {
		t.Errorf("Disk Update: %v", err)
	}

	// Test Update builders
	if UpdateVM("test", client.VirtualMachines()) == nil || UpdateDisk("test", client.Disks()) == nil {
		t.Error("Update builders failed")
	}

	// Test validators and constants
	if !IsValidDiskImage(string(DiskImageUbuntu2404)) {
		t.Error("IsValidDiskImage failed")
	}
	if GetValidDiskImagesString() == "" {
		t.Error("GetValidDiskImagesString failed")
	}
	if !IsValidVMSize("c1a.m") {
		t.Error("IsValidVMSize failed")
	}
	if GetValidVMSizesString() == "" {
		t.Error("GetValidVMSizesString failed")
	}
	if FindVMInstanceType(2, 4, 0) == "" {
		t.Error("FindVMInstanceType failed")
	}

	if NewClient(client.rest) == nil {
		t.Error("NewClient failed")
	}

	if _, err := NewHotswapDiskAttachmentBuilder("test", "vm", "disk").Create(ctx, client.HotswapDiskAttachments()); err != nil {
		t.Errorf("Hotswap builder Create: %v", err)
	}
	if _, err := NewPlacementGroupBuilder("test", "spread").Create(ctx, client.PlacementGroups()); err != nil {
		t.Errorf("PG builder Create: %v", err)
	}

	if _, err := client.HotswapDiskAttachments().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("Hotswap Patch: %v", err)
	}
	if _, err := client.HotswapDiskAttachments().Update(ctx, "test", &compute.HotswapDiskAttachment{ApiVersion: builderAPIVersion, Kind: "HotswapDiskAttachment"}); err != nil {
		t.Errorf("Hotswap Update: %v", err)
	}
	if _, err := client.PlacementGroups().Patch(ctx, "test", map[string]interface{}{}); err != nil {
		t.Errorf("PG Patch: %v", err)
	}
	if _, err := client.PlacementGroups().Update(ctx, "test", &compute.PlacementGroup{ApiVersion: builderAPIVersion, Kind: "PlacementGroup"}); err != nil {
		t.Errorf("PG Update: %v", err)
	}
}
