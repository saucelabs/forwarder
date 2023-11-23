---
title: forwarder pac server
weight: 103
---

# Forwarder Pac Server

Usage: `forwarder pac server --pac <file|url> [--protocol <http|https|h2>] [--address <host:port>] [flags]`

Start HTTP server that serves a PAC file.
You can start HTTP, HTTPS or H2 (HTTPS) server.
The server may be protected by basic authentication.
If you start an HTTPS server and you don't provide a certificate,
the server will generate a self-signed certificate on startup.

The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
The PAC file must contain FindProxyForURL or FindProxyForURLEx and must be valid.
Alerts are ignored.


**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.
You can generate a config file by running `forwarder pac server config-file` command.


## Examples

```
  # HTTP server with basic authentication
  forwarder pac server --pac pac.js --basic-auth user:pass

  # HTTPS server with self-signed certificate
  forwarder pac server --pac pac.js --protocol https --address localhost:80443

  # HTTPS server with custom certificate
  forwarder pac server --pac pac.js --protocol https --address localhost:80443 --tls-cert-file cert.pem --tls-key-file key.pem

```

## Server options

### `--address`

Environment variable: `FORWARDER_ADDRESS`

The server address to listen on.
If the host is empty, the server will listen on all available interfaces.

Default value: `:8080`

### `--basic-auth`

Environment variable: `FORWARDER_BASIC_AUTH`

Basic authentication credentials to protect the server.

### `--protocol`

Environment variable: `FORWARDER_PROTOCOL`

The server protocol.
For https and h2 protocols, if TLS certificate is not specified, the server will use a self-signed certificate.

Default value: `http`

### `--read-header-timeout`

Environment variable: `FORWARDER_READ_HEADER_TIMEOUT`

The amount of time allowed to read request headers.

Default value: `1m0s`

### `--tls-cert-file`

Environment variable: `FORWARDER_TLS_CERT_FILE`

TLS certificate to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded certificate.

### `--tls-key-file`

Environment variable: `FORWARDER_TLS_KEY_FILE`

TLS private key to use if the server protocol is https or h2.
Can be a path to a file or "data:" followed by base64 encoded key.

## Proxy options

### `-p, --pac`

Environment variable: `FORWARDER_PAC`

Proxy Auto-Configuration file to use for upstream proxy selection.
It can be a local file or a URL, you can also use '-' to read from stdin.
The data URI scheme is supported, the format is data:base64,<encoded data>.

Default value: `file://pac.js`

## DNS options

### `--dns-round-robin`

Environment variable: `FORWARDER_DNS_ROUND_ROBIN`

If more than one DNS server is specified with the --dns-server flag, passing this flag will enable round-robin selection.


Default value: `false`

### `-n, --dns-server`

Environment variable: `FORWARDER_DNS_SERVER`

DNS server(s) to use instead of system default.
There are two execution policies, when more then one server is specified.
Fallback: the first server in a list is used as primary, the rest are used as fallbacks.
Round robin: the servers are used in a round-robin fashion.
The port is optional, if not specified the default port is 53.

### `--dns-timeout`

Environment variable: `FORWARDER_DNS_TIMEOUT`

Timeout for dialing DNS servers.
Only used if DNS servers are specified.


Default value: `5s`

## HTTP client options

### `--cacert-file`

Environment variable: `FORWARDER_CACERT_FILE`

Add your own CA certificates to verify against.
The system root certificates will be used in addition to any certificates in this list.
Can be a path to a file or "data:" followed by base64 encoded certificate.
Use this flag multiple times to specify multiple CA certificate files.

### `--http-dial-timeout`

Environment variable: `FORWARDER_HTTP_DIAL_TIMEOUT`

The maximum amount of time a dial will wait for a connect to complete.
With or without a timeout, the operating system may impose its own earlier timeout.
For instance, TCP timeouts are often around 3 minutes.


Default value: `10s`

### `--http-idle-conn-timeout`

Environment variable: `FORWARDER_HTTP_IDLE_CONN_TIMEOUT`

The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
Zero means no limit.


Default value: `1m30s`

### `--http-response-header-timeout`

Environment variable: `FORWARDER_HTTP_RESPONSE_HEADER_TIMEOUT`

The amount of time to wait for a server's response headers after fully writing the request (including its body, if any).This time does not include the time to read the response body.
Zero means no limit.


Default value: `0s`

### `--http-tls-handshake-timeout`

Environment variable: `FORWARDER_HTTP_TLS_HANDSHAKE_TIMEOUT`

The maximum amount of time waiting to wait for a TLS handshake.
Zero means no limit.

Default value: `10s`

### `--insecure`

Environment variable: `FORWARDER_INSECURE`

Don't verify the server's certificate chain and host name.
Enable to work with self-signed certificates.


Default value: `false`

## Logging options

### `--log-file`

Environment variable: `FORWARDER_LOG_FILE`

Path to the log file, if empty, logs to stdout.

### `--log-http`

Environment variable: `FORWARDER_LOG_HTTP`

HTTP request and response logging mode.
Setting this to none disables logging.
The short-url mode logs [scheme://]host[/path] instead of the full URL.
The error mode logs request line and headers if status code is greater than or equal to 500.

### `--log-level`

Environment variable: `FORWARDER_LOG_LEVEL`

Log level.

Default value: `info`

