---
title: forwarder pac eval
weight: 102
---

# Forwarder Pac Eval

Usage: `forwarder pac eval --pac <file|url> [flags] <url>...`

Evaluate a PAC file for given URL (or URLs).
The output is a list of proxy strings, one per URL.
The PAC file can be specified as a file path or URL with scheme "file", "http" or "https".
The PAC file must contain FindProxyForURL or FindProxyForURLEx and must be valid.
Alerts are written to stderr.


**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.
You can generate a config file by running `forwarder pac eval config-file` command.


## Examples

```
  # Evaluate PAC file for multiple URLs
  forwarder pac eval --pac pac.js https://www.google.com https://www.facebook.com

```

## Proxy options

### `-p, --pac` {#pac}

Environment variable: `FORWARDER_PAC`

Proxy Auto-Configuration file to use for upstream proxy selection.
It can be a local file or a URL, you can also use '-' to read from stdin.
The data URI scheme is supported, the format is data:base64,<encoded data>.

Default value: `file://pac.js`

## DNS options

### `--dns-round-robin` {#dns-round-robin}

Environment variable: `FORWARDER_DNS_ROUND_ROBIN`

If more than one DNS server is specified with the --dns-server flag, passing this flag will enable round-robin selection.


Default value: `false`

### `-n, --dns-server` {#dns-server}

Environment variable: `FORWARDER_DNS_SERVER`

DNS server(s) to use instead of system default.
There are two execution policies, when more then one server is specified.
Fallback: the first server in a list is used as primary, the rest are used as fallbacks.
Round robin: the servers are used in a round-robin fashion.
The port is optional, if not specified the default port is 53.

### `--dns-timeout` {#dns-timeout}

Environment variable: `FORWARDER_DNS_TIMEOUT`

Timeout for dialing DNS servers.
Only used if DNS servers are specified.


Default value: `5s`

## HTTP client options

### `--cacert-file` {#cacert-file}

Environment variable: `FORWARDER_CACERT_FILE`

Add your own CA certificates to verify against.
The system root certificates will be used in addition to any certificates in this list.
Can be a path to a file or "data:" followed by base64 encoded certificate.
Use this flag multiple times to specify multiple CA certificate files.

### `--http-dial-timeout` {#http-dial-timeout}

Environment variable: `FORWARDER_HTTP_DIAL_TIMEOUT`

The maximum amount of time a dial will wait for a connect to complete.
With or without a timeout, the operating system may impose its own earlier timeout.
For instance, TCP timeouts are often around 3 minutes.


Default value: `10s`

### `--http-idle-conn-timeout` {#http-idle-conn-timeout}

Environment variable: `FORWARDER_HTTP_IDLE_CONN_TIMEOUT`

The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.
Zero means no limit.


Default value: `1m30s`

### `--http-response-header-timeout` {#http-response-header-timeout}

Environment variable: `FORWARDER_HTTP_RESPONSE_HEADER_TIMEOUT`

The amount of time to wait for a server's response headers after fully writing the request (including its body, if any).This time does not include the time to read the response body.
Zero means no limit.


Default value: `0s`

### `--http-tls-handshake-timeout` {#http-tls-handshake-timeout}

Environment variable: `FORWARDER_HTTP_TLS_HANDSHAKE_TIMEOUT`

The maximum amount of time waiting to wait for a TLS handshake.
Zero means no limit.

Default value: `10s`

### `--insecure` {#insecure}

Environment variable: `FORWARDER_INSECURE`

Don't verify the server's certificate chain and host name.
Enable to work with self-signed certificates.


Default value: `false`

