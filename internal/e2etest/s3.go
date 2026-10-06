// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2etest

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// S3Object is a minimal SigV4 client for test fixtures, so e2e tests can
// upload and remove objects without adding an S3 SDK to the module.
type S3Object struct {
	Endpoint  string // host, e.g. s3.se-sto.evroc.com
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
	Key       string
}

// S3Endpoint returns E2E_S3_ENDPOINT when set (needed outside production),
// otherwise the given default.
func S3Endpoint(defaultEndpoint string) string {
	if endpoint := os.Getenv("E2E_S3_ENDPOINT"); endpoint != "" {
		return strings.TrimPrefix(endpoint, "https://")
	}
	return defaultEndpoint
}

// Put uploads body and returns the object version ID, empty if the bucket is unversioned.
func (o S3Object) Put(ctx context.Context, body []byte) (string, error) {
	resp, err := o.do(ctx, http.MethodPut, nil, body)
	if err != nil {
		return "", err
	}
	return resp.Header.Get("x-amz-version-id"), nil
}

// Delete removes the object, or one specific version when versionID is set.
func (o S3Object) Delete(ctx context.Context, versionID string) error {
	query := url.Values{}
	if versionID != "" {
		query.Set("versionId", versionID)
	}
	_, err := o.do(ctx, http.MethodDelete, query, nil)
	return err
}

func (o S3Object) do(ctx context.Context, method string, query url.Values, body []byte) (*http.Response, error) {
	path := "/" + o.Bucket + "/" + strings.TrimPrefix(o.Key, "/")
	u := url.URL{Scheme: "https", Host: o.Endpoint, Path: path, RawQuery: query.Encode()}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	o.sign(req, body, time.Now().UTC())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		msg, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if readErr != nil {
			msg = []byte(readErr.Error())
		}
		return nil, fmt.Errorf("s3 %s %s: %s: %s", method, path, resp.Status, msg)
	}
	return resp, nil
}

func (o S3Object) sign(req *http.Request, body []byte, now time.Time) {
	amzDate := now.Format("20060102T150405Z")
	day := now.Format("20060102")
	payloadHash := sha256Hex(body)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.EscapedPath(),
		req.URL.Query().Encode(),
		"host:" + req.URL.Host + "\nx-amz-content-sha256:" + payloadHash + "\nx-amz-date:" + amzDate + "\n",
		"host;x-amz-content-sha256;x-amz-date",
		payloadHash,
	}, "\n")
	scope := day + "/" + o.Region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + sha256Hex([]byte(canonicalRequest))

	key := hmacSHA256([]byte("AWS4"+o.SecretKey), day)
	key = hmacSHA256(key, o.Region)
	key = hmacSHA256(key, "s3")
	key = hmacSHA256(key, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(key, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=%s",
		o.AccessKey, scope, signature))
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}
