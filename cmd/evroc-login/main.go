// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

// evroc-login authenticates and displays OAuth tokens.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/evroc-oss/evroc-go-sdk/config"
	"github.com/evroc-oss/evroc-go-sdk/internal/auth"
)

const (
	defaultClientID = "evroc-cli"
)

func defaultAuthURL() string {
	return config.DefaultAuthServerURL + "/auth"
}

func defaultTokenURL() string {
	return config.DefaultAuthServerURL + "/token"
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(ctx context.Context) error {
	fmt.Println("=== evroc OAuth Login ===")
	fmt.Println()

	// Generate PKCE verifier and state for security
	verifier := auth.GenerateVerifier()
	state := auth.GenerateState()

	oidcCfg := auth.OIDCConfig{
		AuthURL:  defaultAuthURL(),
		TokenURL: defaultTokenURL(),
		ClientID: defaultClientID,
		Scopes:   []string{"openid", "offline_access"},
	}

	// Get authorization URL
	authURL := auth.GetAuthorizationURL(oidcCfg, verifier, state)

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If the browser doesn't open, visit: %s\n\n", authURL)

	// Open browser
	if err := openBrowser(ctx, authURL); err != nil {
		fmt.Printf("Warning: Failed to open browser: %v\n", err)
	}

	// Start HTTP server to catch callback
	fmt.Println("Waiting for authentication...")
	_, loginResult, err := auth.NewClientWithHTTPServer(ctx, oidcCfg, verifier, state)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Authentication successful!")
	fmt.Println()
	fmt.Println("=== Copy and paste the following into your terminal ===")
	fmt.Printf("export EVROC_TOKEN='%s'\n", loginResult.AccessToken)
	fmt.Printf("export EVROC_REFRESH_TOKEN='%s'\n", loginResult.RefreshToken)
	fmt.Println()

	return nil
}

func openBrowser(ctx context.Context, url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
