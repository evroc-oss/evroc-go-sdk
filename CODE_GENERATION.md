# Code Generation

This SDK uses code generation to create API client code from OpenAPI specifications.

## For SDK Users

You don't need to regenerate code - all generated files are committed to this repository.
Simply import the SDK and use it:

```bash
go get github.com/evroc-oss/evroc-go-sdk
```

## For Contributors

The code generation tools are available in the repository but require access to
evroc's internal OpenAPI specifications. If you're contributing without access to these
specs, please focus on:

- Bug fixes in existing code
- Documentation improvements
- Example code and use cases
- Helper functions and utilities
- Test coverage improvements

For changes that require API regeneration (new endpoints, schema changes), please open
an issue describing the needed changes and the maintainers will handle the regeneration.

## Maintainers

Code is generated from internal OpenAPI specifications using templates in `internal/codegen/`.
The generation process is handled internally and synced to this public repository to ensure
consistency with evroc's API infrastructure.

### Generated Files

The following directories contain generated code:
- `evroc/compute/` - Compute API client
- `evroc/networking/` - Networking API client
- `evroc/storage/` - Storage API client
- `evroc/iam/` - Identity and Access Management API client
- Parts of `internal/rest/` - Shared REST client utilities

These files should not be manually edited. Changes require regeneration from OpenAPI specs.
