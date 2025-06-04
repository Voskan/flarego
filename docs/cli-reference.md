# FlareGo CLI Reference

This document provides a comprehensive reference for all FlareGo CLI commands and their options.

## Global Options

These options are available for all commands:

- `--config string` - Path to configuration file (YAML/TOML/JSON)
- `--log-json` - Enable JSON log output (default is human-friendly console)

## Commands

### attach

Starts a local agent and streams samples to a FlareGo gateway.

```bash
flarego attach [flags]
```

#### Options

- `--gateway string` - FlareGo gateway gRPC address (host:port) (default "localhost:4317")
- `--hz int` - Sampling frequency in Hz (1-10000) (default 100)
- `--duration duration` - Optional run time (e.g., 30s); 0 = run until Ctrl-C

#### Example

```bash
# Attach to local gateway for 30 seconds
flarego attach --gateway localhost:4317 --duration 30s

# Attach with custom sampling rate
flarego attach --hz 500 --gateway localhost:4317
```

### record

Records a local flame graph snapshot to a .fgo file.

```bash
flarego record [flags]
```

#### Options

- `-d, --duration duration` - Recording duration (e.g., 30s, 2m) (default 30s)
- `-o, --output string` - Output .fgo file path (default auto-named)
- `--hz int` - Sampling frequency in Hz (default 100)
- `--no-compress` - Disable gzip compression of output file

#### Example

```bash
# Record for 1 minute with default settings
flarego record --duration 1m

# Record with custom output file
flarego record --output my-profile.fgo --duration 30s
```

### replay

Inspects a recorded .fgo flamegraph file.

```bash
flarego replay <file.fgo> [flags]
```

#### Options

- `--json` - Output full flamegraph JSON instead of summary

#### Example

```bash
# View flamegraph summary
flarego replay my-profile.fgo

# Get full JSON output
flarego replay my-profile.fgo --json
```

### diff

Shows the difference between two .fgo flamegraph files.

```bash
flarego diff <before.fgo> <after.fgo>
```

#### Example

```bash
# Compare two profiles
flarego diff before.fgo after.fgo
```

### ebpf-attach

Attaches to a running Go process using eBPF uprobes (Linux only).

```bash
flarego ebpf-attach <pid>
```

#### Example

```bash
# Attach to process with PID 1234
flarego ebpf-attach 1234
```

### kubectl

Port-forwards and attaches to a Kubernetes Pod.

```bash
flarego kubectl attach -n <namespace> <resource>
```

#### Example

```bash
# Attach to a pod in the default namespace
flarego kubectl attach my-pod

# Attach to a pod in a specific namespace
flarego kubectl attach -n my-namespace my-pod
```

### version

Prints FlareGo version information.

```bash
flarego version [flags]
```

#### Options

- `--json` - Print version information as JSON

#### Example

```bash
# Print version
flarego version

# Get version as JSON
flarego version --json
```

## Configuration File

FlareGo supports configuration via YAML, TOML, or JSON files. The default location is `$HOME/.config/flarego/config.{yaml,toml,json}`.

Example configuration:

```yaml
# Global settings
gateway: localhost:4317
hz: 100
log_json: false

# Command-specific settings
attach:
  duration: 30s
  gateway: localhost:4317

record:
  duration: 30s
  compress: true
```

## Environment Variables

All configuration options can be set via environment variables with the `FLAREGO_` prefix:

- `FLAREGO_GATEWAY` - Gateway address
- `FLAREGO_HZ` - Sampling frequency
- `FLAREGO_LOG_JSON` - Enable JSON logging

## Output Formats

### Flame Graph (.fgo)

The `.fgo` file format is a gzipped JSON representation of a flame graph. It contains:

- Stack traces
- Timing information
- Metadata
- Optional compression

### JSON Output

When using `--json` flag, the output is a structured JSON object containing:

- Version information
- Build metadata
- Timestamps
- Configuration details
