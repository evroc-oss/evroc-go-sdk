// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/compute"
)

var logger *slog.Logger

func init() {
	level := slog.LevelError
	if os.Getenv("EVROC_SDK_DEBUG") != "" {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}

// WaiterOption allows customizing waiter behavior.
type WaiterOption func(*rest.WaiterConfig)

// WithPollingInterval sets a constant polling interval (disables exponential backoff).
func WithPollingInterval(interval time.Duration) WaiterOption {
	return func(cfg *rest.WaiterConfig) {
		cfg.InitialInterval = interval
		cfg.MaxInterval = interval
		cfg.Multiplier = 1.0
	}
}

// WithExponentialBackoff configures exponential backoff for polling.
func WithExponentialBackoff(initial, max time.Duration, multiplier float64) WaiterOption {
	return func(cfg *rest.WaiterConfig) {
		cfg.InitialInterval = initial
		cfg.MaxInterval = max
		cfg.Multiplier = multiplier
	}
}

// WithProgressCallback sets a callback to track wait progress.
func WithProgressCallback(callback func(attempt int, elapsed time.Duration)) WaiterOption {
	return func(cfg *rest.WaiterConfig) {
		cfg.ProgressCallback = callback
	}
}

// WaitForReady polls the disk until it has a Ready condition with status True.
// Returns the ready disk resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *DisksService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*compute.Disk, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "disk"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *compute.Disk
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		disk, err := s.Get(ctx, name)
		if err != nil {
			logger.Debug("disk get error during wait",
				"disk", name,
				"error", err)
			// Transient errors during polling are not fatal
			// The waiter will continue polling until timeout
			return false, nil
		}
		ready := IsDiskReady(disk)

		// Get ready condition status for logging
		readyCondition := "unknown"
		if disk.Status.Conditions != nil {
			for _, cond := range *disk.Status.Conditions {
				if cond.Type == "Ready" {
					readyCondition = string(cond.Status)
					break
				}
			}
		}

		logger.Debug("disk status check",
			"disk", name,
			"ready", ready,
			"ready_condition", readyCondition)

		if ready {
			result = disk
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForReady polls the VM until it has a Ready condition with status True.
// Returns the ready VM resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *VirtualMachinesService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*compute.VirtualMachine, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "vm"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *compute.VirtualMachine
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		vm, err := s.Get(ctx, name)
		if err != nil {
			// Transient errors during polling are not fatal
			return false, nil
		}

		// Check for provisioning failures - these are terminal errors
		if vm.Status.VirtualMachineStatus != nil {
			status := *vm.Status.VirtualMachineStatus
			if status == "ProvisioningFailed" {
				var errMsg string
				if vm.Status.Conditions != nil {
					for _, cond := range *vm.Status.Conditions {
						if cond.Status == "False" {
							errMsg += fmt.Sprintf("\n  - %s: %s (%s)", cond.Type, cond.Message, cond.Reason)
						}
					}
				}
				return false, fmt.Errorf("VM %s provisioning failed:%s", name, errMsg)
			}
		}

		ready := IsVMReady(vm)
		if ready {
			result = vm
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForDeleted polls until the disk returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *DisksService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "disk"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Resource deleted successfully
			return true, nil
		}
		// Still exists, keep polling
		return false, nil
	})
}

// WaitForDeleted polls until the VM returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *VirtualMachinesService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "vm"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Resource deleted successfully
			return true, nil
		}
		// Still exists, keep polling
		return false, nil
	})
}

// WaitForStopped polls the VM until its status is "Stopped".
// Returns the stopped VM resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *VirtualMachinesService) WaitForStopped(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*compute.VirtualMachine, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "vm"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *compute.VirtualMachine
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		vm, err := s.Get(ctx, name)
		if err != nil {
			return false, nil
		}
		if vm.Status.VirtualMachineStatus != nil && *vm.Status.VirtualMachineStatus == "Stopped" {
			result = vm
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForReady polls the hotswap disk attachment until it has a Ready condition.
// Returns the ready hotswap attachment resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *HotswapDiskAttachmentsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*compute.HotswapDiskAttachment, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "hotswap_attachment"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *compute.HotswapDiskAttachment
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		attachment, err := s.Get(ctx, name)
		if err != nil {
			// Transient errors during polling are not fatal
			return false, nil
		}
		ready := IsAttachmentReady(attachment)
		if ready {
			result = attachment
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForReady polls the placement group until it has a Ready condition.
// Returns the ready placement group resource.
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *PlacementGroupsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*compute.PlacementGroup, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "placement_group"
	config.Metrics = s.client.metrics

	// Apply custom options
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *compute.PlacementGroup
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		pg, err := s.Get(ctx, name)
		if err != nil {
			// Transient errors during polling are not fatal
			return false, nil
		}
		ready := IsPlacementGroupReady(pg)
		if ready {
			result = pg
		}
		return ready, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForDeleted polls until the hotswap disk attachment returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *HotswapDiskAttachmentsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "hotswap_attachment"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}

// WaitForDeleted polls until the placement group returns 404 (deleted).
// Respects context cancellation and uses exponential backoff.
// Optional WaiterOption parameters can customize polling behavior.
func (s *PlacementGroupsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "placement_group"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}

// WaitForReady polls the snapshot until it has a Ready condition with status True.
func (s *SnapshotsService) WaitForReady(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) (*compute.Snapshot, error) {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "snapshot"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	var result *compute.Snapshot
	err := rest.WaitFor(ctx, config, func() (bool, error) {
		snap, err := s.Get(ctx, name)
		if err != nil {
			return false, nil
		}
		if IsSnapshotReady(snap) {
			result = snap
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WaitForDeleted polls until the snapshot returns 404 (deleted).
func (s *SnapshotsService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration, opts ...WaiterOption) error {
	config := rest.DefaultWaiterConfig()
	config.Timeout = timeout
	config.ResourceType = "snapshot"
	config.Metrics = s.client.metrics

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&config)
	}

	return rest.WaitFor(ctx, config, func() (bool, error) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			return true, nil
		}
		return false, nil
	})
}
