# FlareGo Architecture

## Overview

FlareGo is a live scheduler flame-visualizer for Go applications consisting of three main components:

```
               ┌────────────┐
               │   IDE UI   │
               └────────────┘
                     ▲ WebSocket / gRPC
┌─────────┐   gRPC   │
│  Agent  │◀─────────┼─────┐
└─────────┘          │     │
   (in-proc)     ┌─────────────┐
                 │ Gateway/API │───► S3 archive (.fgo)
                 └─────────────┘
```

## Components

### Agent (`flarego-agent`)

- **Purpose**: In-process data collection
- **Technology**: Go 1.24, runtime/trace
- **Responsibilities**:
  - Collect goroutine, GC, heap, blocked states
  - Down-sample and encode events
  - Stream to gateway via gRPC
  - Expose health endpoint

### Gateway (`flarego-gateway`)

- **Purpose**: Central aggregation and fan-out
- **Technology**: Go, gRPC-Gateway, Redis (optional)
- **Responsibilities**:
  - Authentication and authorization
  - Fan-out to UI subscribers
  - Buffering and retention (15min default)
  - Alert rule engine
  - WebSocket/HTTP API for UI

### UI Dashboard (`web/`)

- **Purpose**: Interactive visualization
- **Technology**: TypeScript, React + D3 flamegraph
- **Responsibilities**:
  - Render interactive flame graphs
  - Timeline scrubbing and replay
  - Real-time streaming from gateway
  - Save/load session clips

## Data Flow

1. **Collection**: Agent samplers collect runtime data at 1kHz
2. **Encoding**: Data encoded as JSON flamegraph chunks
3. **Transport**: gRPC streaming to gateway with optional compression
4. **Storage**: Gateway maintains in-memory ring buffer + optional Redis
5. **Distribution**: WebSocket/gRPC-Web streaming to UI clients
6. **Visualization**: React components render D3 flame graphs

## Key Design Decisions

### Performance First

- Agent overhead target: <2% CPU, <30MB RAM
- Lock-free sampling where possible
- Copy-on-write flamegraph building
- Adaptive sampling rates based on load

### Real-time Focus

- Target latency: <1 second from event to UI
- Streaming protocols throughout
- Non-blocking fan-out to slow consumers

### Embeddable

- Agent as import: `import _ "github.com/flarego/agent"`
- UI as WebComponent for IDE integration
- Minimal external dependencies

## Security Model

- JWT authentication for agents
- Bearer token auth for UI clients
- TLS-only in production
- Per-namespace RBAC (future)
- Token-scoped permissions (future)

## Deployment Patterns

### Development

```bash
# Local development
make dev
# or
flarego attach --pid 1234 --open
```

### Production

```yaml
# Kubernetes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flarego-gateway
# ... see deployments/kubernetes/
```

### Sidecar

```yaml
# As sidecar container
containers:
  - name: app
    image: myapp:latest
  - name: flarego-agent
    image: ghcr.io/flarego/agent:latest
    args: ["--gateway", "flarego-gateway:4317"]
```

## Extension Points

### Samplers

Custom samplers implement the `Sampler` interface:

```go
type Sampler interface {
    Start()
    Stop()
}
```

### Exporters

Custom exporters implement the `Exporter` interface:

```go
type Exporter interface {
    Export(ctx context.Context, root *flamegraph.Frame) error
    Close() error
}
```

### Alert Sinks

Custom alert sinks for notifications:

```go
type Sink interface {
    Notify(ruleName, msg string)
}
```

## Performance Characteristics

| Scenario                    | Added CPU | Added RSS | Latency to UI |
| --------------------------- | --------- | --------- | ------------- |
| 2k goroutines, 50k events/s | ≤ 1.8%    | ≤ 25MB    | < 800ms       |
| 10k goroutines spike        | ≤ 5%      | ≤ 60MB    | < 1.5s        |

## Future Architecture

### Multi-tenant Gateway

- Per-tenant namespaces
- Resource quotas
- Audit logging

### eBPF Integration

- Dynamic attach without code changes
- Kernel-level event collection
- Cross-language support

### Cloud Storage

- S3/GCS archive integration
- Long-term retention
- Compliance features
