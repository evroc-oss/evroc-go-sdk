# Configuration

The SDK provides multiple ways to configure authentication and settings. It uses a credential chain that automatically checks multiple sources in order.

## Configuration Methods

### 1. evroc CLI Config (Recommended for Development)

The SDK automatically reads the evroc CLI config from your home directory if present. This file is created by the `evroc login` command.

**Config location:**
- Unix/Linux/macOS: `~/.evroc/config.yaml`
- Windows: `%USERPROFILE%\.evroc\config.yaml`

Install the CLI from [docs.evroc.com/cli.html](https://docs.evroc.com/cli.html), then:

```bash
evroc login
```

The SDK automatically detects and uses this configuration:

```go
// Automatically discovers and uses CLI config if available
client, err := evroc.NewFromEnv(ctx)
```

Or explicitly load from CLI config:

```go
// Use default CLI config location (~/.evroc/config.yaml)
client, err := evroc.NewFromCLIConfig(ctx, "")

// Use custom config path
client, err := evroc.NewFromCLIConfig(ctx, "/path/to/config.yaml")
```

Note: The SDK uses the `currentProfile` from the config file automatically.

The CLI config format:

```yaml
formatVersion: v1
profiles:
  default:
    domain: evroc.com
    apiURL: https://api.evroc.com
    issuerURL: https://authn.iam.evroc.com/realms/evroc-customer
    s3URL: https://s3.se-sto.evroc.com
    organization: cf45341a-acbb-4cf4-9e5e-b7eef91caca2
    project: f1dbc774-c978-4e75-bcf9-a28015a181f7
    region: se-sto
    user:
      username: user@evroc.com
      refreshToken: eyJhbG...
currentProfile: default
```

### 2. Environment Variables

Environment variables override CLI config values.

**User or Service Account Credentials:**

```bash
export EVROC_USERNAME="user@evroc.com"          # or service-account-id
export EVROC_PASSWORD="your-password"            # or service-account-secret
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
export EVROC_ORGANIZATION="org-uuid"  # optional
```

**Direct Tokens:**

```bash
# Option A: Just refresh token (recommended - token will be obtained automatically)
export EVROC_REFRESH_TOKEN="your-refresh-token"
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
export EVROC_ORGANIZATION="org-uuid"  # optional

# Option B: Access token with optional refresh token
export EVROC_TOKEN="your-access-token"
export EVROC_REFRESH_TOKEN="your-refresh-token"  # optional, enables auto-refresh
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
export EVROC_ORGANIZATION="org-uuid"  # optional
```

```go
client, err := evroc.NewFromEnv(ctx)
```

### 3. SDK Configuration File

Create a custom `config.yaml` for the SDK (different from CLI config format):

**User or Service Account Credentials:**

```yaml
auth:
  username: "user@evroc.com"    # or service-account-id
  password: "your-password"      # or service-account-secret

context:
  project: "project-uuid"
  region: "se-sto"
  organization: "org-uuid"  # optional
```

**Direct Tokens:**

```yaml
# Option A: Just refresh token (recommended)
auth:
  refresh_token: "your-refresh-token"

context:
  project: "project-uuid"
  region: "se-sto"
  organization: "org-uuid"  # optional

# Option B: Access token with optional refresh token
auth:
  token: "your-access-token"
  refresh_token: "your-refresh-token"  # optional, enables auto-refresh

context:
  project: "project-uuid"
  region: "se-sto"
  organization: "org-uuid"  # optional
```

```go
client, err := evroc.NewFromFile(ctx, "config.yaml")
```

### 4. Programmatic Configuration

```go
cfg := &config.Config{
    Auth: config.AuthConfig{
        Username: "user@evroc.com",    // or service-account-id
        Password: "your-password",      // or service-account-secret
    },
    Context: config.ContextConfig{
        Project:      "project-uuid",
        Region:       "se-sto",
        Organization: "org-uuid",
    },
}

client, err := evroc.New(ctx, cfg)
```

## Configuration Precedence

The SDK follows this priority order (highest to lowest):

1. **Environment variables** - Always override other sources
2. **Explicit configuration** - `NewFromFile()`, `NewFromCLIConfig()`, or `New()`
3. **CLI config fallback** - `NewFromEnv()` automatically checks the CLI config file if env vars are incomplete

## Authentication Methods

### User Authentication (Interactive Development)

Use your evroc user account credentials for development:

```bash
# Method 1: Using evroc CLI (recommended)
evroc login

# Method 2: Using environment variables
export EVROC_USERNAME="user@evroc.com"
export EVROC_PASSWORD="your-password"
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
```

### Service Account Authentication (Production/CI-CD)

Use service account credentials for automated systems:

1. Contact evroc support (support@evroc.com) to obtain service account credentials
2. Use the service account ID as username and secret as password:

```bash
export EVROC_USERNAME="service-account-id"
export EVROC_PASSWORD="service-account-secret"
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
```

Service accounts use the same authentication mechanism as user accounts, just with different credentials.

## Custom HTTP Client (Testing)

For testing or when you need custom HTTP behavior, use the functional options pattern:

```go
// Custom HTTP client for testing
customClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &mockTransport{}, // Your test transport
}

client, err := evroc.NewFromEnv(ctx, evroc.WithHTTPClient(customClient))
if err != nil {
    log.Fatal(err)
}

// Or with programmatic config
client, err := evroc.New(ctx, cfg, evroc.WithHTTPClient(customClient))
```
