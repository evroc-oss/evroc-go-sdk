// Copyright 2026 evroc
// SPDX-License-Identifier: Apache-2.0

// Service Account Lifecycle Example
//
// Creates a service account, assigns IAM roles so it can actually do useful
// work, then creates a credential whose private key can be used for
// non-interactive authentication (CI/CD, Terraform, scripts).
//
// To authenticate with the resulting credential, see examples/service-account-auth.
//
// Usage:
//
//	export EVROC_PROJECT=my-project
//	export EVROC_REGION=se-sto
//	go run main.go            # create everything
//	go run main.go destroy    # clean up
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	evroc "github.com/evroc-oss/evroc-go-sdk"
	"github.com/evroc-oss/evroc-go-sdk/iam"
	iamtypes "github.com/evroc-oss/evroc-go-sdk/types/iam"
)

const (
	saID    = "sdk-example-sa"
	credID  = "sdk-example-cred"
	keyFile = "private.jwk"
)

func main() {
	ctx := context.Background()

	client, err := evroc.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "destroy" {
		destroy(ctx, client)
		return
	}

	project := client.DefaultProject()

	// Step 1: Browse the role catalog
	fmt.Println("1. Available roles:")
	roles, err := client.IAM().RoleBindings().ListRoles(ctx)
	if err != nil {
		log.Fatalf("Failed to list roles: %v", err)
	}
	if roles.Items != nil {
		for _, r := range *roles.Items {
			desc := ""
			if r.Description != nil {
				desc = *r.Description
			}
			fmt.Printf("   %-42s %s\n", r.Id, desc)
		}
	}

	// Step 2: Create a service account
	fmt.Println("\n2. Creating service account...")
	sa, err := iam.NewServiceAccountBuilder(saID, project).
		WithDescription("SDK example service account").
		Create(ctx, client.IAM().ServiceAccounts())
	if err != nil {
		log.Fatalf("Failed to create service account: %v", err)
	}
	fmt.Printf("   Created: %s\n", sa.Metadata.Id)

	// Step 3: Assign roles — without this the SA can authenticate but cannot access any resources
	fmt.Println("\n3. Assigning roles...")
	saRef := client.IAM().ServiceAccountRef(saID)

	for _, role := range []string{
		iamtypes.RoleComputeOperator,
		iamtypes.RoleNetworkingOperator,
		iamtypes.RoleStorageOperator,
	} {
		_, err := client.IAM().RoleBindings().AssignProjectRole(ctx, &iamtypes.AssignRoleRequest{
			Principal: saRef,
			Role:      role,
		})
		if err != nil {
			log.Fatalf("Failed to assign %s: %v", role, err)
		}
		fmt.Printf("   Assigned: %s\n", role)
	}

	// Step 4: Create a credential — the private key is only returned once
	fmt.Println("\n4. Creating credential...")
	cred, err := iam.NewServiceAccountCredentialBuilder(credID, project, saRef, time.Now().AddDate(0, 0, 30)).
		WithDescription("SDK example credential").
		WithAccessTokenLifetime(3600).
		Create(ctx, client.IAM().ServiceAccountCredentials(saID))
	if err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}

	if cred.Status.PrivateKeyJwk == nil {
		log.Fatal("No private key returned — the key is only available at creation time")
	}

	decoded, err := base64.StdEncoding.DecodeString(*cred.Status.PrivateKeyJwk)
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}
	if err := os.WriteFile(keyFile, decoded, 0600); err != nil {
		log.Fatalf("Failed to write key file: %v", err)
	}
	fmt.Printf("   Created: %s (key saved to %s)\n", cred.Metadata.Id, keyFile)

	// Done
	fmt.Println("\n--- Ready ---")
	fmt.Println("The service account can now manage compute, networking, and storage resources.")
	fmt.Println()
	absKeyFile, _ := filepath.Abs(keyFile)
	fmt.Println("To authenticate (project and region are read from ~/.evroc/config.yaml):")
	fmt.Printf("  export EVROC_SERVICE_ACCOUNT_ID=%s\n", saID)
	fmt.Printf("  export EVROC_SERVICE_ACCOUNT_SECRET=%s\n", absKeyFile)
	fmt.Println()
	fmt.Println("To clean up:  go run main.go destroy")
}

func destroy(ctx context.Context, client *evroc.Client) {
	fmt.Println("Cleaning up...")

	if err := client.IAM().ServiceAccountCredentials(saID).Delete(ctx, credID); err != nil {
		log.Printf("  credential: %v", err)
	} else {
		fmt.Println("  Deleted credential")
	}

	if err := client.IAM().ServiceAccounts().Delete(ctx, saID); err != nil {
		log.Printf("  service account: %v", err)
	} else {
		fmt.Println("  Deleted service account")
	}

	os.Remove(keyFile)
	fmt.Println("  Done")
}
