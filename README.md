Caddy Throttle Listener
==========================

A Caddy module that provides bandwidth throttling for incoming connections. Note that the throttling applies for all bytes transmitted on the connection, including TLS handshake and HTTP headers.

## Overview

This module allows you to throttle incoming connections to your Caddy server, preventing abuse and ensuring fair usage of your resources. It provides a simple and effective way to limit the amount of data that can be transferred over a connection, helping to prevent overwhelming your server with excessive traffic.

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

