# Configuration

## Environment Variables

**Option 1: OAuth Tokens (recommended, use evroc-login)**

```bash
export EVROC_TOKEN="your-access-token"
export EVROC_REFRESH_TOKEN="your-refresh-token"
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
export EVROC_ORGANIZATION="org-uuid"  # optional
```

**Option 2: Username/Password**

```bash
export EVROC_USERNAME="user@evroc.com"
export EVROC_PASSWORD="your-password"
export EVROC_PROJECT="project-uuid"
export EVROC_REGION="se-sto"
export EVROC_ORGANIZATION="org-uuid"  # optional
```

```go
client, err := evroc.NewFromEnv(ctx)
```

## Configuration File

Create `config.yaml` with one of the following auth methods:

**Option 1: OAuth Tokens (recommended, use evroc-login)**

```yaml
auth:
  token: "your-access-token"
  refresh_token: "your-refresh-token"

context:
  project: "project-uuid"
  region: "se-sto"
  organization: "org-uuid"  # optional
```

**Option 2: Username/Password**

```yaml
auth:
  username: "user@evroc.com"
  password: "your-password"

context:
  project: "project-uuid"
  region: "se-sto"
  organization: "org-uuid"  # optional
```

```go
client, err := evroc.NewFromFile(ctx, "config.yaml")
```

## Programmatic Configuration

```go
cfg := &config.Config{
    Auth: config.AuthConfig{
        Username: "user@evroc.com",
        Password: "your-password",
    },
    Context: config.ContextConfig{
        Project:      "project-uuid",
        Region:       "se-sto",
        Organization: "org-uuid",
    },
}

client, err := evroc.New(ctx, cfg)
```

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
