# Metrics & Prometheus

NanoCtl supports two ways to integrate with Prometheus/OpenTelemetry:
1.  **Pull (Source)**: Reading temperature **FROM** Prometheus to control the fan.
2.  **Push (Sink)**: Sending its own metrics **TO** an OTLP Collector.

---

## 1. Push Metrics (Sending Data)

NanoCtl can actively push its status (Fan PWM % and measured Temp) to an OpenTelemetry collector.

### Configuration
In `/etc/nanoctl/fan.yaml`:

```yaml
metrics:
  enabled: true
  endpoint: "http://metrics.nanocluster.local/v1/metrics" # HTTP or gRPC endpoint
  insecure: true
  interval: "10s"
  auth:
    username: "nanoctl"
    password: "your-password"
```

### Available Metrics

| Metric Name | Type | Description |
|---|---|---|
| `nanoctl_temperature_celsius` | Gauge | Current temperature reading used by the controller |
| `nanoctl_fan_duty_cycle_percent` | Gauge | Current Fan PWM output (0-100%) |

### PromQL Examples

**Graph Temperature:**
```promql
nanoctl_temperature_celsius
```

**Graph Fan Speed:**
```promql
nanoctl_fan_duty_cycle_percent
```

---

## 2. Pull Temperature (Reading Data)

If you are managing a cluster, you might want NanoCtl to control the fan based on the **hottest node** in the cluster, rather than just the local sensor.

### Configuration

```yaml
temperature:
  source:
    primary: "prometheus"
    fallback: "file"
    prometheus:
      host: "http://prometheus:9090"
      query: 'max(node_hwmon_temp_celsius{sensor="temp0"})'
      timeout: "5s"
```

### Setup Verification
Run the check command to verify NanoCtl can read from your Prometheus:

```bash
sudo nanoctl check-prometheus
```

> **⚠️ Performance Warning**
>
> When using Prometheus as a temperature source, you are likely scraping `node_exporter` metrics.
> To ensure the fan controller responds quickly to temperature spikes, **the scraping interval for `node_exporter` must be set to 15s or less** in your Prometheus configuration.
> A longer scraping interval (e.g., 1m) will cause significant delays in fan reaction, potentially leading to overheating.
