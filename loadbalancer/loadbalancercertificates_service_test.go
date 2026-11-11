// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package loadbalancer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestLoadBalancerCertificateContentActions(t *testing.T) {
	var gotPath, gotType, gotBody string
	client, server := setupTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotPath, gotType, gotBody = r.URL.Path, r.Header.Get("Content-Type"), string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"result":"success"}}`))
	}))
	defer server.Close()
	certs := client.LoadBalancerCertificates()
	base := "/loadbalancer/" + apiVersion + "/projects/test-project/regions/test-region/loadBalancerCertificates/cert/content"

	if err := certs.UploadContent(context.Background(), "cert", "CERT", "KEY"); err != nil {
		t.Fatalf("UploadContent: %v", err)
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(gotBody), &body); err != nil || body["certificate"] != "CERT" || body["privateKey"] != "KEY" {
		t.Errorf("UploadContent body = %s", gotBody)
	}
	if gotPath != base || gotType != "application/json" {
		t.Errorf("UploadContent path/type = %s %s", gotPath, gotType)
	}

	if err := certs.RenewCertificate(context.Background(), "cert", "NEWCERT"); err != nil {
		t.Fatalf("RenewCertificate: %v", err)
	}
	if gotPath != base+"/certificate" || gotType != "application/x-pem-file" || gotBody != "NEWCERT" {
		t.Errorf("RenewCertificate path/type/body = %s %s %q", gotPath, gotType, gotBody)
	}
}
