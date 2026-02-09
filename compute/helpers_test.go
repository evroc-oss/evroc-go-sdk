// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package compute

import (
	"testing"
	"time"

	computetypes "github.com/evroc-oss/evroc-go-sdk/types/compute"
)

func TestIsVMReady(t *testing.T) {
	t.Run("nil vm", func(t *testing.T) {
		if IsVMReady(nil) {
			t.Error("nil VM should not be ready")
		}
	})

	t.Run("vm with no conditions", func(t *testing.T) {
		vm := &computetypes.VirtualMachine{}
		if IsVMReady(vm) {
			t.Error("VM with no conditions should not be ready")
		}
	})

	t.Run("vm with Ready condition True", func(t *testing.T) {
		conditions := []computetypes.VirtualMachineStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "VMReady",
				Message:            "VM is ready",
			},
		}
		vm := &computetypes.VirtualMachine{
			Status: computetypes.VirtualMachineStatus{
				Conditions: &conditions,
			},
		}
		if !IsVMReady(vm) {
			t.Error("VM with Ready=True should be ready")
		}
	})

	t.Run("vm with Ready condition False", func(t *testing.T) {
		conditions := []computetypes.VirtualMachineStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Creating",
				Message:            "VM is being created",
			},
		}
		vm := &computetypes.VirtualMachine{
			Status: computetypes.VirtualMachineStatus{
				Conditions: &conditions,
			},
		}
		if IsVMReady(vm) {
			t.Error("VM with Ready=False should not be ready")
		}
	})
}

func TestIsVMRunning(t *testing.T) {
	t.Run("nil vm", func(t *testing.T) {
		if !IsVMRunning(nil) {
			t.Error("nil VM should default to running")
		}
	})

	t.Run("vm with no running spec", func(t *testing.T) {
		vm := &computetypes.VirtualMachine{}
		if !IsVMRunning(vm) {
			t.Error("VM with no running spec should default to running")
		}
	})

	t.Run("vm with running=true", func(t *testing.T) {
		running := true
		vm := &computetypes.VirtualMachine{
			Spec: computetypes.VirtualMachineSpec{
				Running: &running,
			},
		}
		if !IsVMRunning(vm) {
			t.Error("VM with running=true should be running")
		}
	})

	t.Run("vm with running=false", func(t *testing.T) {
		running := false
		vm := &computetypes.VirtualMachine{
			Spec: computetypes.VirtualMachineSpec{
				Running: &running,
			},
		}
		if IsVMRunning(vm) {
			t.Error("VM with running=false should not be running")
		}
	})
}

func TestIsDiskReady(t *testing.T) {
	t.Run("nil disk", func(t *testing.T) {
		if IsDiskReady(nil) {
			t.Error("nil disk should not be ready")
		}
	})

	t.Run("disk with no conditions", func(t *testing.T) {
		disk := &computetypes.Disk{}
		if IsDiskReady(disk) {
			t.Error("disk with no conditions should not be ready")
		}
	})

	t.Run("disk with Ready condition True", func(t *testing.T) {
		conditions := []computetypes.DiskStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "DiskReady",
				Message:            "Disk is ready",
			},
		}
		disk := &computetypes.Disk{
			Status: computetypes.DiskStatus{
				Conditions: &conditions,
			},
		}
		if !IsDiskReady(disk) {
			t.Error("disk with Ready=True should be ready")
		}
	})

	t.Run("disk with Ready condition False", func(t *testing.T) {
		conditions := []computetypes.DiskStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Creating",
				Message:            "Disk is being created",
			},
		}
		disk := &computetypes.Disk{
			Status: computetypes.DiskStatus{
				Conditions: &conditions,
			},
		}
		if IsDiskReady(disk) {
			t.Error("disk with Ready=False should not be ready")
		}
	})
}

func TestIsAttachmentReady(t *testing.T) {
	t.Run("nil attachment", func(t *testing.T) {
		if IsAttachmentReady(nil) {
			t.Error("nil attachment should not be ready")
		}
	})

	t.Run("attachment with no conditions", func(t *testing.T) {
		attachment := &computetypes.HotswapDiskAttachment{}
		if IsAttachmentReady(attachment) {
			t.Error("attachment with no conditions should not be ready")
		}
	})

	t.Run("attachment with Ready condition True", func(t *testing.T) {
		conditions := []computetypes.HotswapDiskAttachmentStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "AttachmentReady",
				Message:            "Attachment is ready",
			},
		}
		attachment := &computetypes.HotswapDiskAttachment{
			Status: computetypes.HotswapDiskAttachmentStatus{
				Conditions: &conditions,
			},
		}
		if !IsAttachmentReady(attachment) {
			t.Error("attachment with Ready=True should be ready")
		}
	})

	t.Run("attachment with Ready condition False", func(t *testing.T) {
		conditions := []computetypes.HotswapDiskAttachmentStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Attaching",
				Message:            "Disk is being attached",
			},
		}
		attachment := &computetypes.HotswapDiskAttachment{
			Status: computetypes.HotswapDiskAttachmentStatus{
				Conditions: &conditions,
			},
		}
		if IsAttachmentReady(attachment) {
			t.Error("attachment with Ready=False should not be ready")
		}
	})
}

func TestGetVMState(t *testing.T) {
	t.Run("nil vm", func(t *testing.T) {
		state := GetVMState(nil)
		if state != "Unknown" {
			t.Errorf("expected Unknown, got %s", state)
		}
	})

	t.Run("vm with no conditions", func(t *testing.T) {
		vm := &computetypes.VirtualMachine{}
		state := GetVMState(vm)
		if state != "Unknown" {
			t.Errorf("expected Unknown, got %s", state)
		}
	})

	t.Run("stopped vm", func(t *testing.T) {
		running := false
		conditions := []computetypes.VirtualMachineStatusConditionsItem{}
		vm := &computetypes.VirtualMachine{
			Spec: computetypes.VirtualMachineSpec{
				Running: &running,
			},
			Status: computetypes.VirtualMachineStatus{
				Conditions: &conditions,
			},
		}
		state := GetVMState(vm)
		if state != "Stopped" {
			t.Errorf("expected Stopped, got %s", state)
		}
	})

	t.Run("running vm", func(t *testing.T) {
		conditions := []computetypes.VirtualMachineStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "VMReady",
				Message:            "VM is ready",
			},
		}
		vm := &computetypes.VirtualMachine{
			Status: computetypes.VirtualMachineStatus{
				Conditions: &conditions,
			},
		}
		state := GetVMState(vm)
		if state != "Running" {
			t.Errorf("expected Running, got %s", state)
		}
	})

	t.Run("pending vm", func(t *testing.T) {
		conditions := []computetypes.VirtualMachineStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Creating",
				Message:            "VM is being created",
			},
		}
		vm := &computetypes.VirtualMachine{
			Status: computetypes.VirtualMachineStatus{
				Conditions: &conditions,
			},
		}
		state := GetVMState(vm)
		if state != "Creating" {
			t.Errorf("expected Creating, got %s", state)
		}
	})

	t.Run("vm with empty reason", func(t *testing.T) {
		conditions := []computetypes.VirtualMachineStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "",
				Message:            "VM is not ready",
			},
		}
		vm := &computetypes.VirtualMachine{
			Status: computetypes.VirtualMachineStatus{
				Conditions: &conditions,
			},
		}
		state := GetVMState(vm)
		if state != "Pending" {
			t.Errorf("expected Pending, got %s", state)
		}
	})
}
