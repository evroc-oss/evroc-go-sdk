// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package evroc

import (
	"errors"
	"net/http"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
)

// Common API errors that can be checked with errors.Is()
var (
	// ErrNotFound indicates the requested resource was not found (HTTP 404)
	ErrNotFound = rest.ErrNotFound

	// ErrConflict indicates a conflict with an existing resource (HTTP 409)
	ErrConflict = rest.ErrConflict

	// ErrForbidden indicates access to the resource is forbidden (HTTP 403)
	ErrForbidden = rest.ErrForbidden

	// ErrBadRequest indicates the request was invalid (HTTP 400)
	ErrBadRequest = rest.ErrBadRequest
)

// APIError is the structured error returned by the evroc API for failed
// requests. Extract it with errors.As to inspect StatusCode, Reason, Debug.
type APIError = rest.APIError

// ReasonInsufficientQuota is the stable API reason code (HTTP 403) returned
// when a request is rejected because an organization quota was reached.
const ReasonInsufficientQuota = "InsufficientQuota"

// IsQuotaError reports whether err is a quota-exceeded error (HTTP 403 with
// reason ReasonInsufficientQuota), through any fmt.Errorf wrapping.
func IsQuotaError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) &&
		apiErr.StatusCode == http.StatusForbidden &&
		apiErr.Reason == ReasonInsufficientQuota
}
