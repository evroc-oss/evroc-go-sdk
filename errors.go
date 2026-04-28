// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package evroc

import "github.com/evroc-oss/evroc-go-sdk/internal/rest"

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
