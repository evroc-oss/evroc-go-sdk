// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"context"
	"encoding/json"
	"path"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/loadbalancer"
)

func (s *LoadBalancerCertificatesService) contentPath(name string, action ...string) string {
	return path.Join(append([]string{s.client.path.ResourcePath(
		s.client.parent.DefaultProject(),
		s.client.parent.DefaultRegion(),
		resourceLoadBalancerCertificates,
		name), "content"}, action...)...)
}

// UploadContent uploads or replaces the PEM certificate chain and private key of a
// LoadBalancerCertificate. The content is validated asynchronously; wait for the
// certificate's Ready condition before referencing it from a listener.
func (s *LoadBalancerCertificatesService) UploadContent(ctx context.Context, name, certificatePEM, privateKeyPEM string) error {
	// Sent with PostRaw rather than the JSON helpers, which log request bodies
	// when EVROC_SDK_DEBUG is set and would print the private key.
	body, err := json.Marshal(loadbalancer.CertificateContent{Certificate: certificatePEM, PrivateKey: privateKeyPEM})
	if err != nil {
		return err
	}
	resp, err := s.client.rest.PostRaw(ctx, s.contentPath(name), "application/json", body)
	if err != nil {
		return err
	}
	var result loadbalancer.CertificateContentResult
	return rest.DecodeJSON(resp, &result)
}

// RenewCertificate replaces the PEM certificate chain of a LoadBalancerCertificate,
// keeping the stored private key.
func (s *LoadBalancerCertificatesService) RenewCertificate(ctx context.Context, name, certificatePEM string) error {
	resp, err := s.client.rest.PostRaw(ctx, s.contentPath(name, "certificate"), "application/x-pem-file", []byte(certificatePEM))
	if err != nil {
		return err
	}
	var result loadbalancer.CertificateContentResult
	return rest.DecodeJSON(resp, &result)
}
