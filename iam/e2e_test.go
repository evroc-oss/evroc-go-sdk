// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package iam_test

import (
	"context"
	"testing"
	"time"

	"github.com/evroc-oss/evroc-go-sdk/iam"
	"github.com/evroc-oss/evroc-go-sdk/internal/e2etest"
	iamtypes "github.com/evroc-oss/evroc-go-sdk/types/iam"
)

func TestE2E_ServiceAccount_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	project := e2etest.GetProject(t)
	saName := e2etest.RandomName("sa")

	// Create service account
	t.Logf("Creating service account: %s", saName)
	sa, err := iam.NewServiceAccountBuilder(saName, project).
		WithDescription("e2e test service account").
		Create(ctx, client.IAM().ServiceAccounts())
	if err != nil {
		t.Fatalf("failed to create service account: %v", err)
	}

	saID := e2etest.MustGetID(t, sa.Metadata.Id, "service account")
	t.Logf("Created service account: %s", saID)

	saDeleted := false
	e2etest.DeferCleanup(t, ctx, client.IAM().ServiceAccounts().Delete, saID, "service account", &saDeleted)

	// Get service account
	t.Logf("Getting service account: %s", saID)
	retrieved, err := client.IAM().ServiceAccounts().Get(ctx, saID)
	if err != nil {
		t.Fatalf("failed to get service account: %v", err)
	}
	if retrieved.Metadata.Id != saID {
		t.Errorf("expected ID %s, got %s", saID, retrieved.Metadata.Id)
	}

	// List service accounts
	t.Logf("Listing service accounts...")
	saList, err := client.IAM().ServiceAccounts().List(ctx)
	if err != nil {
		t.Fatalf("failed to list service accounts: %v", err)
	}
	e2etest.AssertInList(t, saList.Items, saID, func(s iamtypes.Serviceaccount) string {
		return s.Metadata.Id
	}, "service account")

	// Create credential
	credName := e2etest.RandomName("cred")
	t.Logf("Creating credential: %s", credName)
	cred, err := iam.NewServiceAccountCredentialBuilder(credName, project, client.IAM().ServiceAccountRef(saID), time.Now().AddDate(0, 0, 2)).
		WithDescription("e2e test credential").
		Create(ctx, client.IAM().ServiceAccountCredentials(saID))
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	credID := e2etest.MustGetID(t, cred.Metadata.Id, "credential")
	t.Logf("Created credential: %s", credID)

	credDeleted := false
	e2etest.DeferCleanup(t, ctx, client.IAM().ServiceAccountCredentials(saID).Delete, credID, "credential", &credDeleted)

	if cred.Status.PrivateKeyJwk == nil {
		t.Error("expected private key in create response")
	}

	// Get credential
	t.Logf("Getting credential: %s", credID)
	retrievedCred, err := client.IAM().ServiceAccountCredentials(saID).Get(ctx, credID)
	if err != nil {
		t.Fatalf("failed to get credential: %v", err)
	}
	if retrievedCred.Metadata.Id != credID {
		t.Errorf("expected ID %s, got %s", credID, retrievedCred.Metadata.Id)
	}
	if retrievedCred.Status.PrivateKeyJwk != nil {
		t.Error("private key should not be returned on Get")
	}

	// List credentials
	t.Logf("Listing credentials...")
	credList, err := client.IAM().ServiceAccountCredentials(saID).List(ctx)
	if err != nil {
		t.Fatalf("failed to list credentials: %v", err)
	}
	e2etest.AssertInList(t, credList.Items, credID, func(c iamtypes.Serviceaccountcredential) string {
		return c.Metadata.Id
	}, "credential")

	// Revoke credential
	t.Logf("Revoking credential: %s", credID)
	if err := client.IAM().ServiceAccountCredentials(saID).Revoke(ctx, credID, "e2e test"); err != nil {
		t.Fatalf("failed to revoke credential: %v", err)
	}

	t.Logf("Waiting for credential revocation...")
	revoked := false
	for i := 0; i < int(e2etest.CredentialRevokeTimeout/e2etest.CredentialRevokeInterval); i++ {
		c, err := client.IAM().ServiceAccountCredentials(saID).Get(ctx, credID)
		if err != nil {
			t.Fatalf("failed to get credential: %v", err)
		}
		if c.Status.RevokedAt != nil {
			revoked = true
			t.Logf("Credential revoked at %s", c.Status.RevokedAt)
			break
		}
		time.Sleep(e2etest.CredentialRevokeInterval)
	}
	if !revoked {
		t.Errorf("credential was not revoked after %v", e2etest.CredentialRevokeTimeout)
	}

	// Delete credential
	t.Logf("Deleting credential: %s", credID)
	if err := client.IAM().ServiceAccountCredentials(saID).Delete(ctx, credID); err != nil {
		t.Fatalf("failed to delete credential: %v", err)
	}
	credDeleted = true

	// HMAC-SigV4 credential: the secret is returned on create only
	hmacName := e2etest.RandomName("hmac")
	t.Logf("Creating HMAC credential: %s", hmacName)
	hmac, err := iam.NewServiceAccountCredentialBuilder(hmacName, project, client.IAM().ServiceAccountRef(saID), time.Now().AddDate(0, 0, 2)).
		WithHMACSigV4().
		Create(ctx, client.IAM().ServiceAccountCredentials(saID))
	if err != nil {
		t.Fatalf("failed to create HMAC credential: %v", err)
	}
	hmacID := e2etest.MustGetID(t, hmac.Metadata.Id, "HMAC credential")
	hmacDeleted := false
	e2etest.DeferCleanup(t, ctx, client.IAM().ServiceAccountCredentials(saID).Delete, hmacID, "HMAC credential", &hmacDeleted)

	if hmac.Spec.Type != iamtypes.HmacSigv4 {
		t.Errorf("expected type %s, got %s", iamtypes.HmacSigv4, hmac.Spec.Type)
	}
	if hmac.Status.SecretAccessKey == nil || *hmac.Status.SecretAccessKey == "" {
		t.Error("expected secret access key in create response")
	}
	if hmac.Spec.HmacSigv4 == nil || hmac.Spec.HmacSigv4.AccessKeyId == "" {
		t.Error("expected access key ID in create response")
	}

	retrievedHMAC, err := client.IAM().ServiceAccountCredentials(saID).Get(ctx, hmacID)
	if err != nil {
		t.Fatalf("failed to get HMAC credential: %v", err)
	}
	if retrievedHMAC.Status.SecretAccessKey != nil {
		t.Error("secret access key should not be returned on Get")
	}

	t.Logf("Deleting HMAC credential: %s", hmacID)
	if err := client.IAM().ServiceAccountCredentials(saID).Delete(ctx, hmacID); err != nil {
		t.Fatalf("failed to delete HMAC credential: %v", err)
	}
	hmacDeleted = true

	// Delete service account
	t.Logf("Deleting service account: %s", saID)
	if err := client.IAM().ServiceAccounts().Delete(ctx, saID); err != nil {
		t.Fatalf("failed to delete service account: %v", err)
	}
	saDeleted = true
}

