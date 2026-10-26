// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evroc-oss/evroc-go-sdk/internal/rest"
	"github.com/evroc-oss/evroc-go-sdk/types/iam"
)

func TestDeriveRoleBindingName(t *testing.T) {
	t.Run("user principal", func(t *testing.T) {
		name, err := DeriveRoleBindingName("/iam/users/11111111-1111-1111-1111-111111111111")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := "u-11111111-1111-1111-1111-111111111111"; name != want {
			t.Errorf("expected %q, got %q", want, name)
		}
	})

	t.Run("service account principal", func(t *testing.T) {
		name, err := DeriveRoleBindingName("/iam/projects/my-project/serviceAccounts/cicd-pipeline")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if want := "sa-my-project.cicd-pipeline"; name != want {
			t.Errorf("expected %q, got %q", want, name)
		}
	})

	t.Run("unrecognized principal returns error", func(t *testing.T) {
		_, err := DeriveRoleBindingName("/iam/groups/some-group")
		if err == nil {
			t.Fatal("expected error for unrecognized principal format")
		}
	})
}

func TestCreateProjectRoleBindingOverwritesName(t *testing.T) {
	var capturedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"apiVersion":"` + builderAPIVersion + `","kind":"RoleBinding","metadata":{"id":"sa-my-project.cicd"},"spec":{"principal":"/iam/projects/my-project/serviceAccounts/cicd","roles":[]}}`))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClient(restClient, &mockContextProvider{})

	req := &iam.RolebindingRequest{
		ApiVersion: builderAPIVersion,
		Kind:       "RoleBinding",
		Metadata: iam.ProjectMetadataRequest{
			Id: "whatever-the-caller-typed",
		},
		Spec: iam.RolebindingSpec{
			Principal: "/iam/projects/my-project/serviceAccounts/cicd",
			Roles:     []iam.RoleEntry{},
		},
	}

	if _, err := client.RoleBindings().CreateProjectRoleBinding(context.Background(), req); err != nil {
		t.Fatalf("CreateProjectRoleBinding: %v", err)
	}

	metadata, _ := capturedBody["metadata"].(map[string]interface{})
	if want := "sa-my-project.cicd"; metadata["id"] != want {
		t.Errorf("expected request body metadata.id to be overwritten to %q, got %v", want, metadata["id"])
	}
}

func TestListMyRoleBindings(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"metadata":{"id":"u-1","project":"p1"}},{"metadata":{"id":"u-1","organization":"o1"}}]}`))
	}))
	defer server.Close()

	restClient, _ := rest.NewClient(rest.Config{BaseURL: server.URL, HTTPClient: server.Client()})
	client := NewClient(restClient, &mockContextProvider{})

	list, err := client.RoleBindings().ListMyRoleBindings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "/iam/" + apiVersion + "/me/roleBindings"; gotPath != want {
		t.Errorf("expected path %q, got %q", want, gotPath)
	}
	items := *list.Items
	if len(items) != 2 || *items[0].Metadata.Project != "p1" || *items[1].Metadata.Organization != "o1" {
		t.Errorf("unexpected items: %+v", items)
	}
}
