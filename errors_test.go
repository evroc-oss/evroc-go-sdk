// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package evroc

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIsQuotaError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "wrapped quota error",
			err: fmt.Errorf("failed to create virtual machine: %w", &APIError{
				StatusCode: http.StatusForbidden,
				Reason:     ReasonInsufficientQuota,
				Debug:      "not enough quota to perform request.",
			}),
			want: true,
		},
		{
			name: "forbidden with a different reason",
			err: &APIError{
				StatusCode: http.StatusForbidden,
				Reason:     "Forbidden",
				Debug:      "access denied",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsQuotaError(tt.err); got != tt.want {
				t.Errorf("IsQuotaError() = %v, want %v", got, tt.want)
			}
		})
	}
}
