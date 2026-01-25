# Configuration Guide

NanoCtl uses a YAML configuration file located at `/etc/nanoctl/fan.yaml`.

## Default Configuration

```yaml
# GPIO Configuration (CM5 default)
gpio:
  chip_name: "gpiochip0"
  pin: 13

# Temperature Control
temperature:
  target: 55.0  # Target CPU temperature in Celsius
  source:
    primary: "file"     # "file" or "prometheus"
    fallback: "file"
    file:
      path: "/sys/class/thermal/thermal_zone0/temp"

# PID Controller
pid:
  kp: 5.0
  ki: 0.1
  kd: 0.5

# Monitoring
monitor:
  check_interval: "1s"

# OTLP Metrics (Push)
metrics:
  enabled: false
  endpoint: "localhost:4317"
  interval: "10s"
  # Optional: Basic Auth
  # auth:
  #   username: "nanoctl"
  #   password: "secure-password"
```

## Settings Explained

### GPIO
- `chip_name`: The GPIO chip device (e.g., `gpiochip4` on Pi 5, `gpiochip0` on others).
- `pin`: The BCM pin number controlling the PWM fan.

### Temperature
- `target`: The temperature the PID controller tries to maintain.
- `source`:
  - `primary`: Where to read temperature from (`file` = local sensor, `prometheus` = remote query).
  - `fallback`: Backup source if primary fails.
  - **Note**: If using `prometheus`, ensure your scraping interval is **< 15s** for responsive cooling.

### PID Controller
- `kp`: Proportional gain (reacts to current error).
- `ki`: Integral gain (reacts to past errors/accumulation).
- `kd`: Derivative gain (reacts to rate of change).
- *Tip: The default values work well for most CM5 setups.*

### Metrics (Push)
Configures NanoCtl to **push** its own metrics to an OpenTelemetry collector.

- `enabled`: Set to `true` to enable pushing.
- `endpoint`: The HTTP or gRPC endpoint (e.g., `http://metrics.local/v1/metrics`).
- `insecure`: `true` to disable TLS verification.
- `auth`: (Optional) Basic auth credentials.
  ```yaml
  auth:
    username: "nanoctl"
    password: "secure-password"
  ```
