# FlareGo Usage Guide

This guide provides practical examples and best practices for using FlareGo effectively.

## Basic Usage

### Recording Profiles

1. **Quick Profile**

   ```bash
   # Record a 30-second profile
   flarego record --duration 30s
   ```

2. **Custom Output**

   ```bash
   # Save to specific file
   flarego record --output my-profile.fgo --duration 1m
   ```

3. **Adjust Sampling Rate**
   ```bash
   # Higher frequency for detailed analysis
   flarego record --hz 500 --duration 30s
   ```

### Analyzing Profiles

1. **View Summary**

   ```bash
   # Show profile summary
   flarego replay my-profile.fgo
   ```

2. **Export JSON**

   ```bash
   # Get full JSON data
   flarego replay my-profile.fgo --json > analysis.json
   ```

3. **Compare Profiles**
   ```bash
   # Compare before/after changes
   flarego diff before.fgo after.fgo
   ```

## Advanced Usage

### Live Monitoring

1. **Attach to Process**

   ```bash
   # Monitor local process
   flarego attach --gateway localhost:4317
   ```

2. **Custom Duration**

   ```bash
   # Monitor for specific time
   flarego attach --duration 5m --gateway localhost:4317
   ```

3. **High-Frequency Sampling**
   ```bash
   # Detailed monitoring
   flarego attach --hz 1000 --duration 1m
   ```

### Kubernetes Integration

1. **Attach to Pod**

   ```bash
   # Monitor specific pod
   flarego kubectl attach -n my-namespace my-pod
   ```

2. **Long-term Monitoring**
   ```bash
   # Monitor with periodic recording
   while true; do
     flarego record --duration 5m --output k8s-$(date +%s).fgo
     sleep 1m
   done
   ```

## Best Practices

### Performance Profiling

1. **Baseline Recording**

   ```bash
   # Record baseline
   flarego record --duration 1m --output baseline.fgo
   ```

2. **Load Testing**

   ```bash
   # Record under load
   flarego record --duration 5m --output load-test.fgo
   ```

3. **Compare Results**
   ```bash
   # Analyze differences
   flarego diff baseline.fgo load-test.fgo
   ```

### Production Monitoring

1. **Regular Sampling**

   ```bash
   # Record every hour
   flarego record --duration 5m --output prod-$(date +%Y%m%d-%H).fgo
   ```

2. **High-Frequency During Issues**

   ```bash
   # Detailed recording during incidents
   flarego record --hz 500 --duration 10m --output incident.fgo
   ```

3. **Long-term Storage**
   ```bash
   # Compress and archive
   gzip *.fgo
   mv *.fgo.gz /archive/profiles/
   ```

### Development Workflow

1. **Before Changes**

   ```bash
   # Record baseline
   flarego record --duration 30s --output before.fgo
   ```

2. **After Changes**

   ```bash
   # Record after changes
   flarego record --duration 30s --output after.fgo
   ```

3. **Analyze Impact**
   ```bash
   # Compare changes
   flarego diff before.fgo after.fgo
   ```

## Troubleshooting

### Common Scenarios

1. **High CPU Usage**

   ```bash
   # Reduce sampling frequency
   flarego record --hz 10 --duration 30s
   ```

2. **Large File Size**

   ```bash
   # Enable compression
   flarego record --duration 1m --output compressed.fgo
   ```

3. **Connection Issues**
   ```bash
   # Test connectivity
   flarego attach --gateway localhost:4317 --duration 5s
   ```

### Performance Tips

1. **Sampling Frequency**

   - Use lower frequencies (10-100 Hz) for long-term monitoring
   - Use higher frequencies (500-1000 Hz) for detailed analysis
   - Adjust based on system load

2. **Recording Duration**

   - Short durations (5-30s) for quick analysis
   - Medium durations (1-5m) for performance testing
   - Long durations (30m+) for production monitoring

3. **Storage Management**
   - Enable compression for long-term storage
   - Regular cleanup of old profiles
   - Archive important profiles

## Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/performance.yml
name: Performance Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run performance test
        run: |
          flarego record --duration 30s --output ci-profile.fgo
          # Analyze and fail if performance degraded
```

### Monitoring Dashboard

```bash
# Record profiles for dashboard
while true; do
  flarego record --duration 5m --output dashboard-$(date +%s).fgo
  sleep 5m
done
```

### Alert System

```bash
# Monitor and alert on issues
flarego attach --duration 1h | while read line; do
  if echo "$line" | grep -q "high_latency"; then
    send_alert "High latency detected"
  fi
done
```

## Web Interface

### Running the Web UI

The FlareGo web interface provides real-time visualization of your application's performance data. There are several ways to run it:

1. **Development Mode**

   ```bash
   # Start the development environment
   make dev
   ```

   This will:

   - Start the FlareGo gateway
   - Launch the web UI in development mode
   - Enable hot-reloading for development

2. **Production Mode**

   ```bash
   # Build the web UI
   cd web && npm run build

   # Start the gateway with web UI
   ./bin/flarego-gateway
   ```

3. **Docker**
   ```bash
   # Run using Docker Compose
   docker compose -f deployments/docker-compose.yaml up -d
   ```

### Accessing the Web UI

Once running, the web interface is available at:

- Development: `http://localhost:3000`
- Production: `http://localhost:8080`

### Web UI Features

1. **Real-time Visualization**

   - Live flame graphs
   - Timeline view
   - Metrics dashboard

2. **Interactive Analysis**

   - Zoom and pan
   - Stack trace inspection
   - Metric correlation

3. **Alert Management**
   - View active alerts
   - Configure alert rules
   - Set up notifications

### Stopping the Web UI

```bash
# Stop development environment
make stop-dev

# Or using Docker Compose
docker compose -f deployments/docker-compose.yaml down
```

## Next Steps

- Read the [CLI Reference](cli-reference.md) for detailed command options
- Check the [Architecture](architecture.md) for system design
- Review the [Installation Guide](installation.md) for setup instructions
