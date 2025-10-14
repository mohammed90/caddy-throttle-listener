Caddy Throttle Listener
==========================

A Caddy module that provides bandwidth throttling and chaos engineering features for incoming connections. This module allows you to simulate real-world network conditions by adding bandwidth limits, latency, and jitter. Note that the throttling applies for all bytes transmitted on the connection, including TLS handshake and HTTP headers.

## Overview

This module allows you to throttle incoming connections to your Caddy server to prevent abuse and simulate chaotic network conditions for testing purposes. It provides:

- **Bandwidth Throttling**: Limit upload/download speeds per connection
- **Latency Injection**: Add fixed delays to simulate slow networks
- **Jitter**: Add random delays to simulate unstable networks

Perfect for preventing abuse, ensuring fair resource usage, and testing how your applications handle real-world network conditions.

## Install

Follow the xcaddy install process [here](https://github.com/caddyserver/xcaddy#install).

Then, build Caddy with this Go module plugged in. For example:

```shell
xcaddy build --with github.com/mohammed90/caddy-throttle-listener
```

## Caddyfile Example


```caddyfile
{
	servers {
		listener_wrappers {
			throttle {
				down 1MiB
				up 1MiB
			}
			tls
		}
	}
}
example.com {
	root * /var/www/html
	file_server
}
```

## Configuration Options

### Basic Throttling

```caddyfile
{
	servers {
		listener_wrappers {
			throttle {
				down 512KB    # Download limit per connection
				up 1MB       # Upload limit per connection
			}
		}
	}
}
```

### Chaos Engineering Features

Add network latency and jitter to test application resilience:

```caddyfile
{
	servers {
		listener_wrappers {
			throttle {
				# Bandwidth limits
				down 1MiB
				up 512KB
				
				# Chaos engineering
				latency 100ms    # Fixed delay per read/write
				jitter 50ms      # Random delay (0-50ms) per operation
			}
		}
	}
}
```

### Testing Scenarios

#### Slow Network Simulation
```caddyfile
throttle {
	down 64KB
	up 32KB
	latency 200ms
	jitter 100ms
}
```

#### High-Latency Connection (Satellite/Cellular)
```caddyfile
throttle {
	latency 800ms
	jitter 200ms
}
```

#### Unstable Connection
```caddyfile
throttle {
	jitter 500ms  # High jitter without base latency
}
```

