// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package networking

import (
	"testing"
	"time"

	networkingtypes "github.com/evroc-oss/evroc-go-sdk/types/networking"
)

func TestIsPublicIPReady(t *testing.T) {
	t.Run("nil public ip", func(t *testing.T) {
		if IsPublicIPReady(nil) {
			t.Error("nil public IP should not be ready")
		}
	})

	t.Run("public ip with no conditions", func(t *testing.T) {
		ip := &networkingtypes.PublicIP{}
		if IsPublicIPReady(ip) {
			t.Error("public IP with no conditions should not be ready")
		}
	})

	t.Run("public ip with Ready condition True", func(t *testing.T) {
		conditions := []networkingtypes.PublicIPStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "PublicIPReady",
				Message:            "Public IP is ready",
			},
		}
		ip := &networkingtypes.PublicIP{
			Status: networkingtypes.PublicIPStatus{
				Conditions: &conditions,
			},
		}
		if !IsPublicIPReady(ip) {
			t.Error("public IP with Ready=True should be ready")
		}
	})

	t.Run("public ip with Ready condition False", func(t *testing.T) {
		conditions := []networkingtypes.PublicIPStatusConditionsItem{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Allocating",
				Message:            "Public IP is being allocated",
			},
		}
		ip := &networkingtypes.PublicIP{
			Status: networkingtypes.PublicIPStatus{
				Conditions: &conditions,
			},
		}
		if IsPublicIPReady(ip) {
			t.Error("public IP with Ready=False should not be ready")
		}
	})

	t.Run("public ip with multiple conditions", func(t *testing.T) {
		conditions := []networkingtypes.PublicIPStatusConditionsItem{
			{
				Type:               "Allocated",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "Allocated",
				Message:            "IP allocated",
			},
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: time.Now(),
				Reason:             "Ready",
				Message:            "IP ready",
			},
		}
		ip := &networkingtypes.PublicIP{
			Status: networkingtypes.PublicIPStatus{
				Conditions: &conditions,
			},
		}
		if !IsPublicIPReady(ip) {
			t.Error("public IP with Ready=True should be ready even with multiple conditions")
		}
	})
}

func TestGetPublicIPAddress(t *testing.T) {
	t.Run("nil public ip", func(t *testing.T) {
		addr := GetPublicIPAddress(nil)
		if addr != "" {
			t.Errorf("nil public IP should return empty string, got '%s'", addr)
		}
	})

	t.Run("public ip with no address", func(t *testing.T) {
		ip := &networkingtypes.PublicIP{}
		addr := GetPublicIPAddress(ip)
		if addr != "" {
			t.Errorf("public IP with no address should return empty string, got '%s'", addr)
		}
	})

	t.Run("public ip with address", func(t *testing.T) {
		ipAddr := "203.0.113.42"
		ip := &networkingtypes.PublicIP{
			Status: networkingtypes.PublicIPStatus{
				PublicIPv4Address: &ipAddr,
			},
		}
		addr := GetPublicIPAddress(ip)
		if addr != ipAddr {
			t.Errorf("expected address '%s', got '%s'", ipAddr, addr)
		}
	})

	t.Run("public ip with different address", func(t *testing.T) {
		ipAddr := "192.0.2.1"
		ip := &networkingtypes.PublicIP{
			Status: networkingtypes.PublicIPStatus{
				PublicIPv4Address: &ipAddr,
			},
		}
		addr := GetPublicIPAddress(ip)
		if addr != ipAddr {
			t.Errorf("expected address '%s', got '%s'", ipAddr, addr)
		}
	})
}
