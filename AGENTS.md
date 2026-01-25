# AGENTS.md - NanoCtl Development Guide

This guide is for AI coding agents working on NanoCtl, a Go-based CLI tool for managing Nano Cluster nodes (CM5/Compute Module 5) via GPIO controls on Linux systems.

## Project Overview

- **Language**: Go 1.25
- **Target Platform**: Linux ARM64 (Raspberry Pi systems)
- **Framework**: Cobra CLI framework
- **Hardware Interface**: GPIO via `github.com/warthog618/go-gpiocdev`
- **Module Name**: `github.com/AlejandroPerez92/nanoctl`

## Build Commands

### Standard Builds
```bash
# Build for ARM64 (Raspberry Pi 5, CM5) - default target
make build-arm64

# Build for local development (native architecture)
make build
go build -o nanoctl

# Clean build artifacts
make clean
```

### Production Builds
```bash
# With optimizations (stripped binaries)
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o nanoctl-linux-arm64
```

## Test Commands

**Note**: This project currently has NO unit tests (`*_test.go` files).

### Running Tests (when added)
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test ./pkg/gpio
go test ./pkg/fan
go test ./cmd

# Run a single test
go test -v ./pkg/gpio -run TestPowerOn
go test -v -run ^TestSpecificFunction$ ./path/to/package

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests for modified packages only
go test ./$(git diff --name-only | grep '\.go$' | xargs -I {} dirname {} | sort -u)
```

## Lint and Format Commands

```bash
# Format code (required before commits)
go fmt ./...

# Run go vet for common issues
go vet ./...

# Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck ./...

# Run golangci-lint (recommended, install from https://golangci-lint.run/)
golangci-lint run

# Check module dependencies
go mod tidy
go mod verify
```

## Code Style Guidelines

### Imports
- Group imports in this order: stdlib, third-party, internal packages
- Use blank line to separate groups
- Example:
  ```go
  import (
      "context"
      "fmt"
      "os"
      
      "github.com/spf13/cobra"
      "github.com/warthog618/go-gpiocdev"
      
      "github.com/AlejandroPerez92/nanoctl/pkg/fan"
      "github.com/AlejandroPerez92/nanoctl/pkg/gpio"
  )
  ```

### Naming Conventions
- **Packages**: lowercase, single word (e.g., `gpio`, `fan`, `cmd`)
- **Files**: lowercase with underscores for multi-word (e.g., `gpio.go`, `monitor.go`)
- **Exported Types**: PascalCase (e.g., `Controller`, `MonitorConfig`)
- **Unexported Types**: camelCase (e.g., `boardType`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `fanChip`, `targetTemp`)

### Types and Interfaces
- Define custom types for domain concepts (e.g., `type BoardType string`)
- Use constants for enum-like values
- Example:
  ```go
  type BoardType string
  
  const (
      BoardCM5 BoardType = "cm5"
  )
  ```

### Error Handling
- **Always** check and handle errors explicitly
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Provide context in error messages
- Example:
  ```go
  if err := someFunc(); err != nil {
      return fmt.Errorf("failed to do something: %w", err)
  }
  ```

### Resource Management
- **Always** use `defer` for cleanup (e.g., `defer line.Close()`)
- Place `defer` immediately after resource acquisition
- Example:
  ```go
  line, err := gpiocdev.RequestLine(c.chipName, pin, gpiocdev.AsOutput(0))
  if err != nil {
      return fmt.Errorf("failed to request GPIO %d: %w", pin, err)
  }
  defer line.Close()
  ```

### Context Usage
- Use `context.Context` for cancellation and graceful shutdown
- Pass context as first parameter to functions
- Example:
  ```go
  func RunMonitor(ctx context.Context, config MonitorConfig) error {
      // ...
      for {
          select {
          case <-ctx.Done():
              return nil
          // ...
          }
      }
  }
  ```

### Comments
- Add comments for exported functions, types, and constants
- Use complete sentences starting with the name being documented
- Example:
  ```go
  // Controller handles GPIO operations for node control
  type Controller struct { ... }
  
  // PowerOn powers on a CM5 node
  // Simulates a single short press (power on)
  func (c *Controller) PowerOn(slot int, boardType BoardType) error { ... }
  ```

### Formatting
- Use `gofmt` / `go fmt` (tabs for indentation, not spaces)
- Line length: aim for 100 characters, hard limit 120
- One statement per line

## Project Structure

```
NanoCtl/
├── main.go              # Entry point
├── cmd/                 # Cobra command implementations
│   ├── root.go          # Root command
│   ├── fan.go           # Fan control command
│   ├── poweron.go       # Power on command
│   ├── poweroff.go      # Power off command
│   ├── reset.go         # Reset command
│   └── service.go       # Systemd service installation
├── pkg/                 # Reusable packages
│   ├── gpio/            # GPIO utilities
│   │   └── gpio.go
│   └── fan/             # Fan control logic
│       ├── monitor.go   # Temperature monitoring
│       └── pwm.go       # PWM control
├── go.mod               # Go module definition
├── Makefile             # Build automation
└── .github/workflows/   # CI/CD pipelines
```

## Best Practices for Linux Services

1. **GPIO Permissions**: Ensure proper permissions for GPIO access (typically requires root)
2. **Graceful Shutdown**: Always handle SIGINT and SIGTERM for clean shutdown
3. **Systemd Integration**: Use proper systemd service units with `Type=simple` or `Type=exec`
4. **Error Logging**: Write errors to stderr using `fmt.Fprintf(os.Stderr, ...)`
5. **Resource Limits**: Be mindful of file descriptor limits for GPIO access
6. **Timing Precision**: Use `time.Duration` for all time-related operations
7. **PID Control**: Tune PID parameters (Kp, Ki, Kd) for specific hardware
8. **Temperature Reads**: Read from `/sys/class/thermal/thermal_zone0/temp` (millidegrees)

## Development Workflow

1. Make changes to code
2. Run `go fmt ./...` to format
3. Run `go vet ./...` to check for issues
4. Add tests for new functionality
5. Build with `make build` or target architecture
6. Test on target hardware (Raspberry Pi)
7. Commit changes with descriptive messages

## Common Patterns

### Creating Cobra Commands
```go
var myCmd = &cobra.Command{
    Use:   "command-name",
    Short: "Short description",
    Long:  `Longer description with details`,
    Run: func(cmd *cobra.Command, args []string) {
        // Command implementation
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().StringVar(&variable, "flag", "default", "description")
}
```

### GPIO Operations
```go
// Request GPIO line
line, err := gpiocdev.RequestLine(chipName, pin, gpiocdev.AsOutput(0))
if err != nil {
    return fmt.Errorf("failed to request GPIO: %w", err)
}
defer line.Close()

// Set GPIO value
if err := line.SetValue(1); err != nil {
    return fmt.Errorf("failed to set GPIO: %w", err)
}
```

## Notes for Agents

- This codebase targets embedded Linux systems (Raspberry Pi)
- GPIO operations require root privileges
- Cross-compilation is required (GOOS=linux GOARCH=arm64)
- `gpiocdev` GPIO library used for GPIO control
- PID controller used for temperature-based fan control
- All commands use Cobra framework for CLI parsing
