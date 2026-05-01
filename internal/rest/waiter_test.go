// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitFor_Success(t *testing.T) {
	config := WaiterConfig{
		Timeout:         5 * time.Second,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
	}

	attempts := 0
	condition := func() (bool, error) {
		attempts++
		if attempts >= 3 {
			return true, nil
		}
		return false, nil
	}

	err := WaitFor(context.Background(), config, condition)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestWaitFor_Timeout(t *testing.T) {
	config := WaiterConfig{
		Timeout:         100 * time.Millisecond,
		InitialInterval: 20 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Multiplier:      1.5,
	}

	condition := func() (bool, error) {
		return false, nil // Never becomes ready
	}

	start := time.Now()
	err := WaitFor(context.Background(), config, condition)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected timeout error")
	}

	// Should timeout around the configured timeout
	if elapsed < 90*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("timeout took unexpected time: %v", elapsed)
	}
}

func TestWaitFor_ContextCancellation(t *testing.T) {
	config := WaiterConfig{
		Timeout:         5 * time.Second,
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      1.5,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	attempts := 0
	condition := func() (bool, error) {
		attempts++
		return false, nil
	}

	start := time.Now()
	err := WaitFor(ctx, config, condition)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected cancellation error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}

	// Should cancel quickly after context is cancelled
	if elapsed > 200*time.Millisecond {
		t.Errorf("cancellation took too long: %v", elapsed)
	}

	// Should have made at least 1 attempt but not many
	if attempts < 1 || attempts > 5 {
		t.Errorf("unexpected number of attempts: %d", attempts)
	}
}

func TestWaitFor_ConditionError(t *testing.T) {
	config := WaiterConfig{
		Timeout:         1 * time.Second,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Multiplier:      2.0,
	}

	expectedErr := errors.New("condition check failed")
	condition := func() (bool, error) {
		return false, expectedErr
	}

	err := WaitFor(context.Background(), config, condition)
	if err == nil {
		t.Error("expected error from condition")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped condition error, got: %v", err)
	}
}

func TestWaitFor_ProgressCallback(t *testing.T) {
	var callbackAttempts []int
	var callbackElapsed []time.Duration

	config := WaiterConfig{
		Timeout:         1 * time.Second,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Multiplier:      1.5,
		ProgressCallback: func(attempt int, elapsed time.Duration) {
			callbackAttempts = append(callbackAttempts, attempt)
			callbackElapsed = append(callbackElapsed, elapsed)
		},
	}

	attempts := 0
	condition := func() (bool, error) {
		attempts++
		if attempts >= 3 {
			return true, nil
		}
		return false, nil
	}

	err := WaitFor(context.Background(), config, condition)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	if len(callbackAttempts) != 3 {
		t.Errorf("expected 3 callback invocations, got %d", len(callbackAttempts))
	}

	for i, attempt := range callbackAttempts {
		if attempt != i+1 {
			t.Errorf("callback attempt[%d] = %d, expected %d", i, attempt, i+1)
		}
	}

	// Elapsed times should be increasing
	for i := 1; i < len(callbackElapsed); i++ {
		if callbackElapsed[i] <= callbackElapsed[i-1] {
			t.Errorf("elapsed time not increasing: %v <= %v", callbackElapsed[i], callbackElapsed[i-1])
		}
	}
}

func TestWaitFor_ExponentialBackoff(t *testing.T) {
	config := WaiterConfig{
		Timeout:         2 * time.Second,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
	}

	var pollTimes []time.Time
	attempts := 0

	condition := func() (bool, error) {
		pollTimes = append(pollTimes, time.Now())
		attempts++
		if attempts >= 5 {
			return true, nil
		}
		return false, nil
	}

	err := WaitFor(context.Background(), config, condition)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	// Check that intervals are increasing (exponential backoff)
	for i := 2; i < len(pollTimes); i++ {
		interval1 := pollTimes[i-1].Sub(pollTimes[i-2])
		interval2 := pollTimes[i].Sub(pollTimes[i-1])

		// Second interval should be larger (or capped at max)
		if interval2 < interval1 && interval2 < config.MaxInterval {
			t.Errorf("backoff not increasing: %v -> %v", interval1, interval2)
		}
	}
}

func TestWaitForWithConstantInterval(t *testing.T) {
	attempts := 0
	var pollTimes []time.Time

	condition := func() (bool, error) {
		pollTimes = append(pollTimes, time.Now())
		attempts++
		if attempts >= 4 {
			return true, nil
		}
		return false, nil
	}

	err := WaitForWithConstantInterval(
		context.Background(),
		1*time.Second,
		20*time.Millisecond,
		condition,
	)

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	// Check that intervals are constant
	for i := 2; i < len(pollTimes); i++ {
		interval1 := pollTimes[i-1].Sub(pollTimes[i-2])
		interval2 := pollTimes[i].Sub(pollTimes[i-1])

		// Allow some timing variance (±5ms)
		diff := interval2 - interval1
		if diff < -5*time.Millisecond || diff > 5*time.Millisecond {
			t.Errorf("intervals not constant: %v vs %v", interval1, interval2)
		}
	}
}

func TestDefaultWaiterConfig(t *testing.T) {
	config := DefaultWaiterConfig()

	if config.Timeout != 5*time.Minute {
		t.Errorf("unexpected timeout: %v", config.Timeout)
	}

	if config.InitialInterval != 2*time.Second {
		t.Errorf("unexpected initial interval: %v", config.InitialInterval)
	}

	if config.MaxInterval != 30*time.Second {
		t.Errorf("unexpected max interval: %v", config.MaxInterval)
	}

	if config.Multiplier != 1.5 {
		t.Errorf("unexpected multiplier: %v", config.Multiplier)
	}
}
