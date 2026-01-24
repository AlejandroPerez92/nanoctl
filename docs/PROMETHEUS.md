# Prometheus Integration Guide

NanoCtl supports reading temperature metrics from Prometheus for cluster-wide temperature monitoring. This allows you to control fan speed based on temperatures from multiple nodes or aggregate metrics.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [PromQL Query Examples](#promql-query-examples)
- [Authentication](#authentication)
- [Testing Connection](#testing-connection)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Quick Start

### Minimal Configuration

The minimum required configuration for Prometheus is just the host:

```yaml
temperature:
  target: 55.0
  source:
    primary: "prometheus"
    fallback: "file"
    prometheus:
      host: "http://prometheus:9090"
    file:
      path: "/sys/class/thermal/thermal_zone0/temp"
```

This will:
- Query Prometheus at `http://prometheus:9090`
- Use the default query: `max(node_hwmon_temp_celsius{sensor="temp0"})`
- Use a 5-second timeout
- Fall back to file-based reading if Prometheus fails

### Full Configuration Example

```yaml
temperature:
  target: 60.0
  source:
    primary: "prometheus"
    fallback: "file"
    prometheus:
      host: "https://prometheus.monitoring.svc.cluster.local:9090"
      query: 'max(node_hwmon_temp_celsius{job="node-exporter",sensor="temp0"})'
      timeout: "10s"
      auth:
        username: "admin"
        password: "secret"
    file:
      path: "/sys/class/thermal/thermal_zone0/temp"
```

## Configuration

### Required Fields

- **`host`**: Prometheus server URL (must start with `http://` or `https://`)

### Optional Fields

- **`query`**: PromQL query for temperature metric
  - Default: `max(node_hwmon_temp_celsius{sensor="temp0"})`
  
- **`timeout`**: Query timeout duration
  - Default: `"5s"`
  - Examples: `"5s"`, `"10s"`, `"30s"`
  
- **`auth`**: Basic authentication credentials
  - `username`: Username for basic auth
  - `password`: Password for basic auth
  - Both must be provided or omitted together

## PromQL Query Examples

### Single Node Temperature

Query temperature from a specific node:

```yaml
query: 'node_hwmon_temp_celsius{instance="node1:9100",sensor="temp0"}'
```

### Maximum Temperature Across Cluster

Use the highest temperature from any node (recommended for safety):

```yaml
query: 'max(node_hwmon_temp_celsius{sensor="temp0"})'
```

### Average Temperature Across Cluster

Use the average temperature:

```yaml
query: 'avg(node_hwmon_temp_celsius{sensor="temp0"})'
```

### Specific Job or Namespace

Filter by Prometheus job label:

```yaml
query: 'max(node_hwmon_temp_celsius{job="compute-nodes",sensor="temp0"})'
```

### Multiple Sensors

Query multiple temperature sensors:

```yaml
query: 'max(node_hwmon_temp_celsius{sensor=~"temp0|temp1"})'
```

### Custom Labels

Use custom labels added by node_exporter:

```yaml
query: 'max(node_hwmon_temp_celsius{cluster="production",rack="A1",sensor="temp0"})'
```

## Authentication

### No Authentication

For internal/trusted networks:

```yaml
prometheus:
  host: "http://prometheus:9090"
  # No auth section needed
```

### Basic Authentication

For authenticated Prometheus servers:

```yaml
prometheus:
  host: "https://prometheus.example.com:9090"
  auth:
    username: "monitoring-user"
    password: "your-secure-password"
```

**Security Note**: Passwords are stored in plain text in the config file. Ensure proper file permissions:

```bash
sudo chmod 600 /etc/nanoctl/fan.yaml
```

## Testing Connection

Use the `check-prometheus` command to verify your configuration:

```bash
sudo nanoctl check-prometheus
```

This will:
1. Load your configuration
2. Connect to Prometheus
3. Execute the temperature query
4. Display the current temperature
5. Report connection time and success/failure

Example output:

```
Prometheus Configuration:
  Host:    http://prometheus:9090
  Query:   max(node_hwmon_temp_celsius{sensor="temp0"})
  Timeout: 5s
  Auth:    None

Testing connection...
✓ Prometheus client created successfully

Executing query...
✓ Query successful (took 45ms)

Temperature: 58.50°C

✓ Prometheus connection is working correctly!
```

## Troubleshooting

### Connection Refused

**Error**: `connection refused` or `no such host`

**Solutions**:
- Verify the Prometheus host URL is correct
- Ensure Prometheus is running and accessible
- Check network connectivity: `curl http://prometheus:9090/-/healthy`
- Verify firewall rules allow connections

### Authentication Failed

**Error**: `401 Unauthorized`

**Solutions**:
- Verify username and password are correct
- Ensure the Prometheus server requires authentication
- Try accessing Prometheus web UI with the same credentials

### Query Returns No Data

**Error**: `no temperature data returned from query`

**Solutions**:
- Verify the metric name exists in Prometheus
- Check the query in Prometheus UI: [http://prometheus:9090/graph](http://prometheus:9090/graph)
- Ensure node_exporter is running on the target nodes
- Verify sensor labels match your query

### Timeout Errors

**Error**: `context deadline exceeded`

**Solutions**:
- Increase timeout value: `timeout: "10s"`
- Check Prometheus server performance
- Simplify the query to reduce computation time

### SSL/TLS Errors

**Error**: `x509: certificate signed by unknown authority`

**Solutions**:
- Verify the Prometheus URL uses `https://`
- Ensure valid SSL certificates are installed
- For self-signed certificates, use HTTP instead (if network is trusted)

## Best Practices

### 1. Use Fallback

Always configure file-based fallback for reliability:

```yaml
source:
  primary: "prometheus"
  fallback: "file"
```

### 2. Choose Appropriate Query

- **Safety-critical**: Use `max()` to respond to hottest node
- **Efficiency**: Use `avg()` for balanced cooling
- **Specific node**: Use instance label for targeted monitoring

### 3. Set Reasonable Timeout

- Default `5s` is usually sufficient
- Increase to `10s` for slow networks
- Don't set too high to avoid delayed fan response

### 4. Monitor Logs

Check systemd logs for errors:

```bash
sudo journalctl -u nanoctl-fan -f
```

### 5. Test Before Deployment

Always test with `check-prometheus` before enabling the service:

```bash
sudo nanoctl check-prometheus
```

### 6. Secure Credentials

Set restrictive permissions on the config file:

```bash
sudo chown root:root /etc/nanoctl/fan.yaml
sudo chmod 600 /etc/nanoctl/fan.yaml
```

### 7. Use HTTPS in Production

For production environments, use HTTPS:

```yaml
host: "https://prometheus.example.com:9090"
```

## Integration with Kubernetes

### Using In-Cluster Prometheus

```yaml
prometheus:
  host: "http://prometheus.monitoring.svc.cluster.local:9090"
  query: 'max(node_hwmon_temp_celsius{sensor="temp0"})'
```

### Using Prometheus Operator

```yaml
prometheus:
  host: "http://prometheus-operated.monitoring:9090"
  query: 'max(node_hwmon_temp_celsius{job="node-exporter",sensor="temp0"})'
```

### With ServiceMonitor

Ensure your node_exporter has a ServiceMonitor and is being scraped:

```bash
kubectl get servicemonitor -n monitoring
kubectl get targets -n monitoring
```

## Example Scenarios

### Scenario 1: Home Lab (Single Prometheus, No Auth)

```yaml
temperature:
  target: 55.0
  source:
    primary: "prometheus"
    fallback: "file"
    prometheus:
      host: "http://192.168.1.100:9090"
```

### Scenario 2: Production Cluster (HTTPS + Auth)

```yaml
temperature:
  target: 60.0
  source:
    primary: "prometheus"
    fallback: "file"
    prometheus:
      host: "https://prometheus.prod.example.com:9090"
      query: 'max(node_hwmon_temp_celsius{job="k8s-nodes",sensor="temp0"})'
      timeout: "10s"
      auth:
        username: "nanoctl"
        password: "secure-password-here"
```

### Scenario 3: Edge Computing (Average Temperature)

```yaml
temperature:
  target: 58.0
  source:
    primary: "prometheus"
    fallback: "file"
    prometheus:
      host: "http://edge-prometheus:9090"
      query: 'avg(node_hwmon_temp_celsius{location="datacenter-a",sensor="temp0"})'
      timeout: "8s"
```

## Support

For issues or questions:
- Check logs: `sudo journalctl -u nanoctl-fan -n 100`
- Test connection: `sudo nanoctl check-prometheus`
