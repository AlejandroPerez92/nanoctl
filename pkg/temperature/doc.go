/*
Package temperature provides temperature reading sources for the NanoCtl fan controller.

# Overview

The temperature package defines a common interface for reading temperature data
from various sources. It supports:

  - Local file-based reading (e.g., /sys/class/thermal/thermal_zone0/temp)
  - Remote Prometheus queries for cluster-wide temperature monitoring

# Usage

Create a temperature source using the factory function:

	config := temperature.SourceConfig{
	    Type: temperature.SourceFile,
	    FilePath: "/sys/class/thermal/thermal_zone0/temp",
	}

	source, err := temperature.NewSource(config)
	if err != nil {
	    return err
	}
	defer source.Close()

	temp, err := source.GetTemperature()
	if err != nil {
	    return err
	}

# File Source

The file source reads temperature from a local file in millidegrees Celsius
and converts it to degrees Celsius:

	fileSource := temperature.NewFileSource("/sys/class/thermal/thermal_zone0/temp")
	temp, err := fileSource.GetTemperature()

# Prometheus Source

The Prometheus source queries a Prometheus server using PromQL:

	config := temperature.PrometheusConfig{
	    Host:  "http://prometheus:9090",
	    Query: `max(node_hwmon_temp_celsius{sensor="temp0"})`,
	    Timeout: "5s",
	    Auth: temperature.AuthConfig{
	        Username: "admin",
	        Password: "secret",
	    },
	}

	promSource, err := temperature.NewPrometheusSource(config)
	if err != nil {
	    return err
	}
	temp, err := promSource.GetTemperature()

# Configuration

Only the Prometheus host is required. All other fields are optional:

  - Host (required): Prometheus server URL (http:// or https://)
  - Query (optional): PromQL query, defaults to max(node_hwmon_temp_celsius{sensor="temp0"})
  - Timeout (optional): Query timeout, defaults to "5s"
  - Auth (optional): Basic authentication credentials

# Error Handling

All temperature sources return errors that can be checked and handled.
The fan controller will automatically fall back to file-based reading
if Prometheus queries fail.
*/
package temperature
