// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 evroc

package config

// Default configuration values for the evroc SDK.
const (
	// DefaultAuthServerURL is the base URL for the authentication server.
	// Change this to switch environments (e.g., staging).
	DefaultAuthServerURL = "https://authn.iam.evroc.com/realms/evroc-customer/protocol/openid-connect"

	defaultBaseURL  = "https://api.evroc.com"
	defaultClientID = "evroc-cli"
)

// defaultTokenURL constructs the OAuth2 token endpoint URL.
func defaultTokenURL() string {
	return DefaultAuthServerURL + "/token"
}
