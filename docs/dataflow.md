# FlareGo Data Flow

This document describes how data flows through the FlareGo system, from collection to visualization.

## Overview

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  Agent  │────▶│Gateway  │────▶│  UI     │     │ Storage │
└─────────┘     └─────────┘     └─────────┘     └─────────┘
     │               │               │               ▲
     │               │               │               │
     └───────────────┴───────────────┴───────────────┘
```

## Data Collection

### Agent Level

1. **Sampling**

   - Goroutine states sampled at configurable frequency (1-10000 Hz)
   - GC events and heap statistics collected
   - Stack traces captured for blocked goroutines

2. **Data Processing**

   - Raw samples aggregated into time windows
   - Stack traces symbolized
   - Metrics calculated (e.g., blocked goroutines, heap usage)

3. **Export**
   - Data encoded as JSON flamegraph chunks
   - Compressed for efficient transmission
   - Streamed to gateway via gRPC

### Gateway Level

1. **Ingestion**

   - Receives data from multiple agents
   - Validates and authenticates connections
   - Decompresses and decodes chunks

2. **Processing**

   - Aggregates data from multiple sources
   - Applies alert rules
   - Maintains in-memory ring buffer

3. **Distribution**
   - Streams to connected UI clients
   - Handles client backpressure
   - Manages client subscriptions

## Data Storage

### In-Memory

1. **Ring Buffer**

   - Fixed-size circular buffer
   - Configurable retention period
   - Thread-safe access

2. **Client State**
   - Per-client subscription state
   - Cursor positions
   - Filter settings

### Persistent Storage

1. **File Storage**

   - `.fgo` files for snapshots
   - Gzip compression
   - JSON format

2. **Optional Redis**
   - High-availability setups
   - Shared state between gateway instances
   - Configurable TTL

## Data Visualization

### UI Components

1. **Flame Graph**

   - Interactive D3 visualization
   - Real-time updates
   - Zoom and pan support

2. **Timeline**

   - Historical data navigation
   - Time range selection
   - Playback controls

3. **Metrics**
   - Real-time charts
   - Threshold indicators
   - Alert status

## Alert System

### Rule Evaluation

1. **Expression Language**

   - Simple boolean expressions
   - Metric comparisons
   - Time-based conditions

2. **Alert Processing**
   - Continuous evaluation
   - State tracking
   - Deduplication

### Notification Flow

1. **Sink Types**

   - Log sink (development)
   - Slack integration
   - Webhook support
   - Jira integration

2. **Delivery**
   - Asynchronous notification
   - Retry with backoff
   - Error handling

## Performance Considerations

### Data Volume

1. **Sampling Rate**

   - Default: 100 Hz
   - Adjustable based on load
   - Impact on CPU usage

2. **Compression**
   - gRPC compression
   - JSON optimization
   - Efficient encoding

### Latency

1. **Collection**

   - Agent overhead: <2% CPU
   - Memory usage: <30MB
   - Sampling jitter: <1ms

2. **Transmission**

   - gRPC streaming
   - WebSocket fallback
   - Network optimization

3. **Processing**
   - Lock-free algorithms
   - Copy-on-write updates
   - Efficient data structures

## Security

### Data Protection

1. **Authentication**

   - JWT for agents
   - Bearer tokens for UI
   - TLS encryption

2. **Authorization**
   - Role-based access
   - Namespace isolation
   - Token scoping

### Network Security

1. **Transport**

   - TLS for all connections
   - Certificate validation
   - Secure WebSocket

2. **API Security**
   - Rate limiting
   - Input validation
   - Error handling

## Monitoring

### System Metrics

1. **Agent Metrics**

   - Sampling rate
   - Memory usage
   - CPU overhead

2. **Gateway Metrics**

   - Connected clients
   - Processing latency
   - Storage usage

3. **UI Metrics**
   - Render performance
   - Update frequency
   - Client latency

## Future Improvements

1. **Data Flow**

   - eBPF-based collection
   - Cross-language support
   - Custom samplers

2. **Storage**

   - S3/GCS integration
   - Long-term retention
   - Data compression

3. **Visualization**
   - Custom views
   - Export formats
   - Plugin system
