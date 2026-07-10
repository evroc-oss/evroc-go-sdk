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

	// Delete credential
	t.Logf("Deleting credential: %s", credID)
	if err := client.IAM().ServiceAccountCredentials(saID).Delete(ctx, credID); err != nil {
		t.Fatalf("failed to delete credential: %v", err)
	}
	credDeleted = true

	// Delete service account
	t.Logf("Deleting service account: %s", saID)
	if err := client.IAM().ServiceAccounts().Delete(ctx, saID); err != nil {
		t.Fatalf("failed to delete service account: %v", err)
	}
	saDeleted = true
}
