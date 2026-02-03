// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// Package rest provides HTTP client utilities for REST APIs.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
)

// HTTPClient is an interface for HTTP clients that can execute requests.
// This allows for easy mocking in tests without depending on *http.Client.
// Standard *http.Client implements this interface.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a REST HTTP client with authentication.
type Client struct {
	baseURL    string
	httpClient HTTPClient
}

// Config contains REST client configuration.
type Config struct {
	BaseURL    string
	HTTPClient HTTPClient // Should be from auth.Client.GetHTTPClient() for authenticated requests
}

// NewClient creates a new REST client.
// The HTTP client should be obtained from auth.Client.GetHTTPClient() which
// automatically handles authentication and token refresh.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.HTTPClient == nil {
		return nil, fmt.Errorf("HTTP client is required (use auth.Client.GetHTTPClient())")
	}

	return &Client{
		baseURL:    cfg.BaseURL,
		httpClient: cfg.HTTPClient,
	}, nil
}

// Do performs an HTTP request.
// Authentication is handled automatically by the HTTP client (from auth.Client.GetHTTPClient).
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.Do(ctx, req)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) (*http.Response, error) {
	return c.doJSON(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) (*http.Response, error) {
	return c.doJSON(ctx, http.MethodPut, path, body)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, body any) (*http.Response, error) {
	return c.doJSON(ctx, http.MethodPatch, path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	return c.Do(ctx, req)
}

// doJSON performs an HTTP request with a JSON body.
func (c *Client) doJSON(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	var jsonData []byte
	if body != nil {
		var err error
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	fullURL := c.baseURL + path

	// Debug logging if EVROC_DEBUG is set
	if os.Getenv("EVROC_DEBUG") != "" {
		log.Printf("DEBUG: %s %s", method, fullURL)
		if jsonData != nil {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, jsonData, "", "  "); err == nil {
				log.Printf("DEBUG: Request body:\n%s", prettyJSON.String())
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Do(ctx, req)

	// Debug logging for response
	if os.Getenv("EVROC_DEBUG") != "" && resp != nil {
		log.Printf("DEBUG: Response: %d %s", resp.StatusCode, resp.Status)
	}

	return resp, err
}

// DecodeJSON decodes a JSON response body.
func DecodeJSON(resp *http.Response, v any) error {
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return nil
}

// Sentinel errors for common HTTP status codes.
// Use errors.Is() to check for these:
//
//	if errors.Is(err, rest.ErrNotFound) {
//	    // handle not found
//	}
var (
	ErrNotFound   = &APIError{StatusCode: http.StatusNotFound}
	ErrConflict   = &APIError{StatusCode: http.StatusConflict}
	ErrForbidden  = &APIError{StatusCode: http.StatusForbidden}
	ErrBadRequest = &APIError{StatusCode: http.StatusBadRequest}
)

// APIError represents an error returned by the API.
type APIError struct {
	StatusCode int
	Reason     string
	Debug      string
}

func (e *APIError) Error() string {
	if e.Debug != "" {
		return fmt.Sprintf("API error (%d): %s - %s", e.StatusCode, e.Reason, e.Debug)
	}
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Reason)
}

// Is implements error matching for errors.Is().
// Matches based on HTTP status code only.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Reason:     fmt.Sprintf("failed to read error response body: %v", err),
		}
	}

	// Debug logging for error responses
	if os.Getenv("EVROC_DEBUG") != "" {
		log.Printf("DEBUG: Error response body: %s", string(body))
	}

	apiErr := &APIError{StatusCode: resp.StatusCode}

	var errResp struct {
		Reason string  `json:"reason"`
		Debug  *string `json:"debug,omitempty"`
	}

	if json.Unmarshal(body, &errResp) == nil && errResp.Reason != "" {
		apiErr.Reason = errResp.Reason
		if errResp.Debug != nil {
			apiErr.Debug = *errResp.Debug
		}
	} else {
		apiErr.Reason = string(body)
		if apiErr.Reason == "" {
			apiErr.Reason = http.StatusText(resp.StatusCode)
		}
	}

	return apiErr
}

// GetResource performs a GET request and decodes the response.
func GetResource[T any](ctx context.Context, client *Client, path string) (T, error) {
	var result T
	resp, err := client.Get(ctx, path, nil)
	if err != nil {
		return result, err
	}

	if err := DecodeJSON(resp, &result); err != nil {
		return result, err
	}

	return result, nil
}

// ListResources performs a GET request for a list endpoint.
func ListResources[T any](ctx context.Context, client *Client, path string) (T, error) {
	return GetResource[T](ctx, client, path)
}

// ListResourcesWithQuery performs a GET request for a list endpoint with query parameters.
func ListResourcesWithQuery[T any](ctx context.Context, client *Client, path string, query url.Values) (T, error) {
	var result T
	resp, err := client.Get(ctx, path, query)
	if err != nil {
		return result, err
	}

	if err := DecodeJSON(resp, &result); err != nil {
		return result, err
	}

	return result, nil
}

// CreateResource performs a POST request with a body.
func CreateResource[T any](ctx context.Context, client *Client, path string, body any) (T, error) {
	var result T
	resp, err := client.Post(ctx, path, body)
	if err != nil {
		return result, err
	}

	if err := DecodeJSON(resp, &result); err != nil {
		return result, err
	}

	return result, nil
}

// UpdateResource performs a PUT request with a body.
func UpdateResource[T any](ctx context.Context, client *Client, path string, body any) (T, error) {
	var result T
	resp, err := client.Put(ctx, path, body)
	if err != nil {
		return result, err
	}

	if err := DecodeJSON(resp, &result); err != nil {
		return result, err
	}

	return result, nil
}

// PatchResource performs a PATCH request with a body.
func PatchResource[T any](ctx context.Context, client *Client, path string, patch any) (T, error) {
	var result T
	resp, err := client.Patch(ctx, path, patch)
	if err != nil {
		return result, err
	}

	if err := DecodeJSON(resp, &result); err != nil {
		return result, err
	}

	return result, nil
}

// DeleteResource performs a DELETE request.
func DeleteResource(ctx context.Context, client *Client, path string) error {
	resp, err := client.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// ServicePath represents a base path for a service with API version.
type ServicePath struct {
	service string
	version string
}

// NewServicePath creates a service path builder.
func NewServicePath(service, version string) ServicePath {
	return ServicePath{
		service: service,
		version: version,
	}
}

// CollectionPath builds a path for a resource collection (list endpoint).
func (sp ServicePath) CollectionPath(project, region, collection string) string {
	return path.Join("/", sp.service, sp.version, "projects", project, "regions", region, collection)
}

// ResourcePath builds a path for a specific resource.
func (sp ServicePath) ResourcePath(project, region, collection, name string) string {
	return path.Join("/", sp.service, sp.version, "projects", project, "regions", region, collection, name)
}

// ProjectCollectionPath builds a path for project-scoped resource collections.
func (sp ServicePath) ProjectCollectionPath(project, collection string) string {
	return path.Join("/", sp.service, sp.version, "projects", project, collection)
}

// ProjectResourcePath builds a path for project-scoped resources.
func (sp ServicePath) ProjectResourcePath(project, collection, name string) string {
	return path.Join("/", sp.service, sp.version, "projects", project, collection, name)
}

// GlobalCollectionPath builds a global resource collection path.
func (sp ServicePath) GlobalCollectionPath(collection string) string {
	return path.Join("/", sp.service, sp.version, collection)
}

// GlobalResourcePath builds a global resource path.
func (sp ServicePath) GlobalResourcePath(collection, name string) string {
	return path.Join("/", sp.service, sp.version, collection, name)
}
