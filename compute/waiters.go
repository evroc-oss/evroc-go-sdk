// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package compute

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/rest"
)

// WaitForReady polls the disk until it has a Ready condition with status True.
func (s *DisksService) WaitForReady(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		disk, err := s.Get(ctx, name)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if IsDiskReady(disk) {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("disk %s did not become ready within %v", name, timeout)
}

// WaitForReady polls the VM until it has a Ready condition with status True.
func (s *VirtualMachinesService) WaitForReady(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		vm, err := s.Get(ctx, name)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		// Check for provisioning failures
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
				return fmt.Errorf("VM %s provisioning failed:%s", name, errMsg)
			}
		}

		if IsVMReady(vm) {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("VM %s did not become ready within %v", name, timeout)
}

// WaitForDeleted polls until the disk returns 404 (deleted).
func (s *DisksService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Add extra sleep to ensure deletion is fully propagated
			time.Sleep(3 * time.Second)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("disk %s was not deleted within %v", name, timeout)
}

// WaitForDeleted polls until the VM returns 404 (deleted).
func (s *VirtualMachinesService) WaitForDeleted(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := s.Get(ctx, name)
		if errors.Is(err, rest.ErrNotFound) {
			// Add extra sleep to ensure deletion is fully propagated
			time.Sleep(3 * time.Second)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("VM %s was not deleted within %v", name, timeout)
}

// WaitForReady polls the hotswap disk attachment until it has a Ready condition.
func (s *HotswapDiskAttachmentsService) WaitForReady(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		attachment, err := s.Get(ctx, name)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if IsAttachmentReady(attachment) {
			return nil
		}

		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("hotswap disk attachment %s did not become ready within %v", name, timeout)
}

