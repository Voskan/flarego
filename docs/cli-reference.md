# FlareGo CLI Reference

## Overview

The `flarego` CLI provides commands for capturing, recording, and replaying Go runtime scheduler flame graphs.

## Installation

```bash
# Build from source
make build

# Or download from releases
wget https://github.com/Voskan/flarego/releases/latest/download/flarego_linux_amd64.tar.gz
```

## Commands

### `flarego attach`

Attach to a running Go process and stream samples to a gateway.

```bash
flarego attach [flags]
```

**Flags:**

- `--gateway string` - FlareGo gateway gRPC address (default "localhost:4317")
- `--hz int` - Sampling frequency in Hz (default 100)
- `--duration duration` - Optional run time (0 = run until Ctrl-C)

**Examples:**

```bash
# Attach to current process and stream to local gateway
flarego attach

# Attach with custom gateway and higher frequency
flarego attach --gateway prod-gateway:4317 --hz 500

# Attach for a specific duration
flarego attach --duration 30s
```

### `flarego record`

Record a local flame graph snapshot to a .fgo file.

```bash
flarego record [flags]
```

**Flags:**

- `--duration, -d duration` - Recording duration (default 30s)
- `--output, -o string` - Output .fgo file path (default auto-named)
- `--hz int` - Sampling frequency in Hz (default 100)
- `--no-compress` - Disable gzip compression of output file

**Examples:**

```bash
# Record for 30 seconds
flarego record --duration 30s

# Record to specific file
flarego record --output my-trace.fgo --duration 60s

# Record uncompressed
flarego record --duration 10s --no-compress
```

### `flarego replay`

Inspect a recorded .fgo flamegraph file.

```bash
flarego replay <file.fgo> [flags]
```

**Flags:**

- `--json` - Output full flamegraph JSON instead of summary

**Examples:**

```bash
# Show summary of recorded trace
flarego replay trace.fgo

# Output full JSON
flarego replay trace.fgo --json
```

### `flarego version`

Print FlareGo version information.

```bash
flarego version [flags]
```

**Flags:**

- `--json` - Print version information as JSON

**Examples:**

```bash
# Human-readable version
flarego version

# JSON format
flarego version --json
```

## Global Flags

- `--config string` - Path to configuration file (YAML/TOML/JSON)
- `--log-json` - Enable JSON log output (default is human-friendly console)

## Configuration

FlareGo can be configured via:

1. Command-line flags (highest priority)
2. Environment variables with `FLAREGO_` prefix
3. Configuration file
4. Default values (lowest priority)

### Environment Variables

- `FLAREGO_GATEWAY` - Default gateway address
- `FLAREGO_HZ` - Default sampling frequency
- `FLAREGO_LOG_JSON` - Enable JSON logging

### Configuration File

```yaml
# ~/.config/flarego/config.yaml
gateway: "localhost:4317"
hz: 100
log_json: false
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `130` - Interrupted (Ctrl-C)

## Examples

### Development Workflow

```bash
# Start local development environment
make dev

# In another terminal, attach to a test process
flarego attach --gateway localhost:4317

# Record a baseline before changes
flarego record --duration 60s --output before.fgo

# Make changes, then record again
flarego record --duration 60s --output after.fgo

# Compare (future feature)
flarego diff before.fgo after.fgo
```

### Production Monitoring

```bash
# Monitor production service via gateway
flarego attach --gateway prod.example.com:4317 --duration 5m

# Save critical moments
flarego record --duration 30s --output incident-$(date +%s).fgo
```

### CI/CD Integration

```bash
#!/bin/bash
# Performance regression test
flarego record --duration 30s --output ci-trace.fgo
# Process trace and fail build if regression detected
# (custom analysis script)
```

## Troubleshooting

### Connection Issues

```bash
# Test gateway connectivity
curl -v http://gateway:8080/metrics

# Check agent logs
flarego attach --log-json 2>&1 | jq
```

### Performance Issues

```bash
# Reduce sampling frequency
flarego attach --hz 10

# Record shorter durations
flarego record --duration 5s
```

### File Format Issues

```bash
# Check if file is compressed
file trace.fgo

# Decompress manually if needed
gunzip < trace.fgo.gz > trace.json
```
