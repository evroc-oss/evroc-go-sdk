// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package kms_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	"github.com/evroc-oss/evroc-go-sdk/kms"
	kmstypes "github.com/evroc-oss/evroc-go-sdk/types/kms"
)

// waitForKey polls until cond holds for the key or the timeout expires.
func waitForKey(t *testing.T, ctx context.Context, keys *kms.KeysService, id, what string, cond func(*kmstypes.Key) bool) *kmstypes.Key {
	t.Helper()
	start := time.Now()
	deadline := start.Add(e2etest.KMSStateTimeout)
	var key *kmstypes.Key
	for time.Now().Before(deadline) {
		var err error
		key, err = keys.Get(ctx, id)
		if err != nil {
			t.Fatalf("failed to get key: %v", err)
		}
		if cond(key) {
			t.Logf("  %s reached after %s", what, time.Since(start).Round(time.Millisecond))
			return key
		}
		t.Logf("  waiting for %s: state=%q versions=%s (%s)", what, key.Status.State, versionsSummary(key), time.Since(start).Round(time.Second))
		time.Sleep(e2etest.KMSStateInterval)
	}
	t.Fatalf("key %s: timed out waiting for %s (state %q)", id, what, key.Status.State)
	return nil
}

func versionsSummary(k *kmstypes.Key) string {
	if k.Status.Versions == nil {
		return "[]"
	}
	parts := make([]string, 0, len(*k.Status.Versions))
	for _, v := range *k.Status.Versions {
		parts = append(parts, fmt.Sprintf("v%d=%s", v.Version, v.State))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// pauseForReconcile works around the KMS proxy stamping its reconcile trigger with
// second resolution: a second action within the same second is never reconciled.
func pauseForReconcile() {
	time.Sleep(1100 * time.Millisecond)
}

func inState(state string) func(*kmstypes.Key) bool {
	return func(k *kmstypes.Key) bool { return k.Status.State == state }
}

func versionState(version int64, state string) func(*kmstypes.Key) bool {
	return func(k *kmstypes.Key) bool {
		if k.Status.Versions == nil {
			return false
		}
		for _, v := range *k.Status.Versions {
			if v.Version == version {
				return v.State == state
			}
		}
		return false
	}
}

func TestE2E_Key_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	keys := client.KMS().Keys()
	name := e2etest.RandomName("key")

	// Create key
	t.Logf("Creating key: %s", name)
	createStart := time.Now()
	key, err := keys.Create(ctx, &kmstypes.KeyRequest{
		ApiVersion: "kms/v1",
		Kind:       "Key",
		Metadata:   kmstypes.RegionalMetadataRequest{Id: name},
		Spec:       kmstypes.KeySpec{Algorithm: kmstypes.Aes256Gcm96},
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	keyID := e2etest.MustGetID(t, key.Metadata.Id, "key")
	t.Logf("  create returned after %s, state=%q", time.Since(createStart).Round(time.Millisecond), key.Status.State)
	// Delete only marks the key for deletion; it is purged after the grace period.
	deleted := false
	e2etest.DeferCleanup(t, ctx, keys.Delete, keyID, "key", &deleted)

	key = waitForKey(t, ctx, keys, keyID, "PreActivated or Active", func(k *kmstypes.Key) bool {
		return k.Status.State == "PreActivated" || k.Status.State == "Active"
	})
	if key.Status.State == "PreActivated" {
		t.Logf("Activating key: %s", keyID)
		pauseForReconcile()
		if err := keys.Activate(ctx, keyID); err != nil {
			t.Fatalf("failed to activate key: %v", err)
		}
		waitForKey(t, ctx, keys, keyID, "Active", inState("Active"))
	}

	// List keys
	list, err := keys.List(ctx)
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}
	e2etest.AssertInList(t, list.Items, keyID, func(k kmstypes.Key) string { return k.Metadata.Id }, "key")

	// Suspend and reactivate
	t.Logf("Suspending key: %s", keyID)
	pauseForReconcile()
	if err := keys.Suspend(ctx, keyID); err != nil {
		t.Fatalf("failed to suspend key: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "Suspended", inState("Suspended"))

	t.Logf("Activating key: %s", keyID)
	pauseForReconcile()
	if err := keys.Activate(ctx, keyID); err != nil {
		t.Fatalf("failed to activate key: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "Active", inState("Active"))

	// Rotate, then trim the first version and abort the trim
	t.Logf("Rotating key: %s", keyID)
	pauseForReconcile()
	if err := keys.Rotate(ctx, keyID); err != nil {
		t.Fatalf("failed to rotate key: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "version 2", versionState(2, "Active"))

	t.Logf("Trimming key versions below 2: %s", keyID)
	pauseForReconcile()
	if err := keys.Trim(ctx, keyID, 2); err != nil {
		t.Fatalf("failed to trim key: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "version 1 PendingTrim", versionState(1, "PendingTrim"))

	t.Logf("Aborting trim: %s", keyID)
	pauseForReconcile()
	if err := keys.AbortTrimming(ctx, keyID); err != nil {
		t.Fatalf("failed to abort trim: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "version 1 Active", versionState(1, "Active"))

	// Delete, abort deletion (lands in Suspended), then delete for good
	t.Logf("Deleting key: %s", keyID)
	pauseForReconcile()
	if err := keys.Delete(ctx, keyID); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "PendingDeletion", inState("PendingDeletion"))

	t.Logf("Aborting deletion: %s", keyID)
	pauseForReconcile()
	if err := keys.AbortDeletion(ctx, keyID); err != nil {
		t.Fatalf("failed to abort deletion: %v", err)
	}
	waitForKey(t, ctx, keys, keyID, "Suspended", inState("Suspended"))

	t.Logf("Deleting key: %s", keyID)
	pauseForReconcile()
	if err := keys.Delete(ctx, keyID); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}
	deleted = true
	waitForKey(t, ctx, keys, keyID, "PendingDeletion", inState("PendingDeletion"))
}
