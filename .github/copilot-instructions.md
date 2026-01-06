# Copilot Instructions for fork-split-openfeature-provider-go

## Project Overview

This repository is a Snap One fork of the Split OpenFeature Provider for Go. It provides
integration between the OpenFeature SDK and the Split feature flag platform, enabling
controlled rollouts and feature management across OvrC services.

## Technology Stack

- **Language**: Go 1.19+
- **Feature Flags**: Split.io platform
- **SDK**: OpenFeature Go SDK
- **Testing**: Ginkgo/Gomega BDD testing framework

## Project Structure

```
fork-split-openfeature-provider-go/
├── provider.go         # Main OpenFeature provider implementation
├── provider_test.go    # Provider test suite
├── splitClient.go      # Split SDK client wrapper
├── tools.go            # Build tool dependencies
├── mocks/              # Mock implementations for testing
├── go.mod              # Go module definition
└── README.md           # Setup and usage documentation
```

## Key Components

### Provider

The main provider type implements the OpenFeature provider interface:

```go
// Create with simple initialization
provider, err := splitProvider.NewProviderSimple("YOUR_SDK_API_KEY")

// Or with custom Split client
provider, err := splitProvider.NewProvider(*splitClient)
```

### Supported Flag Types

- Boolean flags (`BooleanValue`)
- String flags (`StringValue`)
- Integer flags (`IntValue`)
- Float flags (`FloatValue`)
- Object flags (`ObjectValue`)

## Code Style Guidelines

### Go Idioms

- Follow standard Go formatting (gofmt)
- Use descriptive variable and function names
- Document exported types and functions with godoc comments
- Handle all errors explicitly

### Provider Usage Pattern

```go
import (
    "github.com/open-feature/go-sdk/pkg/openfeature"
    splitProvider "github.com/snap-one/fork-split-openfeature-provider-go"
)

// Initialize provider
provider, err := splitProvider.NewProviderSimple("API_KEY")
if err != nil {
    log.Fatal(err)
}
openfeature.SetProvider(provider)

// Create client and evaluate flags
client := openfeature.NewClient("my-service")
ctx := openfeature.NewEvaluationContext("user-123", nil)

enabled := client.BooleanValue(context.Background(), "feature-flag", false, ctx)
```

### Evaluation Context

The Split provider requires a targeting key for flag evaluation:

```go
// Set evaluation context with targeting key
evalCtx := openfeature.NewEvaluationContext("user-id-123", map[string]interface{}{
    "email": "user@example.com",
    "plan":  "premium",
})

// Context can be set at different levels
openfeature.SetEvaluationContext(evalCtx)        // API level
client.SetEvaluationContext(evalCtx)              // Client level
client.BooleanValue(ctx, "flag", false, evalCtx)  // Per-evaluation
```

## Testing Guidelines

### Running Tests

```bash
go test -v ./...
go test -race ./...  # Race detection
```

### Test Framework

This project uses Ginkgo/Gomega for BDD-style tests:

```go
var _ = Describe("Provider", func() {
    It("should evaluate boolean flags", func() {
        // Test implementation
        Expect(result).To(BeTrue())
    })
})
```

### Mocking

Use the provided mocks in `mocks/` directory for testing:
- Mock Split client for unit tests
- Avoid calling real Split API in tests

## OvrC Integration

### Feature Flag Patterns

```go
// Check if feature is enabled for a user
if client.BooleanValue(ctx, "new-dashboard", false, userContext) {
    // Serve new dashboard
} else {
    // Serve legacy dashboard
}

// Get configuration value
timeout := client.IntValue(ctx, "api-timeout-ms", 5000, nil)
```

### Environment Setup

```bash
export SPLIT_API_KEY="your-api-key"
```

### Common Flag Types in OvrC

- **Release flags**: Enable/disable new features
- **Experiment flags**: A/B testing variants
- **Operational flags**: Runtime configuration
- **Permission flags**: Feature access control

## Security Considerations

- Store Split API keys in environment variables or secrets manager
- Never log API keys or expose them in client-side code
- Use server-side SDK for backend services
- Implement proper error handling for Split SDK failures

## Error Handling

```go
// Provider gracefully handles errors
value := client.BooleanValue(ctx, "flag", defaultValue, evalCtx)
// If evaluation fails, defaultValue is returned

// For detailed error information
details := client.BooleanValueDetails(ctx, "flag", defaultValue, evalCtx)
if details.Error != nil {
    log.Printf("flag evaluation error: %v", details.Error)
}
```

## Fork Maintenance

This is a fork of github.com/splitio/split-openfeature-provider-go. When updating:
- Review upstream changes for compatibility
- Test with OvrC services before merging upstream changes
- Preserve any Snap One-specific modifications
- Update dependencies carefully to avoid breaking changes

## Performance Considerations

- Split SDK caches flag definitions locally
- `BlockUntilReady` ensures SDK is initialized before evaluation
- Consider SDK initialization timeout in startup sequence
- Flag evaluations are fast after initial sync

## Related Repositories

- `pkg-go-feature-flags`: OvrC feature flag utilities
- `app-go-api-server`: Uses feature flags for gradual rollouts
