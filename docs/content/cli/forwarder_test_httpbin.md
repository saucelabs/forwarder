---
id: httpbin
title: forwarder test httpbin
---

# Forwarder Test Httpbin

Usage: `forwarder test httpbin [--protocol <http|https|h2>] [--address <host:port>] [flags]`

Start HTTP(S) server that serves httpbin.org API

**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.
You can generate a config file by running `forwarder test httpbin config-file` command.


## Server options

### `--address` {#address}

* Environment variable: `FORWARDER_ADDRESS`
* Value Format: `<host:port>`
* Default value: `:8080`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

### `--basic-auth` {#basic-auth}

* Environment variable: `FORWARDER_BASIC_AUTH`
* Value Format: `<username[:password]>`

Basic authentication credentials to protect the server.

### `--idle-timeout` {#idle-timeout}

* Environment variable: `FORWARDER_IDLE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1h0m0s`

The maximum amount of time to wait for the next request before closing connection.

### `--protocol` {#protocol}

* Environment variable: `FORWARDER_PROTOCOL`
* Value Format: `<http|https|h2>`
* Default value: `http`

The server protocol.
For https and h2 protocols, if TLS certificate is not specified, the server will use a self-signed certificate.

### `--read-header-timeout` {#read-header-timeout}

* Environment variable: `FORWARDER_READ_HEADER_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1m0s`

The amount of time allowed to read request headers.

### `--tls-cert-file` {#tls-cert-file}

* Environment variable: `FORWARDER_TLS_CERT_FILE`
* Value Format: `<path or base64>`

TLS certificate to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded certificate.

### `--tls-handshake-timeout` {#tls-handshake-timeout}

* Environment variable: `FORWARDER_TLS_HANDSHAKE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `0s`

The maximum amount of time to wait for a TLS handshake before closing connection.
Zero means no limit.

### `--tls-key-file` {#tls-key-file}

* Environment variable: `FORWARDER_TLS_KEY_FILE`
* Value Format: `<path or base64>`

TLS private key to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded key.

## API server options

### `--api-address` {#api-address}

* Environment variable: `FORWARDER_API_ADDRESS`
* Value Format: `<host:port>`
* Default value: `localhost:10000`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

### `--api-basic-auth` {#api-basic-auth}

* Environment variable: `FORWARDER_API_BASIC_AUTH`
* Value Format: `<username[:password]>`

Basic authentication credentials to protect the server.

### `--api-idle-timeout` {#api-idle-timeout}

* Environment variable: `FORWARDER_API_IDLE_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1h0m0s`

The maximum amount of time to wait for the next request before closing connection.

### `--api-read-header-timeout` {#api-read-header-timeout}

* Environment variable: `FORWARDER_API_READ_HEADER_TIMEOUT`
* Value Format: `<duration>`
* Default value: `1m0s`

The amount of time allowed to read request headers.

## Logging options

### `--log-file` {#log-file}

* Environment variable: `FORWARDER_LOG_FILE`
* Value Format: `<path>`

Path to the log file, if empty, logs to stdout.

### `--log-http` {#log-http}

* Environment variable: `FORWARDER_LOG_HTTP`
* Value Format: `[api|server:]<none|short-url|url|headers|body|errors>,...`

HTTP request and response logging mode.
Setting this to none disables logging.
The short-url mode logs [scheme://]host[/path] instead of the full URL.
The error mode logs request line and headers if status code is greater than or equal to 500.

### `--log-level` {#log-level}

* Environment variable: `FORWARDER_LOG_LEVEL`
* Value Format: `<error|info|debug>`
* Default value: `info`

Log level.

