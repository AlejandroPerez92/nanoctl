# NanoCtl

NanoCtl is a CLI tool for managing Nano Cluster nodes. It provides commands to power on, power off, and reset CM5 nodes in your cluster using GPIO controls.

The tool uses the [go-gpiocdev](https://github.com/warthog618/go-gpiocdev) library for native GPIO access via the Linux GPIO character device.

## Features

- Power on CM5 nodes
- Power off CM5 nodes (graceful shutdown)
- Force power off CM5 nodes (hard shutdown)
- Reset CM5 nodes
- PWM Fan Control with PID algorithm
- **Prometheus Integration** for cluster-wide temperature monitoring
- Native Go implementation using go-gpiocdev (no external dependencies)

## Installation

### From Releases (Recommended)

You can download the latest binary directly from the releases page:

```bash
# For Raspberry Pi 5 / CM5 (ARM64)
wget https://github.com/YOUR_USERNAME/NanoCtl/releases/latest/download/nanoctl-linux-arm64
chmod +x nanoctl-linux-arm64
sudo mv nanoctl-linux-arm64 /usr/local/bin/nanoctl
```

### Building from Source

Since this tool uses Linux-specific GPIO APIs, it must be built on Linux or cross-compiled for Linux.

### Using Make (Recommended)

The project includes a Makefile for easy building:

```bash
# Build for ARM64 (Raspberry Pi 4/5, CM5)
make build-arm64

# Build for ARM 32-bit (Raspberry Pi 3 and earlier)
make build-arm

# Build for both architectures
make build-all
```

### Manual Cross-compilation

For ARM64 (e.g., Raspberry Pi 4/5, CM5):
```bash
GOOS=linux GOARCH=arm64 go build -o nanoctl
```

For ARM 32-bit (e.g., Raspberry Pi 3 and earlier):
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o nanoctl
```

### Building on Linux

If you're already on a Linux system:
```bash
go build -o nanoctl
```

## Installation

After building, you can move the binary to a location in your PATH:

```bash
sudo mv nanoctl /usr/local/bin/
```

## Usage

### Power On

Power on a CM5 node in the specified slot:

```bash
nanoctl poweron 2
```

Or with explicit board type:

```bash
nanoctl poweron 2 --board cm5
```

### Power Off

Gracefully power off a CM5 node (simulates two short presses for Desktop systems):

```bash
nanoctl poweroff 2
```

Force power off (8 second hold for hard shutdown):

```bash
nanoctl poweroff 2 --force
```

### Reset

Reset a CM5 node:

```bash
nanoctl reset 2
```

### Initialization

Initialize the carrier board (resets the switch chip):

```bash
sudo nanoctl init
```

### Fan Control

Start the fan control daemon:

```bash
sudo nanoctl fan
```

Configuration is loaded from `/etc/nanoctl/fan.yaml`. See [Prometheus Integration](docs/PROMETHEUS.md) for advanced monitoring options.

### Automatic Service Installation

To install the fan controller as a systemd service that runs automatically at boot:

```bash
sudo nanoctl install-service
```

This command will:
1. Create a default configuration file at `/etc/nanoctl/fan.yaml`
2. Create `/etc/systemd/system/nanoctl-fan.service`
3. Reload systemd daemon
4. Enable and start the service

### Prometheus Integration

NanoCtl can read temperature metrics from a Prometheus server, allowing you to control fan speed based on cluster-wide temperatures.

To test your Prometheus configuration:

```bash
sudo nanoctl check-prometheus
```

See [docs/PROMETHEUS.md](docs/PROMETHEUS.md) for full configuration details.

## GPIO Details

The tool uses the following GPIO configuration for CM5 boards:

- **GPIO Chip**: gpiochip2
- **Slot GPIO**: GPIO 2

### Operations

- **Power On**: Single short press (1 second pulse)
- **Power Off**: Two short presses (1 second each) for graceful shutdown
- **Force Off**: 8 second hold for hard power off
- **Reset**: Single short press (1 second pulse)

## Requirements

- Linux system with GPIO character device support (kernel 4.8 or later)
- GPIO character device available (typically `/dev/gpiochip*`)
- Root/sudo access for GPIO operations
- For bias control and line reconfiguration features: Linux 5.5 or later
- For debounce and other advanced features: Linux 5.10 or later

The tool uses the native Linux GPIO character device interface via go-gpiocdev, eliminating the need for external utilities like `gpioset`.

## Supported Boards & Disclaimer

**IMPORTANT**: This tool is designed and tested **ONLY** for **CM5 (Compute Module 5)** boards running a **Headless OS** (e.g., Raspberry Pi OS Lite).

Use on other hardware or Desktop/GUI operating systems is not supported and may lead to unexpected behavior (e.g., power cycles instead of shutdowns).

## Examples

```bash
# Power on slot 2
sudo nanoctl poweron 2

# Gracefully power off slot 2
sudo nanoctl poweroff 2

# Force power off slot 2
sudo nanoctl poweroff 2 --force

# Reset slot 2
sudo nanoctl reset 2
```

## Notes

- All GPIO operations require root/sudo privileges
- The slot parameter is currently for display purposes only; the tool uses GPIO 2 on gpiochip2
- For CM5 boards:
  - One short press will shut down headless systems (Raspberry Pi OS Lite)
  - Two short presses will shut down Desktop systems (Raspberry Pi Desktop with GUI)
  - The tool uses two presses by default for compatibility with both system types
- The tool uses the native Linux GPIO character device API via go-gpiocdev
- No external dependencies required at runtime (unlike tools that rely on gpioset)
