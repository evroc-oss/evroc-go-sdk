# Contributing to evroc Go SDK

Thank you for your interest in contributing to the evroc Go SDK!

## How to Contribute

### Reporting Issues

- Use GitHub Issues to report bugs or request features
- Provide clear reproduction steps for bugs
- Include Go version, SDK version, and relevant environment details
- For API issues, include the API endpoint and request/response details

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests: `go test ./...`
5. Run linter: `go vet ./...` (or `golangci-lint run` if available)
6. Ensure code is formatted: `gofmt -s -w .`
7. Commit your changes with clear commit messages
8. Push to your fork
9. Open a Pull Request with a clear description

### Code Style

- Follow standard Go conventions and idioms
- Run `gofmt` to format code
- Use meaningful variable and function names
- Add tests for new functionality
- Update documentation and examples as needed
- Keep PRs focused and atomic

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Examples

When adding new functionality, consider adding an example in the `examples/` directory
to demonstrate usage. Examples should be:
- Self-contained and runnable
- Well-commented
- Show common use cases

### Code Generation

Parts of this SDK are generated from OpenAPI specifications. Generated code should
not be manually edited. If you need changes that require regeneration (new API endpoints,
schema changes), please open an issue describing the requirements.

See [CODE_GENERATION.md](CODE_GENERATION.md) for details on the code generation process.

### Areas for Contribution

We especially welcome contributions in these areas:
- Bug fixes in non-generated code
- Documentation improvements
- Additional examples and use cases
- Helper functions and utilities
- Test coverage improvements
- Performance optimizations
- Error handling improvements

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/evroc-go-sdk.git
cd evroc-go-sdk

# Install dependencies
go mod download

# Run tests
go test ./...

# Build examples
cd examples/simple
go build .
```

## Code of Conduct

Be respectful, professional, and collaborative in all interactions. We're building
this together to create the best possible SDK for the evroc platform.

## Questions?

Feel free to open an issue for any questions about contributing!
