# FlareGo Development Guide

This guide provides information for developers who want to contribute to FlareGo.

## Development Environment

### Prerequisites

1. **Go Environment**

   ```bash
   # Install Go 1.24 or later
   go version

   # Set up GOPATH
   export GOPATH=$HOME/go
   export PATH=$PATH:$GOPATH/bin
   ```

2. **Build Tools**

   ```bash
   # Install build dependencies
   make deps
   ```

3. **Development Tools**
   ```bash
   # Install development tools
   go install golang.org/x/tools/cmd/godoc@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

### Project Structure

```
flarego/
├── cmd/              # Command-line tools
│   └── flarego/     # Main CLI
├── internal/         # Private application code
│   ├── agent/       # Agent implementation
│   ├── exporter/    # Data exporters
│   └── sampler/     # Data samplers
├── pkg/             # Public libraries
│   └── flamegraph/  # Flame graph data structures
├── web/             # Web interface
├── examples/        # Example code
└── docs/            # Documentation
```

## Building

### Local Development

1. **Build Binary**

   ```bash
   make build
   ```

2. **Run Tests**

   ```bash
   make test
   ```

3. **Lint Code**
   ```bash
   make lint
   ```

### Development Workflow

1. **Create Branch**

   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make Changes**

   ```bash
   # Edit code
   vim internal/agent/collector.go

   # Run tests
   make test
   ```

3. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

## Code Style

### Go Code

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing

### Documentation

- Use [godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc) style comments
- Document all exported types and functions
- Include examples for public APIs

## Testing

### Unit Tests

```go
func TestCollector(t *testing.T) {
    // Test implementation
}
```

### Integration Tests

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // Test implementation
}
```

### Benchmarks

```go
func BenchmarkCollector(b *testing.B) {
    // Benchmark implementation
}
```

## Adding New Features

### 1. Agent Extensions

```go
// internal/agent/sampler/sampler.go
type Sampler interface {
    Start()
    Stop()
}

// Example implementation
type MySampler struct {
    // Implementation
}
```

### 2. Exporters

```go
// internal/agent/exporter/exporter.go
type Exporter interface {
    Export(ctx context.Context, root *flamegraph.Frame) error
    Close() error
}

// Example implementation
type MyExporter struct {
    // Implementation
}
```

### 3. CLI Commands

```go
// cmd/flarego/mycmd.go
func newMyCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "mycmd",
        Short: "Description of my command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
}
```

## Performance Considerations

### Memory Usage

- Use object pools for frequently allocated objects
- Implement proper cleanup in `Stop()` methods
- Monitor memory usage in benchmarks

### CPU Usage

- Use efficient data structures
- Implement sampling rate control
- Profile hot paths

### I/O Operations

- Use buffered I/O
- Implement backpressure
- Handle errors gracefully

## Debugging

### Logging

```go
import "github.com/Voskan/flarego/internal/logging"

// Use structured logging
logging.Sugar().Infow("event", "key", value)
```

### Profiling

```bash
# CPU profile
go test -cpuprofile cpu.prof

# Memory profile
go test -memprofile mem.prof

# Analyze profiles
go tool pprof cpu.prof
```

## Release Process

1. **Version Bump**

   ```bash
   # Update version in pkg/version/version.go
   make version
   ```

2. **Build Release**

   ```bash
   make release
   ```

3. **Create Tag**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

## Contributing

### Pull Requests

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Run tests and linting
5. Submit pull request

### Code Review

- All PRs require review
- CI must pass
- Documentation must be updated
- Tests must be added

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Project Wiki](https://github.com/Voskan/flarego/wiki)
- [Issue Tracker](https://github.com/Voskan/flarego/issues)

## Next Steps

- Read the [Architecture](architecture.md) document
- Check the [CLI Reference](cli-reference.md)
- Review the [Usage Guide](usage-guide.md)
