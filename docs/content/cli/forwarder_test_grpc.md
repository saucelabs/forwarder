---
title: forwarder test grpc
---

# Forwarder Test Grpc

Usage: `forwarder test grpc [--address <host:port>] [flags]`

Start HTTP/2 gRPC server for testing

**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.
You can generate a config file by running `forwarder test grpc config-file` command.


## Server options

### `--address` {#address}

* Environment variable: `FORWARDER_ADDRESS`
* Value Format: `<host:port>`
* Default value: `localhost:1443`

Address to listen on.
If the host is empty, the server will listen on all available interfaces.

### `--tls-cert-file` {#tls-cert-file}

* Environment variable: `FORWARDER_TLS_CERT_FILE`
* Value Format: `<path or base64>`

TLS certificate to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded certificate.

### `--tls-key-file` {#tls-key-file}

* Environment variable: `FORWARDER_TLS_KEY_FILE`
* Value Format: `<path or base64>`

TLS private key to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded key.

## Logging options

### `--log-file` {#log-file}

* Environment variable: `FORWARDER_LOG_FILE`
* Value Format: `<path>`

Path to the log file, if empty, logs to stdout.

### `--log-level` {#log-level}

* Environment variable: `FORWARDER_LOG_LEVEL`
* Value Format: `<error|info|debug>`
* Default value: `info`

Log level.

