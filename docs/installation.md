# FlareGo Installation Guide

This guide covers different methods to install FlareGo and its requirements.

## Prerequisites

- Go 1.24 or later
- Linux, macOS, or Windows
- For eBPF features: Linux kernel 4.9 or later
- For Kubernetes integration: kubectl configured

## Installation Methods

### From Source

1. Clone the repository:

   ```bash
   git clone https://github.com/Voskan/flarego.git
   cd flarego
   ```

2. Build the project:

   ```bash
   make build
   ```

3. Install the binary:
   ```bash
   make install
   ```

### Using Go Install

```bash
go install github.com/Voskan/flarego/cmd/flarego@latest
```

### Using Package Managers

#### macOS (Homebrew)

```bash
brew install flarego
```

#### Linux (APT)

```bash
# Add repository
curl -fsSL https://packages.flarego.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/flarego-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/flarego-archive-keyring.gpg] https://packages.flarego.io/apt stable main" | sudo tee /etc/apt/sources.list.d/flarego.list

# Install
sudo apt update
sudo apt install flarego
```

#### Linux (RPM)

```bash
# Add repository
sudo rpm --import https://packages.flarego.io/gpg
sudo tee /etc/yum.repos.d/flarego.repo << EOF
[flarego]
name=FlareGo Repository
baseurl=https://packages.flarego.io/rpm
enabled=1
gpgcheck=1
gpgkey=https://packages.flarego.io/gpg
EOF

# Install
sudo yum install flarego
```

### Using Docker

```bash
docker pull ghcr.io/flarego/flarego:latest
```

## Configuration

1. Create configuration directory:

   ```bash
   mkdir -p ~/.config/flarego
   ```

2. Create configuration file:
   ```bash
   # ~/.config/flarego/config.yaml
   gateway: localhost:4317
   hz: 100
   log_json: false
   ```

## Verification

1. Check installation:

   ```bash
   flarego version
   ```

2. Test basic functionality:
   ```bash
   # Record a short profile
   flarego record --duration 5s
   ```

## Development Setup

1. Install development dependencies:

   ```bash
   make deps
   ```

2. Run tests:

   ```bash
   make test
   ```

3. Build development version:
   ```bash
   make dev
   ```

## Troubleshooting

### Common Issues

1. **Permission Denied**

   ```bash
   # Fix permissions
   sudo chown -R $USER:$USER ~/.config/flarego
   ```

2. **Missing Dependencies**

   ```bash
   # Install build dependencies
   make deps
   ```

3. **eBPF Support**

   ```bash
   # Check kernel version
   uname -r

   # Install eBPF tools
   sudo apt install linux-tools-common linux-tools-generic
   ```

### Logging

Enable debug logging:

```bash
export FLAREGO_LOG_LEVEL=debug
flarego attach --log-json
```

## Uninstallation

### Remove Binary

```bash
# If installed via make
make uninstall

# If installed via package manager
sudo apt remove flarego  # Debian/Ubuntu
sudo yum remove flarego  # RHEL/CentOS
```

### Clean Configuration

```bash
rm -rf ~/.config/flarego
```

## Next Steps

- Read the [CLI Reference](cli-reference.md)
- Check the [Architecture](architecture.md)
- Try the [Usage Guide](usage-guide.md)
