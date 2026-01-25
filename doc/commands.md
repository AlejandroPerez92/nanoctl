# Command Reference

## `nanoctl poweron <slot>`
Powers on a CM5 node.
- **Usage**: `nanoctl poweron 2`
- **Details**: Sends a single short pulse (1s) to the GPIO.

## `nanoctl poweroff <slot>`
Gracefully shuts down a node.
- **Usage**: `nanoctl poweroff 2`
- **Details**: Sends two short pulses (1s each). This triggers a safe shutdown on most OSes (Headless & Desktop).

## `nanoctl poweroff <slot> --force`
Forcefully cuts power to a node.
- **Usage**: `nanoctl poweroff 2 --force`
- **Details**: Holds the button for 8 seconds. **Warning: Potential data loss.**

## `nanoctl reset <slot>`
Resets a node.
- **Usage**: `nanoctl reset 2`

## `nanoctl fan`
Starts the fan control daemon.
- **Usage**: `sudo nanoctl fan`
- **Note**: Usually run as a systemd service (`nanoctl-fan`).

## `nanoctl install-service`
Installs the systemd service and default configuration.
- **Usage**: `sudo nanoctl install-service`

## `nanoctl check-prometheus`
Tests the connection to a Prometheus server defined in `fan.yaml`.
- **Usage**: `sudo nanoctl check-prometheus`

## `nanoctl version`
Prints version information.
