# FlareGo Plugin Development Guide

This guide explains how to develop plugins for FlareGo, including samplers, exporters, and other extensions.

## Overview

FlareGo's plugin system allows extending functionality through:

- Custom samplers for data collection
- Custom exporters for data output
- Custom alert sinks for notifications
- Custom visualizations for the UI

## Plugin Types

### 1. Samplers

Samplers collect runtime data from your application.

```go
// Example sampler plugin
package mysampler

import (
    "github.com/Voskan/flarego/internal/plugins"
)

type MySampler struct{}

func (p *MySampler) Kind() plugins.Kind {
    return "sampler"
}

func (p *MySampler) Name() string {
    return "mysampler"
}

func (p *MySampler) Init() (any, error) {
    // Initialize your sampler
    return nil, nil
}

func init() {
    plugins.Register(&MySampler{})
}
```

### 2. Exporters

Exporters send data to external systems.

```go
// Example exporter plugin
package myexporter

import (
    "context"
    "github.com/Voskan/flarego/internal/plugins"
    "github.com/Voskan/flarego/pkg/flamegraph"
)

type MyExporter struct{}

func (p *MyExporter) Kind() plugins.Kind {
    return "exporter"
}

func (p *MyExporter) Name() string {
    return "myexporter"
}

func (p *MyExporter) Init() (any, error) {
    // Initialize your exporter
    return nil, nil
}

func (p *MyExporter) Export(ctx context.Context, root *flamegraph.Frame) error {
    // Export data
    return nil
}

func (p *MyExporter) Close() error {
    // Cleanup
    return nil
}

func init() {
    plugins.Register(&MyExporter{})
}
```

### 3. Alert Sinks

Alert sinks handle notification delivery.

```go
// Example alert sink plugin
package mysink

import (
    "github.com/Voskan/flarego/internal/plugins"
)

type MySink struct{}

func (p *MySink) Kind() plugins.Kind {
    return "sink"
}

func (p *MySink) Name() string {
    return "mysink"
}

func (p *MySink) Init() (any, error) {
    // Initialize your sink
    return nil, nil
}

func (p *MySink) Notify(ruleName, msg string) {
    // Send notification
}

func init() {
    plugins.Register(&MySink{})
}
```

## Plugin Development

### 1. Project Structure

```
my-flarego-plugin/
├── go.mod
├── go.sum
├── main.go
└── README.md
```

### 2. Dependencies

```go
module github.com/yourusername/flarego-mysampler

go 1.24

require (
    github.com/Voskan/flarego v0.1.0
)
```

### 3. Configuration

```yaml
# config.yaml
plugins:
  mysampler:
    enabled: true
    config:
      interval: "1s"
      threshold: 100
```

### 4. Building

```bash
# Build plugin
go build -buildmode=plugin -o mysampler.so

# Or use Makefile
make build
```

## Best Practices

### 1. Error Handling

```go
func (p *MyPlugin) Init() (any, error) {
    if err := p.validateConfig(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    return nil, nil
}
```

### 2. Resource Management

```go
func (p *MyPlugin) Close() error {
    // Clean up resources
    if p.client != nil {
        return p.client.Close()
    }
    return nil
}
```

### 3. Logging

```go
import "github.com/Voskan/flarego/internal/logging"

func (p *MyPlugin) Notify(rule, msg string) {
    logging.Sugar().Infow("sending notification",
        "rule", rule,
        "msg", msg)
}
```

## Testing

### 1. Unit Tests

```go
func TestMyPlugin(t *testing.T) {
    p := &MyPlugin{}
    if err := p.Init(); err != nil {
        t.Fatal(err)
    }
    // Test functionality
}
```

### 2. Integration Tests

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // Test with FlareGo
}
```

## Deployment

### 1. Local Development

```bash
# Build and install
make install

# Run with plugin
flarego attach --plugin mysampler.so
```

### 2. Docker

```dockerfile
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -buildmode=plugin -o mysampler.so

FROM ghcr.io/flarego/agent:latest
COPY --from=builder /app/mysampler.so /plugins/
```

### 3. Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flarego-agent
spec:
  template:
    spec:
      containers:
        - name: agent
          image: ghcr.io/flarego/agent:latest
          volumeMounts:
            - name: plugins
              mountPath: /plugins
      volumes:
        - name: plugins
          configMap:
            name: flarego-plugins
```

## Security

### 1. Input Validation

```go
func (p *MyPlugin) validateConfig() error {
    if p.config.Interval <= 0 {
        return errors.New("interval must be positive")
    }
    return nil
}
```

### 2. Resource Limits

```go
func (p *MyPlugin) Init() (any, error) {
    // Set resource limits
    p.maxConcurrent = 10
    p.timeout = 5 * time.Second
    return nil, nil
}
```

## Performance

### 1. Efficient Data Structures

```go
type MyPlugin struct {
    cache    *lru.Cache
    metrics  sync.Map
    samples  chan Sample
}
```

### 2. Concurrency

```go
func (p *MyPlugin) processSamples() {
    for sample := range p.samples {
        go p.handleSample(sample)
    }
}
```

## Troubleshooting

### 1. Common Issues

1. **Plugin Loading**

   - Check file permissions
   - Verify Go version match
   - Check build mode

2. **Configuration**

   - Validate YAML syntax
   - Check required fields
   - Verify paths

3. **Performance**
   - Monitor resource usage
   - Check for leaks
   - Profile hot paths

### 2. Debugging

```go
import "github.com/Voskan/flarego/internal/logging"

func (p *MyPlugin) debug() {
    logging.Sugar().Debugw("plugin state",
        "cache_size", p.cache.Len(),
        "metrics_count", p.metricsCount)
}
```

## Future Improvements

1. **Plugin System**

   - Hot reloading
   - Version compatibility
   - Dependency management

2. **Development Tools**

   - Plugin SDK
   - Testing framework
   - Debugging tools

3. **Integration**
   - More plugin types
   - Better documentation
   - Example plugins
