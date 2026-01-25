# Installation Guide

## Quick Install (Recommended)

To install NanoCtl on your Raspberry Pi (CM5/Pi 4/Pi 5), simply run:

```bash
curl -fsSL https://raw.githubusercontent.com/AlejandroPerez92/nanoctl/main/install.sh | sudo bash
```

This script will:
1. Detect your architecture (ARM64/ARM).
2. Download the latest release.
3. Install the binary to `/usr/local/bin/nanoctl`.
4. Automatically set up and start the systemd service.

## Manual Installation

### 1. Download Binary
Download the latest binary from the [Releases Page](https://github.com/AlejandroPerez92/nanoctl/releases).

```bash
# Example for ARM64 (CM5/Pi 5)
wget https://github.com/AlejandroPerez92/nanoctl/releases/latest/download/nanoctl-linux-arm64
chmod +x nanoctl-linux-arm64
sudo mv nanoctl-linux-arm64 /usr/local/bin/nanoctl
```

### 2. Install Service
Run the built-in service installer:

```bash
sudo nanoctl install-service
```

## Building from Source

If you prefer to build from source:

1. **Clone the repository:**
   ```bash
   git clone https://github.com/AlejandroPerez92/nanoctl.git
   cd nanoctl
   ```

2. **Build:**
   ```bash
   # For local architecture
   make build
   
   # For ARM64 (if cross-compiling)
   make build-arm64
   ```

3. **Install:**
   ```bash
   sudo mv nanoctl-linux-arm64 /usr/local/bin/nanoctl
   ```
