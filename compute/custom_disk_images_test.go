// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"net/http"
	"testing"
	"time"

	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

func TestDiskBuilderSourceSelection(t *testing.T) {
	ref := CustomDiskImageRef("/compute/projects/p/regions/r/customDiskImages/my-image")
	setters := map[string]func(*DiskBuilder) *DiskBuilder{
		"standard": func(b *DiskBuilder) *DiskBuilder { return b.WithImage("ubuntu.24-04.1") },
		"custom":   func(b *DiskBuilder) *DiskBuilder { return b.WithCustomImage(ref) },
		"snapshot": func(b *DiskBuilder) *DiskBuilder {
			return b.WithSnapshot("/compute/projects/p/regions/r/snapshots/snap")
		},
		"blank": func(b *DiskBuilder) *DiskBuilder { return b.WithCustomImage("") },
	}
	for first, setFirst := range setters {
		for last, setLast := range setters {
			t.Run(first+" then "+last, func(t *testing.T) {
				req := setLast(setFirst(NewDiskBuilder("disk").WithZone("a").WithSizeGB(50))).Build()
				if req.Spec.Placement.Zone == nil || *req.Spec.Placement.Zone != "a" || req.Spec.DiskSize.Amount != 50 {
					t.Fatal("lost disk settings")
				}
				source := req.Spec.Source
				if last == "blank" {
					if source == nil || source.Type != computetypes.DiskSpecSourceTypeBlank || source.DiskImageRef != nil || source.SnapshotRef != nil {
						t.Fatalf("clearing source should produce a blank disk: %+v", source)
					}
					return
				}
				if source == nil {
					t.Fatal("missing source")
				}
				if last == "snapshot" {
					if source.Type != computetypes.DiskSpecSourceTypeSnapshot || source.DiskImageRef != nil || source.SnapshotRef == nil {
						t.Fatalf("invalid snapshot source: %+v", source)
					}
					return
				}
				want := "/compute/global/diskImages/evroc/ubuntu.24-04.1"
				if last == "custom" {
					want = ref.String()
				}
				if source.Type != computetypes.DiskSpecSourceTypeImage || source.SnapshotRef != nil || source.DiskImageRef == nil || *source.DiskImageRef != want {
					t.Fatalf("invalid image source: %+v, want %s", source, want)
				}
			})
		}
	}
}

func TestCustomDiskImageBuilder(t *testing.T) {
	req := NewCustomDiskImageBuilder("image", "bucket-ref", "account-ref", "images/node.qcow2").
		WithDefaultDiskSizeGB(50).WithObjectVersion("object-v2").WithDescription("my image").
		WithLabels(map[string]string{"team": "infra"}).Build()
	if req.Kind != "CustomDiskImage" || req.Metadata.Id != "image" || req.Spec.Architecture != "amd64" ||
		req.Spec.DefaultDiskSize.Amount != 50 || req.Spec.DefaultDiskSize.Unit != "GB" {
		t.Errorf("unexpected registration: %+v", req)
	}
	source := req.Spec.Source
	if source.BucketRef != "bucket-ref" || source.ServiceAccountRef != "account-ref" || source.Path != "images/node.qcow2" ||
		source.Version == nil || *source.Version != "object-v2" {
		t.Errorf("unexpected source: %+v", source)
	}
	if req.Spec.Description == nil || *req.Spec.Description != "my image" || req.Metadata.UserLabels == nil || (*req.Metadata.UserLabels)["team"] != "infra" {
		t.Errorf("missing metadata: %+v", req)
	}
}

func TestCustomDiskImageWaiters(t *testing.T) {
	ready := `{"metadata":{"id":"image"},"status":{"conditions":[{"type":"Ready","status":"True"}]}}`
	t.Run("ready", func(t *testing.T) {
		client, close := setupWaiter(t, ready, http.StatusOK)
		defer close()
		image, err := client.CustomDiskImages().WaitForReady(context.Background(), "image", time.Second, nil, WithPollingInterval(time.Millisecond))
		if err != nil || !IsCustomDiskImageReady(image) {
			t.Fatalf("image=%+v err=%v", image, err)
		}
	})
	t.Run("deleted", func(t *testing.T) {
		client, close := setupWaiter(t, "", http.StatusNotFound)
		defer close()
		if err := client.CustomDiskImages().WaitForDeleted(context.Background(), "image", time.Second, nil, WithPollingInterval(time.Millisecond)); err != nil {
			t.Fatal(err)
		}
	})
	if IsCustomDiskImageReady(nil) || IsCustomDiskImageReady(&computetypes.CustomDiskImage{}) {
		t.Fatal("absent readiness must not be ready")
	}
}

func TestWithImagePreservesSnapshotFallback(t *testing.T) {
	const snapshot = "/compute/projects/p/regions/r/snapshots/snap"
	req := NewDiskBuilder("disk").WithSnapshot(snapshot).WithImage("ubuntu.24-04.1").WithImage("").Build()
	if req.Spec.Source == nil || req.Spec.Source.SnapshotRef == nil || *req.Spec.Source.SnapshotRef != snapshot {
		t.Fatal("clearing a standard image must preserve the existing snapshot fallback")
	}
}