func TestE2E_RoleBinding_Lifecycle(t *testing.T) {
	e2etest.PreCheck(t)

	ctx := context.Background()
	client := e2etest.NewClient(t)
	project := e2etest.GetProject(t)
	saName := e2etest.RandomName("sa")

	sa, err := iam.NewServiceAccountBuilder(saName, project).
		WithDescription("e2e test role binding principal").
		Create(ctx, client.IAM().ServiceAccounts())
	if err != nil {
		t.Fatalf("failed to create service account: %v", err)
	}
	saID := e2etest.MustGetID(t, sa.Metadata.Id, "service account")
	saDeleted := false
	e2etest.DeferCleanup(t, ctx, client.IAM().ServiceAccounts().Delete, saID, "service account", &saDeleted)

	principal := client.IAM().ServiceAccountRef(saID)
	wantName, err := iam.DeriveRoleBindingName(principal)
	if err != nil {
		t.Fatalf("failed to derive role binding name: %v", err)
	}

	// Create with a non-conforming id: the SDK must replace it with the
	// derived name, which the v1beta2 API enforces.
	t.Logf("Creating role binding for: %s", principal)
	rb, err := client.IAM().RoleBindings().CreateProjectRoleBinding(ctx, &iamtypes.RolebindingRequest{
		ApiVersion: "iam/v1beta2",
		Kind:       "RoleBinding",
		Metadata:   iamtypes.ProjectMetadataRequest{Id: "not-the-derived-name", Project: &project},
		Spec: iamtypes.RolebindingSpec{
			Principal: principal,
			Roles:     []iamtypes.RoleEntry{{Name: iamtypes.RoleComputeViewer}},
		},
	})
	if err != nil {
		t.Fatalf("failed to create role binding: %v", err)
	}
	rbDeleted := false
	e2etest.DeferCleanup(t, ctx, client.IAM().RoleBindings().DeleteProjectRoleBinding, wantName, "role binding", &rbDeleted)

	if rb.Metadata.Id != wantName {
		t.Errorf("expected role binding ID %s, got %s", wantName, rb.Metadata.Id)
	}

	// Get role binding
	retrieved, err := client.IAM().RoleBindings().GetProjectRoleBinding(ctx, wantName)
	if err != nil {
		t.Fatalf("failed to get role binding: %v", err)
	}
	if len(retrieved.Spec.Roles) != 1 || retrieved.Spec.Roles[0].Name != iamtypes.RoleComputeViewer {
		t.Errorf("expected role %s, got %+v", iamtypes.RoleComputeViewer, retrieved.Spec.Roles)
	}

	// List the caller's own bindings: the binding above belongs to the SA, not
	// the caller, so assert the response shape rather than its presence.
	mine, err := client.IAM().RoleBindings().ListMyRoleBindings(ctx)
	if err != nil {
		t.Fatalf("failed to list my role bindings: %v", err)
	}
	if mine.Items == nil || len(*mine.Items) == 0 {
		t.Fatal("expected the caller to have at least one role binding")
	}
	for _, b := range *mine.Items {
		if (b.Metadata.Project == nil) == (b.Metadata.Organization == nil) {
			t.Errorf("role binding %s: expected exactly one of project or organization", b.Metadata.Id)
		}
	}

	// Delete role binding
	t.Logf("Deleting role binding: %s", wantName)
	if err := client.IAM().RoleBindings().DeleteProjectRoleBinding(ctx, wantName); err != nil {
		t.Fatalf("failed to delete role binding: %v", err)
	}
	rbDeleted = true

	if err := client.IAM().ServiceAccounts().Delete(ctx, saID); err != nil {
		t.Fatalf("failed to delete service account: %v", err)
	}
	saDeleted = true
}
